package request

import (
	"github.com/inexio/thola/internal/device"
)

// ReadInterfacesRequest
//
// ReadInterfacesRequest is the request struct for the read interfaces request.
//
// swagger:model
type ReadInterfacesRequest struct {
	// If you only want specific values of the interfaces you can specify them here.
	Values []string `yaml:"values" json:"values" xml:"values"`
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
