//go:build !client
// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

func (r *CheckSIEMRequest) process(ctx context.Context) (Response, error) {
	r.init()

	com, err := GetCommunicator(ctx, r.BaseRequest)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while getting communicator", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	siem, err := com.GetSIEMComponent(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while reading siem component", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	if siem.SIEM != nil {
		msg := "SIEM: " + *siem.SIEM

		if siem.SystemVersion != nil {
			msg += " (Version " + *siem.SystemVersion + ")"
		}

		r.mon.UpdateStatus(monitoringplugin.OK, msg)
	}

	if siem.LastRecordedMessagesPerSecondNormalizer != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("last_recorded_messages_per_second_normalizer", *siem.LastRecordedMessagesPerSecondStoreHandler))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.AverageMessagesPerSecondLast5minNormalizer != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("average_messages_per_second_last_5_min_normalizer", *siem.AverageMessagesPerSecondLast5minNormalizer))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.LastRecordedMessagesPerSecondStoreHandler != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("last_recorded_messages_per_second_store_handler", *siem.LastRecordedMessagesPerSecondStoreHandler))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.AverageMessagesPerSecondLast5minStoreHandler != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("average_messages_per_second_last_5_min_store_handler", *siem.AverageMessagesPerSecondLast5minStoreHandler))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.ServicesCurrentlyDown != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("services_currently_down", *siem.ServicesCurrentlyDown))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	//cpu
	if siem.CpuConsumptionCollection != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("cpu_consumption_collection", *siem.CpuConsumptionCollection))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.CpuConsumptionNormalization != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("cpu_consumption_normalization", *siem.CpuConsumptionNormalization))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.CpuConsumptionEnrichment != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("cpu_consumption_enrichment", *siem.CpuConsumptionEnrichment))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.CpuConsumptionIndexing != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("cpu_consumption_indexing", *siem.CpuConsumptionIndexing))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.CpuConsumptionDashboardAlerts != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("cpu_consumption_dashboard_alerts", *siem.CpuConsumptionDashboardAlerts))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	// memory consumption
	if siem.MemoryConsumptionCollection != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("memory_consumption_collection", *siem.MemoryConsumptionCollection))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.MemoryConsumptionNormalization != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("memory_consumption_normalization", *siem.MemoryConsumptionNormalization))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.MemoryConsumptionEnrichment != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("memory_consumption_enrichment", *siem.MemoryConsumptionEnrichment))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.MemoryConsumptionIndexing != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("memory_consumption_indexing", *siem.MemoryConsumptionIndexing))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.MemoryConsumptionDashboardAlerts != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("memory_consumption_dashboard_alerts", *siem.MemoryConsumptionDashboardAlerts))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	//queue
	if siem.QueueCollection != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("queue_collection", *siem.QueueCollection))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.QueueNormalization != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("queue_normalization", *siem.QueueNormalization))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.QueueEnrichment != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("queue_enrichment", *siem.QueueEnrichment))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.QueueIndexing != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("queue_indexing", *siem.QueueIndexing))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if siem.QueueDashboardAlerts != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("queue_dashboard_alerts", *siem.QueueDashboardAlerts))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
