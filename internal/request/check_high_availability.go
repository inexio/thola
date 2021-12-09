package request

import (
	"context"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
	"github.com/pkg/errors"
)

// CheckHighAvailabilityRequest
//
// CheckHighAvailabilityRequest is the request struct for the check high-availability request.
//
// swagger:model
type CheckHighAvailabilityRequest struct {
	CheckDeviceRequest
	Role            *string                     `yaml:"role" json:"role" xml:"role"`
	NodesThresholds monitoringplugin.Thresholds `yaml:"nodes_thresholds" json:"nodes_thresholds" xml:"nodes_thresholds"`
}

func (r *CheckHighAvailabilityRequest) validate(ctx context.Context) error {
	if r.Role != nil && *r.Role != "master" && *r.Role != "slave" {
		return fmt.Errorf("invalid high-availability role '%s'", *r.Role)
	}
	if err := r.NodesThresholds.Validate(); err != nil {
		return errors.Wrap(err, "nodes thresholds are invalid")
	}
	return r.CheckDeviceRequest.validate(ctx)
}
