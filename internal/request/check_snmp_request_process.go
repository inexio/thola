//go:build !client
// +build !client

package request

import (
	"context"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/utility"
)

func (r *CheckSNMPRequest) process(ctx context.Context) (Response, error) {
	r.init()
	r.mon.SetOutputDelimiter(" - ")
	var res CheckSNMPResponse
	con, err := r.setupSNMPConnection(ctx)
	if !r.mon.UpdateStatusOnError(err, monitoringplugin.CRITICAL, "failed to create snmp connection", false) {
		version := con.SnmpClient.GetVersion()
		if version == "3" {
			res.SuccessfulSnmpCredentials = &network.SNMPCredentials{
				Version:       version,
				Port:          con.SnmpClient.GetPort(),
				V3Level:       utility.IfThenElse(con.SnmpClient.GetV3Level() == nil, "", *con.SnmpClient.GetV3Level()).(string),
				V3ContextName: utility.IfThenElse(con.SnmpClient.GetV3ContextName() == nil, "", *con.SnmpClient.GetV3ContextName()).(string),
			}
			r.mon.UpdateStatus(monitoringplugin.OK, fmt.Sprintf("version: '%s'; port: '%d'; level: '%s'; context_name: '%s'", res.SuccessfulSnmpCredentials.Version, res.SuccessfulSnmpCredentials.Port, res.SuccessfulSnmpCredentials.V3Level, res.SuccessfulSnmpCredentials.V3ContextName))
		} else {
			res.SuccessfulSnmpCredentials = &network.SNMPCredentials{
				Version:   con.SnmpClient.GetVersion(),
				Community: con.SnmpClient.GetCommunity(),
				Port:      con.SnmpClient.GetPort(),
			}
			r.mon.UpdateStatus(monitoringplugin.OK, fmt.Sprintf("version: '%s'; community: '%s'; port: '%d'", res.SuccessfulSnmpCredentials.Version, res.SuccessfulSnmpCredentials.Community, res.SuccessfulSnmpCredentials.Port))
		}
	}
	res.CheckResponse = CheckResponse{r.mon.GetInfo()}
	return &res, nil
}
