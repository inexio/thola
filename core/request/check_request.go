package request

import (
	"errors"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/core/value"
)

// CheckRequest
//
// CheckRequest is a generic response struct for the check request.
//
// swagger:model
type CheckRequest struct {
	mon                  *monitoringplugin.Response
	PrintPerformanceData bool `yaml:"print_performance_data" json:"print_performance_data" xml:"print_performance_data"`
	JSONMetrics          bool `yaml:"json_metrics" json:"json_metrics" xml:"json_metrics"`
}

func (r *CheckRequest) init() {
	r.mon = monitoringplugin.NewResponse("checked")
	r.mon.PrintPerformanceData(r.PrintPerformanceData)
	r.mon.SetPerformanceDataJSONLabel(r.JSONMetrics)
}

func (r *CheckRequest) handlePreProcessError(err error) (Response, error) {
	r.init()
	r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, err.Error(), false)
	return &CheckResponse{r.mon.GetInfo()}, nil
}

// CheckResponse
//
// CheckResponse is a generic response struct for the check plugin format.
//
// swagger:model
type CheckResponse struct {
	monitoringplugin.ResponseInfo
}

// ToCheckPluginOutput returns the response in checkplugin format.
func (c *CheckResponse) ToCheckPluginOutput() ([]byte, error) {
	return []byte(c.RawOutput), nil
}

// GetExitCode returns the exit code of the response.
func (c *CheckResponse) GetExitCode() int {
	return c.ResponseInfo.StatusCode
}

type CheckThresholds struct {
	WarningMin  *float64 `json:"warningMin" xml:"warningMin"`
	WarningMax  *float64 `json:"warningMax" xml:"warningMax"`
	CriticalMin *float64 `json:"criticalMin" xml:"criticalMin"`
	CriticalMax *float64 `json:"criticalMax" xml:"criticalMax"`
}

func (c *CheckThresholds) validate() error {
	if c.WarningMin != nil && c.WarningMax != nil && *c.WarningMin >= *c.WarningMax {
		return errors.New("warning min and max are invalid")
	}

	if c.CriticalMin != nil && c.CriticalMax != nil && *c.CriticalMin >= *c.CriticalMax {
		return errors.New("critical min and max are invalid")
	}

	if c.CriticalMin != nil && c.WarningMin != nil && *c.CriticalMin >= *c.WarningMin {
		return errors.New("critical and warning min are invalid")
	}

	if c.WarningMax != nil && c.CriticalMax != nil && *c.WarningMax >= *c.CriticalMax {
		return errors.New("critical and warning max are invalid")
	}

	return nil
}

func (c *CheckThresholds) isEmpty() bool {
	return c.WarningMin == nil && c.WarningMax == nil && c.CriticalMin == nil && c.CriticalMax == nil
}

func (c *CheckThresholds) checkValue(v value.Value) int {
	if c.CriticalMin != nil {
		if res, err := value.New(*c.CriticalMin).Cmp(v); err != nil || res != -1 {
			return monitoringplugin.CRITICAL
		}
	}
	if c.CriticalMax != nil {
		if res, err := value.New(*c.CriticalMax).Cmp(v); err != nil || res != 1 {
			return monitoringplugin.CRITICAL
		}
	}
	if c.WarningMin != nil {
		if res, err := value.New(*c.WarningMin).Cmp(v); err != nil || res != -1 {
			return monitoringplugin.WARNING
		}
	}
	if c.WarningMax != nil {
		if res, err := value.New(*c.WarningMax).Cmp(v); err != nil || res != 1 {
			return monitoringplugin.WARNING
		}
	}
	return monitoringplugin.OK
}
