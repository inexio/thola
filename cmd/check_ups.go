package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
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
