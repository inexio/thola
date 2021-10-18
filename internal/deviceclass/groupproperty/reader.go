package groupproperty

import (
	"context"
	"fmt"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/value"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func Interface2Reader(i interface{}, parentReader Reader) (Reader, error) {
	m, ok := i.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("failed to convert group properties to map[interface{}]interface{}")
	}
	if _, ok := m["detection"]; !ok {
		return nil, errors.New("detection is missing in group properties")
	}
	stringDetection, ok := m["detection"].(string)
	if !ok {
		return nil, errors.New("property detection needs to be a string")
	}
	switch stringDetection {
	case "snmpwalk":
		var index OIDReader
		if idx, ok := m["index"]; ok {
			idxString, ok := idx.(string)
			if !ok {
				return nil, errors.New("index needs to be string (oid)")
			}
			oid := network.OID(idxString)
			if err := oid.Validate(); err != nil {
				return nil, errors.Wrap(err, "index needs to be an oid")
			}
			devClassOid := deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: oid,
				},
			}
			index = &devClassOid
		}

		if _, ok := m["values"]; !ok {
			return nil, errors.New("values are missing")
		}
		reader, err := Interface2OIDReader(m["values"])
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse oid reader")
		}

		devClassOIDs, ok := reader.(*deviceClassOIDs)
		if !ok {
			return nil, errors.New("oid reader is no list of oids")
		}

		inheritValuesFromParent := true
		if b, ok := m["inherit_values"]; ok {
			bb, ok := b.(bool)
			if !ok {
				return nil, errors.New("inherit_values needs to be a boolean")
			}
			inheritValuesFromParent = bb
		}

		//overwrite parent
		if inheritValuesFromParent && parentReader != nil {
			parentBaseReader, ok := parentReader.(*baseReader)
			if !ok {
				return nil, errors.New("parent group property reader is not of type base group property reader")
			}

			parentSNMPReader, ok := parentBaseReader.reader.(*snmpReader)
			if !ok {
				return nil, errors.New("can't merge SNMP group property reader with property reader of different type")
			}

			parentSNMPReaderOIDs, ok := parentSNMPReader.oids.(*deviceClassOIDs)
			if !ok {
				return nil, errors.New("parent SNMP group property reader oids is not of type 'deviceClassOIDs'")
			}

			devClassOIDsMerged := parentSNMPReaderOIDs.merge(*devClassOIDs)
			devClassOIDs = &devClassOIDsMerged

			if index == nil {
				index = parentSNMPReader.index
			}
		}

		return &baseReader{
			reader: &snmpReader{
				index: index,
				oids:  devClassOIDs,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown detection type '%s'", stringDetection)
	}
}

type propertyGroup map[string]interface{}

func (g *propertyGroup) decode(destination interface{}) error {
	return mapstructure.WeakDecode(g, destination)
}

func (g *propertyGroup) encode(data interface{}) error {
	return mapstructure.WeakDecode(data, g)
}

type PropertyGroups []propertyGroup

func (g *PropertyGroups) Decode(destination interface{}) error {
	return mapstructure.WeakDecode(g, destination)
}

func (g *PropertyGroups) Encode(data interface{}) error {
	return mapstructure.WeakDecode(data, g)
}

type Reader interface {
	GetProperty(ctx context.Context, filter ...Filter) (PropertyGroups, []value.Value, error)
}

type baseReader struct {
	reader reader
}

func (b baseReader) GetProperty(ctx context.Context, filter ...Filter) (PropertyGroups, []value.Value, error) {
	var r = b.reader
	var err error
	for _, fil := range filter {
		r, err = r.applyFilter(ctx, fil)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to apply filter")
		}
	}
	return r.getProperty(ctx)
}

type reader interface {
	getProperty(ctx context.Context) (PropertyGroups, []value.Value, error)
	applyFilter(ctx context.Context, filter Filter) (reader, error)
}

type snmpReader struct {
	index           OIDReader
	wantedIndices   map[string]struct{}
	filteredIndices map[string]struct{}
	oids            OIDReader
}

func (s snmpReader) getProperty(ctx context.Context) (PropertyGroups, []value.Value, error) {
	var wantedIndices []string

	useSNMPGetsInsteadOfWalk, ok := network.SNMPGetsInsteadOfWalkFromContext(ctx)
	if !ok {
		log.Ctx(ctx).Debug().Msg("SNMPGetsInsteadOfWalk not found in context, using walks")
	}

	if useSNMPGetsInsteadOfWalk {
		var indices map[string]struct{}
		if len(s.wantedIndices) > 0 {
			indices = s.wantedIndices
		} else {
			var err error
			indices, err = s.getIndices(ctx)
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to get indices")
			}
		}

		for index := range indices {
			wantedIndices = append(wantedIndices, index)
		}
	}

	groups, err := s.oids.readOID(ctx, wantedIndices, true)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to read oids")
	}

	var res PropertyGroups
	var indices []value.Value

	// this sorts the groups after their index
	//TODO efficiency
	size := len(groups)
	for i := 0; i < size; i++ {
		var smallestIndex string
		firstRun := true
		for index := range groups {
			if firstRun {
				smallestIndex = index
				firstRun = false
				continue
			}
			cmp, err := network.OID(index).Cmp(network.OID(smallestIndex))
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to compare indices")
			}
			if cmp == -1 {
				smallestIndex = index
			}
		}
		x, ok := groups[smallestIndex].(map[string]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("oidReader for index '%s' returned unexpected data type: %T", smallestIndex, groups[smallestIndex])
		}

		delete(groups, smallestIndex)
		if !useSNMPGetsInsteadOfWalk {
			if _, ok := s.filteredIndices[smallestIndex]; ok {
				continue
			}
		}
		res = append(res, x)
		indices = append(indices, value.New(smallestIndex))
	}

	return res, indices, nil
}

func (s snmpReader) applyFilter(ctx context.Context, filter Filter) (reader, error) {
	return filter.applySNMP(ctx, s)
}

func (s snmpReader) getIndices(ctx context.Context) (map[string]struct{}, error) {
	if s.index == nil {
		return nil, errors.New("indices reader is empty")
	}

	indices, err := s.index.readOID(ctx, nil, false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read indices")
	}
	res := make(map[string]struct{})
	for index := range indices {
		res[index] = struct{}{}
	}

	return res, nil
}
