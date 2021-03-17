// +build !client

package request

import (
	"context"
	"fmt"
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
			p := monitoringplugin.NewPerformanceDataPoint("disk_available", *storage.Available).SetUnit("KB").SetLabel(*storage.Description)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}

			p = monitoringplugin.NewPerformanceDataPoint("disk_used", *storage.Used).SetUnit("KB").SetLabel(*storage.Description)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}

			// get percentage of free part on the storage
			free := fmt.Sprintf("%.2f", 100-float64(*storage.Used)/float64(*storage.Available)*100)
			p = monitoringplugin.NewPerformanceDataPoint("disk_free", free).SetUnit("%").
				SetLabel(*storage.Description).
				SetThresholds(r.DiskThresholds)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
