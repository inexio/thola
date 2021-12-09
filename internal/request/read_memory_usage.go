package request

import "github.com/inexio/thola/internal/device"

// ReadMemoryUsageRequest
//
// ReadMemoryUsageRequest is the request struct for the read memory usage request.
//
// swagger:model
type ReadMemoryUsageRequest struct {
	ReadRequest
}

// ReadMemoryUsageResponse
//
// ReadMemoryUsageResponse is the response struct for the read memory usage request.
//
// swagger:model
type ReadMemoryUsageResponse struct {
	MemoryPools []device.MemoryPool `yaml:"memory_pools" json:"memory_pools" xml:"memory_pools"`
	ReadResponse
}
