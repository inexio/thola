package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkServerCMD)
	checkCMD.AddCommand(checkServerCMD)
}

var checkServerCMD = &cobra.Command{
	Use:   "server",
	Short: "Check the server specific metrics of a device",
	Long: "Checks the server specific metrics of a device.\n\n" +
		"The usage will be printed as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckServerRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
		}
		handleRequest(&r)
	},
}
