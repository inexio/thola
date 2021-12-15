package groupproperty

import (
	"context"
	"github.com/gosnmp/gosnmp"
	"github.com/inexio/thola/internal/network"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroupFilter_ApplyPropertyGroups(t *testing.T) {
	filter := GetGroupFilter([]string{"ifDescr"}, "Ethernet .*")

	groups := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"ifDescr": "Ethernet #1",
		},
		propertyGroup{
			"ifIndex": "2",
			"ifDescr": "Mgmt",
		},
	}

	filteredGroup, err := filter.ApplyPropertyGroups(context.Background(), groups)
	assert.NoError(t, err)

	expected := PropertyGroups{
		propertyGroup{
			"ifIndex": "2",
			"ifDescr": "Mgmt",
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestGroupFilter_ApplyPropertyGroups_noMatch(t *testing.T) {
	filter := GetGroupFilter([]string{"ifDescr"}, "Ethernet #2")

	groups := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"ifDescr": "Ethernet #1",
		},
		propertyGroup{
			"ifIndex": "2",
			"ifDescr": "Mgmt",
		},
	}

	filteredGroup, err := filter.ApplyPropertyGroups(context.Background(), groups)
	assert.NoError(t, err)

	assert.Equal(t, groups, filteredGroup)
}

func TestGroupFilter_ApplyPropertyGroups_nested(t *testing.T) {
	filter := GetGroupFilter([]string{"radio", "level_in"}, "10")

	groups := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"ifDescr": "Ethernet #1",
			"radio": propertyGroup{
				"level_in": "10",
			},
		},
		propertyGroup{
			"ifIndex": "2",
			"ifDescr": "Mgmt",
		},
	}

	filteredGroup, err := filter.ApplyPropertyGroups(context.Background(), groups)
	assert.NoError(t, err)

	expected := PropertyGroups{
		propertyGroup{
			"ifIndex": "2",
			"ifDescr": "Mgmt",
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestGroupFilter_applySNMP(t *testing.T) {
	filter := GetGroupFilter([]string{"ifDescr"}, "Ethernet .*")

	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, network.OID("1")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.1", gosnmp.OctetString, "Ethernet #1"),
			network.NewSNMPResponse("1.2", gosnmp.OctetString, "Mgmt"),
		}, nil)

	reader := snmpReader{
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
		},
	}

	filteredGroup, err := filter.applySNMP(ctx, reader)
	assert.NoError(t, err)

	expected := snmpReader{
		wantedIndices: map[string]struct{}{
			"2": {},
		},
		filteredIndices: map[string]struct{}{
			"1": {},
		},
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestGroupFilter_applySNMP_nested(t *testing.T) {
	filter := GetGroupFilter([]string{"radio", "level_in"}, "10")

	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, network.OID("1")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.1", gosnmp.OctetString, "Ethernet #1"),
			network.NewSNMPResponse("1.2", gosnmp.OctetString, "Mgmt"),
		}, nil).
		On("SNMPWalk", ctx, network.OID("2")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("2.1", gosnmp.OctetString, "10"),
			network.NewSNMPResponse("2.2", gosnmp.OctetString, "7"),
		}, nil)

	reader := snmpReader{
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
			"radio": &deviceClassOIDs{
				"level_in": &deviceClassOID{
					SNMPGetConfiguration: network.SNMPGetConfiguration{
						OID: "2",
					},
				},
			},
		},
	}

	filteredGroup, err := filter.applySNMP(ctx, reader)
	assert.NoError(t, err)

	expected := snmpReader{
		wantedIndices: map[string]struct{}{
			"2": {},
		},
		filteredIndices: map[string]struct{}{
			"1": {},
		},
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
			"radio": &deviceClassOIDs{
				"level_in": &deviceClassOID{
					SNMPGetConfiguration: network.SNMPGetConfiguration{
						OID: "2",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestGroupFilter_applySNMP_noMatch(t *testing.T) {
	filter := GetGroupFilter([]string{"ifDescr"}, "Ethernet #3")

	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, network.OID("1")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.1", gosnmp.OctetString, "Ethernet #1"),
			network.NewSNMPResponse("1.2", gosnmp.OctetString, "Mgmt"),
		}, nil)

	reader := snmpReader{
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
		},
	}

	filteredGroup, err := filter.applySNMP(ctx, reader)
	assert.NoError(t, err)

	expected := snmpReader{
		wantedIndices: map[string]struct{}{
			"1": {},
			"2": {},
		},
		filteredIndices: map[string]struct{}{},
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestValueFilter_CheckMatch(t *testing.T) {
	valueFilter := GetValueFilter([]string{"radio", "level_in"})

	filter, ok := valueFilter.(ValueFilter)
	assert.True(t, ok, "value filter is not implementing the ValueFilter interface")

	assert.False(t, filter.CheckMatch([]string{"radio"}))
	assert.False(t, filter.CheckMatch([]string{"ifDescr"}))
	assert.True(t, filter.CheckMatch([]string{"radio", "level_in"}))
	assert.False(t, filter.CheckMatch([]string{"radio", "level_in", "test"}))
	assert.False(t, filter.CheckMatch([]string{"radio", "level_out"}))
	assert.False(t, filter.CheckMatch([]string{"radio", "level_out", "test"}))
}

func TestValueFilter_ApplyPropertyGroups(t *testing.T) {
	filter := GetValueFilter([]string{"ifDescr"})

	groups := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"ifDescr": "Ethernet #1",
		},
	}

	filteredGroup, err := filter.ApplyPropertyGroups(context.Background(), groups)
	assert.NoError(t, err)

	expected := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestValueFilter_ApplyPropertyGroups_noMatch(t *testing.T) {
	filter := GetValueFilter([]string{"ifOperStatus"})

	groups := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"ifDescr": "Ethernet #1",
		},
	}

	filteredGroup, err := filter.ApplyPropertyGroups(context.Background(), groups)
	assert.NoError(t, err)

	assert.Equal(t, groups, filteredGroup)
}

func TestValueFilter_ApplyPropertyGroups_nested(t *testing.T) {
	filter := GetValueFilter([]string{"radio", "level_in"})

	groups := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"ifDescr": "Ethernet #1",
			"radio": propertyGroup{
				"level_in":  "10",
				"level_out": "10",
			},
		},
	}

	filteredGroup, err := filter.ApplyPropertyGroups(context.Background(), groups)
	assert.NoError(t, err)

	expected := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"ifDescr": "Ethernet #1",
			"radio": propertyGroup{
				"level_out": "10",
			},
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestValueFilter_ApplyPropertyGroups_nestedArray(t *testing.T) {
	filter := GetValueFilter([]string{"radio", "level_in"})

	groups := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"ifDescr": "Ethernet #1",
			"radio": PropertyGroups{
				{
					"level_in":  "10",
					"level_out": "10",
				},
				{
					"level_in":  "7",
					"level_out": "5",
				},
			},
		},
	}

	filteredGroup, err := filter.ApplyPropertyGroups(context.Background(), groups)
	assert.NoError(t, err)

	expected := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"ifDescr": "Ethernet #1",
			"radio": PropertyGroups{
				{
					"level_out": "10",
				},
				{
					"level_out": "5",
				},
			},
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestValueFilter_ApplyPropertyGroups_nestedWholeMatch(t *testing.T) {
	filter := GetValueFilter([]string{"radio"})

	groups := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"ifDescr": "Ethernet #1",
			"radio": propertyGroup{
				"level_in":  "10",
				"level_out": "10",
			},
		},
	}

	filteredGroup, err := filter.ApplyPropertyGroups(context.Background(), groups)
	assert.NoError(t, err)

	expected := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"ifDescr": "Ethernet #1",
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestValueFilter_applySNMP(t *testing.T) {
	filter := GetValueFilter([]string{"ifOperStatus"})

	reader := snmpReader{
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
			"ifOperStatus": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "2",
				},
			},
		},
	}

	filteredGroup, err := filter.applySNMP(context.Background(), reader)
	assert.NoError(t, err)

	expected := snmpReader{
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestValueFilter_applySNMP_nested(t *testing.T) {
	filter := GetValueFilter([]string{"radio", "level_in"})

	reader := snmpReader{
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
			"radio": &deviceClassOIDs{
				"level_in": &deviceClassOID{
					SNMPGetConfiguration: network.SNMPGetConfiguration{
						OID: "2",
					},
				},
				"level_out": &deviceClassOID{
					SNMPGetConfiguration: network.SNMPGetConfiguration{
						OID: "3",
					},
				},
			},
		},
	}

	filteredGroup, err := filter.applySNMP(context.Background(), reader)
	assert.NoError(t, err)

	expected := snmpReader{
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
			"radio": &deviceClassOIDs{
				"level_out": &deviceClassOID{
					SNMPGetConfiguration: network.SNMPGetConfiguration{
						OID: "3",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestValueFilter_applySNMP_noMatch(t *testing.T) {
	filter := GetValueFilter([]string{"ifSpeed"})

	reader := snmpReader{
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
			"ifOperStatus": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "2",
				},
			},
		},
	}

	filteredGroup, err := filter.applySNMP(context.Background(), reader)
	assert.NoError(t, err)

	assert.Equal(t, reader, filteredGroup)
}

func TestExclusiveValueFilter_CheckMatch(t *testing.T) {
	valueFilter := GetExclusiveValueFilter([][]string{{"radio", "level_in"}})

	filter, ok := valueFilter.(ValueFilter)
	assert.True(t, ok, "exclusive value filter is not implementing the ValueFilter interface")

	assert.False(t, filter.CheckMatch([]string{"radio"}))
	assert.True(t, filter.CheckMatch([]string{"ifDescr"}))
	assert.True(t, filter.CheckMatch([]string{"radio", "level_out"}))
	assert.False(t, filter.CheckMatch([]string{"radio", "level_in"}))
	assert.False(t, filter.CheckMatch([]string{"radio", "level_in", "test"}))
}

func TestExclusiveValueFilter_ApplyPropertyGroups(t *testing.T) {
	filter := GetExclusiveValueFilter([][]string{{"ifDescr"}})

	groups := PropertyGroups{
		propertyGroup{
			"ifIndex":      "1",
			"ifDescr":      "Ethernet #1",
			"ifOperStatus": "1",
		},
	}

	filteredGroup, err := filter.ApplyPropertyGroups(context.Background(), groups)
	assert.NoError(t, err)

	expected := PropertyGroups{
		propertyGroup{
			"ifDescr": "Ethernet #1",
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestExclusiveValueFilter_ApplyPropertyGroups_multiple(t *testing.T) {
	filter := GetExclusiveValueFilter([][]string{{"ifIndex"}, {"ifDescr"}})

	groups := PropertyGroups{
		propertyGroup{
			"ifIndex":      "1",
			"ifDescr":      "Ethernet #1",
			"ifOperStatus": "1",
		},
	}

	filteredGroup, err := filter.ApplyPropertyGroups(context.Background(), groups)
	assert.NoError(t, err)

	expected := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"ifDescr": "Ethernet #1",
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestExclusiveValueFilter_ApplyPropertyGroups_nested(t *testing.T) {
	filter := GetExclusiveValueFilter([][]string{{"radio", "level_in"}})

	groups := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"radio": propertyGroup{
				"level_in": "10",
			},
		},
	}

	filteredGroup, err := filter.ApplyPropertyGroups(context.Background(), groups)
	assert.NoError(t, err)

	expected := PropertyGroups{
		propertyGroup{
			"radio": propertyGroup{
				"level_in": "10",
			},
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestExclusiveValueFilter_ApplyPropertyGroups_nestedArray(t *testing.T) {
	filter := GetExclusiveValueFilter([][]string{{"radio", "level_in"}})

	groups := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"ifDescr": "Ethernet #1",
			"radio": PropertyGroups{
				{
					"level_in":  "10",
					"level_out": "10",
				},
				{
					"level_in":  "7",
					"level_out": "5",
				},
			},
		},
	}

	filteredGroup, err := filter.ApplyPropertyGroups(context.Background(), groups)
	assert.NoError(t, err)

	expected := PropertyGroups{
		propertyGroup{
			"radio": PropertyGroups{
				{
					"level_in": "10",
				},
				{
					"level_in": "7",
				},
			},
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestExclusiveValueFilter_ApplyPropertyGroups_multipleNested(t *testing.T) {
	filter := GetExclusiveValueFilter([][]string{{"radio", "level_in"}, {"radio", "level_out"}})

	groups := PropertyGroups{
		propertyGroup{
			"ifIndex": "1",
			"radio": propertyGroup{
				"level_in":       "10",
				"level_out":      "10",
				"max_bitrate_in": "100000",
			},
		},
	}

	filteredGroup, err := filter.ApplyPropertyGroups(context.Background(), groups)
	assert.NoError(t, err)

	expected := PropertyGroups{
		propertyGroup{
			"radio": propertyGroup{
				"level_in":  "10",
				"level_out": "10",
			},
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestExclusiveValueFilter_applySNMP(t *testing.T) {
	filter := GetExclusiveValueFilter([][]string{{"ifDescr"}})

	reader := snmpReader{
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
			"ifOperStatus": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "2",
				},
			},
		},
	}

	filteredGroup, err := filter.applySNMP(context.Background(), reader)
	assert.NoError(t, err)

	expected := snmpReader{
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestExclusiveValueFilter_applySNMP_nested(t *testing.T) {
	filter := GetExclusiveValueFilter([][]string{{"radio", "level_in"}})

	reader := snmpReader{
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
			"radio": &deviceClassOIDs{
				"level_in": &deviceClassOID{
					SNMPGetConfiguration: network.SNMPGetConfiguration{
						OID: "2",
					},
				},
				"level_out": &deviceClassOID{
					SNMPGetConfiguration: network.SNMPGetConfiguration{
						OID: "3",
					},
				},
			},
		},
	}

	filteredGroup, err := filter.applySNMP(context.Background(), reader)
	assert.NoError(t, err)

	expected := snmpReader{
		oids: &deviceClassOIDs{
			"radio": &deviceClassOIDs{
				"level_in": &deviceClassOID{
					SNMPGetConfiguration: network.SNMPGetConfiguration{
						OID: "2",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, filteredGroup)
}

func TestExclusiveValueFilter_applySNMP_noMatch(t *testing.T) {
	filter := GetExclusiveValueFilter([][]string{{"ifSpeed"}})

	reader := snmpReader{
		oids: &deviceClassOIDs{
			"ifDescr": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "1",
				},
			},
			"ifOperStatus": &deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID: "2",
				},
			},
		},
	}

	filteredGroup, err := filter.applySNMP(context.Background(), reader)
	assert.NoError(t, err)

	expected := snmpReader{
		oids: &deviceClassOIDs{},
	}

	assert.Equal(t, expected, filteredGroup)
}
