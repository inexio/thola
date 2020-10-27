package request

// CheckMetricsRequest
//
// CheckRequest is a the request struct for the check metrics request.
//
// swagger:model
type CheckMetricsRequest struct {
	InterfaceFilter []string `yaml:"interface_filter" json:"interface_filter" xml:"interface_filter"`
	CheckDeviceRequest
}
