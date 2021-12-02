package request

import "github.com/inexio/thola/internal/device"

// ReadHardwareHealthRequest
//
// ReadHardwareHealthRequest is the request struct for the read hardware health request.
//
// swagger:model
type ReadHardwareHealthRequest struct {
	ReadRequest
}

// ReadHardwareHealthResponse
//
// ReadHardwareHealthResponse is the response struct for the read hardware health request.
//
// swagger:model
type ReadHardwareHealthResponse struct {
	HardwareHealth device.HardwareHealthComponent `yaml:"hardware_health" json:"hardware_health" xml:"hardware_health"`
	ReadResponse
}
