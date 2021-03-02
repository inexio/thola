package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

func init() {
	addDeviceFlags(checkInterfaceMetricsCMD)
	checkCMD.AddCommand(checkInterfaceMetricsCMD)

	checkInterfaceMetricsCMD.Flags().Bool("print-interfaces", false, "Print interfaces to plugin output")
	checkInterfaceMetricsCMD.Flags().StringSlice("filter", []string{}, "Filter out interfaces which ifType matches the given types")

	err := viper.BindPFlag("checkInterfaceMetrics.print-interfaces", checkInterfaceMetricsCMD.Flags().Lookup("print-interfaces"))
	if err != nil {
		os.Exit(3)
	}

	err = viper.BindPFlag("checkInterfaceMetrics.filter", checkInterfaceMetricsCMD.Flags().Lookup("filter"))
	if err != nil {
		os.Exit(3)
	}
}

var checkInterfaceMetricsCMD = &cobra.Command{
	Use:   "interface-metrics",
	Short: "Reads all interface metrics and prints them as performance data",
	Long:  "Reads all interface metrics and prints them as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckInterfaceMetricsRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
			PrintInterfaces:    viper.GetBool("checkInterfaceMetrics.print-interfaces"),
			Filter:             viper.GetStringSlice("checkInterfaceMetrics.filter"),
		}
		handleRequest(&r)
	},
}
