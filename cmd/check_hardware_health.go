package cmd

import (
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/request"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkHardwareHealthCMD)
	checkCMD.AddCommand(checkHardwareHealthCMD)
}

var checkHardwareHealthCMD = &cobra.Command{
	Use:   "hardware-health",
	Short: "Check hardware-health of a device.",
	Long: "Check hardware-health of a device and return various performance data.\n" +
		"Performance data include states for temperatures, power supply, fans, etc. with the following meanings:\n" +
		"\t0: " + string(device.HardwareHealthComponentStateInitial) + "\n" +
		"\t1: " + string(device.HardwareHealthComponentStateNormal) + "\n" +
		"\t2: " + string(device.HardwareHealthComponentStateWarning) + "\n" +
		"\t3: " + string(device.HardwareHealthComponentStateCritical) + "\n" +
		"\t4: " + string(device.HardwareHealthComponentStateShutdown) + "\n" +
		"\t5: " + string(device.HardwareHealthComponentStateNotPresent) + "\n" +
		"\t6: " + string(device.HardwareHealthComponentStateNotFunctioning) + "\n" +
		"\t7: " + string(device.HardwareHealthComponentStateUnknown),
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckHardwareHealthRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
		}
		handleRequest(&r)
	},
}
