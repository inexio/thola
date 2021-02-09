package request

import "context"

// CheckSBCRequest
//
// CheckSBCRequest is a the request struct for the check sbc request.
//
// swagger:model
type CheckSBCRequest struct {
	CheckDeviceRequest
	SystemHealthScoreThresholds CheckThresholds
}

func (r *CheckSBCRequest) validate(ctx context.Context) error {
	if err := r.SystemHealthScoreThresholds.validate(); err != nil {
		return err
	}
	return r.CheckDeviceRequest.validate(ctx)
}
