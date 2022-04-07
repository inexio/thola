package codecommunicator

import (
	"context"
	"fmt"
	"github.com/inexio/thola/internal/communicator"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/pkg/errors"
)

type codeCommunicator struct {
	deviceClass communicator.Communicator
	parent      communicator.Communicator
}

// GetCodeCommunicator returns the code communicator for the given device class
func GetCodeCommunicator(deviceClass communicator.Communicator, parentNetworkDeviceCommunicator communicator.Communicator) (communicator.Functions, error) {
	if deviceClass == nil {
		return nil, errors.New("device class is empty")
	}
	var base = codeCommunicator{
		deviceClass: deviceClass,
		parent:      parentNetworkDeviceCommunicator,
	}
	classIdentifier := deviceClass.GetIdentifier()
	switch classIdentifier {
	case "ceraos/ip10":
		return &ceraosIP10Communicator{base}, nil
	case "ceraos/ip20":
		return &ceraosIP20Communicator{base}, nil
	case "powerone/acc":
		return &poweroneACCCommunicator{base}, nil
	case "powerone/pcc":
		return &poweronePCCCommunicator{base}, nil
	case "ironware":
		return &ironwareCommunicator{base}, nil
	case "ios":
		return &iosCommunicator{base}, nil
	case "ekinops":
		return &ekinopsCommunicator{base}, nil
	case "adva_fsp3kr7":
		return &advaCommunicator{base}, nil
	case "timos/sas":
		return &timosSASCommunicator{base}, nil
	case "timos":
		return &timosCommunicator{base}, nil
	case "junos":
		return &junosCommunicator{base}, nil
	case "aviat":
		return &aviatCommunicator{base}, nil
	case "fortigate":
		return &fortigateCommunicator{base}, nil
	case "linux":
		return &linuxCommunicator{base}, nil
	case "vmware-esxi":
		return &vmwareESXiCommunicator{base}, nil
	case "aruba":
		return &arubaCommunicator{base}, nil
	case "linux/logpoint":
		return &linuxLogpointCommunicator{base}, nil
	}
	return nil, tholaerr.NewNotFoundError(fmt.Sprintf("no code communicator found for device class identifier '%s'", classIdentifier))
}

func (c *codeCommunicator) GetVendor(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetModel(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetModelSeries(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSerialNumber(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetOSVersion(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetInterfaces(_ context.Context, _ ...groupproperty.Filter) ([]device.Interface, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetCountInterfaces(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetCPUComponentCPULoad(_ context.Context) ([]device.CPU, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetMemoryComponentMemoryUsage(_ context.Context) ([]device.MemoryPool, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetServerComponentProcs(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetServerComponentUsers(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetDiskComponentStorages(_ context.Context) ([]device.DiskComponentStorage, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentAlarmLowVoltageDisconnect(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentBatteryAmperage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentBatteryCapacity(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentBatteryCurrent(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentBatteryRemainingTime(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentBatteryTemperature(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentBatteryVoltage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentCurrentLoad(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentMainsVoltageApplied(_ context.Context) (bool, error) {
	return false, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentRectifierCurrent(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentAgents(_ context.Context) ([]device.SBCComponentAgent, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentRealms(_ context.Context) ([]device.SBCComponentRealm, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentSystemVoltage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentGlobalCallPerSecond(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentGlobalConcurrentSessions(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentActiveLocalContacts(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentTranscodingCapacity(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentLicenseCapacity(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentSystemRedundancy(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetHardwareHealthComponentEnvironmentMonitorState(_ context.Context) (device.HardwareHealthComponentState, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetHardwareHealthComponentFans(_ context.Context) ([]device.HardwareHealthComponentFan, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetHardwareHealthComponentTemperature(_ context.Context) ([]device.HardwareHealthComponentTemperature, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetHardwareHealthComponentVoltage(_ context.Context) ([]device.HardwareHealthComponentVoltage, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetHardwareHealthComponentPowerSupply(_ context.Context) ([]device.HardwareHealthComponentPowerSupply, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentSystemHealthScore(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetHighAvailabilityComponentState(_ context.Context) (device.HighAvailabilityComponentState, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetHighAvailabilityComponentRole(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetHighAvailabilityComponentNodes(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentLastRecordedMessagesPerSecondNormalizer(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentAverageMessagesPerSecondLast5minNormalizer(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentLastRecordedMessagesPerSecondStoreHandler(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentAverageMessagesPerSecondLast5minStoreHandler(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentServicesCurrentlyDown(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentSystemVersion(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentSIEM(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentCpuConsumptionCollection(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentCpuConsumptionNormalization(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentCpuConsumptionEnrichment(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentCpuConsumptionIndexing(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentCpuConsumptionDashboardAlerts(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentMemoryConsumptionCollection(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentMemoryConsumptionNormalization(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentMemoryConsumptionEnrichment(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentMemoryConsumptionIndexing(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentMemoryConsumptionDashboardAlerts(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentQueueCollection(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentQueueNormalization(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentQueueEnrichment(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentQueueIndexing(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentQueueDashboardAlerts(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentActiveSearchProcesses(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentDiskUsageDashboardAlerts(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentZFSPools(_ context.Context) ([]device.SIEMComponentZFSPool, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentRepositories(_ context.Context) ([]device.SIEMComponentRepository, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerVersion(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerIOWait(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerVMSwapiness(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerClusterSize(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerStorageCPUUsage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerStorageMemoryUsage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentConfiguredCapacity(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerStorageAvailableCapacity(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerStorageDFSUsed(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerStorageUnderReplicatedBlocks(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerStorageLiveDataNodes(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerAuthenticatorCPUUsage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerAuthenticatorMemoryUsage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerAuthenticatorServiceStatus(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerAuthenticatorAdminServiceStatus(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerZFSPools(_ context.Context) ([]device.SIEMComponentZFSPool, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentAPIServerVersion(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentAPIServerIOWait(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentAPIServerVMSwapiness(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentAPIServerCPUUsage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentAPIServerMemoryUsage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerProxyCpuUsage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerProxyMemoryUsage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerProxyNumberOfAliveConnections(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerProxyState(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSIEMComponentFabricServerProxyNodesCount(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func filterInterfaces(ctx context.Context, interfaces []device.Interface, filter []groupproperty.Filter) ([]device.Interface, error) {
	if len(filter) == 0 {
		return interfaces, nil
	}

	var propertyGroups groupproperty.PropertyGroups
	err := propertyGroups.Encode(interfaces)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode interfaces to property groups")
	}

	for _, fil := range filter {
		propertyGroups, err = fil.ApplyPropertyGroups(ctx, propertyGroups)
		if err != nil {
			return nil, errors.Wrap(err, "failed to apply filter on property groups")
		}
	}

	var res []device.Interface
	err = propertyGroups.Decode(&res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode property groups to interfaces")
	}

	return res, nil
}
