package deviceclass

import (
	"context"
	"fmt"
	"github.com/inexio/thola/internal/communicator/filter"
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

type groupPropertyReader interface {
	getProperty(ctx context.Context, filter ...filter.PropertyFilter) (propertyGroups, []value.Value, error)
}

type snmpGroupPropertyReader struct {
	index oidReader
	oids  deviceClassOIDs
}

func (s *snmpGroupPropertyReader) getProperty(ctx context.Context, filter ...filter.PropertyFilter) (propertyGroups, []value.Value, error) {
	wantedIndices, filteredIndices, err := s.getFilteredIndices(ctx, filter...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to filter oid indices")
	}

	useSNMPGetsInsteadOfWalk, ok := network.SNMPGetsInsteadOfWalkFromContext(ctx)
	if !ok {
		log.Ctx(ctx).Debug().Msg("SNMPGetsInsteadOfWalk not found in context, using walks")
	}

	if !useSNMPGetsInsteadOfWalk {
		wantedIndices = nil
	}

	groups, err := s.oids.readOID(ctx, wantedIndices, true)
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

		delete(groups, smallestIndex)
		if !useSNMPGetsInsteadOfWalk {
			if _, ok := filteredIndices[strconv.Itoa(smallestIndex)]; ok {
				continue
			}
		}
		res = append(res, x)
		indices = append(indices, value.New(smallestIndex))
	}

	return res, indices, nil
}

func (s *snmpGroupPropertyReader) getFilteredIndices(ctx context.Context, filter ...filter.PropertyFilter) ([]value.Value, map[string]struct{}, error) {
	indices := make(map[string]struct{})
	filteredIndices := make(map[string]struct{})

	if s.index != nil {
		res, err := s.index.readOID(ctx, nil, false)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to read out index oid")
		}
		for idx := range res {
			indices[strconv.Itoa(idx)] = struct{}{}
		}
	}

	for _, f := range filter {
		// compile filter regex
		regex, err := regexp.Compile(f.Regex)
		if err != nil {
			return nil, nil, errors.Wrap(err, "filter regex failed to compile")
		}

		// find filter oid
		attrs := strings.Split(f.Key, "/")
		reader := oidReader(&s.oids)
		for _, attr := range attrs {
			// check if current oid reader contains multiple OIDs
			multipleReader, ok := reader.(*deviceClassOIDs)
			if !ok || multipleReader == nil {
				return nil, nil, errors.New("filter attribute does not exist")
			}

			// check if oid reader contains OID(s) for the current attribute name
			if reader, ok = (*multipleReader)[attr]; !ok {
				return nil, nil, errors.New("filter attribute does not exist")
			}
		}

		// check if the current oid reader contains only a single oid
		singleReader, ok := reader.(*deviceClassOID)
		if !ok || singleReader == nil {
			return nil, nil, errors.New("filter attribute does not exist")
		}

		results, err := singleReader.readOID(ctx, nil, false)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to read out filter oid")
		}

		for index, result := range results {
			// add to indices map
			indices[strconv.Itoa(index)] = struct{}{}
			if regex.MatchString(result.(value.Value).String()) {
				// if filter matches add to filtered indices map
				filteredIndices[strconv.Itoa(index)] = struct{}{}
				log.Ctx(ctx).Debug().Str("filter_key", f.Key).Str("filter_regex", f.Regex).
					Str("received_value", result.(value.Value).String()).
					Msgf("filter matched on index '%d'", index)
			}
		}
	}

	var res []value.Value
	for index := range indices {
		if _, ok := filteredIndices[index]; !ok {
			res = append(res, value.New(index))
		}
	}

	return res, filteredIndices, nil
}

type oidReader interface {
	readOID(context.Context, []value.Value, bool) (map[int]interface{}, error)
}

// deviceClassOIDs is a recursive data structure which maps labels to either a single OID (deviceClassOID) or another deviceClassOIDs
type deviceClassOIDs map[string]oidReader

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
	operators      propertyOperators
	indicesMapping *deviceClassOID
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
	if indices != nil {
		log.Ctx(ctx).Debug().Msg("indices given, using SNMP Gets instead of Walk")

		//change requested indices if necessary
		if d.indicesMapping != nil {
			mappingIndices, err := d.indicesMapping.readOID(ctx, nil, true)
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
		logger := log.Ctx(ctx).With().Str("oid", response.GetOID()).Logger()
		ctx = logger.WithContext(ctx)

		res, err := response.GetValueBySNMPGetConfiguration(d.SNMPGetConfiguration)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("couldn't get value from response")
			continue
		}
		if res != "" || !skipEmpty {
			resNormalized, err := d.operators.apply(ctx, value.New(res))
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

func (n *emptyOIDReader) readOID(context.Context, []value.Value, bool) (map[int]interface{}, error) {
	return nil, tholaerr.NewComponentNotFoundError("oid is ignored")
}
