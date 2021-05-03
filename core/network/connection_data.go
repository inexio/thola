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
	// The snmp community string(s) for the device.
	//
	// example: public
	Communities []string `json:"communities" xml:"communities" yaml:"communities"`
	// The snmp version(s) of the device.
	//
	// example: 2c
	Versions []string `json:"versions" xml:"versions" yaml:"versions"`
	// The snmp port(s) of the device.
	//
	// example: 161
	Ports []int `json:"ports" xml:"ports" yaml:"ports"`
	// The amount of parallel connection requests used while trying to get a valid SNMP connection.
	//
	// example: 5
	DiscoverParallelRequests *int `json:"discoverParallelRequests" xml:"discoverParallelRequests" yaml:"discoverParallelRequests"`
	// The timeout in seconds used while trying to get a valid SNMP connection.
	//
	// example: 2
	DiscoverTimeout *int `json:"discoverTimeout" xml:"discoverTimeout" yaml:"discoverTimeout"`
	// The retries used while trying to get a valid SNMP connection.
	//
	// example: 0
	DiscoverRetries *int `json:"discoverRetries" xml:"discoverRetries" yaml:"discoverRetries"`
	// The data required for an SNMP v3 connection.
	V3Data SNMPv3ConnectionData `json:"v3_data" xml:"v3_data" yaml:"v3_data"`
}

// SNMPv3ConnectionData
//
// SNMPv3ConnectionData includes all snmp v3 specific connection data.
//
// swagger:model
type SNMPv3ConnectionData struct {
	// The security level of the SNMP connection.
	//
	// example: authPriv
	Level *string `json:"level" xml:"level" yaml:"level"`
	// The context name of the SNMP connection.
	//
	// example: bridge1
	ContextName *string `json:"context_name" xml:"context_name" yaml:"context_name"`
	// The user of the SNMP connection.
	//
	// example: user
	User *string `json:"user" xml:"user" yaml:"user"`
	// The authentication protocol passphrase of the SNMP connection.
	//
	// example: passphrase
	AuthKey *string `json:"auth_key" xml:"auth_key" yaml:"auth_key"`
	// The authentication protocol of the SNMP connection.
	//
	// example: MD5
	AuthProtocol *string `json:"auth_protocol" xml:"auth_protocol" yaml:"auth_protocol"`
	// The privacy protocol passphrase of the SNMP connection.
	//
	// example: passphrase
	PrivKey *string `json:"priv_key" xml:"priv_key" yaml:"priv_key"`
	// The privacy protocol of the SNMP connection.
	//
	// example: DES
	PrivProtocol *string `json:"priv_protocol" xml:"priv_protocol" yaml:"priv_protocol"`
}

// SNMPCredentials includes all credential information of the snmp connection.
// V3 values are nil if no snmp v3 is being used.
type SNMPCredentials struct {
	Version       string `yaml:"version" json:"version" xml:"version"`
	Community     string `yaml:"community" json:"community" xml:"community"`
	Port          int    `yaml:"port" json:"port" xml:"port"`
	V3Level       string `yaml:"v3Level" json:"v3Level" xml:"v3Level"`
	V3ContextName string `yaml:"v3ContextName" json:"v3ContextName" xml:"v3ContextName"`
}

// HTTPConnectionData
//
// HTTPConnectionData includes all http connection data for a device.
//
// swagger:model
type HTTPConnectionData struct {
	// The http port(s) of the device.
	//
	// example: 80
	HTTPPorts []int `json:"http_ports" xml:"http_ports" yaml:"http_ports"`
	// The https port(s) of the device.
	//
	// example: 443
	HTTPSPorts []int `json:"https_ports" xml:"https_ports" yaml:"https_ports"`
	// The username for authorization on the device.
	//
	// example: username
	AuthUsername *string `json:"auth_username" xml:"auth_username" yaml:"auth_username"`
	// The password for authorization on the device.
	//
	// example: password
	AuthPassword *string `json:"auth_password" xml:"auth_password" yaml:"auth_password"`
}
