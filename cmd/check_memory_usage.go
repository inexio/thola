package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/inexio/thola/core/value"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkMemoryUsage)
	checkCMD.AddCommand(checkMemoryUsage)

	checkMemoryUsage.Flags().String("warning", "", "warning max threshold for memory usage")
	checkMemoryUsage.Flags().String("critical", "", "critical max threshold for system voltage")
}

var checkMemoryUsage = &cobra.Command{
	Use:   "memory-usage",
	Short: "Checks whether a UPS device has its main voltage applied",
	Long: "Checks whether a UPS device has its main voltage applied.\n\n" +
		"All UPS statistics will be printed as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckMemoryUsageRequest{
			CheckDeviceRequest: getCheckDeviceRequest(),
			MemoryUsageThresholds: request.CheckThresholds{
				WarningMax:  value.New(cmd.Flags().Lookup("warning").Value),
				CriticalMax: value.New(cmd.Flags().Lookup("critical").Value),
			},
		}
		handleRequest(&r)
	},
}
