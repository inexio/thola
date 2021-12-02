package request

// ReadAvailableComponentsRequest
//
// ReadAvailableComponentsRequest is the request struct for the read available-components request.
//
// swagger:model
type ReadAvailableComponentsRequest struct {
	ReadRequest
}

// ReadAvailableComponentsResponse
//
// ReadAvailableComponentsResponse is the response struct for the read available-components response.
//
// swagger:model
type ReadAvailableComponentsResponse struct {
	AvailableComponents []string `yaml:"availableComponents" json:"availableComponents" xml:"availableComponents"`
	ReadResponse
}
