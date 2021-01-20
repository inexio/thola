package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkSBCCMD)
	checkCMD.AddCommand(checkSBCCMD)
}

var checkSBCCMD = &cobra.Command{
	Use:   "sbc",
	Short: "Read out sbc specific metrics as performance data",
	Long:  "Read out sbc specific metrics as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckSBCRequest{
			CheckDeviceRequest: getCheckDeviceRequest(),
		}
		handleRequest(&r)
	},
}
