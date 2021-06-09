package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkCpuLoad)
	checkCMD.AddCommand(checkCpuLoad)

	checkCpuLoad.Flags().Float64("warning", 0, "warning threshold for cpu load")
	checkCpuLoad.Flags().Float64("critical", 0, "critical threshold for cpu load")
}

var checkCpuLoad = &cobra.Command{
	Use:   "cpu-load",
	Short: "Check the cpu load of a device",
	Long: "Checks the cpu load of a device.\n\n" +
		"The usage will be printed as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckCPULoadRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
			CPULoadThresholds:  generateCheckThresholds(cmd, "", "warning", "", "critical", true),
		}
		handleRequest(&r)
	},
}
