package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(readInterfacesCMD)
	readCMD.AddCommand(readInterfacesCMD)

	readInterfacesCMD.Flags().StringSlice("value", []string{}, "Only read out these values of the interfaces (e.g. 'ifDescr')")
}

var readInterfacesCMD = &cobra.Command{
	Use:   "interfaces",
	Short: "Read out interface information of a device",
	Long: "Read out interface information of a device.\n\n" +
		"Also reads special values based on the interface type.",
	Run: func(cmd *cobra.Command, args []string) {
		values, err := cmd.Flags().GetStringSlice("value")
		if err != nil {
			log.Fatal().Err(err).Msg("value needs to be a string")
		}

		request := request.ReadInterfacesRequest{
			Values:      values,
			ReadRequest: getReadRequest(args[0]),
		}
		handleRequest(&request)
	},
}
