package deviceclass

import (
	"context"
	"github.com/gosnmp/gosnmp"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/network/mocks"
	"github.com/inexio/thola/internal/value"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestDeviceClassOID(t *testing.T) {
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
		}, nil)

	sut := deviceClassOID{
		SNMPGetConfiguration: network.SNMPGetConfiguration{
			OID: "1",
		},
	}

	expected := make(map[int]interface{})
	for i := 1; i <= 3; i++ {
		name := value.New("Port " + strconv.Itoa(i))
		expected[i] = name
	}

	res, err := sut.readOID(ctx, nil, true)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}
