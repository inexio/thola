package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(readCountInterfacesCMD)
	readCMD.AddCommand(readCountInterfacesCMD)
}

var readCountInterfacesCMD = &cobra.Command{
	Use:   "count-interfaces",
	Short: "Count interfaces of a device",
	Long:  "Count the interfaces of a device.",
	Run: func(cmd *cobra.Command, args []string) {
		request := request.ReadCountInterfacesRequest{
			ReadRequest: getReadRequest(args[0]),
		}
		handleRequest(&request)
	},
}
