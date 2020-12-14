package request

import (
	"encoding/json"
	"github.com/dgraph-io/badger/v2"
	"github.com/go-redis/redis/v7"
	_ "github.com/go-sql-driver/mysql" //needed for sql driver
	"github.com/huandu/go-sqlbuilder"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/parser"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var db struct {
	sync.Once
	database
}

var cacheExpiration time.Duration

type database interface {
	SetIdentifyData(ip string, data *network.ConnectionData, response *IdentifyResponse) error
	GetIdentifyData(ip string) (*IdentifyResponse, error)
	SetConnectionData(ip string, data *network.ConnectionData) error
	GetConnectionData(ip string) (*network.ConnectionData, error)
}

type badgerDatabase struct {
	db *badger.DB
}

type sqlDatabase struct {
	db *sqlx.DB
}

type redisDatabase struct {
	db *redis.Client
}

type emptyDatabase struct{}

var mysqlSchemaArr = []string{
	`DROP TABLE IF EXISTS cache;`,

	`CREATE TABLE cache (
		id INTEGER PRIMARY KEY,
		ip varchar(255) NOT NULL,
		datatype varchar(255) NOT NULL,
		data text NOT NULL,
		time datetime DEFAULT current_timestamp,
		UNIQUE KEY 'unique_entries' (ip, datatype)
		);`,
	`ALTER TABLE cache MODIFY id int(11) NOT NULL AUTO_INCREMENT;`,
}

type sqlSelectResults []struct {
	Time     string
	Data     string
	Datatype string
}

func initDB() error {
	if viper.GetBool("db.no-cache") {
		db.database = &emptyDatabase{}
		return nil
	}

	var err error
	cacheExpiration, err = time.ParseDuration(viper.GetString("db.duration"))
	if err != nil {
		return errors.Wrap(err, "failed to parse cache expiration")
	}

	if viper.GetString("db.drivername") == "built-in" {
		badgerDB := badgerDatabase{}
		u, err := user.Current()
		if err != nil {
			return err
		}
		badgerDB.db, err = badger.Open(badger.DefaultOptions(filepath.Join(os.TempDir(), "thola-"+u.Username+"-cache")).WithLogger(nil))
		if err != nil {
			return errors.Wrap(err, "error while setting up database")
		}
		if viper.GetBool("db.rebuild") {
			err = badgerDB.db.DropAll()
			if err != nil {
				return errors.Wrap(err, "failed to rebuild the db")
			}
		}
		db.database = &badgerDB
	} else if viper.GetString("db.drivername") == "mysql" {
		checkIfTableExistsQuery := "SHOW TABLES LIKE 'cache';"
		sqlDB := sqlDatabase{}
		if viper.GetString("db.sql.datasourcename") != "" {
			sqlDB.db, err = sqlx.Connect(viper.GetString("db.drivername"), viper.GetString("db.sql.datasourcename"))
			if err != nil {
				return err
			}
		} else {
			return errors.New("no datasourcename set")
		}

		tableNotExist := true
		rows, err := sqlDB.db.Query(checkIfTableExistsQuery)
		if rows != nil {
			tableNotExist = !rows.Next()
			err := rows.Close()
			if err != nil {
				return errors.Wrap(err, "failed to close sql rows")
			}
		}
		if err != nil || tableNotExist || viper.GetBool("db.rebuild") { //!rows.Next() == table does not exist
			err = sqlDB.setupDatabase()
			if err != nil {
				return errors.Wrap(err, "error while setting up database")
			}
		}
		db.database = &sqlDB
	} else if viper.GetString("db.drivername") == "redis" {
		redisDB := redisDatabase{
			db: redis.NewClient(&redis.Options{
				Addr:     viper.GetString("db.redis.addr"),
				Password: viper.GetString("db.redis.password"),
				DB:       viper.GetInt("db.redis.db"),
			}),
		}
		_, err := redisDB.db.Ping().Result()
		if err != nil {
			return errors.Wrap(err, "failed to ping redis db")
		}
		if viper.GetBool("db.rebuild") {
			redisDB.db.FlushAll()
		}
		db.database = &redisDB
	} else {
		return errors.New("invalid drivername, only 'built-in', 'mysql' and 'redis' supported")
	}
	return nil
}

func (d sqlDatabase) setupDatabase() error {
	for _, query := range mysqlSchemaArr {
		_, err := d.db.Exec(query)
		if err != nil {
			_, _ = d.db.Exec(`DROP TABLE IF EXISTS cache;`)
			return errors.Wrap(err, "Could not set up database schema - query: "+query)
		}
	}
	return nil
}

func getDB() (database, error) {
	var err error
	db.Do(func() {
		err = initDB()
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize DB")
	}
	if db.database == nil {
		return nil, errors.New("database was not initialized")
	}
	return db.database, nil
}

func (d *badgerDatabase) SetIdentifyData(ip string, data *network.ConnectionData, response *IdentifyResponse) error {
	txn := d.db.NewTransaction(true)
	defer txn.Discard()

	JSONData, err := parser.ToJSON(&response)
	if err != nil {
		return errors.Wrap(err, "failed to marshall response")
	}
	entry := badger.Entry{
		Key:       []byte("IdentifyResponse-" + ip),
		Value:     JSONData,
		ExpiresAt: uint64(time.Now().Add(cacheExpiration).Unix()),
	}

	err = txn.SetEntry(&entry)
	if err != nil {
		return errors.Wrap(err, "failed to store identify data")
	}

	err = txn.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to store identify data")
	}
	return d.SetConnectionData(ip, data)
}

func (d *sqlDatabase) SetIdentifyData(ip string, data *network.ConnectionData, response *IdentifyResponse) error {
	err := d.insertReplaceQuery(response, ip, "IdentifyResponse")
	if err != nil {
		return errors.Wrap(err, "failed to store identify data")
	}
	err = d.SetConnectionData(ip, data)
	if err != nil {
		return errors.Wrap(err, "failed to store identify data")
	}
	return nil
}

func (d *redisDatabase) SetIdentifyData(ip string, data *network.ConnectionData, response *IdentifyResponse) error {
	JSONData, err := parser.ToJSON(&response)
	if err != nil {
		return errors.Wrap(err, "failed to marshall response")
	}
	d.db.Set("IdentifyResponse-"+ip, JSONData, cacheExpiration)
	return d.SetConnectionData(ip, data)
}

func (d *emptyDatabase) SetIdentifyData(_ string, _ *network.ConnectionData, _ *IdentifyResponse) error {
	return nil
}

func (d *badgerDatabase) GetIdentifyData(ip string) (*IdentifyResponse, error) {
	txn := d.db.NewTransaction(false)
	defer txn.Discard()

	item, err := txn.Get([]byte("IdentifyResponse-" + ip))
	if err != nil {
		return nil, tholaerr.NewNotFoundError("cannot find cache entry")
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get value from db item")
	}

	data := IdentifyResponse{}
	err = json.Unmarshal(value, &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshall identifyResponse")
	}
	return &data, nil
}

func (d *sqlDatabase) GetIdentifyData(ip string) (*IdentifyResponse, error) {
	var identifyResponse IdentifyResponse
	err := d.getEntry(&identifyResponse, ip, "IdentifyResponse")
	if err != nil {
		return nil, err
	}
	return &identifyResponse, nil
}

func (d *redisDatabase) GetIdentifyData(ip string) (*IdentifyResponse, error) {
	value, err := d.db.Get("IdentifyResponse-" + ip).Result()
	if err != nil {
		return nil, tholaerr.NewNotFoundError("cannot find cache entry")
	}
	data := IdentifyResponse{}
	err = json.Unmarshal([]byte(value), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshall identifyResponse")
	}
	return &data, nil
}

func (d *emptyDatabase) GetIdentifyData(_ string) (*IdentifyResponse, error) {
	return nil, tholaerr.NewNotFoundError("no db available")
}

func (d *badgerDatabase) SetConnectionData(ip string, data *network.ConnectionData) error {
	txn := d.db.NewTransaction(true)
	defer txn.Discard()

	JSONData, err := parser.ToJSON(&data)
	if err != nil {
		return errors.Wrap(err, "failed to marshall connection data")
	}
	entry := badger.Entry{
		Key:       []byte("ConnectionData-" + ip),
		Value:     JSONData,
		ExpiresAt: uint64(time.Now().Add(cacheExpiration).Unix()),
	}

	err = txn.SetEntry(&entry)
	if err != nil {
		return errors.Wrap(err, "failed to store connection data")
	}

	err = txn.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to store connection data")
	}
	return nil
}

func (d *sqlDatabase) SetConnectionData(ip string, data *network.ConnectionData) error {
	return d.insertReplaceQuery(data, ip, "ConnectionData")
}

func (d *redisDatabase) SetConnectionData(ip string, data *network.ConnectionData) error {
	JSONData, err := parser.ToJSON(&data)
	if err != nil {
		return errors.Wrap(err, "failed to marshall connectionData")
	}
	d.db.Set("ConnectionData-"+ip, JSONData, cacheExpiration)
	return nil
}

func (d *emptyDatabase) SetConnectionData(_ string, _ *network.ConnectionData) error {
	return nil
}

func (d *badgerDatabase) GetConnectionData(ip string) (*network.ConnectionData, error) {
	txn := d.db.NewTransaction(false)
	defer txn.Discard()

	item, err := txn.Get([]byte("ConnectionData-" + ip))
	if err != nil {
		return nil, tholaerr.NewNotFoundError("cannot find cache entry")
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get value from db item")
	}

	data := network.ConnectionData{}
	err = json.Unmarshal(value, &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshall connectionData")
	}
	return &data, nil
}

func (d *sqlDatabase) GetConnectionData(ip string) (*network.ConnectionData, error) {
	var connectionData network.ConnectionData
	err := d.getEntry(&connectionData, ip, "ConnectionData")
	if err != nil {
		return nil, err
	}
	return &connectionData, nil
}

func (d *redisDatabase) GetConnectionData(ip string) (*network.ConnectionData, error) {
	value, err := d.db.Get("ConnectionData-" + ip).Result()
	if err != nil {
		return nil, tholaerr.NewNotFoundError("cannot find cache entry")
	}
	data := network.ConnectionData{}
	err = json.Unmarshal([]byte(value), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshall connectionData")
	}
	return &data, nil
}

func (d *emptyDatabase) GetConnectionData(_ string) (*network.ConnectionData, error) {
	return nil, tholaerr.NewNotFoundError("no db available")
}

func (d *sqlDatabase) selectQuery(dest interface{}, ip, dataType string) error {
	return d.db.Select(dest, d.db.Rebind("SELECT DATE_FORMAT(time, '%Y-%m-%d %H:%i:%S') as time, data, datatype FROM cache WHERE ip=? AND datatype=?;"), ip, dataType)
}

func (d *sqlDatabase) insertReplaceQuery(data interface{}, ip, dataType string) error {
	JSONData, err := parser.ToJSON(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshall data")
	}

	sb := sqlbuilder.NewInsertBuilder()
	sb.ReplaceInto("cache") // works for insert and replace
	sb.Cols("ip", "datatype", "data")
	sb.Values(ip, dataType, string(JSONData))
	sql, args := sb.Build()
	query, err := sqlbuilder.MySQL.Interpolate(sql, args)
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	_, err = d.db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "failed to exec sql query")
	}
	return nil
}

func (d *sqlDatabase) getEntry(dest interface{}, ip, dataType string) error {
	var results sqlSelectResults
	err := d.selectQuery(&results, ip, dataType)
	if err != nil {
		return errors.Wrap(err, "db select failed")
	}
	if results == nil || len(results) == 0 {
		return tholaerr.NewNotFoundError("cache entry not found")
	}

	res := results[0]
	t, err := time.Parse("2006-01-02 15:04:05", res.Time)
	if err != nil {
		return errors.Wrap(err, "failed to parse timestamp")
	}
	if time.Since(t) > cacheExpiration {
		_, err = d.db.Exec(d.db.Rebind("DELETE FROM cache WHERE ip=? AND datatype=?;"), ip, "IdentifyResponse")
		if err != nil {
			return errors.Wrap(err, "failed to delete expired cache element")
		}
		return tholaerr.NewNotFoundError("found only expired cache entry")
	}

	dataString := `"` + res.Data + `"`
	dataString, err = strconv.Unquote(dataString)
	if err != nil {
		return errors.Wrap(err, "failed to unquote connection data")
	}

	err = json.Unmarshal([]byte(dataString), dest)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshall entry data")
	}
	return nil
}
