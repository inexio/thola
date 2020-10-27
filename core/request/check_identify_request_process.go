// +build !client

package request

import (
	"context"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/utility"
	"github.com/pkg/errors"
)

func (r *CheckIdentifyRequest) process(ctx context.Context) (Response, error) {
	r.init()
	r.failedExpectations = make(map[string]IdentifyExpectationResult)

	identifyRequest := IdentifyRequest{r.BaseRequest}
	response, err := identifyRequest.process(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing identify request", true) {
		return &CheckIdentifyResponse{
			CheckResponse:  CheckResponse{r.mon.GetInfo()},
			IdentifyResult: nil,
		}, nil
	}

	identifyResponse := response.(*IdentifyResponse)
	r.compareExpectations(identifyResponse)

	return &CheckIdentifyResponse{
		CheckResponse:      CheckResponse{r.mon.GetInfo()},
		IdentifyResult:     &identifyResponse.Device,
		FailedExpectations: r.failedExpectations,
	}, nil
}

func (r *CheckIdentifyRequest) compareExpectations(response *IdentifyResponse) {
	if r.Expectations.Class != "" {
		r.mon.UpdateStatusIf(response.Class != r.Expectations.Class, utility.IfThenElseInt(r.OsDiffWarning, monitoringplugin.WARNING, monitoringplugin.CRITICAL), fmt.Sprintf("OS: expected: \"%s\", got: \"%s\"", r.Expectations.Class, response.Class))
	}
	if r.Expectations.Properties.Vendor != nil {
		var failed bool
		var got string
		if response.Properties.Vendor == nil {
			failed = true
			got = "null"
		} else if *response.Properties.Vendor != *r.Expectations.Properties.Vendor {
			failed = true
			got = *response.Properties.Vendor
		}
		if r.mon.UpdateStatusIf(failed, utility.IfThenElseInt(r.VendorDiffWarning, monitoringplugin.WARNING, monitoringplugin.CRITICAL), fmt.Sprintf("Vendor: expected: \"%s\", got: \"%s\"", *r.Expectations.Properties.Vendor, got)) {
			r.failedExpectations["vendor"] = IdentifyExpectationResult{
				Expected: *r.Expectations.Properties.Vendor,
				Got:      got,
			}
		}
	}
	if r.Expectations.Properties.Model != nil {
		var failed bool
		var got string
		if response.Properties.Model == nil {
			failed = true
			got = "null"
		} else if *response.Properties.Model != *r.Expectations.Properties.Model {
			failed = true
			got = *response.Properties.Model
		}
		if r.mon.UpdateStatusIf(failed, utility.IfThenElseInt(r.ModelDiffWarning, monitoringplugin.WARNING, monitoringplugin.CRITICAL), fmt.Sprintf("Model: expected: \"%s\", got: \"%s\"", *r.Expectations.Properties.Model, got)) {
			r.failedExpectations["model"] = IdentifyExpectationResult{
				Expected: *r.Expectations.Properties.Model,
				Got:      got,
			}
		}
	}
	if r.Expectations.Properties.ModelSeries != nil {
		var failed bool
		var got string
		if response.Properties.ModelSeries == nil {
			failed = true
			got = "null"
		} else if *response.Properties.ModelSeries != *r.Expectations.Properties.ModelSeries {
			failed = true
			got = *response.Properties.ModelSeries
		}
		if r.mon.UpdateStatusIf(failed, utility.IfThenElseInt(r.ModelSeriesDiffWarning, monitoringplugin.WARNING, monitoringplugin.CRITICAL), fmt.Sprintf("ModelSeries: expected: \"%s\", got: \"%s\"", *r.Expectations.Properties.ModelSeries, got)) {
			r.failedExpectations["model_series"] = IdentifyExpectationResult{
				Expected: *r.Expectations.Properties.ModelSeries,
				Got:      got,
			}
		}
	}
	if r.Expectations.Properties.SerialNumber != nil {
		var failed bool
		var got string
		if response.Properties.SerialNumber == nil {
			failed = true
			got = "null"
		} else if *response.Properties.SerialNumber != *r.Expectations.Properties.SerialNumber {
			failed = true
			got = *response.Properties.SerialNumber
		}
		if r.mon.UpdateStatusIf(failed, utility.IfThenElseInt(r.SerialNumberDiffWarning, monitoringplugin.WARNING, monitoringplugin.CRITICAL), fmt.Sprintf("SerialNumber: expected: \"%s\", got: \"%s\"", *r.Expectations.Properties.SerialNumber, got)) {
			r.failedExpectations["serial_number"] = IdentifyExpectationResult{
				Expected: *r.Expectations.Properties.SerialNumber,
				Got:      got,
			}
		}
	}
	if r.Expectations.Properties.OSVersion != nil {
		var failed bool
		var got string
		if response.Properties.OSVersion == nil {
			failed = true
			got = "null"
		} else if *response.Properties.OSVersion != *r.Expectations.Properties.OSVersion {
			failed = true
			got = *response.Properties.OSVersion
		}
		if r.mon.UpdateStatusIf(failed, utility.IfThenElseInt(r.OsVersionDiffWarning, monitoringplugin.WARNING, monitoringplugin.CRITICAL), fmt.Sprintf("OSVersion: expected: \"%s\", got: \"%s\"", *r.Expectations.Properties.OSVersion, got)) {
			r.failedExpectations["version"] = IdentifyExpectationResult{
				Expected: *r.Expectations.Properties.OSVersion,
				Got:      got,
			}
		}
	}
}

func (r *CheckIdentifyRequest) handlePreProcessError(err error) (Response, error) {
	r.init()
	r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, err.Error(), false)
	return &CheckIdentifyResponse{
		CheckResponse:      CheckResponse{r.mon.GetInfo()},
		IdentifyResult:     nil,
		FailedExpectations: nil,
	}, nil
}

func (r *CheckIdentifyRequest) validate() error {
	err := r.BaseRequest.validate()
	if err != nil {
		return errors.Wrap(err, "base request is not valid")
	}
	if r.Expectations == (device.Device{}) {
		return errors.New("no expectations given")
	}
	return nil
}
