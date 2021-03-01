package cmd

import (
	"context"
	"fmt"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/parser"
	"github.com/inexio/thola/core/request"
	"github.com/inexio/thola/core/utility"
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
