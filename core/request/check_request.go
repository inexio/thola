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
	WarningMin  value.Value `json:"warningMin" xml:"warningMin"`
	WarningMax  value.Value `json:"warningMax" xml:"warningMax"`
	CriticalMin value.Value `json:"criticalMin" xml:"criticalMin"`
	CriticalMax value.Value `json:"criticalMax" xml:"criticalMax"`
}

func (c *CheckThresholds) validate() error {
	if !c.WarningMin.IsEmpty() && !c.WarningMax.IsEmpty() {
		if cmp, err := c.WarningMin.Cmp(c.WarningMax); err != nil || cmp != -1 {
			return errors.New("warning min and max are invalid")
		}
	}

	if !c.CriticalMin.IsEmpty() && !c.CriticalMax.IsEmpty() {
		if cmp, err := c.CriticalMin.Cmp(c.CriticalMax); err != nil || cmp != -1 {
			return errors.New("critical min and max are invalid")
		}
	}

	if !c.CriticalMin.IsEmpty() && !c.WarningMin.IsEmpty() {
		if cmp, err := c.CriticalMin.Cmp(c.WarningMin); err != nil || cmp != -1 {
			return errors.New("critical and warning min are invalid")
		}
	}

	if !c.WarningMax.IsEmpty() && !c.CriticalMax.IsEmpty() {
		if cmp, err := c.WarningMax.Cmp(c.CriticalMax); err != nil || cmp != -1 {
			return errors.New("warning and critical max are invalid")
		}
	}

	return nil
}

func (c *CheckThresholds) isEmpty() bool {
	return c.WarningMin.IsEmpty() && c.WarningMax.IsEmpty() && c.CriticalMin.IsEmpty() && c.CriticalMax.IsEmpty()
}

func (c *CheckThresholds) checkValue(v value.Value) int {
	if !c.CriticalMin.IsEmpty() {
		if res, err := c.CriticalMin.Cmp(v); err != nil || res != -1 {
			return monitoringplugin.CRITICAL
		}
	}
	if !c.CriticalMax.IsEmpty() {
		if res, err := c.CriticalMax.Cmp(v); err != nil || res != 1 {
			return monitoringplugin.CRITICAL
		}
	}
	if !c.WarningMin.IsEmpty() {
		if res, err := c.WarningMin.Cmp(v); err != nil || res != -1 {
			return monitoringplugin.WARNING
		}
	}
	if !c.WarningMax.IsEmpty() {
		if res, err := c.WarningMax.Cmp(v); err != nil || res != 1 {
			return monitoringplugin.WARNING
		}
	}
	return monitoringplugin.OK
}
