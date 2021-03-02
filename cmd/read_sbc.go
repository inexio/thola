package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
)

func init() {
	readCMD.AddCommand(readSBCCMD)
}

var readSBCCMD = &cobra.Command{
	Use:   "sbc",
	Short: "Read out SBC specific information of a device",
	Long:  "Read out SPC specific information of a device like global call per second or active local contacts, including information per agent and per realm.",
	Run: func(cmd *cobra.Command, args []string) {
		request := request.ReadSBCRequest{
			ReadRequest: getReadRequest(args[0]),
		}
		handleRequest(&request)
	},
}
