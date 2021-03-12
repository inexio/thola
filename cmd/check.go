package cmd

import (
	"fmt"
	"github.com/inexio/go-monitoringplugin"
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

		if !cmd.Flags().Changed("loglevel") {
			zerolog.SetGlobalLevel(zerolog.Disabled)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.UsageString())
	},
}

func getCheckDeviceRequest(host string) request.CheckDeviceRequest {
	return request.CheckDeviceRequest{
		BaseRequest:  getBaseRequest(host),
		CheckRequest: getCheckRequest(),
	}
}

func getCheckRequest() request.CheckRequest {
	return request.CheckRequest{
		JSONMetrics:          viper.GetBool("check.json-metrics"),
		PrintPerformanceData: true,
	}
}

func generateCheckThresholds(cmd *cobra.Command, warningMin, warningMax, criticalMin, criticalMax string, setMinToZeroIfEmpty bool) monitoringplugin.Thresholds {
	var thresholds monitoringplugin.Thresholds
	if flagName := warningMin; flagName != "" && cmd.Flags().Changed(flagName) {
		v, err := cmd.Flags().GetFloat64(flagName)
		if err != nil {
			log.Fatal().Err(err).Msgf("flag '%s' is not a float64", flagName)
		}
		thresholds.WarningMin = v
	}
	if flagName := warningMax; flagName != "" && cmd.Flags().Changed(flagName) {
		v, err := cmd.Flags().GetFloat64(flagName)
		if err != nil {
			log.Fatal().Err(err).Msgf("flag '%s' is not a float64", flagName)
		}
		thresholds.WarningMax = v
	}
	if flagName := criticalMin; flagName != "" && cmd.Flags().Changed(flagName) {
		v, err := cmd.Flags().GetFloat64(flagName)
		if err != nil {
			log.Fatal().Err(err).Msgf("flag '%s' is not a float64", flagName)
		}
		thresholds.CriticalMin = v
	}
	if flagName := criticalMax; flagName != "" && cmd.Flags().Changed(flagName) {
		v, err := cmd.Flags().GetFloat64(flagName)
		if err != nil {
			log.Fatal().Err(err).Msgf("flag '%s' is not a float64", flagName)
		}
		thresholds.CriticalMax = v
	}

	if setMinToZeroIfEmpty {
		if thresholds.HasWarning() && thresholds.WarningMin == nil {
			thresholds.WarningMin = 0
		}
		if thresholds.HasCritical() && thresholds.CriticalMin == nil {
			thresholds.CriticalMin = 0
		}
	}

	return thresholds
}
