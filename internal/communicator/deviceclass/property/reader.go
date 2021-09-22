package property

import (
	"context"
	"github.com/inexio/thola/internal/communicator/deviceclass/condition"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/inexio/thola/internal/value"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func InterfaceSlice2Reader(i []interface{}, task condition.RelatedTask, parentProperty Reader) (Reader, error) {
	var readerSet readerSet
	for _, i := range i {
		reader, err := interface2PReader(i, task)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert yaml identify property")
		}
		readerSet = append(readerSet, reader)
	}
	if parentProperty != nil {
		readerSet = append(readerSet, parentProperty)
	}
	return &readerSet, nil
}

func interface2PReader(i interface{}, task condition.RelatedTask) (Reader, error) {
	m, ok := i.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("failed to convert interface to map[interface{}]interface{}")
	}
	if _, ok := m["detection"]; !ok {
		return nil, errors.New("detection is missing in property")
	}
	stringDetection, ok := m["detection"].(string)
	if !ok {
		return nil, errors.New("property detection needs to be a string")
	}
	var basePropReader baseReader
	switch stringDetection {
	case "snmpget":
		var pr snmpGetReader
		err := mapstructure.Decode(i, &pr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode constant reader")
		}
		basePropReader.reader = &pr
	case "constant":
		v, ok := m["value"]
		if !ok {
			return nil, errors.New("value is missing in constant property reader")
		}
		var pr constantReader
		if _, ok := v.(map[interface{}]interface{}); ok {
			return nil, errors.New("value must not be a map")
		}
		if _, ok := v.([]interface{}); ok {
			return nil, errors.New("value must not be an array")
		}
		pr.Value = value.New(v)
		basePropReader.reader = &pr
	case "SysObjectID":
		var pr sysObjectIDReader
		err := mapstructure.Decode(i, &pr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode sysObjectIDReader")
		}
		basePropReader.reader = &pr
	case "SysDescription":
		var pr sysDescriptionReader
		err := mapstructure.Decode(i, &pr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode sysDescriptionReader")
		}
		basePropReader.reader = &pr
	case "Vendor":
		if task <= condition.PropertyVendor {
			return nil, errors.New("cannot use vendor property, model series is not available here yet")
		}
		var pr vendorReader
		err := mapstructure.Decode(i, &pr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode vendor Reader")
		}
		basePropReader.reader = &pr
	case "Model":
		if task <= condition.PropertyModel {
			return nil, errors.New("cannot use model property, model series is not available here yet")
		}
		var pr modelReader
		err := mapstructure.Decode(i, &pr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode model Reader")
		}
		basePropReader.reader = &pr
	case "ModelSeries":
		if task <= condition.PropertyModelSeries {
			return nil, errors.New("cannot use model series property, model series is not available here yet")
		}
		var pr modelSeriesReader
		err := mapstructure.Decode(i, &pr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode model series Reader")
		}
		basePropReader.reader = &pr

	default:
		return nil, errors.New("invalid detection type " + stringDetection)
	}
	if operators, ok := m["operators"]; ok {
		operatorSlice, ok := operators.([]interface{})
		if !ok {
			return nil, errors.New("operators has to be an array")
		}
		operators, err := InterfaceSlice2Operators(operatorSlice, task)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert interface slice to string operators")
		}
		basePropReader.operators = operators
	}
	if preConditionInterface, ok := m["pre_condition"]; ok {
		preCondition, err := condition.Interface2Condition(preConditionInterface, task)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert pre condition interface 2 condition")
		}
		basePropReader.preCondition = preCondition
	}
	return &basePropReader, nil
}

type Reader interface {
	GetProperty(ctx context.Context) (value.Value, error)
}

type readerSet []Reader

func (p *readerSet) GetProperty(ctx context.Context) (value.Value, error) {
	log.Ctx(ctx).Debug().Msg("starting with property reader set")
	for _, reader := range *p {
		property, err := reader.GetProperty(ctx)
		if err == nil {
			return property, nil
		}
	}
	return value.Empty(), tholaerr.NewNotFoundError("failed to read out property")
}

type baseReader struct {
	reader       Reader
	operators    Operators
	preCondition condition.Condition
}

func (b *baseReader) GetProperty(ctx context.Context) (value.Value, error) {
	if b.preCondition != nil {
		conditionsMatched, err := b.preCondition.Check(ctx)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("error during pre condition check")
			return value.Empty(), errors.Wrap(err, "an error occurred while checking preconditions")
		}
		if !conditionsMatched {
			log.Ctx(ctx).Debug().Err(err).Msg("pre condition not fulfilled")
			return value.Empty(), errors.New("pre condition failed")
		}
	}
	v, err := b.reader.GetProperty(ctx)
	if err != nil {
		return value.Empty(), err
	}
	v, err = b.applyOperators(ctx, v)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("error while applying operators")
		return value.Empty(), errors.Wrap(err, "error while applying operators")
	}
	log.Ctx(ctx).Debug().Msgf("property determined (%v)", v)
	return v, nil
}

func (b *baseReader) applyOperators(ctx context.Context, v value.Value) (value.Value, error) {
	return b.operators.Apply(ctx, v)
}

type constantReader struct {
	Value value.Value
}

func (c *constantReader) GetProperty(ctx context.Context) (value.Value, error) {
	log.Ctx(ctx).Debug().Str("property_reader", "constant").Msg("setting constant property")
	return c.Value, nil
}

type sysObjectIDReader struct {
}

func (c *sysObjectIDReader) GetProperty(ctx context.Context) (value.Value, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return value.Empty(), errors.New("snmp data is missing, SysObjectID property cannot be read")
	}

	sysObjectID, err := con.SNMP.GetSysObjectID(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Str("property_reader", "SysObjectID").Msg("failed to get sys object id")
		return value.Empty(), errors.New("failed to get sys object id")
	}
	log.Ctx(ctx).Debug().Str("property_reader", "SysObjectID").Msg("received SysObjectID successfully")

	return value.New(sysObjectID), nil
}

type sysDescriptionReader struct{}

func (c *sysDescriptionReader) GetProperty(ctx context.Context) (value.Value, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return value.Empty(), errors.New("snmp data is missing, SysDescription property cannot be read")
	}

	sysDescription, err := con.SNMP.GetSysDescription(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Str("property_reader", "SysDescription").Msg("failed to get sys description")
		return value.Empty(), errors.New("failed to get sys description")
	}
	log.Ctx(ctx).Debug().Str("property_reader", "SysDescription").Msg("received SysDescription successfully")

	return value.New(sysDescription), nil
}

type snmpGetReader struct {
	network.SNMPGetConfiguration `yaml:",inline" mapstructure:",squash"`
}

func (s *snmpGetReader) GetProperty(ctx context.Context) (value.Value, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil || con.SNMP.SnmpClient == nil {
		return value.Empty(), errors.New("No SNMP Data available!")
	}
	oid := string(s.OID)
	result, err := con.SNMP.SnmpClient.SNMPGet(ctx, oid)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Str("property_reader", "snmpget").Msg("snmpget on oid " + oid + " failed")
		return value.Empty(), errors.Wrap(err, "snmpget failed")
	}

	var val interface{}
	if s.UseRawResult {
		val, err = result[0].GetValueStringRaw()
	} else {
		val, err = result[0].GetValue()
	}
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Str("property_reader", "snmpget").Msg("snmpget failed")
		return value.Empty(), err
	}
	log.Ctx(ctx).Debug().Str("property_reader", "snmpget").Msg("snmpget successful")
	return value.New(val), nil
}

type vendorReader struct{}

func (v *vendorReader) GetProperty(ctx context.Context) (value.Value, error) {
	properties, ok := device.DevicePropertiesFromContext(ctx)
	if !ok {
		return value.Empty(), errors.New("no properties found in context")
	}

	if properties.Properties.Vendor == nil {
		log.Ctx(ctx).Debug().Str("property_reader", "vendor").Msg("vendor has not yet been determined")
		return value.Empty(), tholaerr.NewPreConditionError("vendor has not yet been determined")
	}
	return value.New(*properties.Properties.Vendor), nil
}

type modelReader struct{}

func (m *modelReader) GetProperty(ctx context.Context) (value.Value, error) {
	properties, ok := device.DevicePropertiesFromContext(ctx)
	if !ok {
		return value.Empty(), errors.New("no properties found in context")
	}

	if properties.Properties.Model == nil {
		log.Ctx(ctx).Debug().Str("property_reader", "model").Msg("model has not yet been determined")
		return value.Empty(), tholaerr.NewPreConditionError("model has not yet been determined")
	}
	return value.New(*properties.Properties.Model), nil
}

type modelSeriesReader struct{}

func (m *modelSeriesReader) GetProperty(ctx context.Context) (value.Value, error) {
	properties, ok := device.DevicePropertiesFromContext(ctx)
	if !ok {
		return value.Empty(), errors.New("no properties found in context")
	}

	if properties.Properties.ModelSeries == nil {
		log.Ctx(ctx).Debug().Str("property_reader", "model_series").Msg("model series has not yet been determined")
		return value.Empty(), tholaerr.NewPreConditionError("model series has not yet been determined")
	}
	return value.New(*properties.Properties.ModelSeries), nil
}
