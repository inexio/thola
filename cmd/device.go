// +build !client

package cmd

import (
	"context"
	"fmt"
	"github.com/inexio/thola/core/database"
	"github.com/inexio/thola/core/parser"
	"github.com/inexio/thola/core/request"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
)

func addBinarySpecificDeviceFlags(fs *flag.FlagSet) {
	fs.StringSlice("snmp-community", defaultSNMPCommunity, "Community strings for SNMP to use")
	fs.StringSlice("snmp-version", defaultSNMPVersion, "SNMP versions to use (1, 2c or 3)")
	fs.IntSlice("snmp-port", defaultSNMPPort, "Ports for SNMP to use")
}

func handleRequest(r request.Request) {
	logger := log.With().Str("request_id", xid.New().String()).Logger()
	ctx := logger.WithContext(context.Background())

	db, err := database.GetDB(ctx)
	if err != nil {
		handleError(ctx, err)
		os.Exit(3)
	}

	resp, err := request.ProcessRequest(ctx, r)
	if err != nil {
		handleError(ctx, err)
		_ = db.CloseConnection(ctx)
		os.Exit(3)
	}

	err = db.CloseConnection(ctx)
	if err != nil {
		handleError(ctx, err)
		os.Exit(3)
	}

	b, err := parser.Parse(resp, viper.GetString("format"))
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Request successful, but failed to parse response")
		os.Exit(3)
	}

	fmt.Printf("%s\n", b)
	os.Exit(resp.GetExitCode())
}
