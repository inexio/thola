package cmd

import (
	"context"
	"fmt"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/parser"
	"github.com/inexio/thola/core/request"
	"github.com/inexio/thola/core/utility"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var deviceFlagSet = buildDeviceFlagSet()

var (
	defaultRequestTimeout          = 0
	defaultSNMPCommunity           = []string{"public"}
	defaultSNMPVersion             = []string{"2c", "1"}
	defaultSNMPPort                = []int{161}
	defaultSNMPDiscoverParRequests = 5
	defaultSNMPDiscoverTimeout     = 2
	defaultSNMPDiscoverRetries     = 0
)

func setDeviceDefaults() {
	viper.SetDefault("request.timeout", defaultRequestTimeout)
	viper.SetDefault("device.snmp-communities", defaultSNMPCommunity)
	viper.SetDefault("device.snmp-versions", defaultSNMPVersion)
	viper.SetDefault("device.snmp-ports", defaultSNMPPort)
	viper.SetDefault("device.snmp-discover-par-requests", defaultSNMPDiscoverParRequests)
	viper.SetDefault("device.snmp-discover-timeout", defaultSNMPDiscoverTimeout)
	viper.SetDefault("device.snmp-discover-retries", defaultSNMPDiscoverRetries)
}

func buildDeviceFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("device_flags", flag.ContinueOnError)

	addBinarySpecificDeviceFlags(fs)
	fs.String("ip", "", "IP of the device")

	fs.Int("timeout", defaultRequestTimeout, "Timeout for the request in seconds (0 => no timeout)")
	fs.Int("snmp-discover-par-requests", defaultSNMPDiscoverParRequests, "The amount of parallel connection requests used while trying to get a valid SNMP connection")
	fs.Int("snmp-discover-timeout", defaultSNMPDiscoverTimeout, "The timeout in seconds used while trying to get a valid SNMP connection")
	fs.Int("snmp-discover-retries", defaultSNMPDiscoverRetries, "The retries used while trying to get a valid SNMP connection")
	fs.IntSlice("http-port", nil, "Ports for HTTP to use")
	fs.IntSlice("https-port", nil, "Ports for HTTPS to use")
	fs.String("http-username", "", "Username for HTTP/HTTPS authorization")
	fs.String("http-password", "", "Password for HTTP/HTTPS authorization")

	return fs
}

func addDeviceFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().AddFlagSet(deviceFlagSet)

	err := cmd.MarkPersistentFlagRequired("ip")
	if err != nil {
		log.Error().
			AnErr("Error", err).
			Msg("Can't make flag ip required")
		return
	}
}

func bindDeviceFlags(cmd *cobra.Command) error {
	if x := cmd.Flags().Lookup("ip"); x != nil {
		err := viper.BindPFlag("device.ip", x)
		if err != nil {
			log.Error().
				AnErr("Error", err).
				Msg("Can't bind flag ip")
			return err
		}
	}
	if x := cmd.Flags().Lookup("timeout"); x != nil {
		err := viper.BindPFlag("request.timeout", x)
		if err != nil {
			log.Error().
				AnErr("Error", err).
				Msg("Can't bind flag timeout")
			return err
		}
	}
	if x := cmd.Flags().Lookup("snmp-discover-par-requests"); x != nil {
		err := viper.BindPFlag("device.snmp-discover-par-requests", x)
		if err != nil {
			log.Error().
				AnErr("Error", err).
				Msg("Can't bind flag snmp-discover-par-requests")
			return err
		}
	}
	if x := cmd.Flags().Lookup("snmp-discover-timeout"); x != nil {
		err := viper.BindPFlag("device.snmp-discover-timeout", x)
		if err != nil {
			log.Error().
				AnErr("Error", err).
				Msg("Can't bind flag snmp-discover-timeout")
			return err
		}
	}
	if x := cmd.Flags().Lookup("snmp-discover-retries"); x != nil {
		err := viper.BindPFlag("device.snmp-discover-retries", x)
		if err != nil {
			log.Error().
				AnErr("Error", err).
				Msg("Can't bind flag snmp-discover-retries")
			return err
		}
	}
	if x := cmd.Flags().Lookup("snmp-community"); x != nil {
		err := viper.BindPFlag("device.snmp-communities", x)
		if err != nil {
			log.Error().
				AnErr("Error", err).
				Msg("Can't bind flag snmp-community-strings")
			return err
		}
	}
	if x := cmd.Flags().Lookup("snmp-version"); x != nil {
		err := viper.BindPFlag("device.snmp-versions", x)
		if err != nil {
			log.Error().
				AnErr("Error", err).
				Msg("Can't bind flag snmp-version")
			return err
		}
	}
	if x := cmd.Flags().Lookup("snmp-port"); x != nil {
		err := viper.BindPFlag("device.snmp-ports", x)
		if err != nil {
			log.Error().
				AnErr("Error", err).
				Msg("Can't bind flag snmp-port")
			return err
		}
	}
	if x := cmd.Flags().Lookup("http-port"); x != nil {
		err := viper.BindPFlag("device.http-ports", x)
		if err != nil {
			log.Error().
				AnErr("Error", err).
				Msg("Can't bind flag http-port")
			return err
		}
	}
	if x := cmd.Flags().Lookup("https-port"); x != nil {
		err := viper.BindPFlag("device.https-ports", x)
		if err != nil {
			log.Error().
				AnErr("Error", err).
				Msg("Can't bind flag https-port")
			return err
		}
	}
	if x := cmd.Flags().Lookup("http-username"); x != nil {
		err := viper.BindPFlag("device.http-username", x)
		if err != nil {
			log.Error().
				AnErr("Error", err).
				Msg("Can't bind flag http-username")
			return err
		}
	}
	if x := cmd.Flags().Lookup("http-password"); x != nil {
		err := viper.BindPFlag("device.http-password", x)
		if err != nil {
			log.Error().
				AnErr("Error", err).
				Msg("Can't bind flag http-password")
			return err
		}
	}
	return nil
}

func getBaseRequest() request.BaseRequest {
	var nullInt *int
	var nullString *string
	timeout := viper.GetInt("request.timeout")
	parallelRequests := viper.GetInt("device.snmp-discover-par-requests")
	discoverTimeout := viper.GetInt("device.snmp-discover-timeout")
	retries := viper.GetInt("device.snmp-discover-retries")
	authUsername := viper.GetString("device.http-username")
	authPassword := viper.GetString("device.http-password")
	return request.BaseRequest{
		Timeout: utility.IfThenElse(deviceFlagSet.Changed("timeout"), &timeout, nullInt).(*int),
		DeviceData: request.DeviceData{
			IPAddress: viper.GetString("device.ip"),
			ConnectionData: network.ConnectionData{
				SNMP: &network.SNMPConnectionData{
					Communities:              utility.IfThenElse(deviceFlagSet.Changed("snmp-community"), viper.GetStringSlice("device.snmp-communities"), []string{}).([]string),
					Versions:                 utility.IfThenElse(deviceFlagSet.Changed("snmp-version"), viper.GetStringSlice("device.snmp-versions"), []string{}).([]string),
					Ports:                    utility.IfThenElse(deviceFlagSet.Changed("snmp-port"), viper.GetIntSlice("device.snmp-ports"), []int{}).([]int),
					DiscoverParallelRequests: utility.IfThenElse(deviceFlagSet.Changed("snmp-discover-par-requests"), &parallelRequests, nullInt).(*int),
					DiscoverTimeout:          utility.IfThenElse(deviceFlagSet.Changed("snmp-discover-timeout"), &discoverTimeout, nullInt).(*int),
					DiscoverRetries:          utility.IfThenElse(deviceFlagSet.Changed("snmp-discover-retries"), &retries, nullInt).(*int),
				},
				HTTP: &network.HTTPConnectionData{
					HTTPPorts:    utility.IfThenElse(deviceFlagSet.Changed("http-port"), viper.GetIntSlice("device.http-ports"), []int{}).([]int),
					HTTPSPorts:   utility.IfThenElse(deviceFlagSet.Changed("https-port"), viper.GetIntSlice("device.https-ports"), []int{}).([]int),
					AuthUsername: utility.IfThenElse(deviceFlagSet.Changed("http-username"), &authUsername, nullString).(*string),
					AuthPassword: utility.IfThenElse(deviceFlagSet.Changed("http-password"), &authPassword, nullString).(*string),
				},
			},
		},
	}
}

func handleError(ctx context.Context, err error) {
	b, err := parser.Parse(err, viper.GetString("format"))
	if err != nil {
		log.Ctx(ctx).Error().AnErr("parse_error", err).AnErr("original_error", err).Msg("failed to parse error")
	} else {
		fmt.Printf("%s\n", b)
	}
}
