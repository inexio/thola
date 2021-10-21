package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(readInterfacesCMD)
	addInterfaceOptionsFlags(readInterfacesCMD)
	readCMD.AddCommand(readInterfacesCMD)
}

var readInterfacesCMD = &cobra.Command{
	Use:   "interfaces",
	Short: "Read out interface information of a device",
	Long: "Read out interface information of a device.\n\n" +
		"Also reads special values based on the interface type.",
	Run: func(cmd *cobra.Command, args []string) {
		request := request.ReadInterfacesRequest{
			InterfaceOptions: getInterfaceOptions(),
			ReadRequest:      getReadRequest(args[0]),
		}
		handleRequest(&request)
	},
}
