package request

import (
	"context"
	"github.com/inexio/thola/core/network"
)

// CheckTholaServerRequest
//
// CheckTholaServerRequest is the request struct for the check thola server request.
//
// swagger:model
type CheckTholaServerRequest struct {
	CheckRequest
	Timeout *int `json:"timeout" xml:"timeout"`
}

func (r *CheckTholaServerRequest) setupConnection(_ context.Context) (*network.RequestDeviceConnection, error) {
	return &network.RequestDeviceConnection{}, nil
}

func (r *CheckTholaServerRequest) getTimeout() *int {
	return r.Timeout
}

func (r *CheckTholaServerRequest) validate(_ context.Context) error {
	return nil
}

// GetDeviceData returns the device data of the request.
func (r *CheckTholaServerRequest) GetDeviceData() *DeviceData {
	return nil
}
