package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

// CheckServerRequest
//
// CheckServerRequest is a the request struct for the check server request.
//
// swagger:model
type CheckServerRequest struct {
	CheckDeviceRequest
	UsersThreshold monitoringplugin.Thresholds `json:"usersThreshold" xml:"usersThreshold"`
	ProcsThreshold monitoringplugin.Thresholds `json:"procsThreshold" xml:"procsThreshold"`
}

func (r *CheckServerRequest) validate(ctx context.Context) error {
	if err := r.UsersThreshold.Validate(); err != nil {
		return err
	}

	if err := r.ProcsThreshold.Validate(); err != nil {
		return err
	}

	return r.CheckDeviceRequest.validate(ctx)
}
