package deviceclass

import (
	"context"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/inexio/thola/core/utility"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"regexp"
	"strings"
)

type condition interface {
	check(ctx context.Context) (bool, error)
}

// Condition is a single condition.
type Condition struct {
	Type      string    `yaml:"type"`
	MatchMode matchMode `yaml:"match_mode" mapstructure:"match_mode"`
	Value     []string  `yaml:"values" mapstructure:"values"`
}

// ConditionSet defines a set of conditions.
type ConditionSet struct {
	LogicalOperator logicalOperator
	Conditions      []condition
}

func (c *ConditionSet) check(ctx context.Context) (bool, error) {
	log.Ctx(ctx).Trace().Msg("starting with matching condition set (OR)")
	for _, condition := range c.Conditions {
		match, err := condition.check(ctx)
		if err != nil {
			return false, errors.Wrap(err, "error during match condition")
		}
		if match && c.LogicalOperator == "OR" {
			log.Ctx(ctx).Trace().Msg("condition set matches (one OR condition matched)")
			return true, nil
		}
		if !match && c.LogicalOperator == "AND" {
			log.Ctx(ctx).Trace().Msg("condition set does not match (one AND condition does not match)")
			return false, nil
		}
	}
	if c.LogicalOperator == "AND" {
		log.Ctx(ctx).Trace().Msg("condition set matches (all AND condition matched)")
		return true, nil
	}
	//c.logicalOperator == OR
	log.Ctx(ctx).Trace().Msg("condition set does not match (no OR condition matched)")
	return false, nil
}

// SnmpCondition is a condition based on snmp.
type SnmpCondition struct {
	Condition                    `yaml:",inline" mapstructure:",squash"`
	network.SNMPGetConfiguration `yaml:",inline" mapstructure:",squash"`
}

func (s *SnmpCondition) check(ctx context.Context) (bool, error) {
	if s.Type == "snmpget" {
		logger := log.Ctx(ctx).With().Str("condition", "snmp").Str("condition_type", s.Type).Str("match_mode", string(s.MatchMode)).Str("oid", string(s.OID)).Logger()
		ctx = logger.WithContext(ctx)
	} else {
		logger := log.Ctx(ctx).With().Str("condition", "snmp").Str("condition_type", s.Type).Str("match_mode", string(s.MatchMode)).Logger()
		ctx = logger.WithContext(ctx)
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		log.Ctx(ctx).Trace().Bool("condition_matched", false).Msg("no snmp connection data available")
		return false, nil
	}
	var val string
	var err error

	if s.Type == "SysDescription" {
		val, err = con.SNMP.GetSysDescription(ctx)
		if err != nil {
			if tholaerr.IsNotFoundError(err) {
				log.Ctx(ctx).Trace().Err(err).Msg("sysDescription is not available for snmp agent")
				return false, nil
			}
			return false, errors.Wrap(err, "failed to get SysDescription")
		}
	} else if s.Type == "SysObjectID" {
		val, err = con.SNMP.GetSysObjectID(ctx)
		if err != nil {
			if tholaerr.IsNotFoundError(err) {
				log.Ctx(ctx).Trace().Err(err).Msg("sysObjectID is not available for snmp agent")
				return false, nil
			}
			return false, errors.Wrap(err, "failed to get SysObjectID")
		}
	} else if s.Type == "snmpget" {
		oid := string(s.OID)
		response, err := con.SNMP.SnmpClient.SNMPGet(ctx, oid)
		if err != nil {
			if tholaerr.IsNotFoundError(err) {
				log.Ctx(ctx).Trace().Err(err).Msg("snmpget returned no result")
				return false, nil
			}
			log.Ctx(ctx).Error().Err(err).Msg("error during snmpget")
			return false, errors.Wrap(err, "snmpget request returned an error")
		}
		val, err = response[0].GetValueBySNMPGetConfiguration(s.SNMPGetConfiguration)
		if err != nil {
			if tholaerr.IsNotFoundError(err) {
				return false, nil
			}
			return false, err
		}

	} else {
		return false, errors.New("invalid condition type")
	}

	return matchStrings(ctx, val, s.MatchMode, s.Value...)
}

func (s *SnmpCondition) validate() error {
	err := s.MatchMode.validate()
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

// HTTPCondition is a condition based on http.
type HTTPCondition struct {
	Condition `yaml:",inline" mapstructure:",squash"`
	URI       string
}

func (s *HTTPCondition) check(ctx context.Context) (bool, error) {
	logger := log.Ctx(ctx).With().Str("condition", "http").Str("condition_type", s.Type).Str("match_mode", string(s.MatchMode)).Str("uri", s.URI).Logger()
	ctx = logger.WithContext(ctx)

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.HTTP == nil {
		log.Ctx(ctx).Trace().Bool("condition_matched", false).Msg("no http connection data available")
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
					log.Ctx(ctx).Trace().Err(err).Str("protocol", con.HTTP.HTTPClient.GetProtocolString()).Int("port", port).Msg("http(s) request returned error")
					if tholaerr.IsNetworkError(err) {
						continue
					}
					return false, errors.Wrap(err, "non-network error during http(s) request!")
				}
				log.Ctx(ctx).Trace().Str("protocol", con.HTTP.HTTPClient.GetProtocolString()).Int("port", port).Msg("http(s) request was successful")
				value = string(r.Body())

				matched, err := matchStrings(ctx, value, s.MatchMode, s.Value...)
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

func (s *HTTPCondition) validate() error {
	err := s.MatchMode.validate()
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

// VendorCondition is a condition based on a vendor.
type VendorCondition struct {
	Condition `yaml:",inline" mapstructure:",squash"`
}

func (m *VendorCondition) check(ctx context.Context) (bool, error) {
	properties, ok := device.DevicePropertiesFromContext(ctx)
	if !ok {
		return false, errors.New("no properties found in context")
	}
	if properties.Properties.Vendor == nil {
		return false, tholaerr.NewPreConditionError("vendor has not yet been determined")
	}
	return matchStrings(ctx, *properties.Properties.Vendor, m.MatchMode, m.Value...)
}

// ModelCondition is a condition based on a model.
type ModelCondition struct {
	Condition `yaml:",inline" mapstructure:",squash"`
}

func (m *ModelCondition) check(ctx context.Context) (bool, error) {
	properties, ok := device.DevicePropertiesFromContext(ctx)
	if !ok {
		return false, errors.New("no properties found in context")
	}
	if properties.Properties.Model == nil {
		return false, tholaerr.NewPreConditionError("model has not yet been determined")
	}
	return matchStrings(ctx, *properties.Properties.Model, m.MatchMode, m.Value...)
}

// ModelSeriesCondition is a condition based on a model series.
type ModelSeriesCondition struct {
	Condition `yaml:",inline" mapstructure:",squash"`
}

func (m *ModelSeriesCondition) check(ctx context.Context) (bool, error) {
	properties, ok := device.DevicePropertiesFromContext(ctx)
	if !ok {
		return false, errors.New("no properties found in context")
	}
	if properties.Properties.ModelSeries == nil {
		return false, tholaerr.NewPreConditionError("model series has not yet been determined")
	}
	return matchStrings(ctx, *properties.Properties.ModelSeries, m.MatchMode, m.Value...)
}

type alwaysTrueCondition struct {
}

func (a *alwaysTrueCondition) check(_ context.Context) (bool, error) {
	return true, nil
}

func matchStrings(ctx context.Context, str string, mode matchMode, matches ...string) (bool, error) {
	switch mode {
	case "contains":
		for _, match := range matches {
			if strings.Contains(str, match) {
				log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	case "!contains":
		for _, match := range matches {
			if !strings.Contains(str, match) {
				log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	case "startsWith":
		for _, match := range matches {
			if strings.HasPrefix(str, match) {
				log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	case "!startsWith":
		for _, match := range matches {
			if !strings.HasPrefix(str, match) {
				log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	case "regex":
		for _, match := range matches {
			i, err := regexp.MatchString(match, str)
			if err != nil {
				log.Ctx(ctx).Trace().Err(err).Msg("error during regex match")
				return false, errors.Wrap(err, "error during regex match")
			}
			if i {
				log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	case "!regex":
		for _, match := range matches {
			i, err := regexp.MatchString(match, str)
			if err != nil {
				log.Ctx(ctx).Trace().Err(err).Msg("error during regex match")
				return false, errors.Wrap(err, "error during regex match")
			}
			if !i {
				log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	case "equals":
		for _, match := range matches {
			if str == match {
				log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	case "!equals":
		for _, match := range matches {
			if str != match {
				log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", true).Msg("value matched!")
				return true, nil
			}
			log.Ctx(ctx).Trace().Str("match_value", match).Str("received_value", str).Bool("matched", false).Msg("value did not match!")
		}
	default:
		return false, errors.New("invalid match type")
	}
	return false, nil
}
