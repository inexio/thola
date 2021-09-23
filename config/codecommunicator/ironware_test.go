package codecommunicator

import (
	"context"
	"github.com/gosnmp/gosnmp"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIronwareCommunicator_GetCPUComponentCPULoad_100thPercent(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, ".1.3.6.1.4.1.1991.1.1.2.11.1.1.6").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.6.1.1.5", gosnmp.OctetString, "600"),
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.6.1.1.300", gosnmp.OctetString, "600"),
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.6.2.1.5", gosnmp.OctetString, "1100"),
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.6.2.1.300", gosnmp.OctetString, "1100"),
		}, nil).
		On("SNMPWalk", ctx, ".1.3.6.1.4.1.1991.1.1.2.11.1.1.1").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.1.1.1.5", gosnmp.OctetString, "1"),
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.1.1.1.300", gosnmp.OctetString, "1"),
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.1.2.1.5", gosnmp.OctetString, "4"),
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.1.2.1.300", gosnmp.OctetString, "4"),
		}, nil)

	sut := ironwareCommunicator{codeCommunicator{}}
	res, err := sut.GetCPUComponentCPULoad(ctx)

	cpu1Label := "1"
	cpu2Label := "4"
	cpu1Util := 6.0
	cpu2Util := 11.0

	expected := []device.CPU{
		{
			Label: &cpu1Label,
			Load:  &cpu1Util,
		},
		{
			Label: &cpu2Label,
			Load:  &cpu2Util,
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

func TestIronwareCommunicator_GetCPUComponentCPULoad_Value(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, ".1.3.6.1.4.1.1991.1.1.2.11.1.1.6").
		Return(nil, errors.New("No Such Object available on this agent at this OID")).
		On("SNMPWalk", ctx, ".1.3.6.1.4.1.1991.1.1.2.11.1.1.4").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.4.1.1.5", gosnmp.OctetString, "6"),
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.4.1.1.300", gosnmp.OctetString, "6"),
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.4.2.1.5", gosnmp.OctetString, "11"),
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.4.2.1.300", gosnmp.OctetString, "11"),
		}, nil).
		On("SNMPWalk", ctx, ".1.3.6.1.4.1.1991.1.1.2.11.1.1.1").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.1.1.1.5", gosnmp.OctetString, "1"),
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.1.1.1.300", gosnmp.OctetString, "1"),
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.1.2.1.5", gosnmp.OctetString, "4"),
			network.NewSNMPResponse(".1.3.6.1.4.1.1991.1.1.2.11.1.1.1.2.1.300", gosnmp.OctetString, "4"),
		}, nil)

	sut := ironwareCommunicator{codeCommunicator{}}
	res, err := sut.GetCPUComponentCPULoad(ctx)

	cpu1Label := "1"
	cpu2Label := "4"
	cpu1Util := 6.0
	cpu2Util := 11.0

	expected := []device.CPU{
		{
			Label: &cpu1Label,
			Load:  &cpu1Util,
		},
		{
			Label: &cpu2Label,
			Load:  &cpu2Util,
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}
