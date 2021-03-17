package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

// CheckCPULoadRequest
//
// CheckCPULoadRequest is a the request struct for the check cpu load request.
//
// swagger:model
type CheckCPULoadRequest struct {
	CheckDeviceRequest
	CPULoadThresholds monitoringplugin.Thresholds `json:"cpuLoadThresholds" xml:"cpuLoadThresholds"`
}

func (r *CheckCPULoadRequest) validate(ctx context.Context) error {
	if err := r.CPULoadThresholds.Validate(); err != nil {
		return err
	}
	return r.CheckDeviceRequest.validate(ctx)
}
