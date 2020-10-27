// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
	"thola/api/statistics"
	"time"
)

func (r *CheckTholaServerRequest) process(_ context.Context) (Response, error) {
	r.init()

	stats, err := statistics.GetStatistics()
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "failed to get statistics", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	r.mon.UpdateStatus(monitoringplugin.OK, "thola server is running since "+stats.UpSince.Format(time.UnixDate))

	err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("total_request_counter", stats.Requests.TotalCount, "c"))
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("successful_request_counter", stats.Requests.SuccessfulCounter, "c"))
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("failed_request_counter", stats.Requests.FailedCounter, "c"))
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("average_response_time", stats.Requests.AverageResponseTime, "s"))
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("snmp_request_counter", stats.Requests.SNMPRequests, "c"))
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
