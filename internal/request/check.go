package request

import (
	"github.com/inexio/go-monitoringplugin"
	"strconv"
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
	_ = r.mon.SetInvalidCharacterBehavior(monitoringplugin.InvalidCharacterReplaceWithErrorAndSetUNKNOWN, "")
	r.mon.PrintPerformanceData(r.PrintPerformanceData)
	r.mon.SetPerformanceDataJSONLabel(r.JSONMetrics)
}

func (r *CheckRequest) HandlePreProcessError(err error) (Response, error) {
	r.init()
	r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, err.Error(), false)
	return &CheckResponse{r.mon.GetInfo()}, nil
}

type labelCounter struct {
	duplicated bool
	current    int
}

type duplicateLabelChecker map[string]*labelCounter

func (d *duplicateLabelChecker) addLabel(label *string) {
	l := ""
	if label != nil {
		l = *label
	}
	if lc, ok := (*d)[l]; ok {
		lc.duplicated = true
	} else {
		(*d)[l] = &labelCounter{
			duplicated: false,
			current:    1,
		}
	}
}

func (d *duplicateLabelChecker) getModifiedLabel(label *string) string {
	l := ""
	if label != nil {
		l = *label
	}
	if lc, ok := (*d)[l]; ok {
		if lc.duplicated {
			if l != "" {
				l += "_"
			}
			l += strconv.Itoa(lc.current)
			lc.current++
		}
	}
	return l
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
