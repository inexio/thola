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

//TestIosCommunicator_GetCPUComponentCPULoad: 1 CPU with no label, rev and dep OID both return the same value (behavior of most devices)
func TestIosCommunicator_GetCPUComponentCPULoad(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.8").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.8.1", gosnmp.Gauge32, uint(10)),
		}, nil).
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.5").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.5.1", gosnmp.Gauge32, uint(10)),
		}, nil).
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.2").
		Return(nil, errors.New("no such oid"))

	sut := iosCommunicator{codeCommunicator{}}

	load := 10.0
	expected := []device.CPU{
		{
			Label: nil,
			Load:  &load,
		},
	}

	res, err := sut.GetCPUComponentCPULoad(ctx)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

//TestIosCommunicator_GetCPUComponentCPULoad_onlyDepOID: 1 CPU with no label, only dep OID returns value (behavior of old cisco devices)
func TestIosCommunicator_GetCPUComponentCPULoad_onlyDepOID(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.8").
		Return(nil, errors.New("no such oid")).
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.5").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.5.1", gosnmp.Gauge32, uint(10)),
		}, nil).
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.2").
		Return(nil, errors.New("no such oid"))

	sut := iosCommunicator{codeCommunicator{}}

	load := 10.0
	expected := []device.CPU{
		{
			Label: nil,
			Load:  &load,
		},
	}

	res, err := sut.GetCPUComponentCPULoad(ctx)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

//TestIosCommunicator_GetCPUComponentCPULoad_onlyRevOID: 1 CPU with no label, only rev OID returns value
func TestIosCommunicator_GetCPUComponentCPULoad_onlyRevOID(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.8").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.8.1", gosnmp.Gauge32, uint(10)),
		}, nil).
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.5").
		Return(nil, errors.New("no such oid")).
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.2").
		Return(nil, errors.New("no such oid"))

	sut := iosCommunicator{codeCommunicator{}}

	load := 10.0
	expected := []device.CPU{
		{
			Label: nil,
			Load:  &load,
		},
	}

	res, err := sut.GetCPUComponentCPULoad(ctx)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

//TestIosCommunicator_GetCPUComponentCPULoad_withLabel: 1 CPU with label, rev and dep OID both return the same value (behavior of most devices)
func TestIosCommunicator_GetCPUComponentCPULoad_withLabel(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.8").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.8.1", gosnmp.Gauge32, uint(10)),
		}, nil).
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.5").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.5.1", gosnmp.Gauge32, uint(10)),
		}, nil).
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.2").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.2.1", gosnmp.Integer, 1),
		}, nil).
		On("SNMPGet", ctx, "1.3.6.1.2.1.47.1.1.1.1.7.1").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.3.6.1.2.1.47.1.1.1.1.7.1", gosnmp.OctetString, "cpu1"),
		}, nil)

	sut := iosCommunicator{codeCommunicator{}}

	load := 10.0
	cpu1 := "cpu1"
	expected := []device.CPU{
		{
			Label: &cpu1,
			Load:  &load,
		},
	}

	res, err := sut.GetCPUComponentCPULoad(ctx)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

//TestIosCommunicator_GetCPUComponentCPULoad_multipleCPUs: 3 CPU with no label, rev and dep OID both return the same value (behavior of most devices)
func TestIosCommunicator_GetCPUComponentCPULoad_multipleCPUs(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.8").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.8.1", gosnmp.Gauge32, uint(10)),
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.8.2", gosnmp.Gauge32, uint(20)),
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.8.3", gosnmp.Gauge32, uint(30)),
		}, nil).
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.5").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.5.1", gosnmp.Gauge32, uint(10)),
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.5.2", gosnmp.Gauge32, uint(20)),
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.5.3", gosnmp.Gauge32, uint(30)),
		}, nil).
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.2").
		Return(nil, errors.New("no such oid"))

	sut := iosCommunicator{codeCommunicator{}}

	load1 := 10.0
	load2 := 20.0
	load3 := 30.0
	expected := []device.CPU{
		{
			Label: nil,
			Load:  &load1,
		},
		{
			Label: nil,
			Load:  &load2,
		},
		{
			Label: nil,
			Load:  &load3,
		},
	}

	res, err := sut.GetCPUComponentCPULoad(ctx)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

//TestIosCommunicator_GetCPUComponentCPULoad_multipleCPUsWithLabel: 3 CPU with label, rev and dep OID both return the same value (behavior of most devices)
func TestIosCommunicator_GetCPUComponentCPULoad_multipleCPUsWithLabel(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.8").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.8.1", gosnmp.Gauge32, uint(10)),
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.8.2", gosnmp.Gauge32, uint(20)),
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.8.3", gosnmp.Gauge32, uint(30)),
		}, nil).
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.5").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.5.1", gosnmp.Gauge32, uint(10)),
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.5.2", gosnmp.Gauge32, uint(20)),
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.5.3", gosnmp.Gauge32, uint(30)),
		}, nil).
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.2").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.2.1", gosnmp.Integer, 3),
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.2.2", gosnmp.Integer, 4),
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.2.3", gosnmp.Integer, 5),
		}, nil).
		On("SNMPGet", ctx, "1.3.6.1.2.1.47.1.1.1.1.7.3").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.3.6.1.2.1.47.1.1.1.1.7.3", gosnmp.OctetString, "cpu1"),
		}, nil).
		On("SNMPGet", ctx, "1.3.6.1.2.1.47.1.1.1.1.7.4").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.3.6.1.2.1.47.1.1.1.1.7.4", gosnmp.OctetString, "cpu2"),
		}, nil).
		On("SNMPGet", ctx, "1.3.6.1.2.1.47.1.1.1.1.7.5").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse("1.3.6.1.2.1.47.1.1.1.1.7.5", gosnmp.OctetString, "cpu3"),
		}, nil)

	sut := iosCommunicator{codeCommunicator{}}

	load1 := 10.0
	cpu1 := "cpu1"
	load2 := 20.0
	cpu2 := "cpu2"
	load3 := 30.0
	cpu3 := "cpu3"
	expected := []device.CPU{
		{
			Label: &cpu1,
			Load:  &load1,
		},
		{
			Label: &cpu2,
			Load:  &load2,
		},
		{
			Label: &cpu3,
			Load:  &load3,
		},
	}

	res, err := sut.GetCPUComponentCPULoad(ctx)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}

//TestIosCommunicator_GetCPUComponentCPULoad_prioritiseRevOID checks if dev oid is prioritised over dep oid
func TestIosCommunicator_GetCPUComponentCPULoad_prioritiseRevOID(t *testing.T) {
	var snmpClient network.MockSNMPClient
	ctx := network.NewContextWithDeviceConnection(context.Background(), &network.RequestDeviceConnection{
		SNMP: &network.RequestDeviceConnectionSNMP{
			SnmpClient: &snmpClient,
		},
	})

	snmpClient.
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.8").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.8.1", gosnmp.Gauge32, uint(10)),
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.8.3", gosnmp.Gauge32, uint(10)),
		}, nil).
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.5").
		Return([]network.SNMPResponse{
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.5.1", gosnmp.Gauge32, uint(20)),
			network.NewSNMPResponse(".1.3.6.1.4.1.9.9.109.1.1.1.1.8.2", gosnmp.Gauge32, uint(20)),
		}, nil).
		On("SNMPWalk", ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.2").
		Return(nil, errors.New("no such oid"))

	sut := iosCommunicator{codeCommunicator{}}

	load1 := 10.0
	load2 := 10.0
	load3 := 20.0
	expected := []device.CPU{
		{
			Label: nil,
			Load:  &load1,
		},
		{
			Label: nil,
			Load:  &load2,
		},
		{
			Label: nil,
			Load:  &load3,
		},
	}

	res, err := sut.GetCPUComponentCPULoad(ctx)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, res)
	}
}
