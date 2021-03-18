package device

import (
	"context"
	"errors"
)

type ctxKey int

const devicePropertiesKey ctxKey = iota + 1

// Status represents an interface status.
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
	// Class of the device.
	//
	// example: routerOS
	Class string `yaml:"class" json:"class" xml:"class"`
	// Properties of the device.
	Properties Properties `yaml:"properties" json:"properties" xml:"properties"`
}

// Properties
//
// Properties are properties that can be determined for a device.
//
// swagger:model
type Properties struct {
	// Vendor of the device.
	//
	// example: Mikrotik
	Vendor *string `yaml:"vendor" json:"vendor" xml:"vendor"`
	// Model of the device.
	//
	// example: CHR
	Model *string `yaml:"model" json:"model" xml:"model"`
	// ModelSeries of the device.
	//
	// example: null
	ModelSeries *string `yaml:"model_series" json:"model_series" xml:"model_series"`
	// SerialNumber of the device.
	//
	// example: null
	SerialNumber *string `yaml:"serial_number" json:"serial_number" xml:"serial_number"`
	// OSVersion of the device.
	//
	// example: 6.44.6
	OSVersion *string `yaml:"os_version" json:"os_version" xml:"os_version"`
}

// Interface
//
// Interface represents all interface values which can be read.
//
// swagger:model
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

	// SubType is not set per default and cannot be read out through a device class.
	// It is used to internally specify a port type, without changing the actual ifType.
	SubType *string `yaml:"-" json:"-" xml:"-"`

	EthernetLike       *EthernetLikeInterface       `yaml:"ethernet_like,omitempty" json:"ethernet_like,omitempty" xml:"ethernet_like,omitempty"`
	Radio              *RadioInterface              `yaml:"radio,omitempty" json:"radio,omitempty" xml:"radio,omitempty" `
	DWDM               *DWDMInterface               `yaml:"dwdm,omitempty" json:"dwdm,omitempty" xml:"dwdm,omitempty"`
	OpticalTransponder *OpticalTransponderInterface `yaml:"optical_transponder,omitempty" json:"optical_transponder,omitempty" xml:"optical_transponder,omitempty"`
	OpticalAmplifier   *OpticalAmplifierInterface   `yaml:"optical_amplifier,omitempty" json:"optical_amplifier,omitempty" xml:"optical_amplifier,omitempty"`
	OpticalOPM         *OpticalOPMInterface         `yaml:"optical_opm,omitempty" json:"optical_opm,omitempty" xml:"optical_opm,omitempty"`
}

//
// Special interface types are defined here.
//

// EthernetLikeInterface
//
// EthernetLikeInterface represents an ethernet like interface.
//
// swagger:model
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

// RadioInterface
//
// RadioInterface represents a radio interface.
//
// swagger:model
type RadioInterface struct {
	LevelOut      *int64  `yaml:"level_out,omitempty" json:"level_out,omitempty" xml:"level_out,omitempty" mapstructure:"level_out"`
	LevelIn       *int64  `yaml:"level_in,omitempty" json:"level_in,omitempty" xml:"level_in,omitempty" mapstructure:"level_in"`
	MaxbitrateOut *uint64 `yaml:"maxbitrate_out,omitempty" json:"maxbitrate_out,omitempty" xml:"maxbitrate_out,omitempty" mapstructure:"maxbitrate_out"`
	MaxbitrateIn  *uint64 `yaml:"maxbitrate_in,omitempty" json:"maxbitrate_in,omitempty" xml:"maxbitrate_in,omitempty" mapstructure:"maxbitrate_in"`
}

// DWDMInterface
//
// DWDMInterface represents a DWDM interface.
//
// swagger:model
type DWDMInterface struct {
	RXPower     *float64         `yaml:"rx_power,omitempty" json:"rx_power,omitempty" xml:"rx_power,omitempty" mapstructure:"rx_power"`
	TXPower     *float64         `yaml:"tx_power,omitempty" json:"tx_power,omitempty" xml:"tx_power,omitempty" mapstructure:"tx_power"`
	RXPower100G *float64         `yaml:"rx_power_100_g,omitempty" json:"rx_power_100_g,omitempty" xml:"rx_power_100_g,omitempty" mapstructure:"rx_power_100_g"`
	TXPower100G *float64         `yaml:"tx_power_100_g,omitempty" json:"tx_power_100_g,omitempty" xml:"tx_power_100_g,omitempty" mapstructure:"tx_power_100_g"`
	Channels    []OpticalChannel `yaml:"channels,omitempty" json:"channels,omitempty" xml:"channels,omitempty" mapstructure:"channels"`
}

// OpticalTransponderInterface
//
// OpticalTransponderInterface represents an optical transponder interface.
//
// swagger:model
type OpticalTransponderInterface struct {
	Identifier     *string  `yaml:"identifier,omitempty" json:"identifier,omitempty" xml:"identifier,omitempty" mapstructure:"identifier"`
	Label          *string  `yaml:"label,omitempty" json:"label,omitempty" xml:"label,omitempty" mapstructure:"label"`
	RXPower        *float64 `yaml:"rx_power,omitempty" json:"rx_power,omitempty" xml:"rx_power,omitempty" mapstructure:"rx_power"`
	TXPower        *float64 `yaml:"tx_power,omitempty" json:"tx_power,omitempty" xml:"tx_power,omitempty" mapstructure:"tx_power"`
	CorrectedFEC   *uint64  `yaml:"corrected_fec,omitempty" json:"corrected_fec,omitempty" xml:"corrected_fec,omitempty" mapstructure:"corrected_fec"`
	UncorrectedFEC *uint64  `yaml:"uncorrected_fec,omitempty" json:"uncorrected_fec,omitempty" xml:"uncorrected_fec,omitempty" mapstructure:"uncorrected_fec"`
}

// OpticalAmplifierInterface
//
// OpticalAmplifierInterface represents an optical amplifier interface.
//
// swagger:model
type OpticalAmplifierInterface struct {
	Identifier *string  `yaml:"identifier,omitempty" json:"identifier,omitempty" xml:"identifier,omitempty" mapstructure:"identifier"`
	Label      *string  `yaml:"label,omitempty" json:"label,omitempty" xml:"label,omitempty" mapstructure:"label"`
	RXPower    *float64 `yaml:"rx_power,omitempty" json:"rx_power,omitempty" xml:"rx_power,omitempty" mapstructure:"rx_power"`
	TXPower    *float64 `yaml:"tx_power,omitempty" json:"tx_power,omitempty" xml:"tx_power,omitempty" mapstructure:"tx_power"`
	Gain       *float64 `yaml:"gain,omitempty" json:"gain,omitempty" xml:"gain,omitempty" mapstructure:"gain"`
}

// OpticalOPMInterface
//
// OpticalOPMInterface represents an optical opm interface.
//
// swagger:model
type OpticalOPMInterface struct {
	Identifier *string          `yaml:"identifier,omitempty" json:"identifier,omitempty" xml:"identifier,omitempty" mapstructure:"identifier"`
	Label      *string          `yaml:"label,omitempty" json:"label,omitempty" xml:"label,omitempty" mapstructure:"label"`
	RXPower    *float64         `yaml:"rx_power,omitempty" json:"rx_power,omitempty" xml:"rx_power,omitempty" mapstructure:"rx_power"`
	Channels   []OpticalChannel `yaml:"channels,omitempty" json:"channels,omitempty" xml:"channels,omitempty" mapstructure:"channels"`
}

// OpticalChannel
//
// OpticalChannel represents an optical channel.
//
// swagger:model
type OpticalChannel struct {
	Channel string   `yaml:"channel,omitempty" json:"channel,omitempty" xml:"channel,omitempty" mapstructure:"channel"`
	RXPower *float64 `yaml:"rx_power,omitempty" json:"rx_power,omitempty" xml:"rx_power,omitempty" mapstructure:"rx_power"`
	TXPower *float64 `yaml:"tx_power,omitempty" json:"tx_power,omitempty" xml:"tx_power,omitempty" mapstructure:"tx_power"`
}

//
// Special device components are defined here.
//

// CPUComponent
//
// CPUComponent represents a CPU component.
//
// swagger:model
type CPUComponent struct {
	Load        []float64 `yaml:"load" json:"load" xml:"load"`
	Temperature []float64 `yaml:"temperature" json:"temperature" xml:"temperature"`
}

// DiskComponent
//
// DiskComponent represents a disk component.
//
// swagger:model
type DiskComponent struct {
	Storages []DiskComponentStorage `yaml:"storages" json:"storages" xml:"storages"`
}

// DiskComponentStorage
//
// DiskComponentStorage contains information per storage.
//
// swagger:model
type DiskComponentStorage struct {
	Type        *string `yaml:"type" json:"type" xml:"type"`
	Description *string `yaml:"description" json:"description" xml:"description"`
	Available   *int    `yaml:"available" json:"available" xml:"available"`
	Used        *int    `yaml:"used" json:"used" xml:"used"`
}

// UPSComponent
//
// UPSComponent represents a UPS component.
//
// swagger:model
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

// ServerComponent
//
// ServerComponent represents a server component.
//
// swagger:model
type ServerComponent struct {
	Procs *int `yaml:"procs" json:"procs" xml:"procs"`
	Users *int `yaml:"users" json:"users" xml:"users"`
}

// SBCComponent
//
// SBCComponent represents a SBC component.
//
// swagger:model
type SBCComponent struct {
	Agents                   []SBCComponentAgent `yaml:"agents" json:"agents" xml:"agents"`
	Realms                   []SBCComponentRealm `yaml:"realms" json:"realms" xml:"realms"`
	GlobalCallPerSecond      *int                `yaml:"global_call_per_second" json:"global_call_per_second" xml:"global_call_per_second"`
	GlobalConcurrentSessions *int                `yaml:"global_concurrent_sessions " json:"global_concurrent_sessions " xml:"global_concurrent_sessions"`
	ActiveLocalContacts      *int                `yaml:"active_local_contacts" json:"active_local_contacts" xml:"active_local_contacts"`
	TranscodingCapacity      *int                `yaml:"transcoding_capacity" json:"transcoding_capacity" xml:"transcoding_capacity"`
	LicenseCapacity          *int                `yaml:"license_capacity" json:"license_capacity" xml:"license_capacity"`
	SystemRedundancy         *int                `yaml:"system_redundancy" json:"system_redundancy" xml:"system_redundancy"`
	SystemHealthScore        *int                `yaml:"system_health_score" json:"system_health_score" xml:"system_health_score"`
}

// SBCComponentAgent
//
// SBCComponentAgent contains information per agent. (Voice)
//
// swagger:model
type SBCComponentAgent struct {
	Hostname                      string `yaml:"hostname" json:"hostname" xml:"hostname" mapstructure:"hostname"`
	CurrentActiveSessionsInbound  *int   `yaml:"current_active_sessions_inbound" json:"current_active_sessions_inbound" xml:"current_active_sessions_inbound" mapstructure:"current_active_sessions_inbound"`
	CurrentSessionRateInbound     *int   `yaml:"current_session_rate_inbound" json:"current_session_rate_inbound" xml:"current_session_rate_inbound" mapstructure:"current_session_rate_inbound"`
	CurrentActiveSessionsOutbound *int   `yaml:"current_active_sessions_outbound" json:"current_active_sessions_outbound" xml:"current_active_sessions_outbound" mapstructure:"current_active_sessions_outbound"`
	CurrentSessionRateOutbound    *int   `yaml:"current_session_rate_outbound" json:"current_session_rate_outbound" xml:"current_session_rate_outbound" mapstructure:"current_session_rate_outbound"`
	PeriodASR                     *int   `yaml:"period_asr" json:"period_asr" xml:"period_asr" mapstructure:"period_asr"`
	Status                        *int   `yaml:"status" json:"status" xml:"status" mapstructure:"status"`
}

// SBCComponentRealm
//
// SBCComponentRealm contains information per realm. (Voice)
//
// swagger:model
type SBCComponentRealm struct {
	Name                          string `yaml:"name" json:"name" xml:"name"`
	CurrentActiveSessionsInbound  *int   `yaml:"current_active_sessions_inbound" json:"current_active_sessions_inbound" xml:"current_active_sessions_inbound" mapstructure:"current_active_sessions_inbound"`
	CurrentSessionRateInbound     *int   `yaml:"current_session_rate_inbound" json:"current_session_rate_inbound" xml:"current_session_rate_inbound" mapstructure:"current_session_rate_inbound"`
	CurrentActiveSessionsOutbound *int   `yaml:"current_active_sessions_outbound" json:"current_active_sessions_outbound" xml:"current_active_sessions_outbound" mapstructure:"current_active_sessions_outbound"`
	CurrentSessionRateOutbound    *int   `yaml:"current_session_rate_outbound" json:"current_session_rate_outbound" xml:"current_session_rate_outbound" mapstructure:"current_session_rate_outbound"`
	PeriodASR                     *int   `yaml:"period_asr" json:"period_asr" xml:"period_asr" mapstructure:"d_asr"`
	ActiveLocalContacts           *int   `yaml:"active_local_contacts" json:"active_local_contacts" xml:"active_local_contacts" mapstructure:"active_local_contacts"`
	Status                        *int   `yaml:"status" json:"status" xml:"status" mapstructure:"status"`
}

// HardwareHealthComponent
//
// HardwareHealthComponent represents hardware health information of a device.
//
// swagger:model
type HardwareHealthComponent struct {
	EnvironmentMonitorState *int                                 `yaml:"environment_monitor_state" json:"environment_monitor_state" xml:"environment_monitor_state"`
	Fans                    []HardwareHealthComponentFan         `yaml:"fans" json:"fans" xml:"fans"`
	PowerSupply             []HardwareHealthComponentPowerSupply `yaml:"power_supply" json:"power_supply" xml:"power_supply"`
}

// HardwareHealthComponentFan
//
// HardwareHealthComponentFan represents one fan of a device.
//
// swagger:model
type HardwareHealthComponentFan struct {
	Description *string `yaml:"description" json:"description" xml:"description"`
	State       *int    `yaml:"state" json:"state" xml:"state"`
}

// HardwareHealthComponentPowerSupply
//
// HardwareHealthComponentPowerSupply represents one power supply of a device.
//
// swagger:model
type HardwareHealthComponentPowerSupply struct {
	Description *string `yaml:"description" json:"description" xml:"description"`
	State       *int    `yaml:"state" json:"state" xml:"state"`
}

// NewContextWithDeviceProperties returns a new context with the device properties.
func NewContextWithDeviceProperties(ctx context.Context, properties Device) context.Context {
	return context.WithValue(ctx, devicePropertiesKey, properties)
}

// DevicePropertiesFromContext returns the device properties from the context.
func DevicePropertiesFromContext(ctx context.Context) (Device, bool) {
	properties, ok := ctx.Value(devicePropertiesKey).(Device)
	return properties, ok
}

// ToStatusCode returns the status as a code.
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
