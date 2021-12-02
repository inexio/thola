package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

// CheckSBCRequest
//
// CheckSBCRequest is the request struct for the check sbc request.
//
// swagger:model
type CheckSBCRequest struct {
	CheckDeviceRequest
	SystemHealthScoreThresholds monitoringplugin.Thresholds
}

func (r *CheckSBCRequest) validate(ctx context.Context) error {
	if err := r.SystemHealthScoreThresholds.Validate(); err != nil {
		return err
	}
	return r.CheckDeviceRequest.validate(ctx)
}
