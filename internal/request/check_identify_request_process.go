// +build !client

package request

import (
	"context"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/utility"
	"github.com/pkg/errors"
)

func (r *CheckIdentifyRequest) process(ctx context.Context) (Response, error) {
	r.init()
	failedExpectations := make(map[string]IdentifyExpectationResult)

	identifyRequest := IdentifyRequest{r.BaseRequest}
	response, err := identifyRequest.process(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing identify request", true) {
		return &CheckIdentifyResponse{
			CheckResponse:  CheckResponse{r.mon.GetInfo()},
			IdentifyResult: nil,
		}, nil
	}

	identifyResponse := response.(*IdentifyResponse)

	if r.Expectations.Class != "" {
		r.mon.UpdateStatusIf(identifyResponse.Class != r.Expectations.Class, utility.IfThenElseInt(r.OsDiffWarning, monitoringplugin.WARNING, monitoringplugin.CRITICAL), fmt.Sprintf("OS: expected: \"%s\", got: \"%s\"", r.Expectations.Class, identifyResponse.Class))
	}
	if r.Expectations.Properties.Vendor != nil {
		var failed bool
		var empty bool
		var got string
		if identifyResponse.Properties.Vendor == nil {
			failed = true
			empty = true
			got = "no result"
		} else if *identifyResponse.Properties.Vendor != *r.Expectations.Properties.Vendor {
			failed = true
			got = *identifyResponse.Properties.Vendor
		}
		if r.mon.UpdateStatusIf(failed, utility.IfThenElseInt(r.VendorDiffWarning, monitoringplugin.WARNING, monitoringplugin.CRITICAL), fmt.Sprintf("Vendor: expected: \"%s\", got: %s", *r.Expectations.Properties.Vendor, utility.IfThenElseString(empty, got, "\""+got+"\""))) {
			failedExpectations["vendor"] = IdentifyExpectationResult{
				Expected: *r.Expectations.Properties.Vendor,
				Got:      got,
			}
		}
	}
	if r.Expectations.Properties.Model != nil {
		var failed bool
		var empty bool
		var got string
		if identifyResponse.Properties.Model == nil {
			failed = true
			empty = true
			got = "no result"
		} else if *identifyResponse.Properties.Model != *r.Expectations.Properties.Model {
			failed = true
			got = *identifyResponse.Properties.Model
		}
		if r.mon.UpdateStatusIf(failed, utility.IfThenElseInt(r.ModelDiffWarning, monitoringplugin.WARNING, monitoringplugin.CRITICAL), fmt.Sprintf("Model: expected: \"%s\", got: %s", *r.Expectations.Properties.Model, utility.IfThenElseString(empty, got, "\""+got+"\""))) {
			failedExpectations["model"] = IdentifyExpectationResult{
				Expected: *r.Expectations.Properties.Model,
				Got:      got,
			}
		}
	}
	if r.Expectations.Properties.ModelSeries != nil {
		var failed bool
		var empty bool
		var got string
		if identifyResponse.Properties.ModelSeries == nil {
			failed = true
			empty = true
			got = "no result"
		} else if *identifyResponse.Properties.ModelSeries != *r.Expectations.Properties.ModelSeries {
			failed = true
			got = *identifyResponse.Properties.ModelSeries
		}
		if r.mon.UpdateStatusIf(failed, utility.IfThenElseInt(r.ModelSeriesDiffWarning, monitoringplugin.WARNING, monitoringplugin.CRITICAL), fmt.Sprintf("ModelSeries: expected: \"%s\", got: %s", *r.Expectations.Properties.ModelSeries, utility.IfThenElseString(empty, got, "\""+got+"\""))) {
			failedExpectations["model_series"] = IdentifyExpectationResult{
				Expected: *r.Expectations.Properties.ModelSeries,
				Got:      got,
			}
		}
	}
	if r.Expectations.Properties.SerialNumber != nil {
		var failed bool
		var empty bool
		var got string
		if identifyResponse.Properties.SerialNumber == nil {
			failed = true
			empty = true
			got = "no result"
		} else if *identifyResponse.Properties.SerialNumber != *r.Expectations.Properties.SerialNumber {
			failed = true
			got = *identifyResponse.Properties.SerialNumber
		}
		if r.mon.UpdateStatusIf(failed, utility.IfThenElseInt(r.SerialNumberDiffWarning, monitoringplugin.WARNING, monitoringplugin.CRITICAL), fmt.Sprintf("SerialNumber: expected: \"%s\", got: %s", *r.Expectations.Properties.SerialNumber, utility.IfThenElseString(empty, got, "\""+got+"\""))) {
			failedExpectations["serial_number"] = IdentifyExpectationResult{
				Expected: *r.Expectations.Properties.SerialNumber,
				Got:      got,
			}
		}
	}
	if r.Expectations.Properties.OSVersion != nil {
		var failed bool
		var empty bool
		var got string
		if identifyResponse.Properties.OSVersion == nil {
			failed = true
			empty = true
			got = "no result"
		} else if *identifyResponse.Properties.OSVersion != *r.Expectations.Properties.OSVersion {
			failed = true
			got = *identifyResponse.Properties.OSVersion
		}
		if r.mon.UpdateStatusIf(failed, utility.IfThenElseInt(r.OsVersionDiffWarning, monitoringplugin.WARNING, monitoringplugin.CRITICAL), fmt.Sprintf("OSVersion: expected: \"%s\", got: %s", *r.Expectations.Properties.OSVersion, utility.IfThenElseString(empty, got, "\""+got+"\""))) {
			failedExpectations["version"] = IdentifyExpectationResult{
				Expected: *r.Expectations.Properties.OSVersion,
				Got:      got,
			}
		}
	}

	return &CheckIdentifyResponse{
		CheckResponse:      CheckResponse{r.mon.GetInfo()},
		IdentifyResult:     &identifyResponse.Device,
		FailedExpectations: failedExpectations,
	}, nil
}

func (r *CheckIdentifyRequest) validate(ctx context.Context) error {
	err := r.BaseRequest.validate(ctx)
	if err != nil {
		return errors.Wrap(err, "base request is not valid")
	}
	if r.Expectations == (device.Device{}) {
		return errors.New("no expectations given")
	}
	return nil
}
