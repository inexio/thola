package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"thola/core/request"
)

func init() {
	addDeviceFlags(checkMetricsCMD)
	checkCMD.AddCommand(checkMetricsCMD)

	checkMetricsCMD.Flags().StringSlice("interface-filter", []string{}, "Filter out interfaces which ifType matches the given types")

	err := viper.BindPFlag("checkMetrics.interface-filter", checkMetricsCMD.Flags().Lookup("interface-filter"))
	if err != nil {
		os.Exit(3)
	}
}

var checkMetricsCMD = &cobra.Command{
	Use:   "metrics",
	Short: "Reads all possible metrics for the device and prints them as performance data",
	Long:  "Reads all possible metrics for the device and prints them as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckMetricsRequest{
			CheckDeviceRequest: getCheckDeviceRequest(),
			InterfaceFilter:    viper.GetStringSlice("checkMetrics.interface-filter"),
		}
		handleRequest(&r)
	},
}
