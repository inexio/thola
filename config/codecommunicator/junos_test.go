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

func TestJunosCommunicator_GetCPUComponentCPULoad_WithSPU(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, network.OID(".1.3.6.1.4.1.2636.3.1.13.1.5")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.2636.3.1.13.1.5.1.1.0", gosnmp.OctetString, "test"),
			network.NewSNMPResponse(".1.3.6.1.4.1.2636.3.1.13.1.5.2.1.0", gosnmp.OctetString, "Routing Engine 0"),
		}, nil).
		On("SNMPGet", ctx, network.OID(".1.3.6.1.4.1.2636.3.1.13.1.8.2.1.0")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.2636.3.1.13.1.8.2.1.0", gosnmp.OctetString, "26"),
		}, nil).
		On("SNMPWalk", ctx, network.OID(".1.3.6.1.4.1.2636.3.39.1.12.1.1.1.11")).
		Return(nil, errors.New("No Such Object available on this agent at this OID"))

	sut := junosCommunicator{codeCommunicator{}}
	res, err := sut.GetCPUComponentCPULoad(ctx)

	cpuLabel := "Routing Engine 0"
	cpuUtil := 26.0

	expected := []device.CPU{
		{
			Label: &cpuLabel,
			Load:  &cpuUtil,
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

func TestJunosCommunicator_GetCPUComponentCPULoad(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, network.OID(".1.3.6.1.4.1.2636.3.1.13.1.5")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.2636.3.1.13.1.5.1.1.0", gosnmp.OctetString, "test"),
			network.NewSNMPResponse(".1.3.6.1.4.1.2636.3.1.13.1.5.2.1.0", gosnmp.OctetString, "Routing Engine 0"),
		}, nil).
		On("SNMPGet", ctx, network.OID(".1.3.6.1.4.1.2636.3.1.13.1.8.2.1.0")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.2636.3.1.13.1.8.2.1.0", gosnmp.OctetString, "26"),
		}, nil).
		On("SNMPWalk", ctx, network.OID(".1.3.6.1.4.1.2636.3.39.1.12.1.1.1.11")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.2636.3.39.1.12.1.1.1.11.0", gosnmp.OctetString, "single"),
		}, nil).
		On("SNMPWalk", ctx, network.OID(".1.3.6.1.4.1.2636.3.39.1.12.1.1.1.4")).
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.2636.3.39.1.12.1.1.1.4.0", gosnmp.OctetString, "7"),
		}, nil)

	sut := junosCommunicator{codeCommunicator{}}
	res, err := sut.GetCPUComponentCPULoad(ctx)

	cpuLabel := "Routing Engine 0"
	cpuUtil := 26.0
	cpuSPULabel := "spu_single"
	cpuSPU1Util := 7.0

	expected := []device.CPU{
		{
			Label: &cpuLabel,
			Load:  &cpuUtil,
		},
		{
			Label: &cpuSPULabel,
			Load:  &cpuSPU1Util,
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}
