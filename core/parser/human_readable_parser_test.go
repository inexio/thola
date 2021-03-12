package parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestStruct struct {
	Number uint
	Name   string
	Array  []float64
}

func TestToHumanReadableString(t *testing.T) {
	output, err := ToHumanReadable("foobla")
	assert.Nil(t, err)
	assert.Equal(t, "foobla", string(output))
}

func TestToHumanReadableFloat(t *testing.T) {
	output, err := ToHumanReadable(0.1)
	assert.Nil(t, err)
	assert.Equal(t, "0.1", string(output))
}

func TestToHumanReadablePointer1(t *testing.T) {
	var i uint64
	i = 2
	output, err := ToHumanReadable(&i)
	assert.Nil(t, err)
	assert.Equal(t, "2", string(output))
}

func TestToHumanReadablePointer2(t *testing.T) {
	var str []string
	str = nil
	output, err := ToHumanReadable(&str)
	assert.Nil(t, err)
	assert.Equal(t, "", string(output))
}

func TestToHumanReadableArray(t *testing.T) {
	output, err := ToHumanReadable([]string{"a", "b", "c"})
	assert.Nil(t, err)
	assert.Equal(t, "[3] a b c", string(output))
}

func TestToHumanReadableMap1(t *testing.T) {
	output, err := ToHumanReadable(map[int]string{1: "one"})
	assert.Nil(t, err)
	assert.Equal(t, "(1) \n"+
		"1: one", string(output))
}

func TestToHumanReadableMap2(t *testing.T) {
	output, err := ToHumanReadable(map[int]TestStruct{
		1: {Number: 5}})
	assert.Nil(t, err)
	assert.Equal(t, "(1) \n"+
		"1: \n"+
		"  Number: 5", string(output))
}

func TestToHumanReadableStruct(t *testing.T) {
	mystruct := TestStruct{Number: 5, Name: "MyName", Array: []float64{0.1, 0.2}}
	output, err := ToHumanReadable(mystruct)
	assert.Nil(t, err)
	assert.Equal(t, "Number: 5\n"+
		"Name: MyName\n"+
		"Array: [2] 0.1 0.2", string(output))
}
