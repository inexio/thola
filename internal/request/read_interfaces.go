package request

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

// ReadInterfacesRequest
//
// ReadInterfacesRequest is the request struct for the read interfaces request.
//
// swagger:model
type ReadInterfacesRequest struct {
	InterfaceOptions
	ReadRequest
}

func (r *ReadInterfacesRequest) validate(ctx context.Context) error {
	if err := r.InterfaceOptions.validate(); err != nil {
		return err
	}
	return r.ReadRequest.validate(ctx)
}

// ReadInterfacesResponse
//
// ReadInterfacesResponse is the request struct for the read interfaces response.
//
// swagger:model
type ReadInterfacesResponse struct {
	Interfaces []device.Interface `yaml:"interfaces" json:"interfaces" xml:"interfaces"`
	ReadResponse
}

// InterfaceOptions
//
// InterfaceOptions is the request struct for the options of an interface request.
//
// swagger:model
type InterfaceOptions struct {
	// If you only want specific values of the interfaces you can specify them here.
	Values                []string `yaml:"values" json:"values" xml:"values"`
	IfDescrRegex          string   `yaml:"ifDescr_regex" json:"ifDescr_regex" xml:"ifDescr_regex"`
	ifDescrRegex          *regexp.Regexp
	IfDescrRegexReplace   string   `yaml:"ifDescr_regex_replace" json:"ifDescr_regex_replace" xml:"ifDescr_regex_replace"`
	IfTypeFilter          []string `yaml:"ifType_filter" json:"ifType_filter" xml:"ifType_filter"`
	IfNameFilter          []string `yaml:"ifName_filter" json:"ifName_filter" xml:"ifName_filter"`
	IfDescrFilter         []string `yaml:"ifDescr_filter" json:"ifDescr_filter" xml:"ifDescr_filter"`
	SNMPGetsInsteadOfWalk bool     `yaml:"snmp_gets_instead_of_walk" json:"snmp_gets_instead_of_walk" xml:"snmp_gets_instead_of_walk"`
}

func (r *InterfaceOptions) validate() error {
	if r.IfDescrRegex != "" && r.IfDescrRegexReplace == "" ||
		r.IfDescrRegex == "" && r.IfDescrRegexReplace != "" {
		return errors.New("'ifDescr-regex' and 'ifDescr-regex-replace' must be set together")
	}
	if r.IfDescrRegex != "" {
		regex, err := regexp.Compile(r.IfDescrRegex)
		if err != nil {
			return errors.Wrap(err, "compiling ifDescrRegex failed")
		}
		r.ifDescrRegex = regex
	}
	return nil
}

func (r *InterfaceOptions) getFilter() []groupproperty.Filter {
	var res []groupproperty.Filter

	for _, f := range r.IfTypeFilter {
		res = append(res, groupproperty.GetGroupFilter([]string{"ifType"}, f))
	}
	for _, f := range r.IfNameFilter {
		res = append(res, groupproperty.GetGroupFilter([]string{"ifName"}, f))
	}
	for _, f := range r.IfDescrFilter {
		res = append(res, groupproperty.GetGroupFilter([]string{"ifDescr"}, f))
	}

	if len(r.Values) > 0 {
		var values [][]string
		for _, fil := range r.Values {
			values = append(values, strings.Split(fil, "/"))
		}
		res = append(res, groupproperty.GetExclusiveValueFilter(values))
	}

	return res
}
