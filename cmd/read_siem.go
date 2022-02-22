package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(readSIEM)
	readCMD.AddCommand(readSIEM)
}

var readSIEM = &cobra.Command{
	Use:   "siem",
	Short: "Read out the SIEM information of a device",
	Long:  "Read out the SIEM information of a device",
	Run: func(cmd *cobra.Command, args []string) {
		request := request.ReadSIEMRequest{
			ReadRequest: getReadRequest(args[0]),
		}
		handleRequest(&request)
	},
}
