package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(readHighAvailability)
	readCMD.AddCommand(readHighAvailability)
}

var readHighAvailability = &cobra.Command{
	Use:   "high-availability",
	Short: "Read out the high availability status of a device",
	Long:  "Read out the high availability status of a device.",
	Run: func(cmd *cobra.Command, args []string) {
		request := request.ReadHighAvailabilityRequest{
			ReadRequest: getReadRequest(args[0]),
		}
		handleRequest(&request)
	},
}
