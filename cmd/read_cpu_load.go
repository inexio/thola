package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(readCPULoadCMD)
	readCMD.AddCommand(readCPULoadCMD)
}

var readCPULoadCMD = &cobra.Command{
	Use:   "cpu-load",
	Short: "Read out the CPU load of a device",
	Long:  "Read out the CPU load of a device.",
	Run: func(cmd *cobra.Command, args []string) {
		request := request.ReadCPULoadRequest{
			ReadRequest: getReadRequest(args[0]),
		}
		handleRequest(&request)
	},
}
