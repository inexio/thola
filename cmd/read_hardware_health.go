package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
)

func init() {
	readCMD.AddCommand(readHardwareHealth)
}

var readHardwareHealth = &cobra.Command{
	Use:   "hardware-health [host]",
	Short: "Read out the hardware health of a device",
	Long:  "Read out the hardware health of a device.",
	Run: func(cmd *cobra.Command, args []string) {
		request := request.ReadHardwareHealthRequest{
			ReadRequest: getReadRequest(args[0]),
		}
		handleRequest(&request)
	},
}
