package statistics

import (
	"github.com/labstack/echo/v4"
	"math"
	"net/http"
	"sync"
	"time"
)

var stats statistics

type statistics struct {
	sync.Once
	sync.RWMutex

	startTime         time.Time
	successfulCounter int
	failedCounter     int
	totalResponseTime time.Duration
}

// Statistics includes stats of all requests handled by the api
type Statistics struct {
	UpSince             time.Time
	TotalCount          int
	SuccessfulCounter   int
	FailedCounter       int
	AverageResponseTime float64
}

func handlerFunc(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		beginning := time.Now()

		h.ServeHTTP(w, r)

		echoRes := w.(*echo.Response)

		stats.add(beginning, echoRes.Status)
	})
}

// Middleware represents an API middleware
func Middleware() echo.MiddlewareFunc {
	stats.Do(func() {
		stats.init()
	})
	return echo.WrapMiddleware(handlerFunc)
}

func (s *statistics) init() {
	s.startTime = time.Now()
}

func (s *statistics) add(startTime time.Time, statusCode int) {
	s.Lock()
	defer s.Unlock()

	s.totalResponseTime += time.Since(startTime)

	if statusCode >= 200 && statusCode <= 299 {
		s.successfulCounter++
	} else {
		s.failedCounter++
	}
}

// GetStatistics returns the current statistics
func GetStatistics() (Statistics, error) {
	stats.RLock()
	defer stats.RUnlock()

	s := Statistics{
		UpSince:             stats.startTime,
		TotalCount:          stats.successfulCounter + stats.failedCounter,
		SuccessfulCounter:   stats.successfulCounter,
		FailedCounter:       stats.failedCounter,
		AverageResponseTime: 0,
	}

	if s.TotalCount == 0 {
		return s, nil
	} else {
		avgNs := int64(stats.totalResponseTime) / int64(s.TotalCount)
		avgSec := float64(avgNs) / float64(time.Second)
		s.AverageResponseTime = math.Floor(avgSec*1000) / 1000
	}

	return s, nil
}
