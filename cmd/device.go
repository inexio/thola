// +build !client

package cmd

import (
	flag "github.com/spf13/pflag"
)

func addBinarySpecificDeviceFlags(fs *flag.FlagSet) {
	fs.StringSlice("snmp-community", defaultSNMPCommunity, "Community strings for SNMP to use")
	fs.StringSlice("snmp-version", defaultSNMPVersion, "SNMP versions to use (1, 2c or 3)")
	fs.IntSlice("snmp-port", defaultSNMPPort, "Ports for SNMP to use")
}
