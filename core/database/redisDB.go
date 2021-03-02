package database

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/parser"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type redisDatabase struct {
	db *redis.Client
}

func (d *redisDatabase) SetDeviceProperties(ctx context.Context, ip string, data device.Device) error {
	JSONData, err := parser.ToJSON(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshall response")
	}
	_, err = d.db.Set(ctx, "DeviceInfo-"+ip, JSONData, cacheExpiration).Result()
	if err != nil && !db.ignoreFailure {
		return errors.Wrap(err, "failed to store device data")
	}
	return nil
}

func (d *redisDatabase) GetDeviceProperties(ctx context.Context, ip string) (device.Device, error) {
	value, err := d.db.Get(ctx, "DeviceInfo-"+ip).Result()
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
	JSONData, err := parser.ToJSON(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshall connectionData")
	}
	_, err = d.db.Set(ctx, "ConnectionData-"+ip, JSONData, cacheExpiration).Result()
	if err != nil && !db.ignoreFailure {
		return errors.Wrap(err, "failed to store connection data")
	}
	return nil
}

func (d *redisDatabase) GetConnectionData(ctx context.Context, ip string) (network.ConnectionData, error) {
	value, err := d.db.Get(ctx, "ConnectionData-"+ip).Result()
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
	_, err := d.db.Ping(ctx).Result()
	return err
}

func (d *redisDatabase) CloseConnection(ctx context.Context) error {
	log.Ctx(ctx).Trace().Msg("closing connection to redis database")
	return d.db.Close()
}
