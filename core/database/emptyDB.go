package database

import (
	"context"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/tholaerr"
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

func (d *emptyDatabase) CloseConnection(_ context.Context) error {
	return nil
}
