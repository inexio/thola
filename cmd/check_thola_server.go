//go:build client
// +build client

package cmd

import (
	"github.com/inexio/thola/internal/request"
	"github.com/spf13/cobra"
)

func init() {
	checkCMD.AddCommand(checkTholaServerCMD)
}

var checkTholaServerCMD = &cobra.Command{
	Use:   "thola-server",
	Short: "Check whether a Thola server is reachable",
	Long: "Check whether a Thola server is reachable.\n\n" +
		"Also prints statistics about how many requests the server handled.",
	Run: func(cmd *cobra.Command, args []string) {
		r := request.CheckTholaServerRequest{
			CheckRequest: getCheckRequest(),
		}
		handleRequest(&r)
	},
}
