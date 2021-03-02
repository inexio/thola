package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkSBCCMD)
	checkCMD.AddCommand(checkSBCCMD)

	checkSBCCMD.Flags().Float64("system-health-score-warning", 0, "warning threshold for system health score")
	checkSBCCMD.Flags().Float64("system-health-score-critical", 0, "critical threshold for system health score")
}

var checkSBCCMD = &cobra.Command{
	Use:   "sbc",
	Short: "Read out sbc specific metrics as performance data",
	Long:  "Read out sbc specific metrics as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckSBCRequest{
			CheckDeviceRequest:          getCheckDeviceRequest(args[0]),
			SystemHealthScoreThresholds: generateCheckThresholds(cmd, "system-health-score-warning", "", "system-health-score-critical", ""),
		}
		handleRequest(&r)
	},
}
