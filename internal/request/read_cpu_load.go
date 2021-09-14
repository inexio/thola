package request

import "github.com/inexio/thola/internal/device"

// ReadCPULoadRequest
//
// ReadCPULoadRequest is a the request struct for the read cpu request.
//
// swagger:model
type ReadCPULoadRequest struct {
	ReadRequest
}

// ReadCPULoadResponse
//
// ReadCPULoadResponse is a the response struct for the read cpu response.
//
// swagger:model
type ReadCPULoadResponse struct {
	CPUs []device.CPU `yaml:"cpus" json:"cpus" xml:"cpus"`
	ReadResponse
}
