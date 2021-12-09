//go:build !client
// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/internal/utility"
)

func (r *CheckUPSRequest) process(ctx context.Context) (Response, error) {
	r.init()

	com, err := GetCommunicator(ctx, r.BaseRequest)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while getting communicator", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	readUPSResponse, err := com.GetUPSComponent(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while reading ups", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing read ups request", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	if readUPSResponse.AlarmLowVoltageDisconnect != nil {
		err := r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("alarm_low_voltage_disconnect", *readUPSResponse.AlarmLowVoltageDisconnect))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.BatteryAmperage != nil {
		err := r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("batt_amperage", *readUPSResponse.BatteryAmperage))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.BatteryRemainingTime != nil {
		err := r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("batt_remaining_time", *readUPSResponse.BatteryRemainingTime))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.BatteryCapacity != nil {
		err := r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("batt_capacity", *readUPSResponse.BatteryCapacity))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.BatteryCurrent != nil {
		err := r.mon.AddPerformanceDataPoint(
			monitoringplugin.NewPerformanceDataPoint("batt_current", *readUPSResponse.BatteryCurrent).
				SetThresholds(r.BatteryCurrentThresholds))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.BatteryTemperature != nil {
		err := r.mon.AddPerformanceDataPoint(
			monitoringplugin.NewPerformanceDataPoint("batt_temperature", *readUPSResponse.BatteryTemperature).
				SetThresholds(r.BatteryTemperatureThresholds))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.BatteryVoltage != nil {
		err := r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("batt_voltage", *readUPSResponse.BatteryVoltage))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.CurrentLoad != nil {
		err := r.mon.AddPerformanceDataPoint(
			monitoringplugin.NewPerformanceDataPoint("current_load", *readUPSResponse.CurrentLoad).
				SetThresholds(r.CurrentLoadThresholds))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.MainsVoltageApplied != nil {
		err := r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("mains_voltage_applied", utility.IfThenElse(*readUPSResponse.MainsVoltageApplied, 1, 0)))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
		r.mon.UpdateStatusIfNot(*readUPSResponse.MainsVoltageApplied, monitoringplugin.CRITICAL, "Mains voltage is not applied")
	}

	if readUPSResponse.RectifierCurrent != nil {
		err := r.mon.AddPerformanceDataPoint(
			monitoringplugin.NewPerformanceDataPoint("rectifier_current", *readUPSResponse.RectifierCurrent).
				SetThresholds(r.RectifierCurrentThresholds))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.SystemVoltage != nil {
		err := r.mon.AddPerformanceDataPoint(
			monitoringplugin.NewPerformanceDataPoint("sys_voltage", *readUPSResponse.SystemVoltage).
				SetThresholds(r.SystemVoltageThresholds))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
