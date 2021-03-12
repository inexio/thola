package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkUPSCMD)
	checkCMD.AddCommand(checkUPSCMD)

	checkUPSCMD.Flags().Float64("batt-current-warning-min", 0, "Warning min threshold for battery current")
	checkUPSCMD.Flags().Float64("batt-current-warning-max", 0, "Warning max threshold for battery current")
	checkUPSCMD.Flags().Float64("batt-current-critical-min", 0, "Critical min threshold for battery current")
	checkUPSCMD.Flags().Float64("batt-current-critical-max", 0, "Critical max threshold for battery current")

	checkUPSCMD.Flags().Float64("batt-temperature-warning-min", 0, "Warning min threshold for battery temperature")
	checkUPSCMD.Flags().Float64("batt-temperature-warning-max", 0, "Warning max threshold for battery temperature")
	checkUPSCMD.Flags().Float64("batt-temperature-critical-min", 0, "Critical min threshold for battery temperature")
	checkUPSCMD.Flags().Float64("batt-temperature-critical-max", 0, "Critical max threshold for battery temperature")

	checkUPSCMD.Flags().Float64("current-load-warning-min", 0, "Warning min threshold for current load")
	checkUPSCMD.Flags().Float64("current-load-warning-max", 0, "Warning max threshold for current load")
	checkUPSCMD.Flags().Float64("current-load-critical-min", 0, "Critical min threshold for current load")
	checkUPSCMD.Flags().Float64("current-load-critical-max", 0, "Critical max threshold for current load")

	checkUPSCMD.Flags().Float64("rectifier-current-warning-min", 0, "Warning min threshold for rectifier current")
	checkUPSCMD.Flags().Float64("rectifier-current-warning-max", 0, "Warning max threshold for rectifier current")
	checkUPSCMD.Flags().Float64("rectifier-current-critical-min", 0, "Critical min threshold for rectifier current")
	checkUPSCMD.Flags().Float64("rectifier-current-critical-max", 0, "Critical max threshold for rectifier current")

	checkUPSCMD.Flags().Float64("system-voltage-warning-min", 0, "Warning min threshold for system voltage")
	checkUPSCMD.Flags().Float64("system-voltage-warning-max", 0, "Warning max threshold for system voltage")
	checkUPSCMD.Flags().Float64("system-voltage-critical-min", 0, "Critical min threshold for system voltage")
	checkUPSCMD.Flags().Float64("system-voltage-critical-max", 0, "Critical max threshold for system voltage")
}

var checkUPSCMD = &cobra.Command{
	Use:   "ups",
	Short: "Checks whether a UPS device has its main voltage applied",
	Long: "Checks whether a UPS device has its main voltage applied.\n\n" +
		"All UPS statistics will be printed as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckUPSRequest{
			CheckDeviceRequest:           getCheckDeviceRequest(args[0]),
			BatteryCurrentThresholds:     generateCheckThresholds(cmd, "batt-current-warning-min", "batt-current-warning-max", "batt-current-critical-min", "batt-current-critical-max", false),
			BatteryTemperatureThresholds: generateCheckThresholds(cmd, "batt-temperature-warning-min", "batt-temperature-warning-max", "batt-temperature-critical-min", "batt-temperature-critical-max", false),
			CurrentLoadThresholds:        generateCheckThresholds(cmd, "current-load-warning-min", "current-load-warning-max", "current-load-warning-max", "current-load-warning-max", false),
			RectifierCurrentThresholds:   generateCheckThresholds(cmd, "rectifier-current-warning-min", "rectifier-current-warning-max", "rectifier-current-critical-min", "rectifier-current-critical-max", false),
			SystemVoltageThresholds:      generateCheckThresholds(cmd, "system-voltage-warning-min", "system-voltage-warning-max", "system-voltage-critical-min", "system-voltage-critical-max", false),
		}
		handleRequest(&r)
	},
}
