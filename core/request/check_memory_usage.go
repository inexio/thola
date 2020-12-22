package request

import "context"

// CheckMemoryUsageRequest
//
// CheckMemoryUsageRequest is a the request struct for the check memory usage request.
//
// swagger:model
type CheckMemoryUsageRequest struct {
	CheckDeviceRequest
	MemoryUsageThresholds CheckThresholds `json:"memoryUsageThresholds" xml:"memoryUsageThresholds"`
}

func (r *CheckMemoryUsageRequest) validate(ctx context.Context) error {
	if err := r.MemoryUsageThresholds.validate(); err != nil {
		return err
	}
	return r.CheckDeviceRequest.validate(ctx)
}
