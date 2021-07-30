package request

// CheckInterfaceMetricsRequest
//
// CheckInterfaceRequest is a the request struct for the check interface metrics request.
//
// swagger:model
type CheckInterfaceMetricsRequest struct {
	PrintInterfaces bool     `yaml:"print_interfaces" json:"print_interfaces" xml:"print_interfaces"`
	IfTypeFilter    []string `yaml:"ifType_filter" json:"ifType_filter" xml:"ifType_filter"`
	IfNameFilter    []string `yaml:"ifName_filter" json:"ifName_filter" xml:"ifName_filter"`
	IfDescrFilter   []string `yaml:"ifDescr_filter" json:"ifDescr_filter" xml:"ifDescr_filter"`
	CheckDeviceRequest
}
