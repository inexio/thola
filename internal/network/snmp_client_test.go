package network

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOID_Cmp_smaller(t *testing.T) {
	res, err := OID("1.1").Cmp("1.2")
	assert.NoError(t, err)
	assert.Equal(t, -1, res)
}

func TestOID_Cmp_smaller_dot(t *testing.T) {
	res, err := OID(".1.1").Cmp("1.2")
	assert.NoError(t, err)
	assert.Equal(t, -1, res)
}

func TestOID_Cmp_smaller_big(t *testing.T) {
	res, err := OID("1.2").Cmp("1.11")
	assert.NoError(t, err)
	assert.Equal(t, -1, res)
}

func TestOID_Cmp_bigger(t *testing.T) {
	res, err := OID("1.2").Cmp("1.1")
	assert.NoError(t, err)
	assert.Equal(t, 1, res)
}

func TestOID_Cmp_bigger_dot(t *testing.T) {
	res, err := OID("1.2").Cmp(".1.1")
	assert.NoError(t, err)
	assert.Equal(t, 1, res)
}

func TestOID_Cmp_bigger_big(t *testing.T) {
	res, err := OID("1.244").Cmp("1.101")
	assert.NoError(t, err)
	assert.Equal(t, 1, res)
}

func TestOID_Cmp_equals(t *testing.T) {
	res, err := OID("1.1").Cmp("1.1")
	assert.NoError(t, err)
	assert.Equal(t, 0, res)
}

func TestOID_Cmp_shorterOID(t *testing.T) {
	res, err := OID("1").Cmp("1.1")
	assert.NoError(t, err)
	assert.Equal(t, -1, res)
}

func TestOID_Cmp_longerOID(t *testing.T) {
	res, err := OID("1.1").Cmp("1")
	assert.NoError(t, err)
	assert.Equal(t, 1, res)
}

func TestOID_AddIndex(t *testing.T) {
	assert.Equal(t, OID("1.1"), OID("1.").AddIndex("1"))
}

func TestOID_AddIndex_noDot(t *testing.T) {
	assert.Equal(t, OID("1.1"), OID("1").AddIndex("1"))
}

func TestOID_AddIndex_doubleDot(t *testing.T) {
	assert.Equal(t, OID("1.1"), OID("1.").AddIndex(".1"))
}
