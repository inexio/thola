package deviceclass

import (
	"context"
	"fmt"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/inexio/thola/core/value"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"regexp"
	"strings"
)

type propertyOperator interface {
	operate(context.Context, value.Value) (value.Value, error)
	returnOnErrorPropertyOperator
}

type returnOnErrorPropertyOperator interface {
	returnOnError() bool
}

type modifyOperator interface {
	modify(context.Context, value.Value) (value.Value, error)
}

type filterOperator interface {
	filter(context.Context, value.Value) error
}

type switchOperator interface {
	switchOperate(context.Context, value.Value) (value.Value, error)
}

type modifyOperatorAdapter struct {
	operator modifyOperator
}

func (o *modifyOperatorAdapter) operate(ctx context.Context, v value.Value) (value.Value, error) {
	return o.operator.modify(ctx, v)
}

func (o *modifyOperatorAdapter) returnOnError() bool {
	return checkReturnOnError(o.operator)
}

type filterOperatorAdapter struct {
	operator filterOperator
}

func (o *filterOperatorAdapter) operate(ctx context.Context, v value.Value) (value.Value, error) {
	err := o.operator.filter(ctx, v)
	if err != nil {
		return value.Empty(), err
	}
	return v, err
}

func (o *filterOperatorAdapter) returnOnError() bool {
	return checkReturnOnError(o.operator)
}

type switchOperatorAdapter struct {
	operator switchOperator
}

func (o *switchOperatorAdapter) operate(ctx context.Context, v value.Value) (value.Value, error) {
	return o.operator.switchOperate(ctx, v)
}

func (o *switchOperatorAdapter) returnOnError() bool {
	return checkReturnOnError(o.operator)
}

func checkReturnOnError(i interface{}) bool {
	if x, ok := i.(returnOnErrorPropertyOperator); ok {
		return x.returnOnError()
	}
	return false
}

type propertyOperators []propertyOperator

func (o *propertyOperators) apply(ctx context.Context, v value.Value) (value.Value, error) {
	for _, operator := range *o {
		x, err := operator.operate(ctx, v)
		if err != nil {
			// if an error occurs, we check if the current operator is
			// a strFilter should still return the previous value
			if operator.returnOnError() {
				return v, nil
			}
			return value.Empty(), errors.Wrap(err, "operator failed")
		}
		v = x
	}
	return v, nil
}

type multiplyNumberModifier struct {
	value propertyReader
}

func (m *multiplyNumberModifier) modify(ctx context.Context, v value.Value) (value.Value, error) {
	a, err := decimal.NewFromString(v.String())
	if err != nil {
		return value.Empty(), errors.New("cannot covert current value to decimal number")
	}
	valueProperty, err := m.value.getProperty(ctx)
	if err != nil {
		return value.Empty(), errors.New("cannot read argument")
	}
	valueFloat, err := valueProperty.Float64()
	if err != nil {
		return value.Empty(), errors.New("cannot covert argument to decimal number")
	}
	b := decimal.NewFromFloat(valueFloat)
	result := a.Mul(b)
	return value.New(result), nil
}

type divideNumberModifier struct {
	value propertyReader
}

func (m *divideNumberModifier) modify(ctx context.Context, v value.Value) (value.Value, error) {
	a, err := decimal.NewFromString(v.String())
	if err != nil {
		return value.Empty(), errors.New("cannot covert current value to decimal number")
	}
	valueProperty, err := m.value.getProperty(ctx)
	if err != nil {
		return value.Empty(), errors.New("cannot read argument")
	}
	valueFloat, err := valueProperty.Float64()
	if err != nil {
		return value.Empty(), errors.New("cannot covert argument to decimal number")
	}
	b := decimal.NewFromFloat(valueFloat)
	result := a.DivRound(b, 2)
	return value.New(result), nil
}

type baseStringFilter struct {
	Value            string    `mapstructure:"value"`
	FilterMethod     matchMode `mapstructure:"filter_method"`
	returnOnMismatch bool      `mapstructure:"return_on_mismatch"`
}

func (f *baseStringFilter) filter(ctx context.Context, v value.Value) error {
	match, err := matchStrings(ctx, v.String(), f.FilterMethod, f.Value)
	if err != nil {
		return errors.Wrap(err, "error during match strings")
	}
	if !match {
		return errors.New("string is not valid")
	}
	return nil
}

func (f *baseStringFilter) returnOnError() bool {
	return f.returnOnMismatch
}

type toUpperCaseModifier struct{}

func (o *toUpperCaseModifier) modify(_ context.Context, v value.Value) (value.Value, error) {
	return value.New(strings.ToUpper(v.String())), nil
}

type toLowerCaseModifier struct{}

func (o *toLowerCaseModifier) modify(_ context.Context, v value.Value) (value.Value, error) {
	return value.New(strings.ToLower(v.String())), nil
}

type overwriteModifier struct {
	overwriteString string
}

func (o *overwriteModifier) modify(_ context.Context, _ value.Value) (value.Value, error) {
	return value.New(o.overwriteString), nil
}

type addSuffixModifier struct {
	suffix string
}

func (a *addSuffixModifier) modify(_ context.Context, v value.Value) (value.Value, error) {
	return value.New(v.String() + a.suffix), nil
}

type addPrefixModifier struct {
	prefix string
}

func (a *addPrefixModifier) modify(_ context.Context, v value.Value) (value.Value, error) {
	return value.New(a.prefix + v.String()), nil
}

type regexSubmatchModifier struct {
	regex  *regexp.Regexp
	format string
}

func newRegexSubmatchModifier(regex string, format string) (*regexSubmatchModifier, error) {
	var r regexSubmatchModifier
	re, err := regexp.Compile(regex)
	if err != nil {
		return nil, errors.Wrap(err, "regex compile failed")
	}
	r.regex = re
	r.format = format
	return &r, nil
}

func (o *regexSubmatchModifier) modify(_ context.Context, v value.Value) (value.Value, error) {
	subMatches := o.regex.FindStringSubmatch(v.String())
	if subMatches == nil {
		return value.Empty(), errors.New("regex does not match")
	}
	return value.New(o.regex.ReplaceAllString(subMatches[0], o.format)), nil
}

type regexReplaceModifier struct {
	regex   *regexp.Regexp
	replace string
}

func newRegexReplaceModifier(regex, replace string) (*regexReplaceModifier, error) {
	var r regexReplaceModifier
	re, err := regexp.Compile(regex)
	if err != nil {
		return nil, errors.Wrap(err, "regex compile failed")
	}
	r.regex = re
	r.replace = replace
	return &r, nil
}

func (r *regexReplaceModifier) modify(_ context.Context, v value.Value) (value.Value, error) {
	return value.New(r.regex.ReplaceAllString(v.String(), r.replace)), nil
}

type insertReadValueModifier struct {
	readValueReader propertyReader
	format          string
}

func (r *insertReadValueModifier) modify(ctx context.Context, v value.Value) (value.Value, error) {
	readValue, err := r.readValueReader.getProperty(ctx)
	if err != nil {
		return value.Empty(), errors.Wrap(err, "failed to read out value")
	}
	str := strings.ReplaceAll(r.format, "$property$", v.String())
	str = strings.ReplaceAll(str, "$read_value$", fmt.Sprint(readValue))
	return value.New(str), nil
}

type mapModifier struct {
	ignoreOnMismatch bool
	mappings         map[string]string
}

func (r *mapModifier) modify(_ context.Context, v value.Value) (value.Value, error) {
	if val, ok := r.mappings[v.String()]; ok {
		return value.New(val), nil
	}
	if r.ignoreOnMismatch {
		return value.Empty(), nil
	}
	return value.Empty(), tholaerr.NewNotFoundError("string not found in mapping")
}

type genericStringSwitch struct {
	switchValueGetter stringSwitchValueGetter
	switchMode        matchMode
	cases             []stringSwitchCase
}

type stringSwitchCase struct {
	caseString string
	operators  propertyOperators
}

func (w *genericStringSwitch) switchOperate(ctx context.Context, s value.Value) (value.Value, error) {
	switchValue, err := w.switchValueGetter.getSwitchValue(ctx, s)
	if err != nil {
		return value.Empty(), errors.Wrap(err, "failed to get switch string")
	}
	switchString := switchValue.String()
	for _, c := range w.cases {
		b, err := matchStrings(ctx, switchString, w.switchMode, c.caseString)
		if err != nil {
			log.Ctx(ctx).Trace().Err(err).Msg("error during match strings")
			continue
		}
		if !b {
			log.Ctx(ctx).Trace().Err(err).Msg("string does not match")
			continue
		}
		x, err := c.operators.apply(ctx, s)
		if err != nil {
			return value.Empty(), errors.Wrapf(err, "failed to apply operations inside switch case '%s'", c.caseString)
		}
		return value.New(x), nil
	}
	return s, nil
}

type stringSwitchValueGetter interface {
	getSwitchValue(context.Context, value.Value) (value.Value, error)
}

type defaultStringSwitchValueGetter struct {
}

func (w *defaultStringSwitchValueGetter) getSwitchValue(_ context.Context, v value.Value) (value.Value, error) {
	return v, nil
}

type snmpwalkCountStringSwitchValueGetter struct {
	oid             string
	useOidForFilter bool
	filter          *baseStringFilter
}

func (w *snmpwalkCountStringSwitchValueGetter) getSwitchValue(ctx context.Context, _ value.Value) (value.Value, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return value.Empty(), errors.New("no snmp connection available, snmpwalk not possible")
	}
	var i int

	res, err := con.SNMP.SnmpClient.SNMPWalk(ctx, w.oid)
	if err != nil {
		return value.Empty(), errors.Wrap(err, "snmpwalk failed")
	}
	for _, r := range res {
		var str string
		if w.useOidForFilter {
			str = r.GetOID()
		} else {
			str, err = r.GetValueString()
			if err != nil {
				//LOG
				continue
			}
		}

		if w.filter != nil {
			err = w.filter.filter(ctx, value.New(str))
			if err == nil {
				i++
			}
		} else {
			i++
		}
	}
	return value.New(i), nil
}
