package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkDiskCMD)
	checkCMD.AddCommand(checkDiskCMD)

	checkDiskCMD.Flags().Float64("warning", 0, "warning threshold for free disk space")
	checkDiskCMD.Flags().Float64("critical", 0, "critical threshold for free disk space")
}

var checkDiskCMD = &cobra.Command{
	Use:   "disk",
	Short: "Check the disk of a device",
	Long: "Checks the disk of a device.\n\n" +
		"The metrics will be printed as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckDiskRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
			DiskThresholds:     generateCheckThresholds(cmd, "", "warning", "", "critical", true),
		}
		handleRequest(&r)
	},
}
