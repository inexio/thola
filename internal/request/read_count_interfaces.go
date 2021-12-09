package request

// ReadCountInterfacesRequest
//
// ReadCountInterfacesRequest is the request struct for the read count-interfaces request.
//
// swagger:model
type ReadCountInterfacesRequest struct {
	ReadRequest
}

// ReadCountInterfacesResponse
//
// ReadCountInterfacesResponse is the response struct for the read count-interfaces response.
//
// swagger:model
type ReadCountInterfacesResponse struct {
	Count int `yaml:"count" json:"count" xml:"count"`
	ReadResponse
}
