package deviceclass

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/inexio/thola/internal/value"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type propertyReader interface {
	getProperty(ctx context.Context) (value.Value, error)
}

type propertyReaderSet []propertyReader

func (p *propertyReaderSet) getProperty(ctx context.Context) (value.Value, error) {
	log.Ctx(ctx).Trace().Msg("starting with property reader set")
	for _, reader := range *p {
		property, err := reader.getProperty(ctx)
		if err == nil {
			return property, nil
		}
	}
	return value.Empty(), tholaerr.NewNotFoundError("failed to read out property")
}

type basePropertyReader struct {
	propertyReader propertyReader
	operators      propertyOperators
	preCondition   condition
}

func (b *basePropertyReader) getProperty(ctx context.Context) (value.Value, error) {
	if b.preCondition != nil {
		conditionsMatched, err := b.preCondition.check(ctx)
		if err != nil {
			log.Ctx(ctx).Trace().Err(err).Msg("error during pre condition check")
			return value.Empty(), errors.Wrap(err, "an error occurred while checking preconditions")
		}
		if !conditionsMatched {
			log.Ctx(ctx).Trace().Err(err).Msg("pre condition not fulfilled")
			return value.Empty(), errors.New("pre condition failed")
		}
	}
	v, err := b.propertyReader.getProperty(ctx)
	if err != nil {
		return value.Empty(), err
	}
	v, err = b.applyOperators(ctx, v)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("error while applying operators")
		return value.Empty(), errors.Wrap(err, "error while applying operators")
	}
	log.Ctx(ctx).Trace().Msgf("property determined (%v)", v)
	return v, nil
}

func (b *basePropertyReader) applyOperators(ctx context.Context, v value.Value) (value.Value, error) {
	return b.operators.apply(ctx, v)
}

type constantPropertyReader struct {
	Value value.Value
}

func (c *constantPropertyReader) getProperty(ctx context.Context) (value.Value, error) {
	log.Ctx(ctx).Trace().Str("property_reader", "constant").Msg("setting constant property")
	return c.Value, nil
}

type sysObjectIDPropertyReader struct {
}

func (c *sysObjectIDPropertyReader) getProperty(ctx context.Context) (value.Value, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return value.Empty(), errors.New("snmp data is missing, SysObjectID property cannot be read")
	}

	sysObjectID, err := con.SNMP.GetSysObjectID(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Str("property_reader", "SysObjectID").Msg("failed to get sys object id")
		return value.Empty(), errors.New("failed to get sys object id")
	}
	log.Ctx(ctx).Trace().Str("property_reader", "SysObjectID").Msg("received SysObjectID successfully")

	return value.New(sysObjectID), nil
}

type sysDescriptionPropertyReader struct {
}

func (c *sysDescriptionPropertyReader) getProperty(ctx context.Context) (value.Value, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return value.Empty(), errors.New("snmp data is missing, SysDescription property cannot be read")
	}

	sysDescription, err := con.SNMP.GetSysDescription(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Str("property_reader", "SysDescription").Msg("failed to get sys description")
		return value.Empty(), errors.New("failed to get sys description")
	}
	log.Ctx(ctx).Trace().Str("property_reader", "SysDescription").Msg("received SysDescription successfully")

	return value.New(sysDescription), nil
}

type snmpGetPropertyReader struct {
	network.SNMPGetConfiguration `yaml:",inline" mapstructure:",squash"`
}

func (s *snmpGetPropertyReader) getProperty(ctx context.Context) (value.Value, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil || con.SNMP.SnmpClient == nil {
		return value.Empty(), errors.New("No SNMP Data available!")
	}
	oid := string(s.OID)
	result, err := con.SNMP.SnmpClient.SNMPGet(ctx, oid)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Str("property_reader", "snmpget").Msg("snmpget on oid " + oid + " failed")
		return value.Empty(), errors.Wrap(err, "snmpget failed")
	}

	var val interface{}
	if s.UseRawResult {
		val, err = result[0].GetValueStringRaw()
	} else {
		val, err = result[0].GetValue()
	}
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Str("property_reader", "snmpget").Msg("snmpget failed")
		return value.Empty(), err
	}
	log.Ctx(ctx).Trace().Str("property_reader", "snmpget").Msg("snmpget successful")
	return value.New(val), nil
}

type vendorPropertyReader struct{}

func (v *vendorPropertyReader) getProperty(ctx context.Context) (value.Value, error) {
	properties, ok := device.DevicePropertiesFromContext(ctx)
	if !ok {
		return value.Empty(), errors.New("no properties found in context")
	}

	if properties.Properties.Vendor == nil {
		log.Ctx(ctx).Trace().Str("property_reader", "vendor").Msg("vendor has not yet been determined")
		return value.Empty(), tholaerr.NewPreConditionError("vendor has not yet been determined")
	}
	return value.New(*properties.Properties.Vendor), nil
}

type modelPropertyReader struct{}

func (m *modelPropertyReader) getProperty(ctx context.Context) (value.Value, error) {
	properties, ok := device.DevicePropertiesFromContext(ctx)
	if !ok {
		return value.Empty(), errors.New("no properties found in context")
	}

	if properties.Properties.Model == nil {
		log.Ctx(ctx).Trace().Str("property_reader", "model").Msg("model has not yet been determined")
		return value.Empty(), tholaerr.NewPreConditionError("model has not yet been determined")
	}
	return value.New(*properties.Properties.Model), nil
}

type modelSeriesPropertyReader struct{}

func (m *modelSeriesPropertyReader) getProperty(ctx context.Context) (value.Value, error) {
	properties, ok := device.DevicePropertiesFromContext(ctx)
	if !ok {
		return value.Empty(), errors.New("no properties found in context")
	}

	if properties.Properties.ModelSeries == nil {
		log.Ctx(ctx).Trace().Str("property_reader", "model_series").Msg("model series has not yet been determined")
		return value.Empty(), tholaerr.NewPreConditionError("model series has not yet been determined")
	}
	return value.New(*properties.Properties.ModelSeries), nil
}
