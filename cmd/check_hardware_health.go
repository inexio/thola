package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkHardwareHealthCMD)
	checkCMD.AddCommand(checkHardwareHealthCMD)
}

var checkHardwareHealthCMD = &cobra.Command{
	Use:   "hardware-health",
	Short: "Check hardware-health of a device.",
	Long:  "Check hardware-health of a device and return various performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckHardwareHealthRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
		}
		handleRequest(&r)
	},
}
