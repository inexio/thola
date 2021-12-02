package request

import (
	"github.com/inexio/thola/internal/device"
)

// IdentifyRequest
//
// IdentifyRequest is the request struct for the identify request.
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
