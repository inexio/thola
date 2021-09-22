package groupproperty

import (
	"context"
	"fmt"
	relatedTask "github.com/inexio/thola/internal/communicator/deviceclass/condition"
	"github.com/inexio/thola/internal/communicator/deviceclass/property"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/inexio/thola/internal/value"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
)

//go:generate go run github.com/vektra/mockery/v2 --name=OIDReader --inpackage

func Interface2OIDReader(i interface{}) (OIDReader, error) {
	values, ok := i.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("values needs to be a map")
	}

	result := make(deviceClassOIDs)

	for val, data := range values {
		dataMap, ok := data.(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("value data needs to be a map")
		}

		valString, ok := val.(string)
		if !ok {
			return nil, errors.New("key of snmp property reader must be a string")
		}

		if v, ok := dataMap["values"]; ok {
			if len(dataMap) != 1 {
				return nil, errors.New("value with subvalues has to many keys")
			}
			reader, err := Interface2OIDReader(v)
			if err != nil {
				return nil, err
			}
			result[valString] = reader
			continue
		}

		if ignore, ok := dataMap["ignore"]; ok {
			if b, ok := ignore.(bool); ok && b {
				result[valString] = &emptyOIDReader{}
				continue
			}
		}

		var oid yamlComponentsOID
		err := mapstructure.Decode(data, &oid)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode values map to yamlComponentsOIDs")
		}
		err = oid.validate()
		if err != nil {
			return nil, errors.Wrapf(err, "oid reader for %s is invalid", valString)
		}
		devClassOID, err := oid.convert()
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert yaml OID to device class OID")
		}
		result[valString] = &devClassOID
	}
	return &result, nil
}

type OIDReader interface {
	readOID(context.Context, []value.Value, bool) (map[int]interface{}, error)
}

// deviceClassOIDs is a recursive data structure which maps labels to either a single OID (deviceClassOID) or another deviceClassOIDs
type deviceClassOIDs map[string]OIDReader

func (d *deviceClassOIDs) readOID(ctx context.Context, indices []value.Value, skipEmpty bool) (map[int]interface{}, error) {
	result := make(map[int]map[string]interface{})
	for label, reader := range *d {
		res, err := reader.readOID(ctx, indices, skipEmpty)
		if err != nil {
			if tholaerr.IsNotFoundError(err) || tholaerr.IsComponentNotFoundError(err) {
				log.Ctx(ctx).Debug().Err(err).Msgf("failed to get value '%s'", label)
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
	operators      property.Operators
	indicesMapping OIDReader
}

func (d *deviceClassOID) readOID(ctx context.Context, indices []value.Value, skipEmpty bool) (map[int]interface{}, error) {
	result := make(map[int]interface{})

	logger := log.Ctx(ctx).With().Str("oid", string(d.OID)).Logger()
	ctx = logger.WithContext(ctx)

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		log.Ctx(ctx).Debug().Msg("snmp client is empty")
		return nil, errors.New("snmp client is empty")
	}

	var snmpResponse []network.SNMPResponse
	var err error
	if len(indices) > 0 {
		log.Ctx(ctx).Debug().Msg("indices given, using SNMP Gets instead of Walk")

		//change requested indices if necessary
		if d.indicesMapping != nil {
			mappingIndices, err := d.indicesMapping.readOID(ctx, nil, true)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read indices")
			}

			ifIndexRelIndex := make(map[string]value.Value)
			for relIndex, ifIndex := range mappingIndices {
				ifIndexValue, ok := ifIndex.(value.Value)
				if !ok {
					return nil, errors.New("index mapping oid didn't return a result of type 'value'")
				}
				ifIndexString := ifIndexValue.String()
				if idx, ok := ifIndexRelIndex[ifIndexString]; ok {
					return nil, fmt.Errorf("index mapping resulted in duplicate ifIndex mapping on '%s'", idx.String())
				}
				ifIndexRelIndex[ifIndexString] = value.New(relIndex)
			}

			var newIndices []value.Value
			for _, ifIndex := range indices {
				if relIndex, ok := ifIndexRelIndex[ifIndex.String()]; ok {
					newIndices = append(newIndices, relIndex)
				}
			}

			indices = newIndices
		}

		oid := string(d.OID)
		if !strings.HasSuffix(oid, ".") {
			oid += "."
		}
		var oids []string
		for _, index := range indices {
			oids = append(oids, oid+index.String())
		}
		snmpResponse, err = con.SNMP.SnmpClient.SNMPGet(ctx, oids...)
	} else {
		snmpResponse, err = con.SNMP.SnmpClient.SNMPWalk(ctx, string(d.OID))
	}
	if err != nil {
		if tholaerr.IsNotFoundError(err) {
			return nil, err
		}
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get oid value of interface")
		return nil, errors.Wrap(err, "failed to get oid value")
	}

	for _, response := range snmpResponse {
		logger := log.Ctx(ctx).With().Str("oid", response.GetOID()).Logger()
		ctx = logger.WithContext(ctx)

		res, err := response.GetValueBySNMPGetConfiguration(d.SNMPGetConfiguration)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("couldn't get value from response")
			continue
		}
		if res != "" || !skipEmpty {
			resNormalized, err := d.operators.Apply(ctx, value.New(res))
			if err != nil {
				if tholaerr.IsDidNotMatchError(err) {
					continue
				}
				log.Ctx(ctx).Debug().Err(err).Msgf("response couldn't be normalized (response: %s)", res)
				return nil, errors.Wrapf(err, "response couldn't be normalized (response: %s)", res)
			}
			oid := strings.Split(response.GetOID(), ".")
			index, err := strconv.Atoi(oid[len(oid)-1])
			if err != nil {
				log.Ctx(ctx).Debug().Err(err).Msg("index isn't an integer")
				return nil, errors.Wrap(err, "index isn't an integer")
			}
			result[index] = resNormalized
		}
	}

	//change indices if necessary
	if d.indicesMapping != nil {
		mappingIndices, err := d.indicesMapping.readOID(ctx, nil, true)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read mapping indices")
		}
		mappedResult := make(map[int]interface{})

		for k, v := range result {
			mappedIdx, ok := mappingIndices[k]
			if !ok {
				continue
			}
			idxValue, ok := mappedIdx.(value.Value)
			if !ok {
				return nil, errors.New("index mapping oid didn't return a result of type 'value'")
			}
			idx, err := idxValue.Int()
			if err != nil {
				return nil, errors.Wrap(err, "failed to convert Value to int")
			}

			if _, ok := mappedResult[idx]; ok {
				return nil, fmt.Errorf("index mapping resulted in duplicate index '%d'", idx)
			}

			mappedResult[idx] = v
		}
		result = mappedResult
	}
	return result, nil
}

type emptyOIDReader struct{}

func (n *emptyOIDReader) readOID(context.Context, []value.Value, bool) (map[int]interface{}, error) {
	return nil, tholaerr.NewComponentNotFoundError("oid is ignored")
}

type yamlComponentsOID struct {
	network.SNMPGetConfiguration `mapstructure:",squash"`
	Operators                    []interface{}
	IndicesMapping               *yamlComponentsOID `mapstructure:"indices_mapping"`
}

func (y *yamlComponentsOID) convert() (deviceClassOID, error) {
	res := deviceClassOID{
		SNMPGetConfiguration: network.SNMPGetConfiguration{
			OID:          y.OID,
			UseRawResult: y.UseRawResult,
		},
	}

	if y.IndicesMapping != nil {
		mappings, err := y.IndicesMapping.convert()
		if err != nil {
			return deviceClassOID{}, errors.New("failed to convert indices mappings")
		}
		res.indicesMapping = &mappings
	}

	if y.Operators != nil {
		operators, err := property.InterfaceSlice2Operators(y.Operators, relatedTask.PropertyDefault)
		if err != nil {
			return deviceClassOID{}, errors.Wrap(err, "failed to read yaml oids operators")
		}
		res.operators = operators
	}

	return res, nil
}

func (y *yamlComponentsOID) validate() error {
	if err := y.OID.Validate(); err != nil {
		return errors.Wrap(err, "oid is invalid")
	}
	return nil
}
