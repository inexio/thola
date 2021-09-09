package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/inexio/thola/internal/utility"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkInterfaceMetricsCMD)
	checkCMD.AddCommand(checkInterfaceMetricsCMD)

	checkInterfaceMetricsCMD.Flags().Bool("print-interfaces", false, "Print interfaces to plugin output")
	checkInterfaceMetricsCMD.Flags().Bool("print-interfaces-csv", false, "Print interfaces to plugin output as CSV")
	checkInterfaceMetricsCMD.Flags().Bool("snmp-gets-instead-of-walk", false, "Use SNMP Gets instead of Walks")
	checkInterfaceMetricsCMD.Flags().String("ifDescr-regex", "", "Apply a regex on the ifDescr of the interfaces. Use it together with the 'ifDescr-regex-replace' flag")
	checkInterfaceMetricsCMD.Flags().String("ifDescr-regex-replace", "", "Apply a regex on the ifDescr of the interfaces. Use it together with the 'ifDescr-regex' flag")
	checkInterfaceMetricsCMD.Flags().StringSlice("ifType-filter", []string{}, "Filter out interfaces which ifType equals the given types")
	checkInterfaceMetricsCMD.Flags().StringSlice("ifName-filter", []string{}, "Filter out interfaces which ifName matches the given regex")
	checkInterfaceMetricsCMD.Flags().StringSlice("ifDescr-filter", []string{}, "Filter out interfaces which ifDescription matches the given regex")
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
		printInterfacesCSV, err := cmd.Flags().GetBool("print-interfaces-csv")
		if err != nil {
			log.Fatal().Err(err).Msg("print-interfaces-csv needs to be a boolean")
		}
		snmpGetsInsteadOfWalk, err := cmd.Flags().GetBool("snmp-gets-instead-of-walk")
		if err != nil {
			log.Fatal().Err(err).Msg("snmp-gets-instead-of-walk needs to be a boolean")
		}
		ifDescrRegex, err := cmd.Flags().GetString("ifDescr-regex")
		if err != nil {
			log.Fatal().Err(err).Msg("ifDescr-regex needs to be a string")
		}
		ifDescrRegexReplace, err := cmd.Flags().GetString("ifDescr-regex-replace")
		if err != nil {
			log.Fatal().Err(err).Msg("ifDescr-regex-replace needs to be a string")
		}
		ifTypeFilter, err := cmd.Flags().GetStringSlice("ifType-filter")
		if err != nil {
			log.Fatal().Err(err).Msg("ifType-filter needs to be a string")
		}
		ifNameFilter, err := cmd.Flags().GetStringSlice("ifName-filter")
		if err != nil {
			log.Fatal().Err(err).Msg("ifName-filter needs to be a string")
		}
		ifDescrFilter, err := cmd.Flags().GetStringSlice("ifDescr-filter")
		if err != nil {
			log.Fatal().Err(err).Msg("ifDescr-filter needs to be a string")
		}

		var nullString *string
		r := request.CheckInterfaceMetricsRequest{
			CheckDeviceRequest:    getCheckDeviceRequest(args[0]),
			PrintInterfaces:       printInterfaces,
			PrintInterfacesCSV:    printInterfacesCSV,
			IfDescrRegex:          utility.IfThenElse(cmd.Flags().Changed("ifDescr-regex"), &ifDescrRegex, nullString).(*string),
			IfDescrRegexReplace:   utility.IfThenElse(cmd.Flags().Changed("ifDescr-regex-replace"), &ifDescrRegexReplace, nullString).(*string),
			IfTypeFilter:          ifTypeFilter,
			IfNameFilter:          ifNameFilter,
			IfDescrFilter:         ifDescrFilter,
			SNMPGetsInsteadOfWalk: snmpGetsInsteadOfWalk,
		}

		handleRequest(&r)
	},
}
