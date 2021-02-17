package communicator

import (
	"context"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/pkg/errors"
)

// ekinopsModuleReader is an interface with one function that receives an array of device.Interface and
// adds the module specific metrics to it. It returns a new
type ekinopsModuleReader interface {
	readModuleMetrics(context.Context, []device.Interface) ([]device.Interface, error)

	getSlotIdentifier() string
	getModuleName() string
}

func ekinopsGetModuleReader(slotIdentifier, module string) (ekinopsModuleReader, error) {
	moduleData := ekinopsModuleData{slotIdentifier, module}
	switch module {
	case "PM_OAIL-HCS", "PM_OAIL-HCS-17", "PM-OABP-HC", "PM-OAIL-HC":
		return &ekinopsModuleReaderWrapper{&ekinopsModuleReaderAmplifier{moduleData}}, nil
	case "PM_100G-AGG":
		return &ekinopsModuleReaderWrapper{&ekinopsModuleReaderTransponder{
			ekinopsModuleData: moduleData,
			networkLinePortsOIDs: ekinopsTransponderOIDs{
				identifierOID:      ".1.3.6.1.4.1.20044.91.7.1.2.1.2",
				labelOID:           ".1.3.6.1.4.1.20044.91.9.3.2.1.3",
				txPowerOID:         ".1.3.6.1.4.1.20044.91.3.3.160.1.2",
				rxPowerOID:         ".1.3.6.1.4.1.20044.91.3.3.192.1.2",
				correctedFEC:       ".1.3.6.1.4.1.20044.91.4.3.193.1.2",
				uncorrectedFEC:     ".1.3.6.1.4.1.20044.91.4.3.197.1.2",
				powerTransformFunc: ekinopsPowerTransform10Log10XMinus40,
			},
			clientPortsOIDs: ekinopsTransponderOIDs{
				identifierOID:      ".1.3.6.1.4.1.20044.91.7.1.1.1.2",
				labelOID:           ".1.3.6.1.4.1.20044.91.9.3.1.1.3",
				txPowerOID:         ".1.3.6.1.4.1.20044.91.3.2.80.1.2",
				rxPowerOID:         ".1.3.6.1.4.1.20044.91.3.2.112.1.2",
				correctedFEC:       "",
				uncorrectedFEC:     "",
				powerTransformFunc: ekinopsPowerTransform10Log10XMinus40,
			},
		}}, nil
	case "PM_1604":
		return &ekinopsModuleReaderWrapper{&ekinopsModuleReaderTransponder{
			ekinopsModuleData: moduleData,
			networkLinePortsOIDs: ekinopsTransponderOIDs{
				identifierOID:      ".1.3.6.1.4.1.20044.77.7.1.2.1.2",
				labelOID:           ".1.3.6.1.4.1.20044.77.9.3.2.1.3",
				txPowerOID:         ".1.3.6.1.4.1.20044.77.3.3.44.1.2",
				rxPowerOID:         ".1.3.6.1.4.1.20044.77.3.3.52.1.2",
				correctedFEC:       "",
				uncorrectedFEC:     "",
				powerTransformFunc: ekinopsPowerTransform10Log10XMinus40,
			},
			clientPortsOIDs: ekinopsTransponderOIDs{
				identifierOID:      ".1.3.6.1.4.1.20044.77.7.1.1.1.2",
				labelOID:           ".1.3.6.1.4.1.20044.77.9.3.1.1.3",
				txPowerOID:         ".1.3.6.1.4.1.20044.77.3.2.40.1.2",
				rxPowerOID:         ".1.3.6.1.4.1.20044.77.3.2.48.1.2",
				correctedFEC:       "",
				uncorrectedFEC:     "",
				powerTransformFunc: ekinopsPowerTransform10Log10XMinus40,
			},
		}}, nil
	case "PM_200FRS02":
		return &ekinopsModuleReaderWrapper{&ekinopsModuleReaderTransponder{
			ekinopsModuleData: moduleData,
			networkLinePortsOIDs: ekinopsTransponderOIDs{
				identifierOID:      ".1.3.6.1.4.1.20044.90.7.1.2.1.2",
				labelOID:           ".1.3.6.1.4.1.20044.90.9.3.2.1.3",
				txPowerOID:         ".1.3.6.1.4.1.20044.90.3.3.144.1.2",
				rxPowerOID:         ".1.3.6.1.4.1.20044.90.3.3.156.1.2",
				correctedFEC:       ".1.3.6.1.4.1.20044.90.4.3.193.1.2",
				uncorrectedFEC:     ".1.3.6.1.4.1.20044.90.4.3.197.1.2",
				powerTransformFunc: ekionopsPowerTransformShiftDivideBy100,
			},
			clientPortsOIDs: ekinopsTransponderOIDs{
				identifierOID:      ".1.3.6.1.4.1.20044.90.7.1.1.1.2",
				labelOID:           ".1.3.6.1.4.1.20044.90.9.3.1.1.3",
				txPowerOID:         ".1.3.6.1.4.1.20044.90.3.2.256.1.2",
				rxPowerOID:         ".1.3.6.1.4.1.20044.90.3.2.288.1.2",
				correctedFEC:       "",
				uncorrectedFEC:     "",
				powerTransformFunc: ekinopsPowerTransform10Log10XMinus40,
			},
		}}, nil
	case "PM_400FRS04-SF":
		return &ekinopsModuleReaderWrapper{&ekinopsModuleReaderTransponder{
			ekinopsModuleData: moduleData,
			networkLinePortsOIDs: ekinopsTransponderOIDs{
				identifierOID:      ".1.3.6.1.4.1.20044.100.7.1.2.1.2",
				labelOID:           ".1.3.6.1.4.1.20044.100.9.3.2.1.3",
				txPowerOID:         ".1.3.6.1.4.1.20044.100.3.3.144.1.2",
				rxPowerOID:         ".1.3.6.1.4.1.20044.100.3.3.156.1.2",
				correctedFEC:       ".1.3.6.1.4.1.20044.100.4.3.196.1.2",
				uncorrectedFEC:     ".1.3.6.1.4.1.20044.100.4.3.197.1.2",
				powerTransformFunc: ekionopsPowerTransformShiftDivideBy100,
			},
			clientPortsOIDs: ekinopsTransponderOIDs{
				identifierOID:      ".1.3.6.1.4.1.20044.100.7.1.1.1.2",
				labelOID:           ".1.3.6.1.4.1.20044.100.9.3.1.1.3",
				txPowerOID:         ".1.3.6.1.4.1.20044.100.3.2.256.1.2",
				rxPowerOID:         ".1.3.6.1.4.1.20044.100.3.2.288.1.2",
				correctedFEC:       "",
				uncorrectedFEC:     "",
				powerTransformFunc: ekinopsPowerTransform10Log10XMinus40,
			},
		}}, nil
	case "PM_O6006MP":
		return &ekinopsModuleReaderWrapper{&ekinopsModuleReaderTransponder{
			ekinopsModuleData: moduleData,
			networkLinePortsOIDs: ekinopsTransponderOIDs{
				identifierOID:      ".1.3.6.1.4.1.20044.70.7.1.2.1.2",
				labelOID:           ".1.3.6.1.4.1.20044.70.9.2.2.1.3",
				txPowerOID:         ".1.3.6.1.4.1.20044.70.3.3.80.1.2",
				rxPowerOID:         ".1.3.6.1.4.1.20044.70.3.3.128.1.2",
				correctedFEC:       ".1.3.6.1.4.1.20044.70.4.3.200.1.2",
				uncorrectedFEC:     "",
				powerTransformFunc: ekinopsPowerTransform10Log10XDivideBy10000,
			},
			clientPortsOIDs: ekinopsTransponderOIDs{
				identifierOID:      ".1.3.6.1.4.1.20044.70.7.1.1.1.2",
				labelOID:           ".1.3.6.1.4.1.20044.70.9.2.1.1.3",
				txPowerOID:         ".1.3.6.1.4.1.20044.70.3.2.48.1.2",
				rxPowerOID:         ".1.3.6.1.4.1.20044.70.3.2.64.1.2",
				correctedFEC:       "",
				uncorrectedFEC:     "",
				powerTransformFunc: ekinopsPowerTransform10Log10XDivideBy10000,
			},
		}}, nil
	case "PM_1001RR":
		return &ekinopsModuleReaderWrapper{&ekinopsModuleReaderTransponder{
			ekinopsModuleData: moduleData,
			networkLinePortsOIDs: ekinopsTransponderOIDs{
				identifierOID:      ".1.3.6.1.4.1.20044.8.7.7",
				labelOID:           ".1.3.6.1.4.1.20044.8.9.3.2.1.3",
				txPowerOID:         ".1.3.6.1.4.1.20044.8.3.3.27",
				rxPowerOID:         ".1.3.6.1.4.1.20044.8.3.3.28",
				correctedFEC:       "",
				uncorrectedFEC:     "",
				powerTransformFunc: ekinopsPowerTransform10Log10XMinus40,
			},
			clientPortsOIDs: ekinopsTransponderOIDs{
				identifierOID:      ".1.3.6.1.4.1.20044.8.7.6",
				labelOID:           ".1.3.6.1.4.1.20044.8.9.3.1.1.3",
				txPowerOID:         ".1.3.6.1.4.1.20044.8.3.2.19",
				rxPowerOID:         ".1.3.6.1.4.1.20044.8.3.2.20",
				correctedFEC:       "",
				uncorrectedFEC:     "",
				powerTransformFunc: ekinopsPowerTransform10Log10XMinus40,
			},
		}}, nil
	case "PM_OPM8":
		return &ekinopsModuleReaderWrapper{&ekinopsModuleReaderOPM8{
			ekinopsModuleData: moduleData,
			ports: ekinopsOPMOIDs{
				identifierOID: ".1.3.6.1.4.1.20044.66.7.1.1.1.2",
				labelOID:      ".1.3.6.1.4.1.20044.66.9.2.7.1.3",
				rxPowerOID:    ".1.3.6.1.4.1.20044.66.3.2.784.1.2",
				channelsOid:   ".1.3.6.1.4.1.20044.66.3.2",
				powerTransformFunc: func(f float64) float64 {
					if f < 32768 {
						return f / 256
					} else {
						return f/256 - 256
					}
				},
			},
		}}, nil
	}
	return nil, errors.New("no module reader available")
}

type ekinopsModuleData struct {
	slotIdentifier string
	moduleName     string
}

func (e *ekinopsModuleData) getSlotIdentifier() string {
	return e.slotIdentifier
}

func (e *ekinopsModuleData) getModuleName() string {
	return e.moduleName
}

type ekinopsModuleReaderWrapper struct {
	ekinopsModuleReader
}

func (m *ekinopsModuleReaderWrapper) readModuleMetrics(ctx context.Context, interfaces []device.Interface) ([]device.Interface, error) {
	// switch snmp community
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}
	community := con.SNMP.SnmpClient.GetCommunity()
	con.SNMP.SnmpClient.SetCommunity(community + m.getSlotIdentifier())
	defer con.SNMP.SnmpClient.SetCommunity(community)

	return m.ekinopsModuleReader.readModuleMetrics(ctx, interfaces)
}
