// +build client

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/inexio/thola/core/request"
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
