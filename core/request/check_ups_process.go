// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/core/utility"
)

func (r *CheckUPSRequest) process(ctx context.Context) (Response, error) {
	r.init()

	readUPSResponse, err := r.getData(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing read ups request", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	if readUPSResponse.UPS.AlarmLowVoltageDisconnect != nil {
		err := r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("alarm_low_voltage_disconnect", *readUPSResponse.UPS.AlarmLowVoltageDisconnect))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.UPS.BatteryAmperage != nil {
		err := r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("batt_amperage", *readUPSResponse.UPS.BatteryAmperage))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.UPS.BatteryRemainingTime != nil {
		err := r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("batt_remaining_time", *readUPSResponse.UPS.BatteryRemainingTime))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.UPS.BatteryCapacity != nil {
		err := r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("batt_capacity", *readUPSResponse.UPS.BatteryCapacity))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.UPS.BatteryCurrent != nil {
		err := r.mon.AddPerformanceDataPoint(
			monitoringplugin.NewPerformanceDataPoint("batt_current", *readUPSResponse.UPS.BatteryCurrent).
				SetThresholds(r.BatteryCurrentThresholds))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.UPS.BatteryTemperature != nil {
		err := r.mon.AddPerformanceDataPoint(
			monitoringplugin.NewPerformanceDataPoint("batt_temperature", *readUPSResponse.UPS.BatteryTemperature).
				SetThresholds(r.BatteryTemperatureThresholds))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.UPS.BatteryVoltage != nil {
		err := r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("batt_voltage", *readUPSResponse.UPS.BatteryVoltage))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.UPS.CurrentLoad != nil {
		err := r.mon.AddPerformanceDataPoint(
			monitoringplugin.NewPerformanceDataPoint("current_load", *readUPSResponse.UPS.CurrentLoad).
				SetThresholds(r.CurrentLoadThresholds))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.UPS.MainsVoltageApplied != nil {
		err := r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("mains_voltage_applied", utility.IfThenElse(*readUPSResponse.UPS.MainsVoltageApplied, 1, 0)))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
		r.mon.UpdateStatusIfNot(*readUPSResponse.UPS.MainsVoltageApplied, monitoringplugin.CRITICAL, "Mains voltage is not applied")
	}

	if readUPSResponse.UPS.RectifierCurrent != nil {
		err := r.mon.AddPerformanceDataPoint(
			monitoringplugin.NewPerformanceDataPoint("rectifier_current", *readUPSResponse.UPS.RectifierCurrent).
				SetThresholds(r.RectifierCurrentThresholds))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if readUPSResponse.UPS.SystemVoltage != nil {
		err := r.mon.AddPerformanceDataPoint(
			monitoringplugin.NewPerformanceDataPoint("sys_voltage", *readUPSResponse.UPS.SystemVoltage).
				SetThresholds(r.SystemVoltageThresholds))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}

func (r *CheckUPSRequest) getData(ctx context.Context) (*ReadUPSResponse, error) {
	readUPSRequest := ReadUPSRequest{ReadRequest{r.BaseRequest}}
	response, err := readUPSRequest.process(ctx)
	if err != nil {
		return nil, err
	}

	readUPSResponse := response.(*ReadUPSResponse)
	return readUPSResponse, nil
}
