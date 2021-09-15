package codecommunicator

import (
	"context"
	"github.com/gosnmp/gosnmp"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/network/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJunosCommunicator_GetCPUComponentCPULoad(t *testing.T) {
	var snmpClient mocks.SNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, ".1.3.6.1.4.1.2636.3.1.13.1.5").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.2636.3.1.13.1.5.1.1.0", gosnmp.OctetString, "test"),
			network.NewSNMPResponse(".1.3.6.1.4.1.2636.3.1.13.1.5.2.1.0", gosnmp.OctetString, "Routing Engine 0"),
		}, nil).
		On("SNMPGet", ctx, ".1.3.6.1.4.1.2636.3.1.13.1.8.2.1.0").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.2636.3.1.13.1.8.2.1.0", gosnmp.OctetString, "26"),
		}, nil)

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
