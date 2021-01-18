package value

import (
	"fmt"
	"github.com/pkg/errors"
	"math/big"
	"strconv"
)

// Value
//
// Value represents a value that was read out from a device.
//
// swagger:model
type Value string

// New creates a new value
func New(i interface{}) Value {
	var v Value
	switch t := i.(type) {
	case []byte:
		v = Value(t)
	case string:
		v = Value(t)
	default:
		v = Value(fmt.Sprint(t))
	}
	return v
}

// Empty returns the an empty value.
func Empty() Value {
	return ""
}

// String returns the value as a string
func (v *Value) String() string {
	return string(*v)
}

// Float64 returns the value as a float 64
func (v *Value) Float64() (float64, error) {
	return strconv.ParseFloat(string(*v), 64)
}

// Int returns the value as an int
func (v *Value) Int() (int, error) {
	return strconv.Atoi(string(*v))
}

// Bool returns the value as a bool
func (v *Value) Bool() (bool, error) {
	return strconv.ParseBool(string(*v))
}

func (v *Value) IsEmpty() bool {
	return v == nil || v.String() == ""
}

func (v *Value) Cmp(val Value) (int, error) {
	var v1, v2 big.Float
	_, _, err := v1.Parse(v.String(), 10)
	if err != nil {
		return 0, errors.Wrap(err, "can't parse value")
	}

	_, _, err = v2.Parse(val.String(), 10)
	if err != nil {
		return 0, errors.Wrap(err, "can't parse value")
	}

	return v1.Cmp(&v2), nil
}
