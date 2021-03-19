package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkServerCMD)
	checkCMD.AddCommand(checkServerCMD)

	checkServerCMD.Flags().Float64("procs-warning", 0, "warning threshold for procs count")
	checkServerCMD.Flags().Float64("procs-critical", 0, "critical threshold for procs count")
	checkServerCMD.Flags().Float64("users-warning", 0, "warning threshold for users count")
	checkServerCMD.Flags().Float64("users-critical", 0, "critical threshold for users count")
}

var checkServerCMD = &cobra.Command{
	Use:   "server",
	Short: "Check the server specific metrics of a device",
	Long: "Checks the server specific metrics of a device.\n\n" +
		"The usage will be printed as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckServerRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
			UsersThreshold:     generateCheckThresholds(cmd, "", "users-warning", "", "users-critical", true),
			ProcsThreshold:     generateCheckThresholds(cmd, "", "procs-warning", "", "procs-critical", true),
		}
		handleRequest(&r)
	},
}
