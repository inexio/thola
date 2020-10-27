package value

import (
	"fmt"
	"strconv"
)

// Value represents a value that was read out from a device.
type Value string

// New creates a new value
func New(i interface{}) Value {
	return Value(fmt.Sprint(i))
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
