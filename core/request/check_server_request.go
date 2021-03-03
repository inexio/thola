package request

import "context"

// CheckServerRequest
//
// CheckServerRequest is a the request struct for the check server request.
//
// swagger:model
type CheckServerRequest struct {
	CheckDeviceRequest
	ServerThresholds CheckThresholds `json:"serverThresholds" xml:"serverThresholds"`
}

func (r *CheckServerRequest) validate(ctx context.Context) error {
	if err := r.ServerThresholds.validate(); err != nil {
		return err
	}
	return r.CheckDeviceRequest.validate(ctx)
}
