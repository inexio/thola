// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/api/statistics"
	"github.com/inexio/thola/core/database"
	"time"
)

func (r *CheckTholaServerRequest) process(ctx context.Context) (Response, error) {
	r.init()

	stats, err := statistics.GetStatistics()
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "failed to get statistics", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	r.mon.UpdateStatus(monitoringplugin.OK, "thola server is running since "+stats.UpSince.Format(time.UnixDate))

	db, err := database.GetDB(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "failed to get database", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	err = db.CheckConnection(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.CRITICAL, "database is not alive", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("total_request_counter", stats.TotalCount).SetUnit("c"))
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("successful_request_counter", stats.SuccessfulCounter).SetUnit("c"))
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("failed_request_counter", stats.FailedCounter).SetUnit("c"))
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("average_response_time", stats.AverageResponseTime).SetUnit("s"))
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
