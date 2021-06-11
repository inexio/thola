package request

import (
	"github.com/inexio/thola/internal/device"
)

// CheckIdentifyRequest
//
// CheckIdentifyRequest is a the request struct for the check identify request.
//
// swagger:model
type CheckIdentifyRequest struct {
	CheckDeviceRequest
	Expectations device.Device `yaml:"expectations" json:"expectations" xml:"expectations"`

	OsDiffWarning           bool `yaml:"os_diff_warning" json:"os_diff_warning" xml:"os_diff_warning"`
	VendorDiffWarning       bool `yaml:"vendor_diff_warning" json:"vendor_diff_warning" xml:"vendor_diff_warning"`
	ModelDiffWarning        bool `yaml:"model_diff_warning" json:"model_diff_warning" xml:"model_diff_warning"`
	ModelSeriesDiffWarning  bool `yaml:"model_series_diff_warning" json:"model_series_diff_warning" xml:"model_series_diff_warning"`
	OsVersionDiffWarning    bool `yaml:"os_version_diff_warning" json:"os_version_diff_warning" xml:"os_version_diff_warning"`
	SerialNumberDiffWarning bool `yaml:"serial_number_diff_warning" json:"serial_number_diff_warning" xml:"serial_number_diff_warning"`
}

// CheckIdentifyResponse
//
// CheckIdentifyResponse is a response struct for the check identify request.
//
// swagger:model
type CheckIdentifyResponse struct {
	CheckResponse
	IdentifyResult     *device.Device                       `yaml:"identify_result" json:"identify_result" xml:"identify_result"`
	FailedExpectations map[string]IdentifyExpectationResult `yaml:"failed_expectations" json:"failed_expectations" xml:"failed_expectations"`
}

// IdentifyExpectationResult is a response struct for the check identify request.
type IdentifyExpectationResult struct {
	Expected string `yaml:"expected" json:"expected" xml:"expected"`
	Got      string `yaml:"got" json:"got" xml:"got"`
}
