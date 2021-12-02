package request

import "github.com/inexio/thola/internal/device"

// ReadSBCRequest
//
// ReadSBCRequest is the request struct for the read sbc request.
//
// swagger:model
type ReadSBCRequest struct {
	ReadRequest
}

// ReadSBCResponse
//
// ReadSBCResponse is the response struct for the read sbc response.
//
// swagger:model
type ReadSBCResponse struct {
	SBC device.SBCComponent `yaml:"sbc" json:"sbc" xml:"sbc"`
	ReadResponse
}
