package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkSIEMCMD)
	checkCMD.AddCommand(checkSIEMCMD)
}

var checkSIEMCMD = &cobra.Command{
	Use:   "siem",
	Short: "Check the siem specific metrics of a device",
	Long: "Checks the siem specific metrics of a device.\n\n" +
		"The usage will be printed as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckSIEMRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
		}
		handleRequest(&r)
	},
}
