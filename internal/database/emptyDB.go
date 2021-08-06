package database

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/rs/zerolog/log"
)

type emptyDatabase struct{}

func (d *emptyDatabase) SetDeviceProperties(_ context.Context, _ string, _ device.Device) error {
	return nil
}

func (d *emptyDatabase) GetDeviceProperties(_ context.Context, _ string) (device.Device, error) {
	return device.Device{}, tholaerr.NewNotFoundError("no db available")
}

func (d *emptyDatabase) SetConnectionData(_ context.Context, _ string, _ network.ConnectionData) error {
	return nil
}

func (d *emptyDatabase) GetConnectionData(_ context.Context, _ string) (network.ConnectionData, error) {
	return network.ConnectionData{}, tholaerr.NewNotFoundError("no db available")
}

func (d *emptyDatabase) CheckConnection(_ context.Context) error {
	return nil
}

func (d *emptyDatabase) CloseConnection(ctx context.Context) error {
	log.Ctx(ctx).Debug().Msg("closing connection to empty database")
	return nil
}
