package cmd

import (
	"context"
	"fmt"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/parser"
	"github.com/inexio/thola/internal/request"
	"github.com/inexio/thola/internal/utility"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func getBaseRequest(host string) request.BaseRequest {
	var nullInt *int
	var nullString *string
	timeout := viper.GetInt("request.timeout")
	parallelRequests := viper.GetInt("device.snmp-discover-par-requests")
	discoverTimeout := viper.GetInt("device.snmp-discover-timeout")
	retries := viper.GetInt("device.snmp-discover-retries")
	authUsername := viper.GetString("device.http-username")
	authPassword := viper.GetString("device.http-password")
	v3Level := viper.GetString("device.snmp-v3-level")
	v3ContextName := viper.GetString("device.snmp-v3-context")
	v3User := viper.GetString("device.snmp-v3-user")
	v3AuthKey := viper.GetString("device.snmp-v3-auth-key")
	v3AuthProto := viper.GetString("device.snmp-v3-auth-proto")
	v3PrivKey := viper.GetString("device.snmp-v3-priv-key")
	v3PrivProto := viper.GetString("device.snmp-v3-priv-proto")
	return request.BaseRequest{
		Timeout: utility.IfThenElse(deviceFlagSet.Changed("timeout"), &timeout, nullInt).(*int),
		DeviceData: request.DeviceData{
			IPAddress: host,
			ConnectionData: network.ConnectionData{
				SNMP: &network.SNMPConnectionData{
					Communities:              utility.IfThenElse(deviceFlagSet.Changed("snmp-community"), viper.GetStringSlice("device.snmp-communities"), []string{}).([]string),
					Versions:                 utility.IfThenElse(deviceFlagSet.Changed("snmp-version"), viper.GetStringSlice("device.snmp-versions"), []string{}).([]string),
					Ports:                    utility.IfThenElse(deviceFlagSet.Changed("snmp-port"), viper.GetIntSlice("device.snmp-ports"), []int{}).([]int),
					DiscoverParallelRequests: utility.IfThenElse(deviceFlagSet.Changed("snmp-discover-par-requests"), &parallelRequests, nullInt).(*int),
					DiscoverTimeout:          utility.IfThenElse(deviceFlagSet.Changed("snmp-discover-timeout"), &discoverTimeout, nullInt).(*int),
					DiscoverRetries:          utility.IfThenElse(deviceFlagSet.Changed("snmp-discover-retries"), &retries, nullInt).(*int),
					V3Data: network.SNMPv3ConnectionData{
						Level:        utility.IfThenElse(deviceFlagSet.Changed("snmp-v3-level"), &v3Level, nullString).(*string),
						ContextName:  utility.IfThenElse(deviceFlagSet.Changed("snmp-v3-context"), &v3ContextName, nullString).(*string),
						User:         utility.IfThenElse(deviceFlagSet.Changed("snmp-v3-user"), &v3User, nullString).(*string),
						AuthKey:      utility.IfThenElse(deviceFlagSet.Changed("snmp-v3-auth-key"), &v3AuthKey, nullString).(*string),
						AuthProtocol: utility.IfThenElse(deviceFlagSet.Changed("snmp-v3-auth-proto"), &v3AuthProto, nullString).(*string),
						PrivKey:      utility.IfThenElse(deviceFlagSet.Changed("snmp-v3-priv-key"), &v3PrivKey, nullString).(*string),
						PrivProtocol: utility.IfThenElse(deviceFlagSet.Changed("snmp-v3-priv-proto"), &v3PrivProto, nullString).(*string),
					},
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
