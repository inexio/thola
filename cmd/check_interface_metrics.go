package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkInterfaceMetricsCMD)
	checkCMD.AddCommand(checkInterfaceMetricsCMD)

	checkInterfaceMetricsCMD.Flags().Bool("print-interfaces", false, "Print interfaces to plugin output")
	checkInterfaceMetricsCMD.Flags().StringSlice("ifType-filter", []string{}, "Filter out interfaces which ifType equals the given types")
	checkInterfaceMetricsCMD.Flags().StringSlice("ifName-filter", []string{}, "Filter out interfaces which ifType matches the given regex")
}

var checkInterfaceMetricsCMD = &cobra.Command{
	Use:   "interface-metrics",
	Short: "Reads all interface metrics and prints them as performance data",
	Long:  "Reads all interface metrics and prints them as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		printInterfaces, err := cmd.Flags().GetBool("print-interfaces")
		if err != nil {
			log.Fatal().Err(err).Msg("print-interfaces needs to be a boolean")
		}
		ifTypeFilter, err := cmd.Flags().GetStringSlice("ifType-filter")
		if err != nil {
			log.Fatal().Err(err).Msg("ifType-filter needs to be a string")
		}
		ifNameFilter, err := cmd.Flags().GetStringSlice("ifName-filter")
		if err != nil {
			log.Fatal().Err(err).Msg("ifName-filter needs to be a string")
		}
		r := request.CheckInterfaceMetricsRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
			PrintInterfaces:    printInterfaces,
			IfTypeFilter:       ifTypeFilter,
			IfNameFilter:       ifNameFilter,
		}
		handleRequest(&r)
	},
}
