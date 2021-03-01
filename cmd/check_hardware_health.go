package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkHardwareHealthCMD)
	checkCMD.AddCommand(checkHardwareHealthCMD)
}

var checkHardwareHealthCMD = &cobra.Command{
	Use:   "hardware-health [host]",
	Short: "Check hardware-health of a device.",
	Long:  "Check hardware-health of a device and return various performance data.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckHardwareHealthRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
		}
		handleRequest(&r)
	},
}
