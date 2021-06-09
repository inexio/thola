package request

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
	MemoryUsage float64 `yaml:"memory_usage" json:"memory_usage" xml:"memory_usage"`
	ReadResponse
}
