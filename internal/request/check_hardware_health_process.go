// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/internal/device"
)

func (r *CheckHardwareHealthRequest) process(ctx context.Context) (Response, error) {
	r.init()

	com, err := GetCommunicator(ctx, r.BaseRequest)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while getting communicator", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	res, err := com.GetHardwareHealthComponent(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while reading hardware-health", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	if res.EnvironmentMonitorState != nil {
		stateInt, err := (*res.EnvironmentMonitorState).GetInt()
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "read out invalid environment monitor state", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("environment_monitor_state", stateInt))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}

		r.mon.UpdateStatusIf((*res.EnvironmentMonitorState) != device.HardwareHealthComponentStateNormal, monitoringplugin.CRITICAL, "environment monitor state is critical")
	}

	// check duplicate labels
	duplicateLabelCheckerFans := make(duplicateLabelChecker)
	for _, fan := range res.Fans {
		duplicateLabelCheckerFans.addLabel(fan.Description)
	}
	for _, fan := range res.Fans {
		if fan.State == nil {
			continue
		}

		stateInt, err := (*fan.State).GetInt()
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "read out invalid hardware health component state for fan", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}

		p := monitoringplugin.NewPerformanceDataPoint("fan_state", stateInt)

		outputDescription := "fan state"
		if label := duplicateLabelCheckerFans.getModifiedLabel(fan.Description); label != "" {
			p.SetLabel(label)
			outputDescription += " (" + label + ")"
		}

		err = r.mon.AddPerformanceDataPoint(p)
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}

		r.mon.UpdateStatusIf(*fan.State == device.HardwareHealthComponentStateWarning, monitoringplugin.WARNING, outputDescription+" is warning")
		r.mon.UpdateStatusIf(*fan.State == device.HardwareHealthComponentStateCritical, monitoringplugin.CRITICAL, outputDescription+" is critical")
	}

	// check duplicate labels
	duplicateLabelCheckerPS := make(duplicateLabelChecker)
	for _, ps := range res.PowerSupply {
		duplicateLabelCheckerPS.addLabel(ps.Description)
	}
	for _, powerSupply := range res.PowerSupply {
		if powerSupply.State == nil {
			continue
		}

		stateInt, err := (*powerSupply.State).GetInt()
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "read out invalid hardware health component state for power supply", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}

		p := monitoringplugin.NewPerformanceDataPoint("power_supply_state", stateInt)

		outputDescription := "power supply state"
		if label := duplicateLabelCheckerPS.getModifiedLabel(powerSupply.Description); label != "" {
			p.SetLabel(label)
			outputDescription += " (" + label + ")"
		}

		err = r.mon.AddPerformanceDataPoint(p)
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}

		r.mon.UpdateStatusIf(*powerSupply.State == device.HardwareHealthComponentStateWarning, monitoringplugin.WARNING, outputDescription+" is warning")
		r.mon.UpdateStatusIf(*powerSupply.State == device.HardwareHealthComponentStateCritical, monitoringplugin.CRITICAL, outputDescription+" is critical")
	}

	// check duplicate labels
	duplicateLabelCheckerTemp := make(duplicateLabelChecker)
	for _, t := range res.Temperature {
		duplicateLabelCheckerTemp.addLabel(t.Description)
	}
	for _, temp := range res.Temperature {
		if temp.State == nil && temp.Temperature == nil {
			continue
		}

		if temp.Temperature != nil {
			p := monitoringplugin.NewPerformanceDataPoint("temperature", *temp.Temperature)

			if label := duplicateLabelCheckerTemp.getModifiedLabel(temp.Description); label != "" {
				p.SetLabel(label)
			}

			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}
		if temp.State != nil {
			stateInt, err := (*temp.State).GetInt()
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "read out invalid hardware health component state for temperature", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}

			p := monitoringplugin.NewPerformanceDataPoint("temperature_state", stateInt)

			outputDescription := "temperature state"
			if label := duplicateLabelCheckerTemp.getModifiedLabel(temp.Description); label != "" {
				p.SetLabel(label)
				outputDescription += " (" + label + ")"
			}

			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}

			r.mon.UpdateStatusIf(*temp.State == device.HardwareHealthComponentStateWarning, monitoringplugin.WARNING, outputDescription+" is warning")
			r.mon.UpdateStatusIf(*temp.State == device.HardwareHealthComponentStateCritical, monitoringplugin.CRITICAL, outputDescription+" is critical")
		}
	}

	// check duplicate labels
	duplicateLabelCheckerVolt := make(duplicateLabelChecker)
	for _, v := range res.Voltage {
		duplicateLabelCheckerVolt.addLabel(v.Description)
	}
	for _, volt := range res.Voltage {
		if volt.State == nil && volt.Voltage == nil {
			continue
		}

		if volt.Voltage != nil {
			p := monitoringplugin.NewPerformanceDataPoint("voltage", *volt.Voltage)

			if label := duplicateLabelCheckerVolt.getModifiedLabel(volt.Description); label != "" {
				p.SetLabel(label)
			}

			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}
		if volt.State != nil {
			stateInt, err := (*volt.State).GetInt()
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "read out invalid hardware health component state for voltage", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}

			p := monitoringplugin.NewPerformanceDataPoint("voltage_state", stateInt)

			outputDescription := "voltage state"
			if label := duplicateLabelCheckerVolt.getModifiedLabel(volt.Description); label != "" {
				p.SetLabel(label)
				outputDescription += " (" + label + ")"
			}

			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}

			r.mon.UpdateStatusIf(*volt.State == device.HardwareHealthComponentStateWarning, monitoringplugin.WARNING, outputDescription+" is warning")
			r.mon.UpdateStatusIf(*volt.State == device.HardwareHealthComponentStateCritical, monitoringplugin.CRITICAL, outputDescription+" is critical")
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
