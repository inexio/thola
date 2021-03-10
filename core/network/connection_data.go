package network

// ConnectionData
//
// ConnectionData includes all connection data for a device.
//
// swagger:model
type ConnectionData struct {
	// Data of the snmp connection to the device
	SNMP *SNMPConnectionData `json:"snmp" xml:"snmp" yaml:"snmp"`
	// Data of the http connection to the device
	HTTP *HTTPConnectionData `json:"http" xml:"http" yaml:"http"`
}

// SNMPConnectionData
//
// SNMPConnectionData includes all snmp connection data for a device.
//
// swagger:model
type SNMPConnectionData struct {
	// The snmp community string(s) for the device
	//
	// example: public
	Communities []string `json:"communities" xml:"communities" yaml:"communities"`
	// The snmp version(s) of the device
	//
	// example: 2c
	Versions []string `json:"versions" xml:"versions" yaml:"versions"`
	// The snmp port(s) of the device
	//
	// example: 161
	Ports []int `json:"ports" xml:"ports" yaml:"ports"`
	// The amount of parallel connection requests used while trying to get a valid SNMP connection
	//
	// example: 5
	DiscoverParallelRequests *int `json:"discoverParallelRequests" xml:"discoverParallelRequests" yaml:"discoverParallelRequests"`
	// The timeout in seconds used while trying to get a valid SNMP connection
	//
	// example: 2
	DiscoverTimeout *int `json:"discoverTimeout" xml:"discoverTimeout" yaml:"discoverTimeout"`
	// The retries used while trying to get a valid SNMP connection
	//
	// example: 0
	DiscoverRetries *int `json:"discoverRetries" xml:"discoverRetries" yaml:"discoverRetries"`
}

// SNMPCredentials includes all credential information of the snmp connection.
type SNMPCredentials struct {
	Version   string `yaml:"version" json:"version" xml:"version"`
	Community string `yaml:"community" json:"community" xml:"community"`
	Port      int    `yaml:"port" json:"port" xml:"port"`
}

// HTTPConnectionData
//
// HTTPConnectionData includes all http connection data for a device.
//
// swagger:model
type HTTPConnectionData struct {
	// The http port(s) of the device
	//
	// example: 80
	HTTPPorts []int `json:"http_ports" xml:"http_ports" yaml:"http_ports"`
	// The https port(s) of the device
	//
	// example: 443
	HTTPSPorts []int `json:"https_ports" xml:"https_ports" yaml:"https_ports"`
	// The username for authorization on the device
	//
	// example: username
	AuthUsername *string `json:"auth_username" xml:"auth_username" yaml:"auth_username"`
	// The password for authorization on the device
	//
	// example: password
	AuthPassword *string `json:"auth_password" xml:"auth_password" yaml:"auth_password"`
}
