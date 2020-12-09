package request

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
	CPULoad []float64 `yaml:"cpu_load" json:"cpu_load" xml:"cpu_load"`
	ReadResponse
}
