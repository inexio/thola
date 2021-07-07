// +build !client

package cmd

import (
	"errors"
	"github.com/inexio/thola/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCMD.AddCommand(apiCMD)

	apiCMD.Flags().Int("port", 8237, "Port for the API")
	apiCMD.Flags().Bool("no-ip-lock", false, "Allow multiple requests at a time for one IP")
	apiCMD.Flags().String("api-format", "json", "API format ('json' or 'xml')")
	apiCMD.Flags().String("username", "", "Username for authorization")
	apiCMD.Flags().String("password", "", "Password for authorization")
	apiCMD.Flags().String("certfile", "", "Cert file for SSL encryption")
	apiCMD.Flags().String("keyfile", "", "Key file for SSL encryption")
	apiCMD.Flags().String("ratelimit", "", "Ratelimit for the API (e.g. 1000 reqs/hour: \"1000-H\")")

	err := viper.BindPFlag("api.port", apiCMD.Flags().Lookup("port"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag port")
		return
	}
	err = viper.BindPFlag("request.no-ip-lock", apiCMD.Flags().Lookup("no-ip-lock"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag no-ip-lock")
		return
	}
	err = viper.BindPFlag("api.format", apiCMD.Flags().Lookup("api-format"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag api-format")
		return
	}
	err = viper.BindPFlag("api.username", apiCMD.Flags().Lookup("username"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag username")
		return
	}
	err = viper.BindPFlag("api.password", apiCMD.Flags().Lookup("password"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag password")
		return
	}
	err = viper.BindPFlag("api.certfile", apiCMD.Flags().Lookup("certfile"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag certfile")
		return
	}
	err = viper.BindPFlag("api.keyfile", apiCMD.Flags().Lookup("keyfile"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag keyfile")
		return
	}
	err = viper.BindPFlag("api.ratelimit", apiCMD.Flags().Lookup("ratelimit"))
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't bind flag ratelimit")
		return
	}
}

var apiCMD = &cobra.Command{
	Use:   "api",
	Short: "Start and configure the API of Thola",
	Long: "Start and configure the API of Thola.\n\n" +
		"You can set a port and authorization for the API. The authorization method is HTTP basic auth.\n" +
		"If the username or password is empty, the API won't use any authorization.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := rootCMD.PersistentPreRunE(cmd, args)
		if err != nil {
			return err
		}

		setDeviceDefaults()

		if !(viper.GetString("api.format") == "json" || viper.GetString("format") == "xml") {
			return errors.New("invalid api format set")
		}
		if viper.GetString("api.username") != "" && viper.GetString("api.password") == "" {
			return errors.New("username but no password for api authorization set")
		}
		if viper.GetString("api.username") == "" && viper.GetString("api.password") != "" {
			return errors.New("password but no username for api authorization set")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		api.StartAPI()
	},
}
