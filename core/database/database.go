package database

import (
	"context"
	"github.com/dgraph-io/badger/v2"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql" //needed for sql driver
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"os/user"
	"path/filepath"
	"sync"
	"time"
)

var db struct {
	sync.Once
	Database

	ignoreFailure bool
}

var cacheExpiration time.Duration

type Database interface {
	SetDeviceProperties(ctx context.Context, ip string, data device.Device) error
	GetDeviceProperties(ctx context.Context, ip string) (device.Device, error)
	SetConnectionData(ctx context.Context, ip string, data network.ConnectionData) error
	GetConnectionData(ctx context.Context, ip string) (network.ConnectionData, error)
	CheckConnection(ctx context.Context) error
	CloseConnection(ctx context.Context) error
}

func initDB(ctx context.Context) error {
	if viper.GetBool("db.no-cache") {
		log.Trace().Msg("initialized empty database")
		db.Database = &emptyDatabase{}
		return nil
	}

	var err error
	cacheExpiration, err = time.ParseDuration(viper.GetString("db.duration"))
	if err != nil {
		return errors.Wrap(err, "failed to parse cache expiration")
	}

	db.ignoreFailure = viper.GetBool("db.ignore-db-failure")

	drivername := viper.GetString("db.drivername")

	if drivername == "built-in" {
		badgerDB := badgerDatabase{}
		u, err := user.Current()
		if err != nil {
			return errors.Wrap(err, "failed to get username")
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
		db.Database = &badgerDB
	} else if drivername == "mysql" {
		checkIfTableExistsQuery := "SHOW TABLES LIKE 'cache';"
		sqlDB := sqlDatabase{}
		if viper.GetString("db.sql.datasourcename") != "" {
			sqlDB.db, err = sqlx.ConnectContext(ctx, viper.GetString("db.drivername"), viper.GetString("db.sql.datasourcename"))
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
		db.Database = &sqlDB
	} else if drivername == "redis" {
		redisDB := redisDatabase{
			db: redis.NewClient(&redis.Options{
				Addr:     viper.GetString("db.redis.addr"),
				Password: viper.GetString("db.redis.password"),
				DB:       viper.GetInt("db.redis.db"),
			}),
		}
		err := redisDB.CheckConnection(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to ping redis db")
		}
		if viper.GetBool("db.rebuild") {
			_, err := redisDB.db.FlushAll(ctx).Result()
			if err != nil {
				return errors.Wrap(err, "failed to rebuild redis db")
			}
		}
		db.Database = &redisDB
	} else {
		return errors.New("invalid drivername, only 'built-in', 'mysql' and 'redis' supported")
	}
	log.Trace().Msg("initialized " + drivername + " database")
	return nil
}

func GetDB(ctx context.Context) (Database, error) {
	var err error
	db.Do(func() {
		err = initDB(ctx)
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize DB")
	}
	if db.Database == nil {
		return nil, errors.New("database was not initialized")
	}
	return db.Database, nil
}
