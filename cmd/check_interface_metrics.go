package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkInterfaceMetricsCMD)
	addInterfaceOptionsFlags(checkInterfaceMetricsCMD)
	checkCMD.AddCommand(checkInterfaceMetricsCMD)

	checkInterfaceMetricsCMD.Flags().Bool("print-interfaces", false, "Print interfaces to plugin output")
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

		r := request.CheckInterfaceMetricsRequest{
			PrintInterfaces:    printInterfaces,
			InterfaceOptions:   getInterfaceOptions(),
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
		}

		handleRequest(&r)
	},
}
