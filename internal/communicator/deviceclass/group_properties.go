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
	"regexp"
	"strconv"
	"strings"
)

type propertyGroup map[string]interface{}

type propertyGroups []propertyGroup

func (g *propertyGroups) Decode(destination interface{}) error {
	return mapstructure.WeakDecode(g, destination)
}

type groupPropertyFilter struct {
	key   string
	regex string
}

type groupPropertyReader interface {
	getProperty(ctx context.Context, filter ...groupPropertyFilter) (propertyGroups, []value.Value, error)
}

type snmpGroupPropertyReader struct {
	oids deviceClassOIDs
}

func (s *snmpGroupPropertyReader) getProperty(ctx context.Context, filter ...groupPropertyFilter) (propertyGroups, []value.Value, error) {
	filteredIndices, err := s.getFilteredIndices(ctx, filter...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to filter oid indices")
	}

	groups, err := s.oids.readOID(ctx, filteredIndices)
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

func (s *snmpGroupPropertyReader) getFilteredIndices(ctx context.Context, filter ...groupPropertyFilter) ([]value.Value, error) {
	filteredIndicesMap := make(map[value.Value]struct{})
	var filteredIndices []value.Value

	for _, f := range filter {
		// compile filter regex
		regex, err := regexp.Compile(f.regex)
		if err != nil {
			return nil, errors.Wrap(err, "filter regex ")
		}

		// find filter oid
		attrs := strings.Split(f.key, "/")
		reader := oidReader(&s.oids)
		for _, attr := range attrs {
			// check if current oid reader contains multiple OIDs
			multipleReader, ok := reader.(*deviceClassOIDs)
			if !ok || multipleReader == nil {
				return nil, errors.New("filter attribute does not exist")
			}

			// check if oid reader contains OID(s) for the current attribute name
			if reader, ok = (*multipleReader)[attr]; !ok {
				return nil, errors.New("filter attribute does not exist")
			}
		}

		// check if the current oid reader contains only a single oid
		singleReader, ok := reader.(*deviceClassOID)
		if !ok || singleReader == nil {
			return nil, errors.New("filter attribute does not exist")
		}

		results, err := singleReader.readOID(ctx, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read out filter oid")
		}

		for index, result := range results {
			if regex.MatchString(result.(value.Value).String()) {
				// filter matches
				delete(filteredIndicesMap, value.New(index))
			} else {
				filteredIndicesMap[value.New(index)] = struct{}{}
			}
		}
	}

	for index := range filteredIndicesMap {
		filteredIndices = append(filteredIndices, index)
	}

	return filteredIndices, nil
}

type oidReader interface {
	readOID(context.Context, []value.Value) (map[int]interface{}, error)
}

// deviceClassOIDs is a recursive data structure which maps labels to either a single OID (deviceClassOID) or another deviceClassOIDs
type deviceClassOIDs map[string]oidReader

func (d *deviceClassOIDs) readOID(ctx context.Context, indices []value.Value) (map[int]interface{}, error) {
	result := make(map[int]map[string]interface{})
	for label, reader := range *d {
		res, err := reader.readOID(ctx, indices)
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
	operators      propertyOperators
	indicesMapping *deviceClassOID
}

func (d *deviceClassOID) readOID(ctx context.Context, indices []value.Value) (map[int]interface{}, error) {
	result := make(map[int]interface{})

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		log.Ctx(ctx).Debug().Str("property", "interface").Msg("snmp client is empty")
		return nil, errors.New("snmp client is empty")
	}

	var snmpResponse []network.SNMPResponse
	var err error
	if indices != nil {
		//change requested indices if necessary
		if d.indicesMapping != nil {
			mappingIndices, err := d.indicesMapping.readOID(ctx, nil)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read indices")
			}

			ifIndexRelIndex := make(map[string]value.Value)
			for relIndex, ifIndex := range mappingIndices {
				if idx, ok := ifIndexRelIndex[ifIndex.(value.Value).String()]; ok {
					return nil, fmt.Errorf("index mapping resulted in duplicate ifIndex mapping on '%s'", idx.String())
				}
				ifIndexRelIndex[ifIndex.(value.Value).String()] = value.New(relIndex)
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
		res, err := response.GetValueBySNMPGetConfiguration(d.SNMPGetConfiguration)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("couldn't get value from response")
			continue
		}
		if res != "" {
			resNormalized, err := d.operators.apply(ctx, value.New(res))
			if err != nil {
				if tholaerr.IsDidNotMatchError(err) {
					continue
				}
				log.Ctx(ctx).Debug().Err(err).Msgf("response couldn't be normalized (oid: %s, response: %s)", response.GetOID(), res)
				return nil, errors.Wrapf(err, "response couldn't be normalized (oid: %s, response: %s)", response.GetOID(), res)
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
		mappingIndices, err := d.indicesMapping.readOID(ctx, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read mapping indices")
		}
		mappedResult := make(map[int]interface{})

		for k, v := range result {
			var idx int
			if _, ok := mappingIndices[k]; ok {
				idx, err = mappingIndices[k].(value.Value).Int()
				if err != nil {
					return nil, errors.Wrap(err, "failed to convert Value to int")
				}
			} else {
				idx = k
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

func (n *emptyOIDReader) readOID(context.Context, []value.Value) (map[int]interface{}, error) {
	return nil, tholaerr.NewComponentNotFoundError("oid is ignored")
}
