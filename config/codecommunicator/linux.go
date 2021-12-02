package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/mapping"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
)

type linuxCommunicator struct {
	codeCommunicator
}

// GetDiskComponentStorages returns the cpu load of ios devices.
func (c *linuxCommunicator) GetDiskComponentStorages(ctx context.Context) ([]device.DiskComponentStorage, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("snmp client is empty")
	}

	typeResponses, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.2.1.25.2.3.1.2")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read storage types")
	}
	descriptionResponses, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.2.1.25.2.3.1.3")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read storage description")
	}
	availableResponses, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.2.1.25.2.3.1.5")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read storage available")
	}
	usedResponses, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.2.1.25.2.3.1.6")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read storage used")
	}
	storageUnitResponses, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.2.1.25.2.3.1.4")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read storage unit")
	}

	var res []device.DiskComponentStorage
	for i := range typeResponses {
		var storage device.DiskComponentStorage

		storageTypeValue, err := typeResponses[i].GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get value from snmp response")
		}
		storageType, err := mapping.GetMappedValue("hrStorageType.yaml", storageTypeValue.String())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get mapped storage type")
		}
		if storageType == "Other" || storageType == "RAM" || storageType == "Virtual Memory" {
			continue
		}
		storage.Type = &storageType

		descriptionValue, err := descriptionResponses[i].GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get value from snmp response")
		}
		description := descriptionValue.String()
		storage.Description = &description

		storageUnitValue, err := storageUnitResponses[i].GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get value from snmp response")
		}
		storageUnit, err := storageUnitValue.UInt64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert value to int")
		}

		availableValue, err := availableResponses[i].GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get value from snmp response")
		}
		available, err := availableValue.UInt64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert value to int")
		}
		availableComputed := available * storageUnit
		storage.Available = &availableComputed

		usedValue, err := usedResponses[i].GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get value from snmp response")
		}
		used, err := usedValue.UInt64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert value to int")
		}
		usedComputed := used * storageUnit
		storage.Used = &usedComputed

		res = append(res, storage)
	}

	return res, nil
}
