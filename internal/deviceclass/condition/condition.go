package condition

import (
	"context"
	"fmt"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/inexio/thola/internal/utility"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"regexp"
	"strings"
)

func Interface2Condition(i interface{}, task RelatedTask) (Condition, error) {
	m, ok := i.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("failed to convert interface to map[interface{}]interface{}")
	}

	var stringType string
	if _, ok := m["type"]; ok {
		stringType, ok = m["type"].(string)
		if !ok {
			return nil, errors.New("condition type needs to be a string")
		}
	} else {
		// if condition type is empty, and it has conditions and optionally a logical operator,
		// and no other attributes, then it will be considered as a conditionSet per default
		if _, ok = m["conditions"]; ok {
			// if there is only "conditions" in the map or only "conditions" and "logical_operator", nothing else
			if _, ok = m["logical_operator"]; (ok && len(m) == 2) || len(m) == 1 {
				stringType = "conditionSet"
			} else {
				return nil, errors.New("no condition type set and attributes do not match conditionSet")
			}
		} else {
			return nil, errors.New("no condition type set and attributes do not match conditionSet")
		}
	}

	if stringType == "conditionSet" {
		var yamlConditionSet yamlConditionSet
		err := mapstructure.Decode(i, &yamlConditionSet)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode conditionSet")
		}
		return yamlConditionSet.convert()
	}
	//SNMP SnmpCondition Types
	if stringType == "SysObjectID" || stringType == "SysDescription" || stringType == "snmpget" {
		var condition snmpCondition
		err := mapstructure.Decode(i, &condition)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode Condition")
		}
		err = condition.validate()
		if err != nil {
			return nil, errors.Wrap(err, "invalid snmp condition")
		}
		return &condition, nil
	}
	//HTTP
	if stringType == "HttpGetBody" {
		var condition httpCondition
		err := mapstructure.Decode(i, &condition)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode condition")
		}
		err = condition.validate()
		if err != nil {
			return nil, errors.Wrap(err, "invalid http condition")
		}
		return &condition, nil
	}

	if stringType == "Vendor" {
		if task <= PropertyVendor {
			return nil, errors.New("cannot use vendor condition, vendor is not available here yet")
		}
		var condition vendorCondition
		err := mapstructure.Decode(i, &condition)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode condition")
		}
		return &condition, nil
	}

	if stringType == "Model" {
		if task <= PropertyModel {
			return nil, errors.New("cannot use model condition, model is not available here yet")
		}
		var condition modelCondition
		err := mapstructure.Decode(i, &condition)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode condition")
		}
		return &condition, nil
	}

	if stringType == "ModelSeries" {
		if task <= PropertyModelSeries {
			return nil, errors.New("cannot use model series condition, model series is not available here yet")
		}
		var condition modelSeriesCondition
		err := mapstructure.Decode(i, &condition)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode condition")
		}
		return &condition, nil
	}
	return nil, fmt.Errorf("invalid condition type '%s'", stringType)
}

type Condition interface {
	Check(ctx context.Context) (bool, error)
	ContainsUniqueRequest() bool
}

// singleCondition is a single condition.
type singleCondition struct {
	Type      string
	MatchMode MatchMode `mapstructure:"match_mode"`
	Value     []string  `mapstructure:"values"`
}

// multipleConditions defines a set of conditions.
type multipleConditions struct {
	LogicalOperator LogicalOperator
	Conditions      []Condition
}

func (c *multipleConditions) Check(ctx context.Context) (bool, error) {
	log.Ctx(ctx).Debug().Msg("starting with matching condition set (OR)")
	for _, condition := range c.Conditions {
		match, err := condition.Check(ctx)
		if err != nil {
			return false, errors.Wrap(err, "error during match condition")
		}
		if match && c.LogicalOperator == "OR" {
			log.Ctx(ctx).Debug().Msg("condition set matches (one OR condition matched)")
			return true, nil
		}
		if !match && c.LogicalOperator == "AND" {
			log.Ctx(ctx).Debug().Msg("condition set does not match (one AND condition does not match)")
			return false, nil
		}
	}
	if c.LogicalOperator == "AND" {
		log.Ctx(ctx).Debug().Msg("condition set matches (all AND condition matched)")
		return true, nil
	}
	//c.logicalOperator == OR
	log.Ctx(ctx).Debug().Msg("condition set does not match (no OR condition matched)")
	return false, nil
}

func (c *multipleConditions) ContainsUniqueRequest() bool {
	for _, con := range c.Conditions {
		if con.ContainsUniqueRequest() {
			return true
		}
	}
	return false
}

// snmpCondition is a condition based on snmp.
type snmpCondition struct {
	singleCondition              `mapstructure:",squash"`
	network.SNMPGetConfiguration `mapstructure:",squash"`
}

func (s *snmpCondition) Check(ctx context.Context) (bool, error) {
	if s.Type == "snmpget" {
		logger := log.Ctx(ctx).With().Str("condition", "snmp").Str("condition_type", s.Type).Str("match_mode", string(s.MatchMode)).Str("oid", string(s.OID)).Logger()
		ctx = logger.WithContext(ctx)
	} else {
		logger := log.Ctx(ctx).With().Str("condition", "snmp").Str("condition_type", s.Type).Str("match_mode", string(s.MatchMode)).Logger()
		ctx = logger.WithContext(ctx)
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		log.Ctx(ctx).Debug().Bool("condition_matched", false).Msg("no snmp connection data available")
		return false, nil
	}
	var val string
	var err error

	if s.Type == "SysDescription" {
		val, err = con.SNMP.GetSysDescription(ctx)
		if err != nil {
			if tholaerr.IsNotFoundError(err) {
				log.Ctx(ctx).Debug().Err(err).Msg("sysDescription is not available for snmp agent")
				return false, nil
			}
			return false, errors.Wrap(err, "failed to get SysDescription")
		}
	} else if s.Type == "SysObjectID" {
		val, err = con.SNMP.GetSysObjectID(ctx)
		if err != nil {
			if tholaerr.IsNotFoundError(err) {
				log.Ctx(ctx).Debug().Err(err).Msg("sysObjectID is not available for snmp agent")
				return false, nil
			}
			return false, errors.Wrap(err, "failed to get SysObjectID")
		}
	} else if s.Type == "snmpget" {
		response, err := con.SNMP.SnmpClient.SNMPGet(ctx, s.OID)
		if err != nil {
			if tholaerr.IsNotFoundError(err) {
				log.Ctx(ctx).Debug().Err(err).Msg("snmpget returned no result")
				return false, nil
			}
			log.Ctx(ctx).Error().Err(err).Msg("error during snmpget")
			return false, errors.Wrap(err, "snmpget request returned an error")
		}
		value, err := response[0].GetValueBySNMPGetConfiguration(s.SNMPGetConfiguration)
		if err != nil {
			if tholaerr.IsNotFoundError(err) {
				return false, nil
			}
			return false, err
		}
		val = value.String()

	} else {
		return false, errors.New("invalid condition type")
	}

	return MatchStrings(ctx, val, s.MatchMode, s.Value...)
}

func (s *snmpCondition) validate() error {
	err := s.MatchMode.Validate()
	if err != nil {
		return errors.Wrap(err, "invalid matchmode")
	}
	if s.Type != "SysObjectID" && s.Type != "SysDescription" && s.Type != "snmpget" {
		return errors.New("invalid condition type for snmp condition (type = " + s.Type + ")")
	}
	if len(s.Value) == 0 {
		return errors.New("no values defined")
	}
	if s.Type == "snmpget" {
		if s.OID == "" {
			return errors.New("oid is missing (type = snmpget)")
		}
		err = s.OID.Validate()
		if err != nil {
			return errors.New("invalid oid")
		}
	} else {
		// if snmp condition is not "snmpget", SNMPGetConfiguration needs to be a zero value
		if s.SNMPGetConfiguration != (network.SNMPGetConfiguration{}) {
			return errors.New("snmpget configuration data (oid, use_raw_result, ...) are only allowed for snmpget conditions")
		}
	}

	return nil
}

func (s *snmpCondition) ContainsUniqueRequest() bool {
	return s.Type == "snmpget"
}

// httpCondition is a condition based on http.
type httpCondition struct {
	singleCondition `mapstructure:",squash"`
	URI             string
}

func (s *httpCondition) Check(ctx context.Context) (bool, error) {
	logger := log.Ctx(ctx).With().Str("condition", "http").Str("condition_type", s.Type).Str("match_mode", string(s.MatchMode)).Str("uri", s.URI).Logger()
	ctx = logger.WithContext(ctx)

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.HTTP == nil {
		log.Ctx(ctx).Debug().Bool("condition_matched", false).Msg("no http connection data available")
		return false, nil //TODO: throw error and catch it or just return false?
	}
	var value string
	if s.Type == "HttpGetBody" {
		for _, useHTTPS := range []bool{true, false} {
			con.HTTP.HTTPClient.UseHTTPS(useHTTPS)
			for _, port := range utility.IfThenElse(useHTTPS, con.HTTP.ConnectionData.HTTPSPorts, con.HTTP.ConnectionData.HTTPPorts).([]int) {
				con.HTTP.HTTPClient.SetPort(port)
				r, err := con.HTTP.HTTPClient.Request(ctx, "GET", s.URI, "", nil, nil)
				if err != nil {
					log.Ctx(ctx).Debug().Err(err).Str("protocol", con.HTTP.HTTPClient.GetProtocolString()).Int("port", port).Msg("http(s) request returned error")
					if tholaerr.IsNetworkError(err) {
						continue
					}
					return false, errors.Wrap(err, "non-network error during http(s) request!")
				}
				log.Ctx(ctx).Debug().Str("protocol", con.HTTP.HTTPClient.GetProtocolString()).Int("port", port).Msg("http(s) request was successful")
				value = string(r.Body())

				matched, err := MatchStrings(ctx, value, s.MatchMode, s.Value...)
				if err != nil {
					return false, errors.Wrap(err, "error during match Strings")
				}
				//if matched return true
				if matched {
					return true, nil
				}
				//if result does not match, try the next http port
			}
		}
	} else {
		return false, errors.New("invalid condition type")
	}
	return false, nil
}

func (s *httpCondition) ContainsUniqueRequest() bool {
	return true
}

func (s *httpCondition) validate() error {
	err := s.MatchMode.Validate()
	if err != nil {
		return errors.Wrap(err, "invalid matchmode")
	}
	if s.Type != "HttpGetBody" {
		return errors.New("invalid condition type for http condition (type = " + s.Type + ")")
	}
	if len(s.Value) == 0 {
		return errors.New("no values defined")
	}

	return nil
}

// vendorCondition is a condition based on a vendor.
type vendorCondition struct {
	singleCondition `mapstructure:",squash"`
}

func (m *vendorCondition) Check(ctx context.Context) (bool, error) {
	properties, ok := device.DevicePropertiesFromContext(ctx)
	if !ok {
		return false, errors.New("no properties found in context")
	}
	if properties.Properties.Vendor == nil {
		return false, tholaerr.NewPreConditionError("vendor has not yet been determined")
	}
	return MatchStrings(ctx, *properties.Properties.Vendor, m.MatchMode, m.Value...)
}

func (m *vendorCondition) ContainsUniqueRequest() bool {
	return false
}

// modelCondition is a condition based on a model.
type modelCondition struct {
	singleCondition `mapstructure:",squash"`
}

func (m *modelCondition) Check(ctx context.Context) (bool, error) {
	properties, ok := device.DevicePropertiesFromContext(ctx)
	if !ok {
		return false, errors.New("no properties found in context")
	}
	if properties.Properties.Model == nil {
		return false, tholaerr.NewPreConditionError("model has not yet been determined")
	}
	return MatchStrings(ctx, *properties.Properties.Model, m.MatchMode, m.Value...)
}

func (m *modelCondition) ContainsUniqueRequest() bool {
	return false
}

// modelSeriesCondition is a condition based on a model series.
type modelSeriesCondition struct {
	singleCondition `mapstructure:",squash"`
}

func (m *modelSeriesCondition) Check(ctx context.Context) (bool, error) {
	properties, ok := device.DevicePropertiesFromContext(ctx)
	if !ok {
		return false, errors.New("no properties found in context")
	}
	if properties.Properties.ModelSeries == nil {
		return false, tholaerr.NewPreConditionError("model series has not yet been determined")
	}
	return MatchStrings(ctx, *properties.Properties.ModelSeries, m.MatchMode, m.Value...)
}

func (m *modelSeriesCondition) ContainsUniqueRequest() bool {
	return false
}

type alwaysTrueCondition struct {
}

func (a *alwaysTrueCondition) Check(_ context.Context) (bool, error) {
	return true, nil
}

func (a *alwaysTrueCondition) ContainsUniqueRequest() bool {
	return false
}

func GetAlwaysTrueCondition() Condition {
	return &alwaysTrueCondition{}
}

type yamlConditionSet struct {
	LogicalOperator LogicalOperator `mapstructure:"logical_operator"`
	Conditions      []interface{}
}

func (y *yamlConditionSet) convert() (Condition, error) {
	err := y.validate()
	if err != nil {
		return nil, errors.Wrap(err, "invalid yaml condition set")
	}
	var conditionSet multipleConditions
	for _, condition := range y.Conditions {
		matcher, err := Interface2Condition(condition, ClassifyDevice)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert interface to condition")
		}
		conditionSet.Conditions = append(conditionSet.Conditions, matcher)
	}
	conditionSet.LogicalOperator = y.LogicalOperator
	return &conditionSet, nil
}

func (y *yamlConditionSet) validate() error {
	if len(y.Conditions) == 0 {
		return errors.New("empty condition array")
	}
	err := y.LogicalOperator.validate()
	if err != nil {
		if y.LogicalOperator == "" {
			y.LogicalOperator = "OR" // default logical operator is always OR
		}
		return errors.Wrap(err, "invalid logical operator")
	}
	return nil
}

// LogicalOperator represents a logical operator (OR or AND).
type LogicalOperator string

func (l *LogicalOperator) validate() error {
	if *l != "AND" && *l != "OR" {
		return errors.New(string("unknown logical operator '" + *l + "'"))
	}
	return nil
}

// MatchMode represents a match mode that is used to match a condition.
type MatchMode string

func (m *MatchMode) Validate() error {
	if *m != "contains" && *m != "!contains" && *m != "startsWith" && *m != "!startsWith" && *m != "regex" && *m != "!regex" && *m != "equals" && *m != "!equals" {
		return errors.New(string("unknown matchmode \"" + *m + "\""))
	}
	return nil
}

func MatchStrings(ctx context.Context, str string, mode MatchMode, matches ...string) (bool, error) {
	switch mode {
	case "contains":
		for _, match := range matches {
			if strings.Contains(str, match) {
				log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	case "!contains":
		for _, match := range matches {
			if !strings.Contains(str, match) {
				log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	case "startsWith":
		for _, match := range matches {
			if strings.HasPrefix(str, match) {
				log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	case "!startsWith":
		for _, match := range matches {
			if !strings.HasPrefix(str, match) {
				log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	case "regex":
		for _, match := range matches {
			i, err := regexp.MatchString(match, str)
			if err != nil {
				log.Ctx(ctx).Debug().Err(err).Msg("error during regex match")
				return false, errors.Wrap(err, "error during regex match")
			}
			if i {
				log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	case "!regex":
		for _, match := range matches {
			i, err := regexp.MatchString(match, str)
			if err != nil {
				log.Ctx(ctx).Debug().Err(err).Msg("error during regex match")
				return false, errors.Wrap(err, "error during regex match")
			}
			if !i {
				log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	case "equals":
		for _, match := range matches {
			if str == match {
				log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	case "!equals":
		for _, match := range matches {
			if str != match {
				log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Debug().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	default:
		return false, errors.New("invalid match type")
	}
	return false, nil
}
