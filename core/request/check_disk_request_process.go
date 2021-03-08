// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

func (r *CheckDiskRequest) process(ctx context.Context) (Response, error) {
	r.init()

	diskRequest := ReadDiskRequest{ReadRequest{r.BaseRequest}}
	response, err := diskRequest.process(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing read disk request", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}
	disk := response.(*ReadDiskResponse).Disk

	for _, storage := range disk.Storages {
		if storage.Type != nil && storage.Description != nil && storage.Available != nil && storage.Used != nil {
			// ignore non-physical storage types
			if *storage.Type != "Other" && *storage.Type != "RAM" && *storage.Type != "Virtual Memory" {
				p := monitoringplugin.NewPerformanceDataPoint("disk_available", *storage.Available, "KB").SetLabel(*storage.Description)
				err = r.mon.AddPerformanceDataPoint(p)
				if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
					r.mon.PrintPerformanceData(false)
					return &CheckResponse{r.mon.GetInfo()}, nil
				}
				p = monitoringplugin.NewPerformanceDataPoint("disk_used", *storage.Used, "KB").SetLabel(*storage.Description)
				err = r.mon.AddPerformanceDataPoint(p)
				if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
					r.mon.PrintPerformanceData(false)
					return &CheckResponse{r.mon.GetInfo()}, nil
				}
			}
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
