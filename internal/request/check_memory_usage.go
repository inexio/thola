package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

// CheckMemoryUsageRequest
//
// CheckMemoryUsageRequest is a the request struct for the check memory usage request.
//
// swagger:model
type CheckMemoryUsageRequest struct {
	CheckDeviceRequest
	MemoryUsageThresholds monitoringplugin.Thresholds `json:"memoryUsageThresholds" xml:"memoryUsageThresholds"`
}

func (r *CheckMemoryUsageRequest) validate(ctx context.Context) error {
	if err := r.MemoryUsageThresholds.Validate(); err != nil {
		return err
	}
	return r.CheckDeviceRequest.validate(ctx)
}
