package parser

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/pkg/errors"
	"reflect"
)

type jsonParser interface {
	ToJSON() ([]byte, error)
}

type xmlParser interface {
	ToXML() ([]byte, error)
}

type humanReadableParser interface {
	ToHumanReadable() ([]byte, error)
}

type toCheckPluginOutputParser interface {
	ToCheckPluginOutput() ([]byte, error)
}

// Parse parses the object into the desired format
func Parse(i interface{}, format string) ([]byte, error) {
	switch format {
	case "json":
		return ToJSON(i)
	case "xml":
		return ToXML(i)
	case "check-plugin":
		return ToCheckPluginOutput(i)
	default:
		return ToHumanReadable(i)
	}
}

// ToXML parses the object to XML.
func ToXML(i interface{}) ([]byte, error) {
	i = checkIfError(i)
	if p, ok := i.(xmlParser); ok {
		return p.ToXML()
	}
	responseString, err := xml.Marshal(i)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal to xml")
	}
	return responseString, nil
}

// ToJSON parses the object to JSON.
func ToJSON(i interface{}) ([]byte, error) {
	i = checkIfError(i)
	if p, ok := i.(jsonParser); ok {
		return p.ToJSON()
	}
	responseString, err := json.Marshal(i)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal to json")
	}
	return responseString, nil
}

//ToHumanReadable parses the object to a human readable format.
func ToHumanReadable(i interface{}) ([]byte, error) {
	i = checkIfError(i)
	if p, ok := i.(humanReadableParser); ok {
		return p.ToHumanReadable()
	}
	if i == nil {
		return []byte("null"), nil
	}
	readable := toHumanReadable(reflect.ValueOf(i), 0)
	return bytes.TrimSpace(readable), nil
}

// ToCheckPluginOutput parses the object to a check plugin format.
func ToCheckPluginOutput(i interface{}) ([]byte, error) {
	if p, ok := i.(toCheckPluginOutputParser); ok {
		return p.ToCheckPluginOutput()
	}
	return nil, errors.New("object cannot be passed to check plugin output")
}

// ToStruct parses the formatted content into the struct with the correct unmarshal method.
func ToStruct(contents []byte, format string, i interface{}) error {
	switch format {
	case "json":
		d := json.NewDecoder(bytes.NewReader(contents))
		d.UseNumber()
		return d.Decode(i)
	case "xml":
		return xml.Unmarshal(contents, i)
	default:
		return fmt.Errorf("invalid format '%s'", format)
	}
}

func checkIfError(i interface{}) interface{} {
	if err, ok := i.(error); ok {
		i = tholaerr.OutputError{Error: err.Error()}
	}
	return i
}
