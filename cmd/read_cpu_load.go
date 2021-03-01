package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
)

func init() {
	readCMD.AddCommand(readCPULoadCMD)
}

var readCPULoadCMD = &cobra.Command{
	Use:   "cpu-load [host]",
	Short: "Read out the CPU load of a device",
	Long:  "Read out the CPU load of a device.",
	Run: func(cmd *cobra.Command, args []string) {
		request := request.ReadCPULoadRequest{
			ReadRequest: getReadRequest(args[0]),
		}
		handleRequest(&request)
	},
}
