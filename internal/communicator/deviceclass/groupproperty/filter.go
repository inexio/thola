package groupproperty

import (
	"context"
	"github.com/inexio/thola/internal/value"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"regexp"
	"strconv"
	"strings"
)

type Filter interface {
	applySNMP(ctx context.Context, reader snmpReader) (snmpReader, error)
}

type GroupFilter interface {
	GetFilterProperties() (string, string)
}

type groupFilter struct {
	key   string
	regex string
}

func GetGroupFilter(key, regex string) Filter {
	return &groupFilter{
		key:   key,
		regex: regex,
	}
}

func (g *groupFilter) GetFilterProperties() (string, string) {
	return g.key, g.regex
}

func (g *groupFilter) applySNMP(ctx context.Context, reader snmpReader) (snmpReader, error) {
	if len(reader.wantedIndices) == 0 {
		var err error
		reader.wantedIndices, err = reader.getIndices(ctx)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("failed to read indices, ignoring index oid")

		}
	}
	if reader.wantedIndices == nil {
		reader.wantedIndices = make(map[string]struct{})
	}
	if reader.filteredIndices == nil {
		reader.filteredIndices = make(map[string]struct{})
	}

	// compile filter regex
	regex, err := regexp.Compile(g.regex)
	if err != nil {
		return snmpReader{}, errors.Wrap(err, "filter regex failed to compile")
	}

	// find filter oid
	attrs := strings.Split(g.key, "/")
	oidReader := reader.oids
	for _, attr := range attrs {
		// check if current oid reader contains multiple OIDs
		multipleReader, ok := oidReader.(*deviceClassOIDs)
		if !ok || multipleReader == nil {
			return snmpReader{}, errors.New("filter attribute does not exist")
		}

		// check if oid reader contains OID(s) for the current attribute name
		if oidReader, ok = (*multipleReader)[attr]; !ok {
			return snmpReader{}, errors.New("filter attribute does not exist")
		}
	}

	// check if the current oid reader contains only a single oid
	singleReader, ok := oidReader.(*deviceClassOID)
	if !ok || singleReader == nil {
		return snmpReader{}, errors.New("filter attribute does not exist")
	}

	results, err := singleReader.readOID(ctx, nil, false)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Str("oid", string(singleReader.OID)).Msg("failed to read out filter oid, skipping filter")
		return reader, nil
	}

	for index, result := range results {
		indexString := strconv.Itoa(index)
		if regex.MatchString(result.(value.Value).String()) {
			// if filter matches add to filtered indices map and delete from wanted indices
			reader.filteredIndices[indexString] = struct{}{}
			delete(reader.wantedIndices, indexString)
			log.Ctx(ctx).Debug().Str("filter_key", g.key).Str("filter_regex", g.regex).
				Str("received_value", result.(value.Value).String()).
				Msgf("filter matched on index '%d'", index)
		} else {
			// if filter does not match check if index was filtered before
			if _, ok := reader.filteredIndices[indexString]; !ok {
				// if not add it to wanted indices map
				reader.wantedIndices[indexString] = struct{}{}
			}
		}
	}

	return reader, nil
}
