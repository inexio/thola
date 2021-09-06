package request

import (
	"context"
	"github.com/pkg/errors"
	"regexp"
)

// CheckInterfaceMetricsRequest
//
// CheckInterfaceRequest is a the request struct for the check interface metrics request.
//
// swagger:model
type CheckInterfaceMetricsRequest struct {
	PrintInterfaces       bool    `yaml:"print_interfaces" json:"print_interfaces" xml:"print_interfaces"`
	IfDescrRegex          *string `yaml:"ifDescr_regex" json:"ifDescr_regex" xml:"ifDescr_regex"`
	ifDescrRegex          *regexp.Regexp
	IfDescrRegexReplace   *string  `yaml:"ifDescr_regex_replace" json:"ifDescr_regex_replace" xml:"ifDescr_regex_replace"`
	IfTypeFilter          []string `yaml:"ifType_filter" json:"ifType_filter" xml:"ifType_filter"`
	IfNameFilter          []string `yaml:"ifName_filter" json:"ifName_filter" xml:"ifName_filter"`
	IfDescrFilter         []string `yaml:"ifDescr_filter" json:"ifDescr_filter" xml:"ifDescr_filter"`
	SNMPGetsInsteadOfWalk bool     `yaml:"snmp_gets_instead_of_walk" json:"snmp_gets_instead_of_walk" xml:"snmp_gets_instead_of_walk"`
	CheckDeviceRequest
}

func (r *CheckInterfaceMetricsRequest) validate(ctx context.Context) error {
	if r.IfDescrRegex != nil && r.IfDescrRegexReplace == nil ||
		r.IfDescrRegex == nil && r.IfDescrRegexReplace != nil {
		return errors.New("'ifDescr-regex' and 'ifDescr-regex-replace' must be set together")
	}

	if r.IfDescrRegex != nil {
		regex, err := regexp.Compile(*r.IfDescrRegex)
		if err != nil {
			return errors.Wrap(err, "compiling ifDescrRegex failed")
		}
		r.ifDescrRegex = regex
	}

	return r.CheckDeviceRequest.validate(ctx)
}
