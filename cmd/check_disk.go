package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkDiskCMD)
	checkCMD.AddCommand(checkDiskCMD)
}

var checkDiskCMD = &cobra.Command{
	Use:   "disk",
	Short: "Check the disk of a device",
	Long: "Checks the disk of a device.\n\n" +
		"The metrics will be printed as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckDiskRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
		}
		handleRequest(&r)
	},
}
