// +build !client

package request

import (
	"context"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/core/network"
)

func (r *CheckSNMPRequest) process(ctx context.Context) (Response, error) {
	r.init()
	r.mon.SetOutputDelimiter(" - ")
	var res CheckSNMPResponse
	con, err := r.setupSNMPConnection(ctx)
	if !r.mon.UpdateStatusOnError(err, monitoringplugin.CRITICAL, "failed to create snmp connection", false) {
		res.SuccessfulSnmpCredentials = &network.SNMPCredentials{
			Version:   con.SnmpClient.GetVersion(),
			Community: con.SnmpClient.GetCommunity(),
			Port:      con.SnmpClient.GetPort(),
		}
		res.SuccessfulSnmpCredentials.Version = con.SnmpClient.GetVersion()
		r.mon.UpdateStatus(monitoringplugin.OK, fmt.Sprintf("version: '%s'; community: '%s'; port: '%d'", res.SuccessfulSnmpCredentials.Version, res.SuccessfulSnmpCredentials.Community, res.SuccessfulSnmpCredentials.Port))
		res.SuccessfulSnmpCredentials.Version = con.SnmpClient.GetVersion()
	}
	res.CheckResponse = CheckResponse{r.mon.GetInfo()}
	return &res, nil
}
