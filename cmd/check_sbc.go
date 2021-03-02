package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/inexio/thola/core/value"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkSBCCMD)
	checkCMD.AddCommand(checkSBCCMD)

	checkSBCCMD.Flags().String("system-health-score-warning", "", "warning threshold for system health score")
	checkSBCCMD.Flags().String("system-health-score-critical", "", "critical threshold for system health score")
}

var checkSBCCMD = &cobra.Command{
	Use:   "sbc",
	Short: "Read out sbc specific metrics as performance data",
	Long:  "Read out sbc specific metrics as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckSBCRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
			SystemHealthScoreThresholds: request.CheckThresholds{
				WarningMin:  value.New(cmd.Flags().Lookup("system-health-score-warning").Value),
				CriticalMin: value.New(cmd.Flags().Lookup("system-health-score-critical").Value),
			},
		}
		handleRequest(&r)
	},
}
