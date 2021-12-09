package request

import (
	"context"
)

// CheckInterfaceMetricsRequest
//
// CheckInterfaceRequest is the request struct for the check interface metrics request.
//
// swagger:model
type CheckInterfaceMetricsRequest struct {
	PrintInterfaces bool `yaml:"print_interfaces" json:"print_interfaces" xml:"print_interfaces"`
	InterfaceOptions
	CheckDeviceRequest
}

func (r *CheckInterfaceMetricsRequest) validate(ctx context.Context) error {
	if err := r.InterfaceOptions.validate(); err != nil {
		return err
	}
	return r.CheckDeviceRequest.validate(ctx)
}
