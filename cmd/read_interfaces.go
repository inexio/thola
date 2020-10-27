package cmd

import (
	"github.com/spf13/cobra"
	"thola/core/request"
)

func init() {
	readCMD.AddCommand(readInterfacesCMD)
}

var readInterfacesCMD = &cobra.Command{
	Use:   "interfaces",
	Short: "Read out interface information of a device",
	Long: "Read out interface information of a device.\n\n" +
		"Also reads special values based on the interface type.",
	Run: func(cmd *cobra.Command, args []string) {
		request := request.ReadInterfacesRequest{
			ReadRequest: getReadRequest(),
		}
		handleRequest(&request)
	},
}
