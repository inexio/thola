package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(readServerCMD)
	readCMD.AddCommand(readServerCMD)
}

var readServerCMD = &cobra.Command{
	Use:   "server",
	Short: "Read out server specific information of a device",
	Long:  "Read out server specific information of a device like disk usage or process count.",
	Run: func(cmd *cobra.Command, args []string) {
		request := request.ReadServerRequest{
			ReadRequest: getReadRequest(args[0]),
		}
		handleRequest(&request)
	},
}
