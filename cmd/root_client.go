// +build client

package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

func init() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})

	rootCMD.PersistentFlags().String("format", "pretty", "Output format ('json', 'xml' or 'pretty')")
	rootCMD.PersistentFlags().String("target-api", "", "The URL of the target API")
	rootCMD.PersistentFlags().String("target-api-username", "", "The username for authorization on the target API")
	rootCMD.PersistentFlags().String("target-api-password", "", "The password for authorization on the target API")
	rootCMD.PersistentFlags().String("target-api-format", "json", "The format of the target API ('json' or 'xml')")

	rootCMD.PersistentFlags().Bool("insecure-ssl-cert", false, "Allow insecure SSL certificate of the target API")

	rootCMD.Flags().Bool("version", false, "Prints the version of Thola")

	err := rootCMD.MarkPersistentFlagRequired("target-api")
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't make flag target-api required")
		return
	}

	err = viper.BindPFlag("format", rootCMD.PersistentFlags().Lookup("format"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag format")
		return
	}

	err = viper.BindPFlag("target-api", rootCMD.PersistentFlags().Lookup("target-api"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag target-api")
		return
	}

	err = viper.BindPFlag("target-api-username", rootCMD.PersistentFlags().Lookup("target-api-username"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag target-api-username")
		return
	}

	err = viper.BindPFlag("target-api-password", rootCMD.PersistentFlags().Lookup("target-api-password"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag target-api-password")
		return
	}

	err = viper.BindPFlag("target-api-format", rootCMD.PersistentFlags().Lookup("target-api-format"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag target-api-format")
		return
	}

	err = viper.BindPFlag("insecure-ssl-cert", rootCMD.PersistentFlags().Lookup("insecure-ssl-cert"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag insecure-ssl-cert")
		return
	}

	err = viper.BindPFlag("version", rootCMD.PersistentFlags().Lookup("version"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag version")
		return
	}
}

var rootCMD = &cobra.Command{
	Use:   "thola",
	Short: "Thola is an open source tool designed for communicating with different network devices",
	Long: "Thola is an open source tool designed for communicating with different network devices.\n\n" +
		"The main features are identifying, monitoring, requesting infos and statistics of devices.\n" +
		"It has a check plugin mode for running with popular monitoring systems like Icinga and Nagios.\n" +
		"In addition to that it has a REST API mode for integrating with existing IT infrastructure.\n" +
		"This is the client version which only sends requests to an instance of Thola running in API mode.",
	DisableFlagsInUseLine: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := bindDeviceFlags(cmd)
		if err != nil {
			return errors.Wrap(err, "failed to bind device flags")
		}
		if !(viper.GetString("format") == "json" || viper.GetString("format") == "xml" || viper.GetString("format") == "pretty") {
			return errors.New("invalid output format set")
		}
		if !(viper.GetString("target-api-format") == "json" || viper.GetString("target-api-format") == "xml") {
			return errors.New("invalid api format set")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("version") {
			fmt.Println("v0.1.1")
		} else {
			fmt.Print(cmd.UsageString())
		}
	},
}

func Execute() {
	if err := rootCMD.Execute(); err != nil {
		os.Exit(1)
	}
}
