package request

import "github.com/inexio/thola/internal/device"

// ReadSIEMRequest
//
// ReadSIEMRequest is the request struct for the read siem request.
//
// swagger:model
type ReadSIEMRequest struct {
	ReadRequest
}

// ReadSIEMResponse
//
// ReadSIEMResponse is the response struct for the read siem response.
//
// swagger:model
type ReadSIEMResponse struct {
	SIEM device.SIEMComponent `yaml:"siem" json:"siem" xml:"siem"`
	ReadResponse
}
