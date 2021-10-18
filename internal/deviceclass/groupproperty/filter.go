package groupproperty

import (
	"context"
	"github.com/inexio/thola/internal/value"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"regexp"
	"strconv"
)

type Filter interface {
	ApplyPropertyGroups(context.Context, PropertyGroups) (PropertyGroups, error)

	applySNMP(context.Context, snmpReader) (snmpReader, error)
}

type groupFilter struct {
	key   []string
	regex string
}

func GetGroupFilter(key []string, regex string) Filter {
	return &groupFilter{
		key:   key,
		regex: regex,
	}
}

func (g *groupFilter) ApplyPropertyGroups(ctx context.Context, propertyGroups PropertyGroups) (PropertyGroups, error) {
	var res PropertyGroups

	// compile filter regex
	regex, err := regexp.Compile(g.regex)
	if err != nil {
		return nil, errors.Wrap(err, "filter regex failed to compile")
	}

out:
	for i, group := range propertyGroups {
		currentGroup := group

		for i, attr := range g.key {
			if next, ok := currentGroup[attr]; ok {
				if i == len(g.key)-1 {
					break
				}
				var nextGroup propertyGroup
				err = nextGroup.encode(next)
				if err != nil {
					return nil, errors.Wrap(err, "failed to encode next filter key value to property group")
				}
				currentGroup = nextGroup
			} else {
				// current interface does not have the filter key so skip it
				res = append(res, group)
				continue out
			}
		}

		v := currentGroup[g.key[len(g.key)-1]]
		if vString := value.New(v).String(); regex.MatchString(vString) {
			log.Ctx(ctx).Debug().Strs("filter_key", g.key).Str("filter_regex", g.regex).
				Str("received_value", vString).
				Msgf("filter matched on index '%s' of property group", strconv.Itoa(i))
		} else {
			res = append(res, group)
		}
	}

	return res, nil
}

func (g *groupFilter) applySNMP(ctx context.Context, reader snmpReader) (snmpReader, error) {
	if len(reader.wantedIndices) == 0 {
		var err error
		reader.wantedIndices, err = reader.getIndices(ctx)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("failed to read indices, ignoring index oid")
		}
	}
	// TODO copy maps
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
	oidReader := reader.oids
	for _, attr := range g.key {
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
		if regex.MatchString(result.(value.Value).String()) {
			// if filter matches add to filtered indices map and delete from wanted indices
			reader.filteredIndices[index] = struct{}{}
			delete(reader.wantedIndices, index)
			log.Ctx(ctx).Debug().Strs("filter_key", g.key).Str("filter_regex", g.regex).
				Str("received_value", result.(value.Value).String()).
				Msgf("filter matched on index '%s'", index)
		} else {
			// if filter does not match check if index was filtered before
			if _, ok := reader.filteredIndices[index]; !ok {
				// if not add it to wanted indices map
				reader.wantedIndices[index] = struct{}{}
			}
		}
	}

	return reader, nil
}

type ValueFilter interface {
	CheckMatch([]string) bool
}

type valueFilter struct {
	value []string
}

func GetValueFilter(value []string) Filter {
	return &valueFilter{
		value: value,
	}
}

func (g *valueFilter) CheckMatch(value []string) bool {
	for i, k := range value {
		if i == len(g.value) {
			return false
		}
		if k != g.value[i] {
			return false
		}
	}
	return true
}

func (g *valueFilter) ApplyPropertyGroups(ctx context.Context, propertyGroups PropertyGroups) (PropertyGroups, error) {
	var res PropertyGroups

	for _, group := range propertyGroups {
		newGroup, err := filterPropertyGroupKey(ctx, group, g.value, func(a, b string) bool {
			return a == b
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to filter group property")
		}
		res = append(res, newGroup)
	}

	return res, nil
}

func (g *valueFilter) applySNMP(ctx context.Context, reader snmpReader) (snmpReader, error) {
	var err error
	reader.oids, err = filterOIDReaderKey(ctx, reader.oids, g.value, func(a, b string) bool {
		return a == b
	})
	if err != nil {
		return snmpReader{}, err
	}
	return reader, nil
}

func filterPropertyGroupKey(ctx context.Context, group propertyGroup, key []string, compareFunc func(a, b string) bool) (propertyGroup, error) {
	if len(key) == 0 {
		return nil, errors.New("filter key is empty")
	}

	//copy values
	groupCopy := make(propertyGroup)
	for k, v := range group {
		if compareFunc(k, key[0]) {
			if len(key) > 1 {
				var nextGroup propertyGroup
				err := nextGroup.encode(v)
				if err != nil {
					return nil, errors.Wrap(err, "failed to encode next filter key value to property group")
				}
				r, err := filterPropertyGroupKey(ctx, nextGroup, key[1:], compareFunc)
				if err != nil {
					return nil, err
				}
				groupCopy[k] = r
			} else {
				log.Ctx(ctx).Debug().Str("value", k).Msg("filter matched on value in property group")
			}
			continue
		}
		groupCopy[k] = v
	}

	return groupCopy, nil
}

func filterOIDReaderKey(ctx context.Context, reader OIDReader, key []string, compareFunc func(a, b string) bool) (OIDReader, error) {
	if len(key) == 0 {
		return nil, errors.New("filter key is empty")
	}

	// check if current oid reader contains multiple OIDs
	multipleReader, ok := reader.(*deviceClassOIDs)
	if !ok || multipleReader == nil {
		return nil, errors.New("filter attribute does not exist")
	}

	//copy values
	readerCopy := make(deviceClassOIDs)
	for k, v := range *multipleReader {
		if compareFunc(k, key[0]) {
			if len(key) > 1 {
				r, err := filterOIDReaderKey(ctx, v, key[1:], compareFunc)
				if err != nil {
					return nil, err
				}
				readerCopy[k] = r
			} else {
				log.Ctx(ctx).Debug().Str("value", k).Msg("filter matched on value in snmp reader")
			}
			continue
		}
		readerCopy[k] = v
	}

	return &readerCopy, nil
}
