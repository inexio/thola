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
			var p *monitoringplugin.PerformanceDataPoint

			if r.DiskThresholds.HasWarning() && r.DiskThresholds.HasCritical() { // check if max thresholds exist
				warningMax := float64(*storage.Available) * r.DiskThresholds.WarningMax.(float64) / 100
				criticalMax := float64(*storage.Available) * r.DiskThresholds.CriticalMax.(float64) / 100
				thresholds := monitoringplugin.Thresholds{WarningMin: 0, WarningMax: warningMax, CriticalMin: 0, CriticalMax: criticalMax}
				p = monitoringplugin.NewPerformanceDataPoint("disk_used", *storage.Used).SetUnit("KB").SetLabel(*storage.Description).SetThresholds(thresholds).SetMax(float64(*storage.Available))
			} else {
				p = monitoringplugin.NewPerformanceDataPoint("disk_used", *storage.Used).SetUnit("KB").SetLabel(*storage.Description).SetMax(float64(*storage.Available))
			}

			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
