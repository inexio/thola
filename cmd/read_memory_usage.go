package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
)

func init() {
	readCMD.AddCommand(readMemoryUsageCMD)
}

var readMemoryUsageCMD = &cobra.Command{
	Use:   "memory-usage",
	Short: "Read out the memory usage of a device",
	Long:  "Read out the memory usage of a device.",
	Run: func(cmd *cobra.Command, args []string) {
		request := request.ReadMemoryUsageRequest{
			ReadRequest: getReadRequest(args[0]),
		}
		handleRequest(&request)
	},
}
