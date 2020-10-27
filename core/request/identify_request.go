package request

import (
	"thola/core/device"
)

// IdentifyRequest
//
// IdentifyRequest is a the request struct for the identify request.
//
// swagger:model
type IdentifyRequest struct {
	BaseRequest
}

// IdentifyResponse
//
// IdentifyResponse is the response struct that is for identify requests.
//
// swagger:model
type IdentifyResponse struct {
	device.Device `yaml:",inline"`
	BaseResponse  `yaml:",inline"`
}
