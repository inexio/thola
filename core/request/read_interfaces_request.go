package request

import (
	"thola/core/device"
)

// ReadInterfacesRequest
//
// ReadInterfacesRequest is a the request struct for the read interfaces request.
//
// swagger:model
type ReadInterfacesRequest struct {
	ReadRequest
}

// ReadInterfacesResponse
//
// ReadInterfacesResponse is a the request struct for the read interfaces response.
//
// swagger:model
type ReadInterfacesResponse struct {
	Interfaces []device.Interface `yaml:"interfaces" json:"interfaces" xml:"interfaces"`
	ReadResponse
}
