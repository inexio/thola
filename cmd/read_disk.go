package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(readDiskCMD)
	readCMD.AddCommand(readDiskCMD)
}

var readDiskCMD = &cobra.Command{
	Use:   "disk",
	Short: "Read out storage information of a device",
	Long:  "Read out storage information of a device like types and used space of storages.",
	Run: func(cmd *cobra.Command, args []string) {
		request := request.ReadDiskRequest{
			ReadRequest: getReadRequest(args[0]),
		}
		handleRequest(&request)
	},
}
