package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/inexio/thola/internal/utility"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkHighAvailabilityCMD)
	checkCMD.AddCommand(checkHighAvailabilityCMD)

	checkHighAvailabilityCMD.Flags().String("role", "", "Expected role of the device in its high availability setup ('master' or 'slave')")
	checkHighAvailabilityCMD.Flags().Float64("nodes-warning", 0, "warning threshold for number of nodes in high availability setup")
	checkHighAvailabilityCMD.Flags().Float64("nodes-critical", 0, "critical threshold for number of nodes in high availability setup")
}

var checkHighAvailabilityCMD = &cobra.Command{
	Use:   "high-availability",
	Short: "Check the high availability status of a device",
	Long:  "Checks the high availability status of a device.",
	Run: func(cmd *cobra.Command, args []string) {
		var nilString *string
		role := cmd.Flags().Lookup("role").Value.String()
		r := request.CheckHighAvailabilityRequest{
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
			Role:               utility.IfThenElse(cmd.Flags().Changed("role"), &role, nilString).(*string),
			NodesThresholds:    generateCheckThresholds(cmd, "nodes-warning", "", "nodes-critical", "", true),
		}
		handleRequest(&r)
	},
}
