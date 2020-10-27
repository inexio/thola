// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/core/tholaerr"
)

func (r *CheckMetricsRequest) process(ctx context.Context) (Response, error) {
	r.init()

	checkInterfaceMetricsRequest := CheckInterfaceMetricsRequest{
		Filter:             r.InterfaceFilter,
		CheckDeviceRequest: r.CheckDeviceRequest,
	}

	checkInterfaceData, err := checkInterfaceMetricsRequest.getData(ctx)
	if err != nil {
		if !tholaerr.IsComponentNotFound(err) {
			r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing check interface request", true)
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	} else {
		err = addCheckInterfacePerformanceData(checkInterfaceData.Interfaces, r.mon)
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data of check interface request", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	checkUPSRequest := CheckUPSRequest{
		CheckDeviceRequest: r.CheckDeviceRequest,
	}

	checkUPSData, err := checkUPSRequest.getData(ctx)
	if err != nil {
		if !tholaerr.IsComponentNotFound(err) {
			r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing check ups request", true)
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	} else {
		err = addCheckUPSPerformanceData(checkUPSData.UPS, r.mon)
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data of check ups request", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
