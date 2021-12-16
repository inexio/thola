//go:build !client
// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

func (r *CheckDiskRequest) process(ctx context.Context) (Response, error) {
	r.init()

	com, err := GetCommunicator(ctx, r.BaseRequest)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while getting communicator", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	disk, err := com.GetDiskComponent(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while reading disk", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	duplicateLabelCheckerDisk := make(duplicateLabelChecker)
	for _, disk := range disk.Storages {
		duplicateLabelCheckerDisk.addLabel(disk.Description)
	}

	for _, storage := range disk.Storages {
		if storage.Used != nil {
			var p *monitoringplugin.PerformanceDataPoint

			if (r.DiskThresholds.HasWarning() || r.DiskThresholds.HasCritical()) && storage.Available != nil {
				thresholds := monitoringplugin.Thresholds{
					WarningMin:  0,
					CriticalMin: 0,
				}

				if r.DiskThresholds.HasWarning() {
					thresholds.WarningMax = float64(*storage.Available) * r.DiskThresholds.WarningMax.(float64) / 100
				}
				if r.DiskThresholds.HasCritical() {
					thresholds.CriticalMax = float64(*storage.Available) * r.DiskThresholds.CriticalMax.(float64) / 100
				}

				p = monitoringplugin.NewPerformanceDataPoint("disk_used", *storage.Used).SetUnit("B").SetThresholds(thresholds)
			} else {
				p = monitoringplugin.NewPerformanceDataPoint("disk_used", *storage.Used).SetUnit("B")
			}

			if storage.Description != nil {
				p.SetLabel(duplicateLabelCheckerDisk.getModifiedLabel(storage.Description))
			}

			if storage.Available != nil {
				p.SetMax(*storage.Available)
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
