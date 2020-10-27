package statistics

import (
	"context"
	"errors"
	"github.com/inexio/thola/core/network"
	"github.com/labstack/echo"
	"github.com/thoas/stats"
	"net/http"
	"strings"
	"sync"
	"time"
)

var statistics struct {
	sync.Once

	mw        *stats.Stats
	tholaData *tholaData
	startTime time.Time
}

type tholaData struct {
	sync.RWMutex

	SNMPRequests uint64
}

func handlerFunc(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		beginning, _ := statistics.mw.Begin(w)

		h.ServeHTTP(w, r)

		echoRes := w.(*echo.Response)

		statistics.mw.End(beginning, stats.WithStatusCode(echoRes.Status), stats.WithSize(int(echoRes.Size)))
	})
}

// Middleware represents an API middleware
func Middleware() echo.MiddlewareFunc {
	statistics.Do(func() {
		statistics.mw = stats.New()
		statistics.startTime = time.Now()
		statistics.tholaData = &tholaData{}
	})
	return echo.WrapMiddleware(handlerFunc)
}

// Stats includes stats of all requests handled by the api
type Stats struct {
	Pid      int
	UpSince  time.Time
	Requests RequestsStats
}

// RequestsStats includes request stats of all requests handled by the api
type RequestsStats struct {
	TotalCount          int
	SuccessfulCounter   int
	FailedCounter       int
	StatusCodeCount     map[string]int
	SNMPRequests        uint64
	AverageResponseTime float64 // this is the average response time of all time since the api started
	AverageResponseSize int64   // same for response size

}

// GetStatistics returns the current statistics
func GetStatistics() (Stats, error) {
	// this error can be removed when there are other stats than request stats
	if statistics.mw == nil || statistics.tholaData == nil {
		return Stats{}, errors.New("no request stats available")
	}
	data := statistics.mw.Data()
	var s Stats
	s.Pid = data.Pid
	s.UpSince = statistics.startTime

	s.Requests.TotalCount = data.TotalCount
	s.Requests.StatusCodeCount = data.StatusCodeCount
	for code, count := range data.TotalStatusCodeCount {
		if strings.HasPrefix(code, "2") {
			s.Requests.SuccessfulCounter += count
		} else {
			s.Requests.FailedCounter += count
		}
	}

	s.Requests.AverageResponseTime = data.AverageResponseTimeSec
	s.Requests.AverageResponseSize = data.AverageResponseSize

	statistics.tholaData.RLock()
	defer statistics.tholaData.RUnlock()
	s.Requests.SNMPRequests = statistics.tholaData.SNMPRequests

	return s, nil
}

// AddRequestStatistics adds statistics which are only available while processing the request
func AddRequestStatistics(ctx context.Context) {
	if statistics.tholaData == nil {
		return
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return
	}

	statistics.tholaData.Lock()
	defer statistics.tholaData.Unlock()
	statistics.tholaData.SNMPRequests += uint64(con.SNMP.SnmpClient.GetRequestCounter())
}
