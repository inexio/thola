package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkMemoryUsage)
	checkCMD.AddCommand(checkMemoryUsage)

	checkMemoryUsage.Flags().Float64("warning", 0, "warning threshold for memory usage")
	checkMemoryUsage.Flags().Float64("critical", 0, "critical threshold for memory usage")
}

var checkMemoryUsage = &cobra.Command{
	Use:   "memory-usage",
	Short: "Check the memory usage of a device",
	Long: "Checks the memory usage of a device.\n\n" +
		"The usage will be printed as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckMemoryUsageRequest{
			CheckDeviceRequest:    getCheckDeviceRequest(args[0]),
			MemoryUsageThresholds: generateCheckThresholds(cmd, "", "warning", "", "critical", true),
		}
		handleRequest(&r)
	},
}
