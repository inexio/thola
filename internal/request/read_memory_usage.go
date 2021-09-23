package request

import "github.com/inexio/thola/internal/device"

// ReadMemoryUsageRequest
//
// ReadMemoryUsageRequest is a the request struct for the read memory usage request.
//
// swagger:model
type ReadMemoryUsageRequest struct {
	ReadRequest
}

// ReadMemoryUsageResponse
//
// ReadMemoryUsageResponse is a the response struct for the read memory usage request.
//
// swagger:model
type ReadMemoryUsageResponse struct {
	MemoryPools []device.MemoryPool `yaml:"memory_pools" json:"memory_pools" xml:"memory_pools"`
	ReadResponse
}
