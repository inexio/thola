package cmd

import (
	"github.com/spf13/cobra"
	"thola/core/request"
)

func init() {
	addDeviceFlags(checkUPSCMD)
	checkCMD.AddCommand(checkUPSCMD)
}

var checkUPSCMD = &cobra.Command{
	Use:   "ups",
	Short: "Checks whether a UPS device has its main voltage applied",
	Long: "Checks whether a UPS device has its main voltage applied.\n\n" +
		"All UPS statistics will be printed as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckUPSRequest{
			CheckDeviceRequest: getCheckDeviceRequest(),
		}
		handleRequest(&r)
	},
}
