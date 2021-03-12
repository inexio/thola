package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

// CheckDiskRequest
//
// CheckDiskRequest is a the request struct for the check disk request.
//
// swagger:model
type CheckDiskRequest struct {
	CheckDeviceRequest
	DiskThresholds monitoringplugin.Thresholds `json:"diskThresholds" xml:"diskThresholds"`
}

func (r *CheckDiskRequest) validate(ctx context.Context) error {
	if err := r.DiskThresholds.Validate(); err != nil {
		return err
	}
	return r.CheckDeviceRequest.validate(ctx)
}
