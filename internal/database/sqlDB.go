package database

import (
	"context"
	"encoding/json"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/parser"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"time"
)

type sqlDatabase struct {
	db *sqlx.DB
}

type sqlSelectResults []struct {
	Time     string
	Data     string
	Datatype string
}

var mysqlSchemaArr = []string{
	`DROP TABLE IF EXISTS cache;`,

	`CREATE TABLE cache (
		id INTEGER PRIMARY KEY,
		ip varchar(255) NOT NULL,
		datatype varchar(255) NOT NULL,
		data text NOT NULL,
		time datetime DEFAULT current_timestamp NOT NULL,
		CONSTRAINT unique_entries UNIQUE (ip, datatype)
		);`,
	`ALTER TABLE cache MODIFY id int(11) NOT NULL AUTO_INCREMENT;`,
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

func (d *sqlDatabase) SetDeviceProperties(ctx context.Context, ip string, data device.Device) error {
	err := d.insertReplaceQuery(ctx, data, ip, "DeviceInfo")
	if err != nil {
		return errors.Wrap(err, "failed to store device data")
	}
	return nil
}

func (d *sqlDatabase) GetDeviceProperties(ctx context.Context, ip string) (device.Device, error) {
	var identifyResponse device.Device
	err := d.getEntry(ctx, &identifyResponse, ip, "DeviceInfo")
	if err != nil {
		return device.Device{}, err
	}
	return identifyResponse, nil
}

func (d *sqlDatabase) SetConnectionData(ctx context.Context, ip string, data network.ConnectionData) error {
	return d.insertReplaceQuery(ctx, data, ip, "ConnectionData")
}

func (d *sqlDatabase) GetConnectionData(ctx context.Context, ip string) (network.ConnectionData, error) {
	var connectionData network.ConnectionData
	err := d.getEntry(ctx, &connectionData, ip, "ConnectionData")
	if err != nil {
		return network.ConnectionData{}, err
	}
	return connectionData, nil
}

func (d *sqlDatabase) CheckConnection(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *sqlDatabase) CloseConnection(ctx context.Context) error {
	log.Ctx(ctx).Debug().Msg("closing connection to mysql database")
	return d.db.Close()
}

func (d *sqlDatabase) insertReplaceQuery(ctx context.Context, data interface{}, ip, dataType string) error {
	JSONData, err := parser.ToJSON(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshall data")
	}

	_, err = d.db.ExecContext(ctx, d.db.Rebind("REPLACE INTO cache (ip, datatype, data) VALUES (?, ?, ?);"), ip, dataType, string(JSONData))
	if err != nil {
		return errors.Wrap(err, "failed to exec sql query")
	}
	return nil
}

func (d *sqlDatabase) getEntry(ctx context.Context, dest interface{}, ip, dataType string) error {
	var results sqlSelectResults
	err := d.db.SelectContext(ctx, &results, d.db.Rebind("SELECT DATE_FORMAT(time, '%Y-%m-%d %H:%i:%S') as time, data, datatype FROM cache WHERE ip=? AND datatype=?;"), ip, dataType)
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
		_, err = d.db.ExecContext(ctx, d.db.Rebind("DELETE FROM cache WHERE ip=? AND datatype=?;"), ip, "IdentifyResponse")
		if err != nil {
			return errors.Wrap(err, "failed to delete expired cache element")
		}
		return tholaerr.NewNotFoundError("found only expired cache entry")
	}

	err = json.Unmarshal([]byte(res.Data), dest)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshall entry data")
	}
	return nil
}
