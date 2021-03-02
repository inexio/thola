package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/inexio/thola/core/value"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkMemoryUsage)
	checkCMD.AddCommand(checkMemoryUsage)

	checkMemoryUsage.Flags().String("warning", "", "warning threshold for memory usage")
	checkMemoryUsage.Flags().String("critical", "", "critical threshold for system voltage")
}

var checkMemoryUsage = &cobra.Command{
	Use:   "memory-usage",
	Short: "Check the memory usage of a device",
	Long: "Checks the memory usage of a device.\n\n" +
		"The usage will be printed as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckMemoryUsageRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
			MemoryUsageThresholds: request.CheckThresholds{
				WarningMax:  value.New(cmd.Flags().Lookup("warning").Value),
				CriticalMax: value.New(cmd.Flags().Lookup("critical").Value),
			},
		}
		handleRequest(&r)
	},
}
