package cmd

import (
	"github.com/spf13/cobra"
	"thola/core/request"
)

func init() {
	readCMD.AddCommand(readCountInterfacesCMD)
}

var readCountInterfacesCMD = &cobra.Command{
	Use:   "count-interfaces",
	Short: "Count interfaces of a device",
	Long:  "Count the interfaces of a device.",
	Run: func(cmd *cobra.Command, args []string) {
		request := request.ReadCountInterfacesRequest{
			ReadRequest: getReadRequest(),
		}
		handleRequest(&request)
	},
}
