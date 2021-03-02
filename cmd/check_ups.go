package cmd

import (
	"github.com/inexio/thola/core/request"
	"github.com/inexio/thola/core/value"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkUPSCMD)
	checkCMD.AddCommand(checkUPSCMD)

	checkUPSCMD.Flags().String("batt-current-warning-min", "", "Warning min threshold for battery current")
	checkUPSCMD.Flags().String("batt-current-warning-max", "", "Warning max threshold for battery current")
	checkUPSCMD.Flags().String("batt-current-critical-min", "", "Critical min threshold for battery current")
	checkUPSCMD.Flags().String("batt-current-critical-max", "", "Critical max threshold for battery current")

	checkUPSCMD.Flags().String("batt-temperature-warning-min", "", "Warning min threshold for battery temperature")
	checkUPSCMD.Flags().String("batt-temperature-warning-max", "", "Warning max threshold for battery temperature")
	checkUPSCMD.Flags().String("batt-temperature-critical-min", "", "Critical min threshold for battery temperature")
	checkUPSCMD.Flags().String("batt-temperature-critical-max", "", "Critical max threshold for battery temperature")

	checkUPSCMD.Flags().String("current-load-warning-min", "", "Warning min threshold for current load")
	checkUPSCMD.Flags().String("current-load-warning-max", "", "Warning max threshold for current load")
	checkUPSCMD.Flags().String("current-load-critical-min", "", "Critical min threshold for current load")
	checkUPSCMD.Flags().String("current-load-critical-max", "", "Critical max threshold for current load")

	checkUPSCMD.Flags().String("rectifier-current-warning-min", "", "Warning min threshold for rectifier current")
	checkUPSCMD.Flags().String("rectifier-current-warning-max", "", "Warning max threshold for rectifier current")
	checkUPSCMD.Flags().String("rectifier-current-critical-min", "", "Critical min threshold for rectifier current")
	checkUPSCMD.Flags().String("rectifier-current-critical-max", "", "Critical max threshold for rectifier current")

	checkUPSCMD.Flags().String("system-voltage-warning-min", "", "Warning min threshold for system voltage")
	checkUPSCMD.Flags().String("system-voltage-warning-max", "", "Warning max threshold for system voltage")
	checkUPSCMD.Flags().String("system-voltage-critical-min", "", "Critical min threshold for system voltage")
	checkUPSCMD.Flags().String("system-voltage-critical-max", "", "Critical max threshold for system voltage")
}

var checkUPSCMD = &cobra.Command{
	Use:   "ups",
	Short: "Checks whether a UPS device has its main voltage applied",
	Long: "Checks whether a UPS device has its main voltage applied.\n\n" +
		"All UPS statistics will be printed as performance data.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckUPSRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
			BatteryCurrentThresholds: request.CheckThresholds{
				WarningMin:  value.New(cmd.Flags().Lookup("batt-current-warning-min").Value),
				WarningMax:  value.New(cmd.Flags().Lookup("batt-current-warning-max").Value),
				CriticalMin: value.New(cmd.Flags().Lookup("batt-current-critical-min").Value),
				CriticalMax: value.New(cmd.Flags().Lookup("batt-current-critical-max").Value),
			},
			BatteryTemperatureThresholds: request.CheckThresholds{
				WarningMin:  value.New(cmd.Flags().Lookup("batt-temperature-warning-min").Value),
				WarningMax:  value.New(cmd.Flags().Lookup("batt-temperature-warning-max").Value),
				CriticalMin: value.New(cmd.Flags().Lookup("batt-temperature-critical-min").Value),
				CriticalMax: value.New(cmd.Flags().Lookup("batt-temperature-critical-max").Value),
			},
			CurrentLoadThresholds: request.CheckThresholds{
				WarningMin:  value.New(cmd.Flags().Lookup("current-load-warning-min").Value),
				WarningMax:  value.New(cmd.Flags().Lookup("current-load-warning-max").Value),
				CriticalMin: value.New(cmd.Flags().Lookup("current-load-critical-min").Value),
				CriticalMax: value.New(cmd.Flags().Lookup("current-load-critical-max").Value),
			},
			RectifierCurrentThresholds: request.CheckThresholds{
				WarningMin:  value.New(cmd.Flags().Lookup("rectifier-current-warning-min").Value),
				WarningMax:  value.New(cmd.Flags().Lookup("rectifier-current-warning-max").Value),
				CriticalMin: value.New(cmd.Flags().Lookup("rectifier-current-critical-min").Value),
				CriticalMax: value.New(cmd.Flags().Lookup("rectifier-current-critical-max").Value),
			},
			SystemVoltageThresholds: request.CheckThresholds{
				WarningMin:  value.New(cmd.Flags().Lookup("system-voltage-warning-min").Value),
				WarningMax:  value.New(cmd.Flags().Lookup("system-voltage-warning-max").Value),
				CriticalMin: value.New(cmd.Flags().Lookup("system-voltage-critical-min").Value),
				CriticalMax: value.New(cmd.Flags().Lookup("system-voltage-critical-max").Value),
			},
		}
		handleRequest(&r)
	},
}
