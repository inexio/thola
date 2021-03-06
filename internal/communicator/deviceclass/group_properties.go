package deviceclass

import (
	"context"
	"fmt"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/inexio/thola/internal/value"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
)

type propertyGroup map[string]interface{}

type propertyGroups []propertyGroup

func (g *propertyGroups) Decode(destination interface{}) error {
	return mapstructure.WeakDecode(g, destination)
}

type groupPropertyReader interface {
	getProperty(ctx context.Context) (propertyGroups, []value.Value, error)
}

type snmpGroupPropertyReader struct {
	oids deviceClassOIDs
}

func (s *snmpGroupPropertyReader) getProperty(ctx context.Context) (propertyGroups, []value.Value, error) {
	groups, err := s.oids.readOID(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to read oids")
	}

	var res propertyGroups
	var indices []value.Value

	// this sorts the groups after their index
	//TODO efficiency
	size := len(groups)
	for i := 0; i < size; i++ {
		var smallestIndex int
		firstRun := true
		for index := range groups {
			if firstRun {
				smallestIndex = index
				firstRun = false
			}
			if index < smallestIndex {
				smallestIndex = index
			}
		}
		x, ok := groups[smallestIndex].(map[string]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("oidReader for index '%d' returned unexpected data type: %T", smallestIndex, groups[smallestIndex])
		}

		res = append(res, x)
		indices = append(indices, value.New(smallestIndex))
		delete(groups, smallestIndex)
	}

	return res, indices, nil
}

type oidReader interface {
	readOID(context.Context) (map[int]interface{}, error)
}

// deviceClassOIDs is a recursive data structure which maps labels to either a single OID (deviceClassOID) or another deviceClassOIDs
type deviceClassOIDs map[string]oidReader

func (d *deviceClassOIDs) readOID(ctx context.Context) (map[int]interface{}, error) {
	result := make(map[int]map[string]interface{})
	for label, reader := range *d {
		res, err := reader.readOID(ctx)
		if err != nil {
			if tholaerr.IsNotFoundError(err) || tholaerr.IsComponentNotFoundError(err) {
				log.Ctx(ctx).Trace().Err(err).Msgf("value %s", label)
				continue
			}
			return nil, errors.Wrapf(err, "failed to get value '%s'", label)
		}
		for ifIndex, v := range res {
			// ifIndex was not known before, so create a new group
			if _, ok := result[ifIndex]; !ok {
				result[ifIndex] = make(map[string]interface{})
			}
			result[ifIndex][label] = v
		}
	}

	r := make(map[int]interface{})
	for k, v := range result {
		r[k] = v
	}

	return r, nil
}

func (d *deviceClassOIDs) merge(overwrite deviceClassOIDs) deviceClassOIDs {
	devClassOIDsNew := make(deviceClassOIDs)
	for k, v := range *d {
		devClassOIDsNew[k] = v
	}
	for k, v := range overwrite {
		if reader, ok := devClassOIDsNew[k]; ok {
			oidsOld, oldIsOIDs := reader.(*deviceClassOIDs)
			oidsOverwrite, overwriteIsOIDs := v.(*deviceClassOIDs)
			if oldIsOIDs && overwriteIsOIDs {
				mergedOIDs := oidsOld.merge(*oidsOverwrite)
				devClassOIDsNew[k] = &mergedOIDs
				continue
			}
		}
		devClassOIDsNew[k] = v
	}

	return devClassOIDsNew
}

// deviceClassOID represents a single OID which can be read
type deviceClassOID struct {
	network.SNMPGetConfiguration
	operators      propertyOperators
	indicesMapping *deviceClassOID
}

func (d *deviceClassOID) readOID(ctx context.Context) (map[int]interface{}, error) {
	result := make(map[int]interface{})

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		log.Ctx(ctx).Trace().Str("property", "interface").Msg("snmp client is empty")
		return nil, errors.New("snmp client is empty")
	}

	snmpResponse, err := con.SNMP.SnmpClient.SNMPWalk(ctx, string(d.OID))
	if err != nil {
		if tholaerr.IsNotFoundError(err) {
			log.Ctx(ctx).Trace().Err(err).Msgf("oid %s not found on device", d.OID)
			return nil, err
		}
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get oid value of interface")
		return nil, errors.Wrap(err, "failed to get oid value")
	}

	for _, response := range snmpResponse {
		res, err := response.GetValueBySNMPGetConfiguration(d.SNMPGetConfiguration)
		if err != nil {
			log.Ctx(ctx).Trace().Err(err).Msg("couldn't get value from response response")
			return nil, errors.Wrap(err, "couldn't get value from response response")
		}
		if res != "" {
			resNormalized, err := d.operators.apply(ctx, value.New(res))
			if err != nil {
				if tholaerr.IsDidNotMatchError(err) {
					continue
				}
				log.Ctx(ctx).Trace().Err(err).Msgf("response couldn't be normalized (oid: %s, response: %s)", response.GetOID(), res)
				return nil, errors.Wrapf(err, "response couldn't be normalized (oid: %s, response: %s)", response.GetOID(), res)
			}
			oid := strings.Split(response.GetOID(), ".")
			index, err := strconv.Atoi(oid[len(oid)-1])
			if err != nil {
				log.Ctx(ctx).Trace().Err(err).Msg("index isn't an integer")
				return nil, errors.Wrap(err, "index isn't an integer")
			}
			result[index] = resNormalized
		}
	}

	//change indices if necessary
	if d.indicesMapping != nil {
		indices, err := d.indicesMapping.readOID(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read indices")
		}
		mappedResult := make(map[int]interface{})

		for k, v := range result {
			var idx int
			if _, ok := indices[k]; ok {
				idx, err = indices[k].(value.Value).Int()
				if err != nil {
					return nil, errors.Wrap(err, "failed to convert Value to int")
				}
			} else {
				idx = k
			}

			if _, ok := mappedResult[idx]; ok {
				return nil, fmt.Errorf("index mappings resulted in duplicated index '%d'", idx)
			}

			mappedResult[idx] = v
		}
		result = mappedResult
	}
	return result, nil
}

type emptyOIDReader struct{}

func (n *emptyOIDReader) readOID(context.Context) (map[int]interface{}, error) {
	return nil, tholaerr.NewComponentNotFoundError("oid is ignored")
}
