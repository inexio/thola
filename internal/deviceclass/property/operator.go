package property

import (
	"context"
	"fmt"
	condition2 "github.com/inexio/thola/internal/deviceclass/condition"
	"github.com/inexio/thola/internal/mapping"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/inexio/thola/internal/value"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"regexp"
	"strconv"
	"strings"
)

func InterfaceSlice2Operators(i []interface{}, task condition2.RelatedTask) (Operators, error) {
	var propertyOperators Operators
	for _, opInterface := range i {
		m, ok := opInterface.(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("failed to convert interface to map[interface{}]interface{}")
		}
		if _, ok := m["type"]; !ok {
			return nil, errors.New("operator type is missing!")
		}
		stringType, ok := m["type"].(string)
		if !ok {
			return nil, errors.New("operator type needs to be a string")
		}

		switch stringType {
		case "filter":
			var adapter filterOperatorAdapter
			var filter baseStringFilter
			filterMethod, ok := m["filter_method"]
			if ok {
				if filterMethodString, ok := filterMethod.(string); ok {
					filter.FilterMethod = condition2.MatchMode(filterMethodString)
				} else {
					return nil, errors.New("filter method needs to be a string")
				}
				err := filter.FilterMethod.Validate()
				if err != nil {
					return nil, errors.Wrap(err, "invalid filter method")
				}
			} else {
				filter.FilterMethod = "contains"
			}
			val, ok := m["value"]
			if !ok {
				return nil, errors.New("value is missing")
			}
			if valueString, ok := val.(string); ok {
				filter.Value = valueString
			}
			if returnOnMismatchInt, ok := m["return_on_mismatch"]; ok {
				if returnOnMismatch, ok := returnOnMismatchInt.(bool); ok {
					filter.returnOnMismatch = returnOnMismatch
				} else {
					return nil, errors.New("return_on_mismatch needs to be a boolean")
				}
			}
			adapter.operator = &filter
			propertyOperators = append(propertyOperators, &adapter)
		case "modify":
			var modifier modifyOperatorAdapter
			modifyMethod, ok := m["modify_method"]
			if !ok {
				return nil, errors.New("modify method is missing in modify operator")
			}
			modifyMethodString, ok := modifyMethod.(string)
			if !ok {
				return nil, errors.New("modify method isn't a string")
			}
			switch modifyMethodString {
			case "regexSubmatch":
				format, ok := m["format"]
				if !ok {
					return nil, errors.New("format is missing")
				}
				formatString, ok := format.(string)
				if !ok {
					return nil, errors.New("format has to be a string")
				}
				regex, ok := m["regex"]
				if !ok {
					return nil, errors.New("regex is missing")
				}
				regexString, ok := regex.(string)
				if !ok {
					return nil, errors.New("regex has to be a string")
				}
				var returnOnMismatch bool
				if returnOnMismatchInt, ok := m["return_on_mismatch"]; ok {
					if returnOnMismatch, ok = returnOnMismatchInt.(bool); !ok {
						return nil, errors.New("return_on_mismatch needs to be a boolean")
					}
				}
				mod, err := newRegexSubmatchModifier(regexString, formatString, returnOnMismatch)
				if err != nil {
					return nil, errors.Wrap(err, "failed to create new regex submatch modifier")
				}
				modifier.operator = mod
			case "regexReplace":
				replace, ok := m["replace"]
				if !ok {
					return nil, errors.New("replace is missing")
				}
				replaceString, ok := replace.(string)
				if !ok {
					return nil, errors.New("replace has to be a string")
				}
				regex, ok := m["regex"]
				if !ok {
					return nil, errors.New("regex is missing")
				}
				regexString, ok := regex.(string)
				if !ok {
					return nil, errors.New("regex has to be a string")
				}
				mod, err := newRegexReplaceModifier(regexString, replaceString)
				if err != nil {
					return nil, errors.Wrap(err, "failed to create new regex replace modifier")
				}
				modifier.operator = mod
			case "toUpperCase":
				var toUpperCaseModifier toUpperCaseModifier
				modifier.operator = &toUpperCaseModifier
			case "toLowerCase":
				var toLowerCaseModifier toLowerCaseModifier
				modifier.operator = &toLowerCaseModifier
			case "overwrite":
				overwriteString, ok := m["value"].(string)
				if !ok {
					return nil, errors.New("value is missing in overwrite operator, or is not of type string")
				}
				var overwriteModifier overwriteModifier
				overwriteModifier.overwriteString = overwriteString
				modifier.operator = &overwriteModifier
			case "addPrefix":
				prefix, ok := m["value"].(string)
				if !ok {
					return nil, errors.New("value is missing in addPrefix operator, or is not of type string")
				}
				var prefixModifier addPrefixModifier
				prefixModifier.prefix = prefix
				modifier.operator = &prefixModifier
			case "addSuffix":
				suffix, ok := m["value"].(string)
				if !ok {
					return nil, errors.New("value is missing in addSuffix operator, or is not of type string")
				}
				var suffixModifier addSuffixModifier
				suffixModifier.suffix = suffix
				modifier.operator = &suffixModifier
			case "insertReadValue":
				format, ok := m["format"].(string)
				if !ok {
					return nil, errors.New("format is missing in insertReadValue operator, or is not of type string")
				}
				valueReaderInterface, ok := m["read_value"]
				if !ok {
					return nil, errors.New("read value is missing in insertReadValue operator")
				}
				valueReader, err := interface2PReader(valueReaderInterface, task)
				if err != nil {
					return nil, errors.Wrap(err, "failed to convert read_value to reader in insertReadValue operator")
				}
				var irvModifier insertReadValueModifier
				irvModifier.format = format
				irvModifier.readValueReader = valueReader
				modifier.operator = &irvModifier
			case "map":
				mappingsInterface, ok := m["mappings"]
				if !ok {
					return nil, errors.New("mappings is missing in map string modifier")
				}
				var ignoreOnMismatch bool
				ignoreOnMismatchInterface, ok := m["ignore_on_mismatch"]
				if ok {
					ignoreOnMismatchBool, ok := ignoreOnMismatchInterface.(bool)
					if !ok {
						return nil, errors.New("ignore_on_mismatch in map modifier needs to be boolean")
					}
					ignoreOnMismatch = ignoreOnMismatchBool
				}

				var mapModifier mapModifier
				mapModifier.ignoreOnMismatch = ignoreOnMismatch

				mappings, ok := mappingsInterface.(map[interface{}]interface{})
				if !ok {
					file, ok := mappingsInterface.(string)
					if !ok {
						return nil, errors.New("mappings needs to be a map[string]string or string in map string modifier")
					}
					mappingsFile, err := mapping.GetMapping(file)
					if err != nil {
						return nil, errors.Wrap(err, "can't get specified mapping")
					}
					mapModifier.mappings = mappingsFile
				} else {
					mapModifier.mappings = make(map[string]string)
					for k, val := range mappings {
						key := fmt.Sprint(k)
						valString := fmt.Sprint(val)

						mapModifier.mappings[key] = valString
					}
				}
				if len(mapModifier.mappings) == 0 {
					return nil, errors.New("mappings is empty")
				}
				modifier.operator = &mapModifier
			case "add":
				valueReaderInterface := m["value"]
				if !ok {
					return nil, errors.New("value is missing in add")
				}
				valueReader, err := interface2PReader(valueReaderInterface, task)
				if err != nil {
					return nil, errors.New("value is missing in add modify operator, or is not of type float64")
				}
				var addModifier addNumberModifier
				addModifier.value = valueReader
				modifier.operator = &addModifier
			case "subtract":
				valueReaderInterface := m["value"]
				if !ok {
					return nil, errors.New("value is missing in subtract")
				}
				valueReader, err := interface2PReader(valueReaderInterface, task)
				if err != nil {
					return nil, errors.New("value is missing in subtract modify operator, or is not of type float64")
				}
				var subtractModifier subtractNumberModifier
				subtractModifier.value = valueReader
				modifier.operator = &subtractModifier
			case "multiply":
				valueReaderInterface := m["value"]
				if !ok {
					return nil, errors.New("value is missing in multiply")
				}
				valueReader, err := interface2PReader(valueReaderInterface, task)
				if err != nil {
					return nil, errors.New("value is missing in multiply modify operator, or is not of type float64")
				}
				var multiplyModifier multiplyNumberModifier
				multiplyModifier.value = valueReader
				modifier.operator = &multiplyModifier
			case "divide":
				valueReaderInterface := m["value"]
				if !ok {
					return nil, errors.New("value is missing in divide")
				}
				valueReader, err := interface2PReader(valueReaderInterface, task)
				if err != nil {
					return nil, errors.New("value is missing in divide modify operator, or is not of type float64")
				}
				var divideModifier divideNumberModifier
				divideModifier.value = valueReader
				modifier.operator = &divideModifier
			default:
				return nil, fmt.Errorf("invalid modify method '%s'", modifyMethod)
			}
			propertyOperators = append(propertyOperators, &modifier)
		case "switch":
			var sw switchOperatorAdapter
			var switcher genericStringSwitch
			var switchValue string

			// get switch mode, default = equals
			switchMode, ok := m["switch_mode"]
			if ok {
				if switchModeString, ok := switchMode.(string); ok {
					switcher.switchMode = condition2.MatchMode(switchModeString)
				} else {
					return nil, errors.New("filter method needs to be a string")
				}
				err := switcher.switchMode.Validate()
				if err != nil {
					return nil, errors.Wrap(err, "invalid filter method")
				}
			} else {
				switcher.switchMode = "equals"
			}

			// get switch value, default = "default"
			switchValueInterface, ok := m["switch_value"]
			if ok {
				if switchValue, ok = switchValueInterface.(string); !ok {
					return nil, errors.New("switch value needs to be a string")
				}
			} else {
				switchValue = "default"
			}

			// get switchValueGetter
			switch switchValue {
			case "default":
				switcher.switchValueGetter = &defaultStringSwitchValueGetter{}
			case "snmpwalkCount":
				switchValueGetter := snmpwalkCountStringSwitchValueGetter{}
				oid, ok := m["oid"].(string)
				if !ok {
					return nil, errors.New("oid in snmpwalkCount switch operator is missing, or is not a string")
				}
				switchValueGetter.oid = oid
				if filter, ok := m["snmp_result_filter"]; ok {
					var bStrFilter baseStringFilter
					err := mapstructure.Decode(filter, &bStrFilter)
					if err != nil {
						return nil, errors.Wrap(err, "failed to decode snmp_result_filter")
					}
					err = bStrFilter.FilterMethod.Validate()
					if err != nil {
						return nil, errors.Wrap(err, "invalid filter method")
					}
					switchValueGetter.filter = &bStrFilter

					if useOidForFilter, ok := m["use_oid_for_filter"].(bool); ok {
						switchValueGetter.useOidForFilter = useOidForFilter
					}
				}
				switcher.switchValueGetter = &switchValueGetter
			}

			// following operators
			cases, ok := m["cases"].([]interface{})
			if !ok {
				return nil, errors.New("cases are missing in switch operator, or it is not an array")
			}

			for _, cInterface := range cases {
				c, ok := cInterface.(map[interface{}]interface{})
				if !ok {
					return nil, errors.New("switch case needs to be a map")
				}
				caseString, ok := c["case"].(string)
				if !ok {
					caseInt, ok := c["case"].(int)
					if !ok {
						return nil, errors.New("case string is missing in switch operator case, or is not a string or int")
					}
					caseString = strconv.Itoa(caseInt)
				}
				subOperatorsInterface, ok := c["operators"].([]interface{})
				if !ok {
					return nil, fmt.Errorf("operators are missing in switch operator case, or it is not an array, in switch case '%s'", caseString)
				}
				subOperators, err := InterfaceSlice2Operators(subOperatorsInterface, task)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to convert []interface{} to propertyOperators in switch case '%s'", caseString)
				}
				switchCase := stringSwitchCase{
					caseString: caseString,
					operators:  subOperators,
				}
				switcher.cases = append(switcher.cases, switchCase)
			}

			sw.operator = &switcher
			propertyOperators = append(propertyOperators, &sw)
		default:
			return nil, fmt.Errorf("invalid operator type '%s'", stringType)
		}
	}
	return propertyOperators, nil
}

type Operators []operator

func (o *Operators) Apply(ctx context.Context, v value.Value) (value.Value, error) {
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

type operator interface {
	operate(context.Context, value.Value) (value.Value, error)
	returnOnErrorOperator
}

type returnOnErrorOperator interface {
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
	if x, ok := i.(returnOnErrorOperator); ok {
		return x.returnOnError()
	}
	return false
}

type addNumberModifier struct {
	value Reader
}

func (m *addNumberModifier) modify(ctx context.Context, v value.Value) (value.Value, error) {
	a, b, err := getCalculationOperators(ctx, v, m.value)
	if err != nil {
		return nil, err
	}
	result := a.Add(b)
	return value.New(result), nil
}

type subtractNumberModifier struct {
	value Reader
}

func (m *subtractNumberModifier) modify(ctx context.Context, v value.Value) (value.Value, error) {
	a, b, err := getCalculationOperators(ctx, v, m.value)
	if err != nil {
		return nil, err
	}
	result := a.Sub(b)
	return value.New(result), nil
}

type multiplyNumberModifier struct {
	value Reader
}

func (m *multiplyNumberModifier) modify(ctx context.Context, v value.Value) (value.Value, error) {
	a, b, err := getCalculationOperators(ctx, v, m.value)
	if err != nil {
		return nil, err
	}
	result := a.Mul(b)
	return value.New(result), nil
}

type divideNumberModifier struct {
	value Reader
}

func (m *divideNumberModifier) modify(ctx context.Context, v value.Value) (value.Value, error) {
	a, b, err := getCalculationOperators(ctx, v, m.value)
	if err != nil {
		return nil, err
	}
	result := a.DivRound(b, 2)
	return value.New(result), nil
}

func getCalculationOperators(ctx context.Context, v value.Value, value Reader) (decimal.Decimal, decimal.Decimal, error) {
	a, err := decimal.NewFromString(v.String())
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, errors.New("cannot covert current value to decimal number")
	}
	valueProperty, err := value.GetProperty(ctx)
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, errors.New("cannot read argument")
	}
	valueFloat, err := valueProperty.Float64()
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, errors.New("cannot covert argument to decimal number")
	}
	b := decimal.NewFromFloat(valueFloat)

	return a, b, nil
}

type baseStringFilter struct {
	Value            string               `mapstructure:"value"`
	FilterMethod     condition2.MatchMode `mapstructure:"filter_method"`
	returnOnMismatch bool                 `mapstructure:"return_on_mismatch"`
}

func (f *baseStringFilter) filter(ctx context.Context, v value.Value) error {
	match, err := condition2.MatchStrings(ctx, v.String(), f.FilterMethod, f.Value)
	if err != nil {
		return errors.Wrap(err, "error during match strings")
	}
	if !match {
		return tholaerr.NewDidNotMatchError("value didn't match")
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
	regex            *regexp.Regexp
	format           string
	returnOnMismatch bool
}

func newRegexSubmatchModifier(regex string, format string, returnOnMismatch bool) (*regexSubmatchModifier, error) {
	var r regexSubmatchModifier
	re, err := regexp.Compile(regex)
	if err != nil {
		return nil, errors.Wrap(err, "regex compile failed")
	}
	r.regex = re
	r.format = format
	r.returnOnMismatch = returnOnMismatch
	return &r, nil
}

func (o *regexSubmatchModifier) modify(_ context.Context, v value.Value) (value.Value, error) {
	subMatches := o.regex.FindStringSubmatch(v.String())
	if subMatches == nil {
		return value.Empty(), errors.New("regex does not match")
	}
	return value.New(o.regex.ReplaceAllString(subMatches[0], o.format)), nil
}

func (o *regexSubmatchModifier) returnOnError() bool {
	return o.returnOnMismatch
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
	readValueReader Reader
	format          string
}

func (r *insertReadValueModifier) modify(ctx context.Context, v value.Value) (value.Value, error) {
	readValue, err := r.readValueReader.GetProperty(ctx)
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
	switchMode        condition2.MatchMode
	cases             []stringSwitchCase
}

type stringSwitchCase struct {
	caseString string
	operators  Operators
}

func (w *genericStringSwitch) switchOperate(ctx context.Context, s value.Value) (value.Value, error) {
	switchValue, err := w.switchValueGetter.getSwitchValue(ctx, s)
	if err != nil {
		return value.Empty(), errors.Wrap(err, "failed to get switch string")
	}
	switchString := switchValue.String()
	for _, c := range w.cases {
		b, err := condition2.MatchStrings(ctx, switchString, w.switchMode, c.caseString)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("error during match strings")
			continue
		}
		if !b {
			log.Ctx(ctx).Debug().Err(err).Msg("string does not match")
			continue
		}
		x, err := c.operators.Apply(ctx, s)
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
