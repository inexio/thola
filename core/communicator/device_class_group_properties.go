package communicator

import (
	"context"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/inexio/thola/core/value"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
)

type groupPropertyReader interface {
	getProperty(ctx context.Context) ([]map[string]value.Value, error)
}

type snmpGroupPropertyReader struct {
	oids deviceClassOIDs
}

func (s *snmpGroupPropertyReader) getProperty(ctx context.Context) ([]map[string]value.Value, error) {
	networkInterfaces := make(map[int]map[string]value.Value)

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		log.Ctx(ctx).Trace().Str("property", "interface").Msg("snmp client is empty")
		return nil, errors.New("snmp client is empty")
	}

	for name, oid := range s.oids {
		snmpResponse, err := con.SNMP.SnmpClient.SNMPWalk(ctx, string(oid.OID))
		if err != nil {
			if tholaerr.IsNotFoundError(err) {
				log.Ctx(ctx).Trace().Err(err).Msgf("oid %s (%s) not found on device", oid.OID, name)
				continue
			}
			log.Ctx(ctx).Trace().Err(err).Msg("failed to get oid value of interface")
			return nil, errors.Wrap(err, "failed to get oid value")
		}

		for _, response := range snmpResponse {
			res, err := response.GetValueBySNMPGetConfiguration(oid.SNMPGetConfiguration)
			if err != nil {
				log.Ctx(ctx).Trace().Err(err).Msg("couldn't get value from response response")
				return nil, errors.Wrap(err, "couldn't get value from response response")
			}
			if res != "" {
				resNormalized, err := oid.operators.apply(ctx, value.New(res))
				if err != nil {
					log.Ctx(ctx).Trace().Err(err).Msgf("response couldn't be normalized (oid: %s, response: %s)", response.GetOID(), res)
					return nil, errors.Wrapf(err, "response couldn't be normalized (oid: %s, response: %s)", response.GetOID(), res)
				}
				oid := strings.Split(response.GetOID(), ".")
				index, err := strconv.Atoi(oid[len(oid)-1])
				if err != nil {
					log.Ctx(ctx).Trace().Err(err).Msg("index isn't an integer")
					return nil, errors.Wrap(err, "index isn't an integer")
				}
				if _, ok := networkInterfaces[index]; !ok {
					networkInterfaces[index] = make(map[string]value.Value)
				}
				networkInterfaces[index][name] = resNormalized
			}
		}
	}

	var res []map[string]value.Value

	//TODO efficiency
	size := len(networkInterfaces)
	for i := 0; i < size; i++ {
		var smallestIndex int
		firstRun := true
		for index := range networkInterfaces {
			if firstRun {
				smallestIndex = index
				firstRun = false
			}
			if index < smallestIndex {
				smallestIndex = index
			}
		}
		res = append(res, networkInterfaces[smallestIndex])
		delete(networkInterfaces, smallestIndex)
	}

	return res, nil
}
