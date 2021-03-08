package request

import "context"

// CheckDiskRequest
//
// CheckDiskRequest is a the request struct for the check disk request.
//
// swagger:model
type CheckDiskRequest struct {
	CheckDeviceRequest
	DiskThresholds CheckThresholds `json:"diskThresholds" xml:"diskThresholds"`
}

func (r *CheckDiskRequest) validate(ctx context.Context) error {
	if err := r.DiskThresholds.validate(); err != nil {
		return err
	}
	return r.CheckDeviceRequest.validate(ctx)
}
