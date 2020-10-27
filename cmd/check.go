package cmd

import (
	"fmt"
	"github.com/inexio/thola/core/request"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCMD.AddCommand(checkCMD)

	checkCMD.PersistentFlags().Bool("json-metrics", false, "Print all metrics in the JSON format")

	err := viper.BindPFlag("check.json-metrics", checkCMD.PersistentFlags().Lookup("json-metrics"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag config")
		return
	}
}

var checkCMD = &cobra.Command{
	Use:   "check",
	Short: "Use Thola to monitor network devices",
	Long: "Use Thola to monitor network devices.\n\n" +
		"By default the output is in the check plugin format, which is compatible with Nagios or Icinga.\n" +
		"You need to specify the information which you want to check with a subcommand.",
	DisableFlagsInUseLine: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := rootCMD.PersistentPreRunE(cmd, args)
		if err != nil {
			return err
		}

		if !cmd.Flags().Changed("format") {
			viper.Set("format", "check-plugin")
		}

		zerolog.SetGlobalLevel(zerolog.Disabled)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.UsageString())
	},
}

func getCheckDeviceRequest() request.CheckDeviceRequest {
	return request.CheckDeviceRequest{
		BaseRequest:  getBaseRequest(),
		CheckRequest: getCheckRequest(),
	}
}

func getCheckRequest() request.CheckRequest {
	return request.CheckRequest{
		JSONMetrics:          viper.GetBool("check.json-metrics"),
		PrintPerformanceData: true,
	}
}
