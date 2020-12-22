// +build client

package cmd

import (
	"context"
	"fmt"
	"github.com/inexio/thola/core/parser"
	"github.com/inexio/thola/core/request"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
)

func addBinarySpecificDeviceFlags(fs *flag.FlagSet) {
	fs.StringSlice("snmp-community", nil, "Community strings for SNMP to use")
	fs.StringSlice("snmp-version", nil, "SNMP versions to use (1, 2c or 3)")
	fs.IntSlice("snmp-port", nil, "Ports for SNMP to use")
}

func handleRequest(r request.Request) {
	rid := xid.New().String()
	logger := log.With().Str("request_id", rid).Logger()
	ctx := logger.WithContext(request.NewContextWithRequestID(context.Background(), rid))

	log.Ctx(ctx).Trace().Msg("sending request")

	resp, err := request.ProcessRequest(ctx, r)
	if err != nil {
		handleError(ctx, err)
		os.Exit(3)
	}

	log.Ctx(ctx).Trace().Msg("received response")

	b, err := parser.Parse(resp, viper.GetString("format"))
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Request successful, but failed to parse response")
		os.Exit(3)
	}

	fmt.Printf("%s\n", b)
	os.Exit(resp.GetExitCode())
}
