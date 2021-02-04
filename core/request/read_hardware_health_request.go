package request

import "github.com/inexio/thola/core/device"

// ReadHardwareHealthRequest
//
// ReadHardwareHealthRequest is a the request struct for the read hardware health request.
//
// swagger:model
type ReadHardwareHealthRequest struct {
	ReadRequest
}

// ReadHardwareHealthResponse
//
// ReadHardwareHealthResponse is a the response struct for the read hardware health request.
//
// swagger:model
type ReadHardwareHealthResponse struct {
	device.HardwareHealthComponent
	ReadResponse
}
