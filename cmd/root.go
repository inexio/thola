// +build !client

package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var cfgFile string

func init() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	cobra.OnInitialize(initConfig)

	rootCMD.PersistentFlags().StringVar(&cfgFile, "config", "", "The location of the config file")
	rootCMD.PersistentFlags().String("loglevel", "error", "The loglevel")
	rootCMD.PersistentFlags().String("format", "pretty", "Output format ('json', 'xml' or 'pretty')")
	rootCMD.PersistentFlags().String("db-drivername", "built-in", "Database type for caching ('built-in', 'mysql' or 'redis' supported)")
	rootCMD.PersistentFlags().String("db-duration", "60m", "Duration in which the cache stays valid")
	rootCMD.PersistentFlags().String("sql-datasourcename", "", "Data sourcename if using a sql driver")
	rootCMD.PersistentFlags().String("redis-addr", "", "Database address if using the redis driver")
	rootCMD.PersistentFlags().String("redis-pass", "", "Database password if using the redis driver")

	rootCMD.PersistentFlags().Int("redis-db", 0, "Database to use if using the redis driver")

	rootCMD.PersistentFlags().Bool("db-rebuild", false, "Rebuild the cache DB")
	rootCMD.PersistentFlags().Bool("no-cache", false, "Don't use a database cache")
	rootCMD.Flags().Bool("version", false, "Prints the version of Thola")

	err := viper.BindPFlag("config", rootCMD.PersistentFlags().Lookup("config"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag config")
		return
	}

	err = viper.BindPFlag("loglevel", rootCMD.PersistentFlags().Lookup("loglevel"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag loglevel")
		return
	}

	err = viper.BindPFlag("format", rootCMD.PersistentFlags().Lookup("format"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag format")
		return
	}

	err = viper.BindPFlag("db.drivername", rootCMD.PersistentFlags().Lookup("db-drivername"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag db-drivername")
		return
	}

	err = viper.BindPFlag("db.duration", rootCMD.PersistentFlags().Lookup("db-duration"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag db-duration")
		return
	}

	err = viper.BindPFlag("db.sql.datasourcename", rootCMD.PersistentFlags().Lookup("sql-datasourcename"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag sql-datasourcename")
		return
	}

	err = viper.BindPFlag("db.redis.addr", rootCMD.PersistentFlags().Lookup("redis-addr"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag redis-addr")
		return
	}

	err = viper.BindPFlag("db.redis.password", rootCMD.PersistentFlags().Lookup("redis-pass"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag redis-password")
		return
	}

	err = viper.BindPFlag("db.redis.db", rootCMD.PersistentFlags().Lookup("redis-db"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag redis-db")
		return
	}

	err = viper.BindPFlag("db.rebuild", rootCMD.PersistentFlags().Lookup("db-rebuild"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag db-rebuild")
		return
	}

	err = viper.BindPFlag("db.no-cache", rootCMD.PersistentFlags().Lookup("no-cache"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag no-cache")
		return
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(os.ExpandEnv("$HOME/.thola"))
		viper.AddConfigPath("./config")
		viper.AddConfigPath("/var/opt/thola")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.SetEnvPrefix("THOLA")
	viper.AutomaticEnv()

	_ = viper.ReadInConfig()
}

var rootCMD = &cobra.Command{
	Use:   "thola",
	Short: "Thola is an open source tool designed for communicating with different network devices",
	Long: "Thola is an open source tool designed for communicating with different network devices.\n\n" +
		"The main features are identifying, monitoring, requesting infos and statistics of devices.\n" +
		"It has a check plugin mode for running with popular monitoring systems like Icinga and Nagios.\n" +
		"In addition to that it has a REST API mode for integrating with existing IT infrastructure.",
	DisableFlagsInUseLine: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := bindDeviceFlags(cmd)
		if err != nil {
			return errors.Wrap(err, "failed to bind device flags")
		}
		if !(viper.GetString("format") == "json" || viper.GetString("format") == "xml" || viper.GetString("format") == "pretty") {
			return errors.New("invalid output format set")
		}
		loglevel, err := zerolog.ParseLevel(viper.GetString("loglevel"))
		if err != nil {
			return errors.New("invalid loglevel set")
		}
		zerolog.SetGlobalLevel(loglevel)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Lookup("version").Changed {
			fmt.Println("v0.1.2")
		} else {
			fmt.Print(cmd.UsageString())
		}
	},
}

// Execute is the entrypoint for the CLI interface.
func Execute() {
	if err := rootCMD.Execute(); err != nil {
		os.Exit(1)
	}
}
