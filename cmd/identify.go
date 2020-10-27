package cmd

import (
	"github.com/spf13/cobra"
	"thola/core/request"
)

func init() {
	addDeviceFlags(identifyCMD)
	rootCMD.AddCommand(identifyCMD)
}

var identifyCMD = &cobra.Command{
	Use:   "identify",
	Short: "Automatically identify devices",
	Long: "Automatically identify devices.\n\n" +
		"It returns properties like vendor, model, serial number,...",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.IdentifyRequest{
			BaseRequest: getBaseRequest(),
		}
		handleRequest(&r)
	},
}
