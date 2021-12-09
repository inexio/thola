package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

// CheckUPSRequest
//
// CheckUPSRequest is the request struct for the check ups request.
//
// swagger:model
type CheckUPSRequest struct {
	CheckDeviceRequest
	BatteryCurrentThresholds     monitoringplugin.Thresholds `json:"batteryCurrentThresholds" xml:"batteryCurrentThresholds"`
	BatteryTemperatureThresholds monitoringplugin.Thresholds `json:"batteryTemperatureThresholds" xml:"batteryTemperatureThresholds"`
	CurrentLoadThresholds        monitoringplugin.Thresholds `json:"currentLoadThresholds" xml:"currentLoadThresholds"`
	RectifierCurrentThresholds   monitoringplugin.Thresholds `json:"rectifierCurrentThresholds" xml:"rectifierCurrentThresholds"`
	SystemVoltageThresholds      monitoringplugin.Thresholds `json:"systemVoltageThresholds" xml:"systemVoltageThresholds"`
}

func (r *CheckUPSRequest) validate(ctx context.Context) error {
	if err := r.BatteryCurrentThresholds.Validate(); err != nil {
		return err
	}

	if err := r.BatteryTemperatureThresholds.Validate(); err != nil {
		return err
	}

	if err := r.CurrentLoadThresholds.Validate(); err != nil {
		return err
	}

	if err := r.RectifierCurrentThresholds.Validate(); err != nil {
		return err
	}

	if err := r.SystemVoltageThresholds.Validate(); err != nil {
		return err
	}

	return r.CheckDeviceRequest.validate(ctx)
}
