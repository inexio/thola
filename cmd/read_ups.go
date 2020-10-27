package cmd

import (
	"github.com/spf13/cobra"
	"thola/core/request"
)

func init() {
	readCMD.AddCommand(readUPSCMD)
}

var readUPSCMD = &cobra.Command{
	Use:   "ups",
	Short: "Read out UPS information of a device",
	Long:  "Read out UPS information of a device like battery capacity and current usage.",
	Run: func(cmd *cobra.Command, args []string) {
		request := request.ReadUPSRequest{
			ReadRequest: getReadRequest(),
		}
		handleRequest(&request)
	},
}
