package database

import (
	"context"
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/parser"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type redisDatabase struct {
	pool redis.Pool
}

func (d *redisDatabase) SetDeviceProperties(ctx context.Context, ip string, data device.Device) error {
	conn, err := d.pool.GetContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get connection to redis database")
	}
	defer conn.Close()

	JSONData, err := parser.ToJSON(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshall response")
	}
	_, err = conn.Do("SETEX", "DeviceInfo-"+ip, cacheExpiration.Seconds(), JSONData)
	if err != nil && !db.ignoreFailure {
		return errors.Wrap(err, "failed to store device data")
	}
	return nil
}

func (d *redisDatabase) GetDeviceProperties(ctx context.Context, ip string) (device.Device, error) {
	conn, err := d.pool.GetContext(ctx)
	if err != nil {
		return device.Device{}, errors.Wrap(err, "failed to get connection to redis database")
	}
	defer conn.Close()

	value, err := redis.String(conn.Do("GET", "DeviceInfo-"+ip))
	if err != nil {
		return device.Device{}, tholaerr.NewNotFoundError("cannot find cache entry")
	}
	data := device.Device{}
	err = json.Unmarshal([]byte(value), &data)
	if err != nil {
		return device.Device{}, errors.Wrap(err, "failed to unmarshall device properties")
	}
	return data, nil
}

func (d *redisDatabase) SetConnectionData(ctx context.Context, ip string, data network.ConnectionData) error {
	conn, err := d.pool.GetContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get connection to redis database")
	}
	defer conn.Close()

	JSONData, err := parser.ToJSON(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshall connectionData")
	}
	_, err = conn.Do("SETEX", "ConnectionData-"+ip, cacheExpiration.Seconds(), JSONData)
	if err != nil && !db.ignoreFailure {
		return errors.Wrap(err, "failed to store connection data")
	}
	return nil
}

func (d *redisDatabase) GetConnectionData(ctx context.Context, ip string) (network.ConnectionData, error) {
	conn, err := d.pool.GetContext(ctx)
	if err != nil {
		return network.ConnectionData{}, errors.Wrap(err, "failed to get connection to redis database")
	}
	defer conn.Close()

	value, err := redis.String(conn.Do("GET", "ConnectionData-"+ip))
	if err != nil {
		return network.ConnectionData{}, tholaerr.NewNotFoundError("cannot find cache entry")
	}
	data := network.ConnectionData{}
	err = json.Unmarshal([]byte(value), &data)
	if err != nil {
		return network.ConnectionData{}, errors.Wrap(err, "failed to unmarshall connectionData")
	}
	return data, nil
}

func (d *redisDatabase) CheckConnection(ctx context.Context) error {
	conn, err := d.pool.GetContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get connection to redis database")
	}
	defer conn.Close()

	if conn.Err() != nil {
		return errors.Wrap(err, "connection establishment to redis database failed")
	}
	res, err := redis.String(conn.Do("PING"))
	if err != nil {
		return errors.Wrap(err, "sending command to redis database failed")
	}
	if res != "PONG" {
		return errors.New("redis database didn't respond with 'PONG' to 'PING' command")
	}
	return err
}

func (d *redisDatabase) CloseConnection(ctx context.Context) error {
	log.Ctx(ctx).Debug().Msg("closing connection to redis database")
	return d.pool.Close()
}
