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

func TestSNMPGroupPropertyReader_getProperty(t *testing.T) {
	var oidReader mockOidReader
	ctx := context.Background()

	oidReader.
		On("readOID", ctx, []value.Value(nil), true).
		Return(map[int]interface{}{
			1: value.New("Port 1"),
			2: value.New("Port 2"),
			3: value.New("Port 3"),
		}, nil)

	sut := snmpGroupPropertyReader{
		oids: deviceClassOIDs{
			"ifDescr": &oidReader,
		},
	}

	expectedPropertyGroups := propertyGroups{
		propertyGroup{
			"ifDescr": value.New("Port 1"),
		},
		propertyGroup{
			"ifDescr": value.New("Port 2"),
		},
		propertyGroup{
			"ifDescr": value.New("Port 3"),
		},
	}

	expectedIndices := []value.Value{
		value.New("1"),
		value.New("2"),
		value.New("3"),
	}

	res, indices, err := sut.getProperty(ctx)
	if assert.NoError(t, err) {
		assert.Equal(t, expectedPropertyGroups, res)
		assert.Equal(t, expectedIndices, indices)
	}
}

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
