package request

import "github.com/inexio/thola/internal/device"

// ReadServerRequest
//
// ReadServerRequest is a the request struct for the read server request.
//
// swagger:model
type ReadServerRequest struct {
	ReadRequest
}

// ReadServerResponse
//
// ReadServerResponse is a the response struct for the read server response.
//
// swagger:model
type ReadServerResponse struct {
	Server device.ServerComponent `yaml:"server" json:"server" xml:"server"`
	ReadResponse
}
