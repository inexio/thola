package value

import (
	"fmt"
	"github.com/pkg/errors"
	"math/big"
	"strconv"
)

// Value represents a value of any type.
type Value interface {
	String() string
	Float64() (float64, error)
	Int() (int, error)
	UInt64() (uint64, error)
	Bool() (bool, error)
	IsEmpty() bool
	Cmp(val Value) (int, error)
}

// value
//
// value represents a value that was read out from a device.
//
// swagger:model
type value string

// New creates a new value
func New(i interface{}) Value {
	var v value
	switch t := i.(type) {
	case []byte:
		v = value(t)
	case string:
		v = value(t)
	case nil:
		v = ""
	default:
		v = value(fmt.Sprint(t))
	}
	return v
}

// String returns the value as a string
func (v value) String() string {
	return string(v)
}

// Float64 returns the value as a float 64
func (v value) Float64() (float64, error) {
	return strconv.ParseFloat(string(v), 64)
}

// Int returns the value as an int
func (v value) Int() (int, error) {
	return strconv.Atoi(string(v))
}

// UInt64 returns the value as an uint64
func (v value) UInt64() (uint64, error) {
	return strconv.ParseUint(string(v), 10, 64)
}

// Bool returns the value as a bool
func (v value) Bool() (bool, error) {
	return strconv.ParseBool(string(v))
}

// IsEmpty returns if the value is empty
func (v value) IsEmpty() bool {
	return v == ""
}

// Cmp compares two values
//   -1 if receiver < val
//    0 if receiver == val
//   +1 if receiver > val
func (v value) Cmp(val Value) (int, error) {
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
