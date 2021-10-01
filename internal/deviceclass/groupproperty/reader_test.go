package groupproperty

import (
	"context"
	"github.com/gosnmp/gosnmp"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/utility"
	"github.com/inexio/thola/internal/value"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// TestDeviceClassOID_readOID tests deviceClassOID.readOid(...) without indices and skipEmpty = false
func TestDeviceClassOID_readOID(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, network.OID("1")).
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

	expected := map[string]interface{}{
		"1": value.New("Port 1"),
		"2": value.New("Port 2"),
		"3": value.New("Port 3"),
		"4": value.New(""),
	}

	res, err := sut.readOID(ctx, nil, false)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

// TestDeviceClassOID_readOID_skipEmpty tests deviceClassOID.readOid(...) without indices and skipEmpty = true
func TestDeviceClassOID_readOID_skipEmpty(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, network.OID("1")).
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

	expected := map[string]interface{}{
		"1": value.New("Port 1"),
		"2": value.New("Port 2"),
		"3": value.New("Port 3"),
	}

	res, err := sut.readOID(ctx, nil, true)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

// TestDeviceClassOID_readOID_withIndices tests deviceClassOID.readOid(...) with indices and skipEmpty = false
func TestDeviceClassOID_readOID_withIndices(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	ctx = network.NewContextWithSNMPGetsInsteadOfWalk(ctx, true)

	snmpClient.
		On("SNMPGet", ctx, network.OID("1.1"), network.OID("1.2"), network.OID("1.3"), network.OID("1.4")).
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

	expected := map[string]interface{}{
		"1": value.New("Port 1"),
		"2": value.New("Port 2"),
		"3": value.New("Port 3"),
		"4": value.New(""),
	}

	res, err := sut.readOID(ctx, []string{"1", "2", "3", "4"}, false)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

// TestDeviceClassOID_readOID_withIndicesSkipEmpty tests deviceClassOID.readOid(...) with indices and skipEmpty = true
func TestDeviceClassOID_readOID_withIndicesSkipEmpty(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	ctx = network.NewContextWithSNMPGetsInsteadOfWalk(ctx, true)

	snmpClient.
		On("SNMPGet", ctx, network.OID("1.1"), network.OID("1.2"), network.OID("1.3"), network.OID("1.4")).
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

	expected := map[string]interface{}{
		"1": value.New("Port 1"),
		"2": value.New("Port 2"),
		"3": value.New("Port 3"),
	}

	res, err := sut.readOID(ctx, []string{"1", "2", "3", "4"}, true)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

// TestDeviceClassOID_readOID_indicesMapping tests deviceClassOID.readOid(...) with an indices mapping and no given indices
func TestDeviceClassOID_readOID_indicesMapping(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, network.OID("1")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.1", gosnmp.OctetString, "3"),
			network.NewSNMPResponse("1.2", gosnmp.OctetString, "2"),
			network.NewSNMPResponse("1.3", gosnmp.OctetString, "1"),
		}, nil).
		On("SNMPWalk", ctx, network.OID("2")).
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

	expected := map[string]interface{}{
		"1": value.New("Port 3"),
		"2": value.New("Port 2"),
		"3": value.New("Port 1"),
	}

	res, err := sut.readOID(ctx, nil, false)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

// TestDeviceClassOID_readOID_indicesMapping tests deviceClassOID.readOid(...) with an indices mapping and given indices
func TestDeviceClassOID_readOID_indicesMappingWithIndices(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, network.OID("1")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.1", gosnmp.OctetString, "3"),
			network.NewSNMPResponse("1.2", gosnmp.OctetString, "2"),
			network.NewSNMPResponse("1.3", gosnmp.OctetString, "1"),
		}, nil).
		On("SNMPGet", ctx, network.OID("2.3"), network.OID("2.2"), network.OID("2.1")).
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

	expected := map[string]interface{}{
		"1": value.New("Port 3"),
		"2": value.New("Port 2"),
		"3": value.New("Port 1"),
	}

	res, err := sut.readOID(ctx, []string{"1", "2", "3", "4"}, false)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

// TestDeviceClassOIDs_readOID tests deviceClassOIDs.readOid(...)
func TestDeviceClassOIDs_readOID(t *testing.T) {
	var ifIndexOidReader MockOIDReader
	var ifDescrOidReader MockOIDReader
	ctx := context.Background()

	ifIndexOidReader.
		On("readOID", ctx, []string(nil), true).
		Return(map[string]interface{}{
			"1": value.New(1),
			"2": value.New(2),
			"3": value.New(3),
		}, nil)
	ifDescrOidReader.
		On("readOID", ctx, []string(nil), true).
		Return(map[string]interface{}{
			"1": value.New("Port 1"),
			"2": value.New("Port 2"),
			"3": value.New("Port 3"),
		}, nil)

	sut := deviceClassOIDs{
		"ifIndex": &ifIndexOidReader,
		"ifDescr": &ifDescrOidReader,
	}

	expected := map[string]interface{}{
		"1": map[string]interface{}{
			"ifIndex": value.New(1),
			"ifDescr": value.New("Port 1"),
		},
		"2": map[string]interface{}{
			"ifIndex": value.New(2),
			"ifDescr": value.New("Port 2"),
		},
		"3": map[string]interface{}{
			"ifIndex": value.New(3),
			"ifDescr": value.New("Port 3"),
		},
	}

	res, err := sut.readOID(ctx, []string(nil), true)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

// TestDeviceClassOIDs_readOID tests deviceClassOIDs.readOid(...) including multiple level of oidReaders
func TestDeviceClassOIDs_readOID_multipleLevel(t *testing.T) {
	var ifIndexOidReader MockOIDReader
	var ifDescrOidReader MockOIDReader
	var radioInterfaceOidReader MockOIDReader
	ctx := context.Background()

	ifIndexOidReader.
		On("readOID", ctx, []string(nil), true).
		Return(map[string]interface{}{
			"1": value.New("1"),
			"2": value.New("2"),
			"3": value.New("3"),
		}, nil)
	ifDescrOidReader.
		On("readOID", ctx, []string(nil), true).
		Return(map[string]interface{}{
			"1": value.New("Port 1"),
			"2": value.New("Port 2"),
			"3": value.New("Port 3"),
		}, nil)
	radioInterfaceOidReader.
		On("readOID", ctx, []string(nil), true).
		Return(map[string]interface{}{
			"1": map[string]interface{}{
				"level_in":  value.New(1),
				"level_out": value.New(-1),
			},
			"2": map[string]interface{}{
				"level_in":  value.New(2),
				"level_out": value.New(-2),
			},
		}, nil)

	sut := deviceClassOIDs{
		"ifIndex": &ifIndexOidReader,
		"ifDescr": &ifDescrOidReader,
		"radio":   &radioInterfaceOidReader,
	}

	expected := map[string]interface{}{
		"1": map[string]interface{}{
			"ifIndex": value.New(1),
			"ifDescr": value.New("Port 1"),
			"radio": map[string]interface{}{
				"level_in":  value.New(1),
				"level_out": value.New(-1),
			},
		},
		"2": map[string]interface{}{
			"ifIndex": value.New(2),
			"ifDescr": value.New("Port 2"),
			"radio": map[string]interface{}{
				"level_in":  value.New(2),
				"level_out": value.New(-2),
			},
		},
		"3": map[string]interface{}{
			"ifIndex": value.New(3),
			"ifDescr": value.New("Port 3"),
		},
	}

	res, err := sut.readOID(ctx, []string(nil), true)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

func TestSNMPReader_getProperty(t *testing.T) {
	var oidReader MockOIDReader
	ctx := context.Background()

	oidReader.
		On("readOID", ctx, []string(nil), true).
		Return(map[string]interface{}{
			"1": map[string]interface{}{
				"ifIndex": value.New(1),
				"ifDescr": value.New("Port 1"),
			},
			"2": map[string]interface{}{
				"ifIndex": value.New(2),
				"ifDescr": value.New("Port 2"),
			},
			"3": map[string]interface{}{
				"ifIndex": value.New(3),
				"ifDescr": value.New("Port 3"),
			},
		}, nil)

	sut := snmpReader{
		oids: &oidReader,
	}

	expectedPropertyGroups := PropertyGroups{
		propertyGroup{
			"ifIndex": value.New(1),
			"ifDescr": value.New("Port 1"),
		},
		propertyGroup{
			"ifIndex": value.New(2),
			"ifDescr": value.New("Port 2"),
		},
		propertyGroup{
			"ifIndex": value.New(3),
			"ifDescr": value.New("Port 3"),
		},
	}

	expectedIndices := []value.Value{
		value.New(1),
		value.New(2),
		value.New(3),
	}

	res, indices, err := sut.getProperty(ctx)
	if assert.NoError(t, err) {
		assert.Equal(t, expectedPropertyGroups, res)
		assert.Equal(t, expectedIndices, indices)
	}
}

func TestSNMPReader_getProperty_filter(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, network.OID("1")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.1", gosnmp.OctetString, "1"),
			network.NewSNMPResponse("1.2", gosnmp.OctetString, "2"),
			network.NewSNMPResponse("1.3", gosnmp.OctetString, "3"),
		}, nil).
		On("SNMPWalk", ctx, network.OID("2")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("2.1", gosnmp.OctetString, "Port 1"),
			network.NewSNMPResponse("2.2", gosnmp.OctetString, "Port 2"),
			network.NewSNMPResponse("2.3", gosnmp.OctetString, "Port 3"),
		}, nil)

	sut := baseReader{
		reader: &snmpReader{
			oids: &deviceClassOIDs{
				"ifIndex": &deviceClassOID{
					SNMPGetConfiguration: network.SNMPGetConfiguration{
						OID: "1",
					},
				},
				"ifDescr": &deviceClassOID{
					SNMPGetConfiguration: network.SNMPGetConfiguration{
						OID: "2",
					},
				},
			},
		},
	}

	expectedPropertyGroups := PropertyGroups{
		propertyGroup{
			"ifIndex": value.New(1),
			"ifDescr": value.New("Port 1"),
		},
		propertyGroup{
			"ifIndex": value.New(3),
			"ifDescr": value.New("Port 3"),
		},
	}

	expectedIndices := []value.Value{
		value.New(1),
		value.New(3),
	}

	res, indices, err := sut.GetProperty(ctx, &groupFilter{
		key:   "ifDescr",
		regex: "2",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, expectedPropertyGroups, res)
		assert.Equal(t, expectedIndices, indices)
	}
}

func TestSNMPReader_getProperty_getsInsteadOfWalk(t *testing.T) {
	var oidReader MockOIDReader
	var indexOIDReader MockOIDReader
	ctx := network.NewContextWithSNMPGetsInsteadOfWalk(context.Background(), true)

	oidReader.
		On("readOID", ctx, mock.MatchedBy(func(input []string) bool {
			return utility.SameStringSlice(input, []string{
				"1",
				"2",
				"3",
			})
		}), true).
		Return(map[string]interface{}{
			"1": map[string]interface{}{
				"ifIndex": value.New(1),
				"ifDescr": value.New("Port 1"),
			},
			"2": map[string]interface{}{
				"ifIndex": value.New(2),
				"ifDescr": value.New("Port 2"),
			},
			"3": map[string]interface{}{
				"ifIndex": value.New(3),
				"ifDescr": value.New("Port 3"),
			},
		}, nil)

	indexOIDReader.
		On("readOID", ctx, []string(nil), false).
		Return(map[string]interface{}{
			"1": map[string]interface{}{
				"ifIndex": value.New(1),
			},
			"2": map[string]interface{}{
				"ifIndex": value.New(2),
			},
			"3": map[string]interface{}{
				"ifIndex": value.New(3),
			},
		}, nil)

	sut := snmpReader{
		oids:  &oidReader,
		index: &indexOIDReader,
	}

	expectedPropertyGroups := PropertyGroups{
		propertyGroup{
			"ifIndex": value.New(1),
			"ifDescr": value.New("Port 1"),
		},
		propertyGroup{
			"ifIndex": value.New(2),
			"ifDescr": value.New("Port 2"),
		},
		propertyGroup{
			"ifIndex": value.New(3),
			"ifDescr": value.New("Port 3"),
		},
	}

	expectedIndices := []value.Value{
		value.New(1),
		value.New(2),
		value.New(3),
	}

	res, indices, err := sut.getProperty(ctx)
	if assert.NoError(t, err) {
		assert.Equal(t, expectedPropertyGroups, res)
		assert.Equal(t, expectedIndices, indices)
	}
}

func TestSNMPReader_getProperty_getsInsteadOfWalkFilter(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(network.NewContextWithSNMPGetsInsteadOfWalk(context.Background(), true),
		&network.RequestDeviceConnection{
			SNMP: &network.RequestDeviceConnectionSNMP{
				SnmpClient: &snmpClient,
			},
		})

	snmpClient.
		On("SNMPGet", ctx, network.OID("1.1"), network.OID("1.3")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.1", gosnmp.OctetString, "1"),
			network.NewSNMPResponse("1.3", gosnmp.OctetString, "3"),
		}, nil).
		On("SNMPGet", ctx, network.OID("1.3"), network.OID("1.1")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.1", gosnmp.OctetString, "1"),
			network.NewSNMPResponse("1.3", gosnmp.OctetString, "3"),
		}, nil).
		On("SNMPGet", ctx, network.OID("2.1"), network.OID("2.3")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("2.1", gosnmp.OctetString, "Port 1"),
			network.NewSNMPResponse("2.3", gosnmp.OctetString, "Port 3"),
		}, nil).
		On("SNMPGet", ctx, network.OID("2.3"), network.OID("2.1")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("2.1", gosnmp.OctetString, "Port 1"),
			network.NewSNMPResponse("2.3", gosnmp.OctetString, "Port 3"),
		}, nil).
		On("SNMPWalk", ctx, network.OID("2")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("2.1", gosnmp.OctetString, "Port 1"),
			network.NewSNMPResponse("2.2", gosnmp.OctetString, "Port 2"),
			network.NewSNMPResponse("2.3", gosnmp.OctetString, "Port 3"),
		}, nil)

	sut := baseReader{
		reader: &snmpReader{
			oids: &deviceClassOIDs{
				"ifIndex": &deviceClassOID{
					SNMPGetConfiguration: network.SNMPGetConfiguration{
						OID: "1",
					},
				},
				"ifDescr": &deviceClassOID{
					SNMPGetConfiguration: network.SNMPGetConfiguration{
						OID: "2",
					},
				},
			},
		},
	}

	expectedPropertyGroups := PropertyGroups{
		propertyGroup{
			"ifIndex": value.New(1),
			"ifDescr": value.New("Port 1"),
		},
		propertyGroup{
			"ifIndex": value.New(3),
			"ifDescr": value.New("Port 3"),
		},
	}

	expectedIndices := []value.Value{
		value.New(1),
		value.New(3),
	}

	res, indices, err := sut.GetProperty(ctx, &groupFilter{
		key:   "ifDescr",
		regex: "2",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, expectedPropertyGroups, res)
		assert.Equal(t, expectedIndices, indices)
	}
}
