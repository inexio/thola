package cmd

import (
	"github.com/spf13/cobra"
	"thola/core/request"
)

func init() {
	addDeviceFlags(checkSNMPCMD)
	checkCMD.AddCommand(checkSNMPCMD)
}

var checkSNMPCMD = &cobra.Command{
	Use:   "snmp",
	Short: "Check whether a device is reachable over snmp",
	Long: "Check whether a device is reachable over snmp.\n\n" +
		"All configured SNMP options will be tried.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckSNMPRequest{
			CheckDeviceRequest: getCheckDeviceRequest(),
		}
		handleRequest(&r)
	},
}
