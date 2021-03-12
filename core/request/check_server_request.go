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
	ServerThresholds monitoringplugin.Thresholds `json:"serverThresholds" xml:"serverThresholds"`
}

func (r *CheckServerRequest) validate(ctx context.Context) error {
	if err := r.ServerThresholds.Validate(); err != nil {
		return err
	}
	return r.CheckDeviceRequest.validate(ctx)
}
