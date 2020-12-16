package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/inexio/thola/core/value"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkCpuLoad)
	checkCMD.AddCommand(checkCpuLoad)

	checkCpuLoad.Flags().String("warning", "", "warning threshold for cpu load")
	checkCpuLoad.Flags().String("critical", "", "critical threshold for cpu load")
}

var checkCpuLoad = &cobra.Command{
	Use:   "cpu-load",
	Short: "Check the cpu load of a device",
	Long: "Checks the cpu load of a device.\n\n" +
		"The usage will be printed as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckCPULoadRequest{
			CheckDeviceRequest: getCheckDeviceRequest(),
			CPULoadThresholds: request.CheckThresholds{
				WarningMax:  value.New(cmd.Flags().Lookup("warning").Value),
				CriticalMax: value.New(cmd.Flags().Lookup("critical").Value),
			},
		}
		handleRequest(&r)
	},
}
