package deviceclass

import (
	"context"
	"github.com/gosnmp/gosnmp"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/network/mocks"
	"github.com/inexio/thola/internal/value"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestDeviceClassOID_readOID tests deviceClassOID.readOid(...) without indices and skipEmpty = false
func TestDeviceClassOID_readOID(t *testing.T) {
	var snmpClient mocks.SNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, "1").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.1", gosnmp.OctetString, "Port 1"),
			network.NewSNMPResponse("1.2", gosnmp.OctetString, "Port 2"),
			network.NewSNMPResponse("1.3", gosnmp.OctetString, "Port 3"),
			network.NewSNMPResponse("1.4", gosnmp.OctetString, ""),
		}, nil)

	sut := deviceClassOID{
		SNMPGetConfiguration: network.SNMPGetConfiguration{
			OID: "1",
		},
	}

	expected := map[int]interface{}{
		1: value.New("Port 1"),
		2: value.New("Port 2"),
		3: value.New("Port 3"),
		4: value.New(""),
	}

	res, err := sut.readOID(ctx, nil, false)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

// TestDeviceClassOID_readOID_skipEmpty tests deviceClassOID.readOid(...) without indices and skipEmpty = true
func TestDeviceClassOID_readOID_skipEmpty(t *testing.T) {
	var snmpClient mocks.SNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, "1").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.1", gosnmp.OctetString, "Port 1"),
			network.NewSNMPResponse("1.2", gosnmp.OctetString, "Port 2"),
			network.NewSNMPResponse("1.3", gosnmp.OctetString, "Port 3"),
			network.NewSNMPResponse("1.4", gosnmp.OctetString, ""),
		}, nil)

	sut := deviceClassOID{
		SNMPGetConfiguration: network.SNMPGetConfiguration{
			OID: "1",
		},
	}

	expected := map[int]interface{}{
		1: value.New("Port 1"),
		2: value.New("Port 2"),
		3: value.New("Port 3"),
	}

	res, err := sut.readOID(ctx, nil, true)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

// TestDeviceClassOID_readOID_withIndices tests deviceClassOID.readOid(...) with indices and skipEmpty = false
func TestDeviceClassOID_readOID_withIndices(t *testing.T) {
	var snmpClient mocks.SNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	ctx = network.NewContextWithSNMPGetsInsteadOfWalk(ctx, true)

	snmpClient.
		On("SNMPGet", ctx, "1.1", "1.2", "1.3", "1.4").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.1", gosnmp.OctetString, "Port 1"),
			network.NewSNMPResponse(".1.2", gosnmp.OctetString, "Port 2"),
			network.NewSNMPResponse(".1.3", gosnmp.OctetString, "Port 3"),
			network.NewSNMPResponse(".1.4", gosnmp.OctetString, ""),
		}, nil)

	sut := deviceClassOID{
		SNMPGetConfiguration: network.SNMPGetConfiguration{
			OID: "1",
		},
	}

	expected := map[int]interface{}{
		1: value.New("Port 1"),
		2: value.New("Port 2"),
		3: value.New("Port 3"),
		4: value.New(""),
	}

	res, err := sut.readOID(ctx, []value.Value{value.New(1), value.New(2), value.New(3), value.New(4)}, false)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

// TestDeviceClassOID_readOID_withIndicesSkipEmpty tests deviceClassOID.readOid(...) with indices and skipEmpty = true
func TestDeviceClassOID_readOID_withIndicesSkipEmpty(t *testing.T) {
	var snmpClient mocks.SNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	ctx = network.NewContextWithSNMPGetsInsteadOfWalk(ctx, true)

	snmpClient.
		On("SNMPGet", ctx, "1.1", "1.2", "1.3", "1.4").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.1", gosnmp.OctetString, "Port 1"),
			network.NewSNMPResponse(".1.2", gosnmp.OctetString, "Port 2"),
			network.NewSNMPResponse(".1.3", gosnmp.OctetString, "Port 3"),
			network.NewSNMPResponse(".1.4", gosnmp.OctetString, ""),
		}, nil)

	sut := deviceClassOID{
		SNMPGetConfiguration: network.SNMPGetConfiguration{
			OID: "1",
		},
	}

	expected := map[int]interface{}{
		1: value.New("Port 1"),
		2: value.New("Port 2"),
		3: value.New("Port 3"),
	}

	res, err := sut.readOID(ctx, []value.Value{value.New(1), value.New(2), value.New(3), value.New(4)}, true)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

// TestDeviceClassOID_readOID_indicesMapping tests deviceClassOID.readOid(...) with an indices mapping and no given indices
func TestDeviceClassOID_readOID_indicesMapping(t *testing.T) {
	var snmpClient mocks.SNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, "1").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.1", gosnmp.OctetString, "3"),
			network.NewSNMPResponse("1.2", gosnmp.OctetString, "2"),
			network.NewSNMPResponse("1.3", gosnmp.OctetString, "1"),
		}, nil)

	snmpClient.
		On("SNMPWalk", ctx, "2").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("2.1", gosnmp.OctetString, "Port 1"),
			network.NewSNMPResponse("2.2", gosnmp.OctetString, "Port 2"),
			network.NewSNMPResponse("2.3", gosnmp.OctetString, "Port 3"),
			network.NewSNMPResponse("2.4", gosnmp.OctetString, "Port 4"),
		}, nil)

	indicesMappingDeviceClassOid := deviceClassOID{
		SNMPGetConfiguration: network.SNMPGetConfiguration{
			OID: "1",
		},
	}

	sut := deviceClassOID{
		SNMPGetConfiguration: network.SNMPGetConfiguration{
			OID: "2",
		},
		indicesMapping: &indicesMappingDeviceClassOid,
	}

	expected := map[int]interface{}{
		1: value.New("Port 3"),
		2: value.New("Port 2"),
		3: value.New("Port 1"),
	}

	res, err := sut.readOID(ctx, nil, false)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

// TestDeviceClassOID_readOID_indicesMapping tests deviceClassOID.readOid(...) with an indices mapping and given indices
func TestDeviceClassOID_readOID_indicesMappingWithIndices(t *testing.T) {
	var snmpClient mocks.SNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, "1").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.1", gosnmp.OctetString, "3"),
			network.NewSNMPResponse("1.2", gosnmp.OctetString, "2"),
			network.NewSNMPResponse("1.3", gosnmp.OctetString, "1"),
		}, nil)

	snmpClient.
		On("SNMPGet", ctx, "2.3", "2.2", "2.1").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".2.3", gosnmp.OctetString, "Port 3"),
			network.NewSNMPResponse(".2.2", gosnmp.OctetString, "Port 2"),
			network.NewSNMPResponse(".2.1", gosnmp.OctetString, "Port 1"),
		}, nil)

	indicesMappingDeviceClassOid := deviceClassOID{
		SNMPGetConfiguration: network.SNMPGetConfiguration{
			OID: "1",
		},
	}

	sut := deviceClassOID{
		SNMPGetConfiguration: network.SNMPGetConfiguration{
			OID: "2",
		},
		indicesMapping: &indicesMappingDeviceClassOid,
	}

	expected := map[int]interface{}{
		1: value.New("Port 3"),
		2: value.New("Port 2"),
		3: value.New("Port 1"),
	}

	res, err := sut.readOID(ctx, []value.Value{value.New(1), value.New(2), value.New(3), value.New(4)}, false)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

// TestDeviceClassOIDs_readOID tests deviceClassOIDs.readOid(...)
func TestDeviceClassOIDs_readOID(t *testing.T) {
	var ifIndexOidReader mockOidReader
	var ifDescrOidReader mockOidReader
	ctx := context.Background()

	ifIndexOidReader.
		On("readOID", ctx, []value.Value(nil), true).
		Return(map[int]interface{}{
			1: value.New(1),
			2: value.New(2),
			3: value.New(3),
		}, nil)
	ifDescrOidReader.
		On("readOID", ctx, []value.Value(nil), true).
		Return(map[int]interface{}{
			1: value.New("Port 1"),
			2: value.New("Port 2"),
			3: value.New("Port 3"),
		}, nil)

	sut := deviceClassOIDs{
		"ifIndex": &ifIndexOidReader,
		"ifDescr": &ifDescrOidReader,
	}

	expectedPropertyGroups := map[int]interface{}{
		1: map[string]interface{}{
			"ifIndex": value.New(1),
			"ifDescr": value.New("Port 1"),
		},
		2: map[string]interface{}{
			"ifIndex": value.New(2),
			"ifDescr": value.New("Port 2"),
		},
		3: map[string]interface{}{
			"ifIndex": value.New(3),
			"ifDescr": value.New("Port 3"),
		},
	}

	res, err := sut.readOID(ctx, []value.Value(nil), true)
	if assert.NoError(t, err) {
		assert.Equal(t, expectedPropertyGroups, res)
	}
}

// TestDeviceClassOIDs_readOID tests deviceClassOIDs.readOid(...) including multiple level of oidReaders
func TestDeviceClassOIDs_readOID_multipleLevel(t *testing.T) {
	var ifIndexOidReader mockOidReader
	var ifDescrOidReader mockOidReader
	var radioInterfaceOidReader mockOidReader
	ctx := context.Background()

	ifIndexOidReader.
		On("readOID", ctx, []value.Value(nil), true).
		Return(map[int]interface{}{
			1: value.New("1"),
			2: value.New("2"),
			3: value.New("3"),
		}, nil)
	ifDescrOidReader.
		On("readOID", ctx, []value.Value(nil), true).
		Return(map[int]interface{}{
			1: value.New("Port 1"),
			2: value.New("Port 2"),
			3: value.New("Port 3"),
		}, nil)
	radioInterfaceOidReader.
		On("readOID", ctx, []value.Value(nil), true).
		Return(map[int]interface{}{
			1: map[string]interface{}{
				"level_in":  value.New(1),
				"level_out": value.New(-1),
			},
			2: map[string]interface{}{
				"level_in":  value.New(2),
				"level_out": value.New(-2),
			},
		}, nil)

	sut := deviceClassOIDs{
		"ifIndex": &ifIndexOidReader,
		"ifDescr": &ifDescrOidReader,
		"radio":   &radioInterfaceOidReader,
	}

	expectedPropertyGroups := map[int]interface{}{
		1: map[string]interface{}{
			"ifIndex": value.New(1),
			"ifDescr": value.New("Port 1"),
			"radio": map[string]interface{}{
				"level_in":  value.New(1),
				"level_out": value.New(-1),
			},
		},
		2: map[string]interface{}{
			"ifIndex": value.New(2),
			"ifDescr": value.New("Port 2"),
			"radio": map[string]interface{}{
				"level_in":  value.New(2),
				"level_out": value.New(-2),
			},
		},
		3: map[string]interface{}{
			"ifIndex": value.New(3),
			"ifDescr": value.New("Port 3"),
		},
	}

	res, err := sut.readOID(ctx, []value.Value(nil), true)
	if assert.NoError(t, err) {
		assert.Equal(t, expectedPropertyGroups, res)
	}
}
