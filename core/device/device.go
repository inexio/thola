package device

import (
	"context"
	"errors"
)

type ctxKey int

const devicePropertiesKey ctxKey = iota + 1

// Status represents an interface status
type Status string

// All status codes with the corresponding label
const (
	StatusUp             Status = "up"
	StatusDown           Status = "down"
	StatusTesting        Status = "testing"
	StatusUnknown        Status = "unknown"
	StatusDormant        Status = "dormant"
	StatusNotPresent     Status = "notPresent"
	StatusLowerLayerDown Status = "lowerLayerDown"
)

// Device
//
// Device represents a device and has the same structure as Response.
// Response can possibly be removed and replaced by Device.
//
// swagger:model
type Device struct {
	// The os of the device
	//
	// example: routerOS
	Class string `yaml:"class" json:"class" xml:"class"`
	// The properties of the device
	Properties Properties `yaml:"properties" json:"properties" xml:"properties"`
}

// Properties
//
// Properties are properties that can be determined for a device
//
// swagger:model
type Properties struct {
	// Vendor of the device
	//
	// example: Mikrotik
	Vendor *string `yaml:"vendor" json:"vendor" xml:"vendor"`
	// Model of the device
	//
	// example: CHR
	Model *string `yaml:"model" json:"model" xml:"model"`
	// ModelSeries of the device
	//
	// example: null
	ModelSeries *string `yaml:"model_series" json:"model_series" xml:"model_series"`
	// SerialNumber of the device
	//
	// example: null
	SerialNumber *string `yaml:"serial_number" json:"serial_number" xml:"serial_number"`
	// OSVersion of the device
	//
	// example: 6.44.6
	OSVersion *string `yaml:"os_version" json:"os_version" xml:"os_version"`
}

// Interface represents all interface values which can be read
type Interface struct {
	IfIndex              *uint64 `yaml:"ifIndex" json:"ifIndex" xml:"ifIndex"`
	IfDescr              *string `yaml:"ifDescr" json:"ifDescr" xml:"ifDescr"`
	IfType               *string `yaml:"ifType" json:"ifType" xml:"ifType"`
	IfMtu                *uint64 `yaml:"ifMtu" json:"ifMtu" xml:"ifMtu"`
	IfSpeed              *uint64 `yaml:"ifSpeed" json:"ifSpeed" xml:"ifSpeed"`
	IfPhysAddress        *string `yaml:"ifPhysAddress" json:"ifPhysAddress" xml:"ifPhysAddress"`
	IfAdminStatus        *Status `yaml:"ifAdminStatus" json:"ifAdminStatus" xml:"ifAdminStatus"`
	IfOperStatus         *Status `yaml:"ifOperStatus" json:"ifOperStatus" xml:"ifOperStatus"`
	IfLastChange         *uint64 `yaml:"ifLastChange" json:"ifLastChange" xml:"ifLastChange"`
	IfInOctets           *uint64 `yaml:"ifInOctets" json:"ifInOctets" xml:"ifInOctets"`
	IfInUcastPkts        *uint64 `yaml:"ifInUcastPkts" json:"ifInUcastPkts" xml:"ifInUcastPkts"`
	IfInNUcastPkts       *uint64 `yaml:"ifInNUcastPkts" json:"ifInNUcastPkts" xml:"ifInNUcastPkts"`
	IfInDiscards         *uint64 `yaml:"ifInDiscards" json:"ifInDiscards" xml:"ifInDiscards"`
	IfInErrors           *uint64 `yaml:"ifInErrors" json:"ifInErrors" xml:"ifInErrors"`
	IfInUnknownProtos    *uint64 `yaml:"ifInUnknownProtos" json:"ifInUnknownProtos" xml:"ifInUnknownProtos"`
	IfOutOctets          *uint64 `yaml:"ifOutOctets" json:"ifOutOctets" xml:"ifOutOctets"`
	IfOutUcastPkts       *uint64 `yaml:"ifOutUcastPkts" json:"ifOutUcastPkts" xml:"ifOutUcastPkts"`
	IfOutNUcastPkts      *uint64 `yaml:"ifOutNUcastPkts" json:"ifOutNUcastPkts" xml:"ifOutNUcastPkts"`
	IfOutDiscards        *uint64 `yaml:"ifOutDiscards" json:"ifOutDiscards" xml:"ifOutDiscards"`
	IfOutErrors          *uint64 `yaml:"ifOutErrors" json:"ifOutErrors" xml:"ifOutErrors"`
	IfOutQLen            *uint64 `yaml:"ifOutQLen" json:"ifOutQLen" xml:"ifOutQLen"`
	IfSpecific           *string `yaml:"ifSpecific" json:"ifSpecific" xml:"ifSpecific"`
	IfName               *string `yaml:"ifName" json:"ifName" xml:"ifName"`
	IfInMulticastPkts    *uint64 `yaml:"ifInMulticastPkts" json:"ifInMulticastPkts" xml:"ifInMulticastPkts"`
	IfInBroadcastPkts    *uint64 `yaml:"ifInBroadcastPkts" json:"ifInBroadcastPkts" xml:"ifInBroadcastPkts"`
	IfOutMulticastPkts   *uint64 `yaml:"ifOutMulticastPkts" json:"ifOutMulticastPkts" xml:"ifOutMulticastPkts"`
	IfOutBroadcastPkts   *uint64 `yaml:"ifOutBroadcastPkts" json:"ifOutBroadcastPkts" xml:"ifOutBroadcastPkts"`
	IfHCInOctets         *uint64 `yaml:"ifHCInOctets" json:"ifHCInOctets" xml:"ifHCInOctets"`
	IfHCInUcastPkts      *uint64 `yaml:"ifHCInUcastPkts" json:"ifHCInUcastPkts" xml:"ifHCInUcastPkts"`
	IfHCInMulticastPkts  *uint64 `yaml:"ifHCInMulticastPkts" json:"ifHCInMulticastPkts" xml:"ifHCInMulticastPkts"`
	IfHCInBroadcastPkts  *uint64 `yaml:"ifHCInBroadcastPkts" json:"ifHCInBroadcastPkts" xml:"ifHCInBroadcastPkts"`
	IfHCOutOctets        *uint64 `yaml:"ifHCOutOctets" json:"ifHCOutOctets" xml:"ifHCOutOctets"`
	IfHCOutUcastPkts     *uint64 `yaml:"ifHCOutUcastPkts" json:"ifHCOutUcastPkts" xml:"ifHCOutUcastPkts"`
	IfHCOutMulticastPkts *uint64 `yaml:"ifHCOutMulticastPkts" json:"ifHCOutMulticastPkts" xml:"ifHCOutMulticastPkts"`
	IfHCOutBroadcastPkts *uint64 `yaml:"ifHCOutBroadcastPkts" json:"ifHCOutBroadcastPkts" xml:"ifHCOutBroadcastPkts"`
	IfHighSpeed          *uint64 `yaml:"ifHighSpeed" json:"ifHighSpeed" xml:"ifHighSpeed"`
	IfAlias              *string `yaml:"ifAlias" json:"ifAlias" xml:"ifAlias"`

	EthernetLikeInterface `mapstructure:",squash"`
	RadioInterface        `mapstructure:",squash"`
	DWDMInterface         `mapstructure:",squash"`
}

// EthernetLikeInterface represents an ethernet like interface
type EthernetLikeInterface struct {
	Dot3StatsAlignmentErrors           *uint64 `yaml:"dot3StatsAlignmentErrors,omitempty" json:"dot3StatsAlignmentErrors,omitempty" xml:"dot3StatsAlignmentErrors,omitempty"`
	Dot3StatsFCSErrors                 *uint64 `yaml:"dot3StatsFCSErrors,omitempty" json:"dot3StatsFCSErrors,omitempty" xml:"dot3StatsFCSErrors,omitempty"`
	Dot3StatsSingleCollisionFrames     *uint64 `yaml:"dot3StatsSingleCollisionFrames,omitempty" json:"dot3StatsSingleCollisionFrames,omitempty" xml:"dot3StatsSingleCollisionFrames,omitempty"`
	Dot3StatsMultipleCollisionFrames   *uint64 `yaml:"dot3StatsMultipleCollisionFrames,omitempty" json:"dot3StatsMultipleCollisionFrames,omitempty" xml:"dot3StatsMultipleCollisionFrames,omitempty"`
	Dot3StatsSQETestErrors             *uint64 `yaml:"dot3StatsSQETestErrors,omitempty" json:"dot3StatsSQETestErrors,omitempty" xml:"dot3StatsSQETestErrors,omitempty"`
	Dot3StatsDeferredTransmissions     *uint64 `yaml:"dot3StatsDeferredTransmissions,omitempty" json:"dot3StatsDeferredTransmissions,omitempty" xml:"dot3StatsDeferredTransmissions,omitempty"`
	Dot3StatsLateCollisions            *uint64 `yaml:"dot3StatsLateCollisions,omitempty" json:"dot3StatsLateCollisions,omitempty" xml:"dot3StatsLateCollisions,omitempty"`
	Dot3StatsExcessiveCollisions       *uint64 `yaml:"dot3StatsExcessiveCollisions,omitempty" json:"dot3StatsExcessiveCollisions,omitempty" xml:"dot3StatsExcessiveCollisions,omitempty"`
	Dot3StatsInternalMacTransmitErrors *uint64 `yaml:"dot3StatsInternalMacTransmitErrors,omitempty" json:"dot3StatsInternalMacTransmitErrors,omitempty" xml:"dot3StatsInternalMacTransmitErrors,omitempty"`
	Dot3StatsCarrierSenseErrors        *uint64 `yaml:"dot3StatsCarrierSenseErrors,omitempty" json:"dot3StatsCarrierSenseErrors,omitempty" xml:"dot3StatsCarrierSenseErrors,omitempty"`
	Dot3StatsFrameTooLongs             *uint64 `yaml:"dot3StatsFrameTooLongs,omitempty" json:"dot3StatsFrameTooLongs,omitempty" xml:"dot3StatsFrameTooLongs,omitempty"`
	Dot3StatsInternalMacReceiveErrors  *uint64 `yaml:"dot3StatsInternalMacReceiveErrors,omitempty" json:"dot3StatsInternalMacReceiveErrors,omitempty" xml:"dot3StatsInternalMacReceiveErrors,omitempty"`
	Dot3HCStatsFCSErrors               *uint64 `yaml:"dot3HCStatsFCSErrors,omitempty" json:"dot3HCStatsFCSErrors,omitempty" xml:"dot3HCStatsFCSErrors,omitempty"`
	EtherStatsCRCAlignErrors           *uint64 `yaml:"etherStatsCRCAlignErrors ,omitempty" json:"etherStatsCRCAlignErrors,omitempty" xml:"etherStatsCRCAlignErrors,omitempty"`
}

// RadioInterface represents a radio interface
type RadioInterface struct {
	LevelOut      *int64  `yaml:"level_out,omitempty" json:"level_out,omitempty" xml:"level_out,omitempty" mapstructure:"level_out"`
	LevelIn       *int64  `yaml:"level_in,omitempty" json:"level_in,omitempty" xml:"level_in,omitempty" mapstructure:"level_in"`
	MaxbitrateOut *uint64 `yaml:"maxbitrate_out,omitempty" json:"maxbitrate_out,omitempty" xml:"maxbitrate_out,omitempty" mapstructure:"maxbitrate_out"`
	MaxbitrateIn  *uint64 `yaml:"maxbitrate_in,omitempty" json:"maxbitrate_in,omitempty" xml:"maxbitrate_in,omitempty" mapstructure:"maxbitrate_in"`
}

// DWDMInterface represents a DWDM interface
type DWDMInterface struct {
	RXLevel *float64 `yaml:"rx_level,omitempty" json:"rx_level,omitempty" xml:"rx_level,omitempty" mapstructure:"rx_level"`
	TXLevel *float64 `yaml:"tx_level,omitempty" json:"tx_level,omitempty" xml:"tx_level,omitempty" mapstructure:"tx_level"`
}

// UPSComponent represents a UPS component
type UPSComponent struct {
	AlarmLowVoltageDisconnect *int     `yaml:"alarm_low_voltage_disconnect" json:"alarm_low_voltage_disconnect" xml:"alarm_low_voltage_disconnect"`
	BatteryAmperage           *float64 `yaml:"battery_amperage " json:"battery_amperage " xml:"battery_amperage"`
	BatteryCapacity           *float64 `yaml:"battery_capacity" json:"battery_capacity" xml:"battery_capacity"`
	BatteryCurrent            *float64 `yaml:"battery_current" json:"battery_current" xml:"battery_current"`
	BatteryRemainingTime      *float64 `yaml:"battery_remaining_time" json:"battery_remaining_time" xml:"battery_remaining_time"`
	BatteryTemperature        *float64 `yaml:"battery_temperature" json:"battery_temperature" xml:"battery_temperature"`
	BatteryVoltage            *float64 `yaml:"battery_voltage" json:"battery_voltage" xml:"battery_voltage"`
	CurrentLoad               *float64 `yaml:"current_load" json:"current_load" xml:"current_load"`
	MainsVoltageApplied       *bool    `yaml:"mains_voltage_applied" json:"mains_voltage_applied" xml:"mains_voltage_applied"`
	RectifierCurrent          *float64 `yaml:"rectifier_current" json:"rectifier_current" xml:"rectifier_current"`
	SystemVoltage             *float64 `yaml:"system_voltage" json:"system_voltage" xml:"system_voltage"`
}

// CPUComponent represents a CPU component
type CPUComponent struct {
	Load        []float64 `yaml:"load" json:"load" xml:"load"`
	Temperature []float64 `yaml:"temperature" json:"temperature" xml:"temperature"`
}

// NewContextWithDeviceProperties returns a new context with the device properties
func NewContextWithDeviceProperties(ctx context.Context, properties Device) context.Context {
	return context.WithValue(ctx, devicePropertiesKey, properties)
}

// DevicePropertiesFromContext returns the device properties from the context
func DevicePropertiesFromContext(ctx context.Context) (Device, bool) {
	properties, ok := ctx.Value(devicePropertiesKey).(Device)
	return properties, ok
}

// ToStatusCode returns the status as a code
func (s Status) ToStatusCode() (int, error) {
	switch s {
	case StatusUp:
		return 1, nil
	case StatusDown:
		return 2, nil
	case StatusTesting:
		return 3, nil
	case StatusUnknown:
		return 4, nil
	case StatusDormant:
		return 5, nil
	case StatusNotPresent:
		return 6, nil
	case StatusLowerLayerDown:
		return 7, nil
	default:
		return 0, errors.New("invalid status")
	}
}
