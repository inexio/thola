package request

// CheckInterfaceMetricsRequest
//
// CheckInterfaceRequest is a the request struct for the check interface metrics request.
//
// swagger:model
type CheckInterfaceMetricsRequest struct {
	PrintInterfaces bool     `yaml:"print_interfaces" json:"print_interfaces" xml:"print_interfaces"`
	Filter          []string `yaml:"filter" json:"filter" xml:"filter"`
	CheckDeviceRequest
}
