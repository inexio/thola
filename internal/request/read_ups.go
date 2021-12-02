package request

import "github.com/inexio/thola/internal/device"

// ReadUPSRequest
//
// ReadUPSRequest is the request struct for the read ups request.
//
// swagger:model
type ReadUPSRequest struct {
	ReadRequest
}

// ReadUPSResponse
//
// ReadUPSResponse is the response struct for the read ups response.
//
// swagger:model
type ReadUPSResponse struct {
	UPS device.UPSComponent `yaml:"ups" json:"ups" xml:"ups"`
	ReadResponse
}
