package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var interfaceOptionsFlagSet = buildInterfaceOptionsFlagSet()

func buildInterfaceOptionsFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("interface_options", flag.ContinueOnError)

	fs.StringSlice("value", []string{}, "If set only the specified values will be read from the interfaces (e.g. 'ifDescr')")
	fs.Bool("snmp-gets-instead-of-walk", false, "Use SNMP Gets instead of Walks")
	fs.String("ifDescr-regex", "", "Apply a regex on the ifDescr of the interfaces. Use it together with the 'ifDescr-regex-replace' flag")
	fs.String("ifDescr-regex-replace", "", "Apply a regex on the ifDescr of the interfaces. Use it together with the 'ifDescr-regex' flag")
	fs.StringSlice("ifType-filter", []string{}, "Filter out interfaces which ifType equals the given types")
	fs.StringSlice("ifName-filter", []string{}, "Filter out interfaces which ifName matches the given regex")
	fs.StringSlice("ifDescr-filter", []string{}, "Filter out interfaces which ifDescription matches the given regex")

	return fs
}

func addInterfaceOptionsFlags(cmd *cobra.Command) {
	cmd.Flags().AddFlagSet(interfaceOptionsFlagSet)
}

func getInterfaceOptions() request.InterfaceOptions {
	values, err := interfaceOptionsFlagSet.GetStringSlice("value")
	if err != nil {
		log.Fatal().Err(err).Msg("value needs to be a string")
	}
	snmpGetsInsteadOfWalk, err := interfaceOptionsFlagSet.GetBool("snmp-gets-instead-of-walk")
	if err != nil {
		log.Fatal().Err(err).Msg("snmp-gets-instead-of-walk needs to be a boolean")
	}
	ifDescrRegex, err := interfaceOptionsFlagSet.GetString("ifDescr-regex")
	if err != nil {
		log.Fatal().Err(err).Msg("ifDescr-regex needs to be a string")
	}
	ifDescrRegexReplace, err := interfaceOptionsFlagSet.GetString("ifDescr-regex-replace")
	if err != nil {
		log.Fatal().Err(err).Msg("ifDescr-regex-replace needs to be a string")
	}
	ifTypeFilter, err := interfaceOptionsFlagSet.GetStringSlice("ifType-filter")
	if err != nil {
		log.Fatal().Err(err).Msg("ifType-filter needs to be a string")
	}
	ifNameFilter, err := interfaceOptionsFlagSet.GetStringSlice("ifName-filter")
	if err != nil {
		log.Fatal().Err(err).Msg("ifName-filter needs to be a string")
	}
	ifDescrFilter, err := interfaceOptionsFlagSet.GetStringSlice("ifDescr-filter")
	if err != nil {
		log.Fatal().Err(err).Msg("ifDescr-filter needs to be a string")
	}

	return request.InterfaceOptions{
		Values:                values,
		IfDescrRegex:          ifDescrRegex,
		IfDescrRegexReplace:   ifDescrRegexReplace,
		IfTypeFilter:          ifTypeFilter,
		IfNameFilter:          ifNameFilter,
		IfDescrFilter:         ifDescrFilter,
		SNMPGetsInsteadOfWalk: snmpGetsInsteadOfWalk,
	}
}
