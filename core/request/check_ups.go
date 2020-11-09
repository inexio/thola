package request

// CheckUPSRequest
//
// CheckUPSRequest is a the request struct for the check ups request.
//
// swagger:model
type CheckUPSRequest struct {
	CheckDeviceRequest
	BatteryCurrentThresholds     CheckThresholds `json:"batteryCurrentThresholds" xml:"batteryCurrentThresholds"`
	BatteryTemperatureThresholds CheckThresholds `json:"batteryTemperatureThresholds" xml:"batteryTemperatureThresholds"`
	CurrentLoadThresholds        CheckThresholds `json:"currentLoadThresholds" xml:"currentLoadThresholds"`
	RectifierCurrentThresholds   CheckThresholds `json:"rectifierCurrentThresholds" xml:"rectifierCurrentThresholds"`
	SystemVoltageThresholds      CheckThresholds `json:"systemVoltageThresholds" xml:"systemVoltageThresholds"`
}

func (r *CheckUPSRequest) validate() error {
	if err := r.BatteryCurrentThresholds.validate(); err != nil {
		return err
	}

	if err := r.BatteryTemperatureThresholds.validate(); err != nil {
		return err
	}

	if err := r.CurrentLoadThresholds.validate(); err != nil {
		return err
	}

	if err := r.RectifierCurrentThresholds.validate(); err != nil {
		return err
	}

	if err := r.SystemVoltageThresholds.validate(); err != nil {
		return err
	}

	return r.CheckDeviceRequest.validate()
}
