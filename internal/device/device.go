package device

import (
	"context"
	"errors"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
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

// PerformanceDataPointModifier is used to overwrite PerformanceDataPoints
type PerformanceDataPointModifier func(p *monitoringplugin.PerformanceDataPoint)

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
	IfIndex              *uint64 `yaml:"ifIndex" json:"ifIndex" xml:"ifIndex" mapstructure:"ifIndex"`
	IfDescr              *string `yaml:"ifDescr" json:"ifDescr" xml:"ifDescr" mapstructure:"ifDescr"`
	IfType               *string `yaml:"ifType" json:"ifType" xml:"ifType" mapstructure:"ifType"`
	IfMtu                *uint64 `yaml:"ifMtu" json:"ifMtu" xml:"ifMtu" mapstructure:"ifMtu"`
	IfSpeed              *uint64 `yaml:"ifSpeed" json:"ifSpeed" xml:"ifSpeed" mapstructure:"ifSpeed"`
	IfPhysAddress        *string `yaml:"ifPhysAddress" json:"ifPhysAddress" xml:"ifPhysAddress" mapstructure:"ifPhysAddress"`
	IfAdminStatus        *Status `yaml:"ifAdminStatus" json:"ifAdminStatus" xml:"ifAdminStatus" mapstructure:"ifAdminStatus"`
	IfOperStatus         *Status `yaml:"ifOperStatus" json:"ifOperStatus" xml:"ifOperStatus" mapstructure:"ifOperStatus"`
	IfLastChange         *uint64 `yaml:"ifLastChange" json:"ifLastChange" xml:"ifLastChange" mapstructure:"ifLastChange"`
	IfInOctets           *uint64 `yaml:"ifInOctets" json:"ifInOctets" xml:"ifInOctets" mapstructure:"ifInOctets"`
	IfInUcastPkts        *uint64 `yaml:"ifInUcastPkts" json:"ifInUcastPkts" xml:"ifInUcastPkts" mapstructure:"ifInUcastPkts"`
	IfInNUcastPkts       *uint64 `yaml:"ifInNUcastPkts" json:"ifInNUcastPkts" xml:"ifInNUcastPkts" mapstructure:"ifInNUcastPkts"`
	IfInDiscards         *uint64 `yaml:"ifInDiscards" json:"ifInDiscards" xml:"ifInDiscards" mapstructure:"ifInDiscards"`
	IfInErrors           *uint64 `yaml:"ifInErrors" json:"ifInErrors" xml:"ifInErrors" mapstructure:"ifInErrors"`
	IfInUnknownProtos    *uint64 `yaml:"ifInUnknownProtos" json:"ifInUnknownProtos" xml:"ifInUnknownProtos" mapstructure:"ifInUnknownProtos"`
	IfOutOctets          *uint64 `yaml:"ifOutOctets" json:"ifOutOctets" xml:"ifOutOctets" mapstructure:"ifOutOctets"`
	IfOutUcastPkts       *uint64 `yaml:"ifOutUcastPkts" json:"ifOutUcastPkts" xml:"ifOutUcastPkts" mapstructure:"ifOutUcastPkts"`
	IfOutNUcastPkts      *uint64 `yaml:"ifOutNUcastPkts" json:"ifOutNUcastPkts" xml:"ifOutNUcastPkts" mapstructure:"ifOutNUcastPkts"`
	IfOutDiscards        *uint64 `yaml:"ifOutDiscards" json:"ifOutDiscards" xml:"ifOutDiscards" mapstructure:"ifOutDiscards"`
	IfOutErrors          *uint64 `yaml:"ifOutErrors" json:"ifOutErrors" xml:"ifOutErrors" mapstructure:"ifOutErrors"`
	IfOutQLen            *uint64 `yaml:"ifOutQLen" json:"ifOutQLen" xml:"ifOutQLen" mapstructure:"ifOutQLen"`
	IfSpecific           *string `yaml:"ifSpecific" json:"ifSpecific" xml:"ifSpecific" mapstructure:"ifSpecific"`
	IfName               *string `yaml:"ifName" json:"ifName" xml:"ifName" mapstructure:"ifName"`
	IfInMulticastPkts    *uint64 `yaml:"ifInMulticastPkts" json:"ifInMulticastPkts" xml:"ifInMulticastPkts" mapstructure:"ifInMulticastPkts"`
	IfInBroadcastPkts    *uint64 `yaml:"ifInBroadcastPkts" json:"ifInBroadcastPkts" xml:"ifInBroadcastPkts" mapstructure:"ifInBroadcastPkts"`
	IfOutMulticastPkts   *uint64 `yaml:"ifOutMulticastPkts" json:"ifOutMulticastPkts" xml:"ifOutMulticastPkts" mapstructure:"ifOutMulticastPkts"`
	IfOutBroadcastPkts   *uint64 `yaml:"ifOutBroadcastPkts" json:"ifOutBroadcastPkts" xml:"ifOutBroadcastPkts" mapstructure:"ifOutBroadcastPkts"`
	IfHCInOctets         *uint64 `yaml:"ifHCInOctets" json:"ifHCInOctets" xml:"ifHCInOctets" mapstructure:"ifHCInOctets"`
	IfHCInUcastPkts      *uint64 `yaml:"ifHCInUcastPkts" json:"ifHCInUcastPkts" xml:"ifHCInUcastPkts" mapstructure:"ifHCInUcastPkts"`
	IfHCInMulticastPkts  *uint64 `yaml:"ifHCInMulticastPkts" json:"ifHCInMulticastPkts" xml:"ifHCInMulticastPkts" mapstructure:"ifHCInMulticastPkts"`
	IfHCInBroadcastPkts  *uint64 `yaml:"ifHCInBroadcastPkts" json:"ifHCInBroadcastPkts" xml:"ifHCInBroadcastPkts" mapstructure:"ifHCInBroadcastPkts"`
	IfHCOutOctets        *uint64 `yaml:"ifHCOutOctets" json:"ifHCOutOctets" xml:"ifHCOutOctets" mapstructure:"ifHCOutOctets"`
	IfHCOutUcastPkts     *uint64 `yaml:"ifHCOutUcastPkts" json:"ifHCOutUcastPkts" xml:"ifHCOutUcastPkts" mapstructure:"ifHCOutUcastPkts"`
	IfHCOutMulticastPkts *uint64 `yaml:"ifHCOutMulticastPkts" json:"ifHCOutMulticastPkts" xml:"ifHCOutMulticastPkts" mapstructure:"ifHCOutMulticastPkts"`
	IfHCOutBroadcastPkts *uint64 `yaml:"ifHCOutBroadcastPkts" json:"ifHCOutBroadcastPkts" xml:"ifHCOutBroadcastPkts" mapstructure:"ifHCOutBroadcastPkts"`
	IfHighSpeed          *uint64 `yaml:"ifHighSpeed" json:"ifHighSpeed" xml:"ifHighSpeed" mapstructure:"ifHighSpeed"`
	IfAlias              *string `yaml:"ifAlias" json:"ifAlias" xml:"ifAlias" mapstructure:"ifAlias"`

	// MaxSpeedIn and MaxSpeedOut are set if an interface has different values for max speed in / out
	MaxSpeedIn  *uint64 `yaml:"max_speed_in" json:"max_speed_in" xml:"max_speed_in" mapstructure:"max_speed_in"`
	MaxSpeedOut *uint64 `yaml:"max_speed_out" json:"max_speed_out" xml:"max_speed_out" mapstructure:"max_speed_out"`

	// SubType is not set per default and cannot be read out through a device class.
	// It is used to internally specify a port type, without changing the actual ifType.
	SubType *string `yaml:"-" json:"-" xml:"-"`

	EthernetLike       *EthernetLikeInterface       `yaml:"ethernet_like,omitempty" json:"ethernet_like,omitempty" xml:"ethernet_like,omitempty" mapstructure:"ethernet_like,omitempty"`
	Radio              *RadioInterface              `yaml:"radio,omitempty" json:"radio,omitempty" xml:"radio,omitempty" mapstructure:"radio,omitempty"`
	DWDM               *DWDMInterface               `yaml:"dwdm,omitempty" json:"dwdm,omitempty" xml:"dwdm,omitempty" mapstructure:"dwdm,omitempty"`
	OpticalTransponder *OpticalTransponderInterface `yaml:"optical_transponder,omitempty" json:"optical_transponder,omitempty" xml:"optical_transponder,omitempty" mapstructure:"optical_transponder,omitempty"`
	OpticalAmplifier   *OpticalAmplifierInterface   `yaml:"optical_amplifier,omitempty" json:"optical_amplifier,omitempty" xml:"optical_amplifier,omitempty" mapstructure:"optical_amplifier,omitempty"`
	OpticalOPM         *OpticalOPMInterface         `yaml:"optical_opm,omitempty" json:"optical_opm,omitempty" xml:"optical_opm,omitempty" mapstructure:"optical_opm,omitempty"`
	SAP                *SAPInterface                `yaml:"sap,omitempty" json:"sap,omitempty" xml:"sap,omitempty" mapstructure:"sap,omitempty"`
	VLAN               *VLANInformation             `yaml:"vlan,omitempty" json:"vlan,omitempty" xml:"vlan,omitempty" mapstructure:"vlan,omitempty"`
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
	Dot3StatsAlignmentErrors             *uint64 `yaml:"dot3StatsAlignmentErrors" json:"dot3StatsAlignmentErrors" xml:"dot3StatsAlignmentErrors" mapstructure:"dot3StatsAlignmentErrors"`
	Dot3StatsFCSErrors                   *uint64 `yaml:"dot3StatsFCSErrors" json:"dot3StatsFCSErrors" xml:"dot3StatsFCSErrors" mapstructure:"dot3StatsFCSErrors"`
	Dot3StatsSingleCollisionFrames       *uint64 `yaml:"dot3StatsSingleCollisionFrames" json:"dot3StatsSingleCollisionFrames" xml:"dot3StatsSingleCollisionFrames" mapstructure:"dot3StatsSingleCollisionFrames"`
	Dot3StatsMultipleCollisionFrames     *uint64 `yaml:"dot3StatsMultipleCollisionFrames" json:"dot3StatsMultipleCollisionFrames" xml:"dot3StatsMultipleCollisionFrames" mapstructure:"dot3StatsMultipleCollisionFrames"`
	Dot3StatsSQETestErrors               *uint64 `yaml:"dot3StatsSQETestErrors" json:"dot3StatsSQETestErrors" xml:"dot3StatsSQETestErrors" mapstructure:"dot3StatsSQETestErrors"`
	Dot3StatsDeferredTransmissions       *uint64 `yaml:"dot3StatsDeferredTransmissions" json:"dot3StatsDeferredTransmissions" xml:"dot3StatsDeferredTransmissions" mapstructure:"dot3StatsDeferredTransmissions"`
	Dot3StatsLateCollisions              *uint64 `yaml:"dot3StatsLateCollisions" json:"dot3StatsLateCollisions" xml:"dot3StatsLateCollisions" mapstructure:"dot3StatsLateCollisions"`
	Dot3StatsExcessiveCollisions         *uint64 `yaml:"dot3StatsExcessiveCollisions" json:"dot3StatsExcessiveCollisions" xml:"dot3StatsExcessiveCollisions" mapstructure:"dot3StatsExcessiveCollisions"`
	Dot3StatsInternalMacTransmitErrors   *uint64 `yaml:"dot3StatsInternalMacTransmitErrors" json:"dot3StatsInternalMacTransmitErrors" xml:"dot3StatsInternalMacTransmitErrors" mapstructure:"dot3StatsInternalMacTransmitErrors"`
	Dot3StatsCarrierSenseErrors          *uint64 `yaml:"dot3StatsCarrierSenseErrors" json:"dot3StatsCarrierSenseErrors" xml:"dot3StatsCarrierSenseErrors" mapstructure:"dot3StatsCarrierSenseErrors"`
	Dot3StatsFrameTooLongs               *uint64 `yaml:"dot3StatsFrameTooLongs" json:"dot3StatsFrameTooLongs" xml:"dot3StatsFrameTooLongs" mapstructure:"dot3StatsFrameTooLongs"`
	Dot3StatsInternalMacReceiveErrors    *uint64 `yaml:"dot3StatsInternalMacReceiveErrors" json:"dot3StatsInternalMacReceiveErrors" xml:"dot3StatsInternalMacReceiveErrors" mapstructure:"dot3StatsInternalMacReceiveErrors"`
	Dot3HCStatsAlignmentErrors           *uint64 `yaml:"dot3HCStatsAlignmentErrors" json:"dot3HCStatsAlignmentErrors" xml:"dot3HCStatsAlignmentErrors" mapstructure:"dot3HCStatsAlignmentErrors"`
	Dot3HCStatsFCSErrors                 *uint64 `yaml:"dot3HCStatsFCSErrors" json:"dot3HCStatsFCSErrors" xml:"dot3HCStatsFCSErrors" mapstructure:"dot3HCStatsFCSErrors"`
	Dot3HCStatsInternalMacTransmitErrors *uint64 `yaml:"dot3HCStatsInternalMacTransmitErrors" json:"dot3HCStatsInternalMacTransmitErrors" xml:"dot3HCStatsInternalMacTransmitErrors" mapstructure:"dot3HCStatsInternalMacTransmitErrors"`
	Dot3HCStatsFrameTooLongs             *uint64 `yaml:"dot3HCStatsFrameTooLongs" json:"dot3HCStatsFrameTooLongs" xml:"dot3HCStatsFrameTooLongs" mapstructure:"dot3HCStatsFrameTooLongs"`
	Dot3HCStatsInternalMacReceiveErrors  *uint64 `yaml:"dot3HCStatsInternalMacReceiveErrors" json:"dot3HCStatsInternalMacReceiveErrors" xml:"dot3HCStatsInternalMacReceiveErrors" mapstructure:"dot3HCStatsInternalMacReceiveErrors"`
	EtherStatsCRCAlignErrors             *uint64 `yaml:"etherStatsCRCAlignErrors" json:"etherStatsCRCAlignErrors" xml:"etherStatsCRCAlignErrors" mapstructure:"etherStatsCRCAlignErrors"`
}

// RadioInterface
//
// RadioInterface represents a radio interface.
//
// swagger:model
type RadioInterface struct {
	LevelIn       *float64       `yaml:"level_in" json:"level_in" xml:"level_in" mapstructure:"level_in"`
	LevelOut      *float64       `yaml:"level_out" json:"level_out" xml:"level_out" mapstructure:"level_out"`
	MaxbitrateIn  *uint64        `yaml:"maxbitrate_in" json:"maxbitrate_in" xml:"maxbitrate_in" mapstructure:"maxbitrate_in"`
	MaxbitrateOut *uint64        `yaml:"maxbitrate_out" json:"maxbitrate_out" xml:"maxbitrate_out" mapstructure:"maxbitrate_out"`
	RXFrequency   *float64       `yaml:"rx_frequency" json:"rx_frequency" xml:"rx_frequency" mapstructure:"rx_frequency"`
	TXFrequency   *float64       `yaml:"tx_frequency" json:"tx_frequency" xml:"tx_frequency" mapstructure:"tx_frequency"`
	Channels      []RadioChannel `yaml:"channels" json:"channels" xml:"channels" mapstructure:"channels"`
}

// RadioChannel
//
// RadioChannel represents a radio channel.
//
// swagger:model
type RadioChannel struct {
	Channel       *string  `yaml:"channel" json:"channel" xml:"channel" mapstructure:"channel"`
	LevelIn       *float64 `yaml:"level_in" json:"level_in" xml:"level_in" mapstructure:"level_in"`
	LevelOut      *float64 `yaml:"level_out" json:"level_out" xml:"level_out" mapstructure:"level_out"`
	MaxbitrateIn  *uint64  `yaml:"maxbitrate_in" json:"maxbitrate_in" xml:"maxbitrate_in" mapstructure:"maxbitrate_in"`
	MaxbitrateOut *uint64  `yaml:"maxbitrate_out" json:"maxbitrate_out" xml:"maxbitrate_out" mapstructure:"maxbitrate_out"`
	RXFrequency   *float64 `yaml:"rx_frequency" json:"rx_frequency" xml:"rx_frequency" mapstructure:"rx_frequency"`
	TXFrequency   *float64 `yaml:"tx_frequency" json:"tx_frequency" xml:"tx_frequency" mapstructure:"tx_frequency"`
}

// DWDMInterface
//
// DWDMInterface represents a DWDM interface.
//
// swagger:model
type DWDMInterface struct {
	RXPower        *float64         `yaml:"rx_power" json:"rx_power" xml:"rx_power" mapstructure:"rx_power"`
	TXPower        *float64         `yaml:"tx_power" json:"tx_power" xml:"tx_power" mapstructure:"tx_power"`
	CorrectedFEC   []Rate           `yaml:"corrected_fec" json:"corrected_fec" xml:"corrected_fec" mapstructure:"corrected_fec"`
	UncorrectedFEC []Rate           `yaml:"uncorrected_fec" json:"uncorrected_fec" xml:"uncorrected_fec" mapstructure:"uncorrected_fec"`
	Channels       []OpticalChannel `yaml:"channels" json:"channels" xml:"channels" mapstructure:"channels"`
}

// OpticalTransponderInterface
//
// OpticalTransponderInterface represents an optical transponder interface.
//
// swagger:model
type OpticalTransponderInterface struct {
	Identifier     *string  `yaml:"identifier" json:"identifier" xml:"identifier" mapstructure:"identifier"`
	Label          *string  `yaml:"label" json:"label" xml:"label" mapstructure:"label"`
	RXPower        *float64 `yaml:"rx_power" json:"rx_power" xml:"rx_power" mapstructure:"rx_power"`
	TXPower        *float64 `yaml:"tx_power" json:"tx_power" xml:"tx_power" mapstructure:"tx_power"`
	CorrectedFEC   *uint64  `yaml:"corrected_fec" json:"corrected_fec" xml:"corrected_fec" mapstructure:"corrected_fec"`
	UncorrectedFEC *uint64  `yaml:"uncorrected_fec" json:"uncorrected_fec" xml:"uncorrected_fec" mapstructure:"uncorrected_fec"`
}

// OpticalAmplifierInterface
//
// OpticalAmplifierInterface represents an optical amplifier interface.
//
// swagger:model
type OpticalAmplifierInterface struct {
	Identifier *string  `yaml:"identifier" json:"identifier" xml:"identifier" mapstructure:"identifier"`
	Label      *string  `yaml:"label" json:"label" xml:"label" mapstructure:"label"`
	RXPower    *float64 `yaml:"rx_power" json:"rx_power" xml:"rx_power" mapstructure:"rx_power"`
	TXPower    *float64 `yaml:"tx_power" json:"tx_power" xml:"tx_power" mapstructure:"tx_power"`
	Gain       *float64 `yaml:"gain" json:"gain" xml:"gain" mapstructure:"gain"`
}

// OpticalOPMInterface
//
// OpticalOPMInterface represents an optical opm interface.
//
// swagger:model
type OpticalOPMInterface struct {
	Identifier *string          `yaml:"identifier" json:"identifier" xml:"identifier" mapstructure:"identifier"`
	Label      *string          `yaml:"label" json:"label" xml:"label" mapstructure:"label"`
	RXPower    *float64         `yaml:"rx_power" json:"rx_power" xml:"rx_power" mapstructure:"rx_power"`
	Channels   []OpticalChannel `yaml:"channels" json:"channels" xml:"channels" mapstructure:"channels"`
}

// OpticalChannel
//
// OpticalChannel represents an optical channel.
//
// swagger:model
type OpticalChannel struct {
	Channel *string  `yaml:"channel" json:"channel" xml:"channel" mapstructure:"channel"`
	RXPower *float64 `yaml:"rx_power" json:"rx_power" xml:"rx_power" mapstructure:"rx_power"`
	TXPower *float64 `yaml:"tx_power" json:"tx_power" xml:"tx_power" mapstructure:"tx_power"`
}

// SAPInterface
//
// SAPInterface represents a service access point interface.
//
// swagger:model
type SAPInterface struct {
	Inbound  *uint64 `yaml:"inbound" json:"inbound" xml:"inbound" mapstructure:"inbound"`
	Outbound *uint64 `yaml:"outbound" json:"outbound" xml:"outbound" mapstructure:"outbound"`
}

// VLANInformation
//
// VLANInformation includes all information regarding the VLANs of the interface.
//
// swagger:model
type VLANInformation struct {
	VLANs []VLAN `yaml:"vlans" json:"vlans" xml:"vlans" mapstructure:"vlans"`
}

// VLAN
//
// VLAN includes all information about a VLAN.
//
// swagger:model
type VLAN struct {
	Name   *string `yaml:"name" json:"name" xml:"name" mapstructure:"name"`
	Status *string `yaml:"status" json:"status" xml:"status" mapstructure:"status"`
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
	CPUs []CPU `yaml:"cpus" json:"cpus" xml:"cpus" mapstructure:"cpus"`
}

// CPU
//
// CPU contains information per CPU.
//
// swagger:model
type CPU struct {
	Label *string  `yaml:"label" json:"label" xml:"label" mapstructure:"label"`
	Load  *float64 `yaml:"load" json:"load" xml:"load" mapstructure:"load"`
}

// MemoryComponent
//
// MemoryComponent represents a Memory component
//
// swagger:model
type MemoryComponent struct {
	Pools []MemoryPool `yaml:"pools" json:"pools" xml:"pools" mapstructure:"pools"`
}

// MemoryPool
//
// MemoryPool contains information per memory pool.
//
// swagger:model
type MemoryPool struct {
	Label                        *string  `yaml:"label" json:"label" xml:"label" mapstructure:"label"`
	Usage                        *float64 `yaml:"usage" json:"usage" xml:"usage" mapstructure:"usage"`
	PerformanceDataPointModifier `yaml:"-" json:"-" xml:"-" human_readable:"-"`
}

// DiskComponent
//
// DiskComponent represents a disk component.
//
// swagger:model
type DiskComponent struct {
	Storages []DiskComponentStorage `yaml:"storages" json:"storages" xml:"storages" mapstructure:"storages"`
}

// DiskComponentStorage
//
// DiskComponentStorage contains information per storage.
//
// swagger:model
type DiskComponentStorage struct {
	Type        *string `yaml:"type" json:"type" xml:"type" mapstructure:"type"`
	Description *string `yaml:"description" json:"description" xml:"description" mapstructure:"description"`
	Available   *uint64 `yaml:"available" json:"available" xml:"available" mapstructure:"available"`
	Used        *uint64 `yaml:"used" json:"used" xml:"used" mapstructure:"used"`
}

// UPSComponent
//
// UPSComponent represents a UPS component.
//
// swagger:model
type UPSComponent struct {
	AlarmLowVoltageDisconnect *int     `yaml:"alarm_low_voltage_disconnect" json:"alarm_low_voltage_disconnect" xml:"alarm_low_voltage_disconnect" mapstructure:"alarm_low_voltage_disconnect"`
	BatteryAmperage           *float64 `yaml:"battery_amperage " json:"battery_amperage " xml:"battery_amperage" mapstructure:"battery_amperage"`
	BatteryCapacity           *float64 `yaml:"battery_capacity" json:"battery_capacity" xml:"battery_capacity" mapstructure:"battery_capacity"`
	BatteryCurrent            *float64 `yaml:"battery_current" json:"battery_current" xml:"battery_current" mapstructure:"battery_current"`
	BatteryRemainingTime      *float64 `yaml:"battery_remaining_time" json:"battery_remaining_time" xml:"battery_remaining_time" mapstructure:"battery_remaining_time"`
	BatteryTemperature        *float64 `yaml:"battery_temperature" json:"battery_temperature" xml:"battery_temperature" mapstructure:"battery_temperature"`
	BatteryVoltage            *float64 `yaml:"battery_voltage" json:"battery_voltage" xml:"battery_voltage" mapstructure:"battery_voltage"`
	CurrentLoad               *float64 `yaml:"current_load" json:"current_load" xml:"current_load" mapstructure:"current_load"`
	MainsVoltageApplied       *bool    `yaml:"mains_voltage_applied" json:"mains_voltage_applied" xml:"mains_voltage_applied" mapstructure:"mains_voltage_applied"`
	RectifierCurrent          *float64 `yaml:"rectifier_current" json:"rectifier_current" xml:"rectifier_current" mapstructure:"rectifier_current"`
	SystemVoltage             *float64 `yaml:"system_voltage" json:"system_voltage" xml:"system_voltage" mapstructure:"system_voltage"`
}

// ServerComponent
//
// ServerComponent represents a server component.
//
// swagger:model
type ServerComponent struct {
	Procs *int `yaml:"procs" json:"procs" xml:"procs" mapstructure:"procs"`
	Users *int `yaml:"users" json:"users" xml:"users" mapstructure:"users"`
}

// SBCComponent
//
// SBCComponent represents a SBC component.
//
// swagger:model
type SBCComponent struct {
	Agents                   []SBCComponentAgent `yaml:"agents" json:"agents" xml:"agents" mapstructure:"agents"`
	Realms                   []SBCComponentRealm `yaml:"realms" json:"realms" xml:"realms" mapstructure:"realms"`
	GlobalCallPerSecond      *int                `yaml:"global_call_per_second" json:"global_call_per_second" xml:"global_call_per_second" mapstructure:"global_call_per_second"`
	GlobalConcurrentSessions *int                `yaml:"global_concurrent_sessions " json:"global_concurrent_sessions " xml:"global_concurrent_sessions" mapstructure:"global_concurrent_sessions"`
	ActiveLocalContacts      *int                `yaml:"active_local_contacts" json:"active_local_contacts" xml:"active_local_contacts" mapstructure:"active_local_contacts"`
	TranscodingCapacity      *int                `yaml:"transcoding_capacity" json:"transcoding_capacity" xml:"transcoding_capacity" mapstructure:"transcoding_capacity"`
	LicenseCapacity          *int                `yaml:"license_capacity" json:"license_capacity" xml:"license_capacity" mapstructure:"license_capacity"`
	SystemRedundancy         *int                `yaml:"system_redundancy" json:"system_redundancy" xml:"system_redundancy" mapstructure:"system_redundancy"`
	SystemHealthScore        *int                `yaml:"system_health_score" json:"system_health_score" xml:"system_health_score" mapstructure:"system_health_score"`
}

// SBCComponentAgent
//
// SBCComponentAgent contains information per agent. (Voice)
//
// swagger:model
type SBCComponentAgent struct {
	Hostname                      *string `yaml:"hostname" json:"hostname" xml:"hostname" mapstructure:"hostname"`
	CurrentActiveSessionsInbound  *int    `yaml:"current_active_sessions_inbound" json:"current_active_sessions_inbound" xml:"current_active_sessions_inbound" mapstructure:"current_active_sessions_inbound"`
	CurrentSessionRateInbound     *int    `yaml:"current_session_rate_inbound" json:"current_session_rate_inbound" xml:"current_session_rate_inbound" mapstructure:"current_session_rate_inbound"`
	CurrentActiveSessionsOutbound *int    `yaml:"current_active_sessions_outbound" json:"current_active_sessions_outbound" xml:"current_active_sessions_outbound" mapstructure:"current_active_sessions_outbound"`
	CurrentSessionRateOutbound    *int    `yaml:"current_session_rate_outbound" json:"current_session_rate_outbound" xml:"current_session_rate_outbound" mapstructure:"current_session_rate_outbound"`
	PeriodASR                     *int    `yaml:"period_asr" json:"period_asr" xml:"period_asr" mapstructure:"period_asr"`
	Status                        *int    `yaml:"status" json:"status" xml:"status" mapstructure:"status"`
}

// SBCComponentRealm
//
// SBCComponentRealm contains information per realm. (Voice)
//
// swagger:model
type SBCComponentRealm struct {
	Name                          *string `yaml:"name" json:"name" xml:"name"`
	CurrentActiveSessionsInbound  *int    `yaml:"current_active_sessions_inbound" json:"current_active_sessions_inbound" xml:"current_active_sessions_inbound" mapstructure:"current_active_sessions_inbound"`
	CurrentSessionRateInbound     *int    `yaml:"current_session_rate_inbound" json:"current_session_rate_inbound" xml:"current_session_rate_inbound" mapstructure:"current_session_rate_inbound"`
	CurrentActiveSessionsOutbound *int    `yaml:"current_active_sessions_outbound" json:"current_active_sessions_outbound" xml:"current_active_sessions_outbound" mapstructure:"current_active_sessions_outbound"`
	CurrentSessionRateOutbound    *int    `yaml:"current_session_rate_outbound" json:"current_session_rate_outbound" xml:"current_session_rate_outbound" mapstructure:"current_session_rate_outbound"`
	PeriodASR                     *int    `yaml:"period_asr" json:"period_asr" xml:"period_asr" mapstructure:"d_asr"`
	ActiveLocalContacts           *int    `yaml:"active_local_contacts" json:"active_local_contacts" xml:"active_local_contacts" mapstructure:"active_local_contacts"`
	Status                        *int    `yaml:"status" json:"status" xml:"status" mapstructure:"status"`
}

// HardwareHealthComponent
//
// HardwareHealthComponent represents hardware health information of a device.
//
// swagger:model
type HardwareHealthComponent struct {
	EnvironmentMonitorState *HardwareHealthComponentState        `yaml:"environment_monitor_state" json:"environment_monitor_state" xml:"environment_monitor_state" mapstructure:"environment_monitor_state"`
	Fans                    []HardwareHealthComponentFan         `yaml:"fans" json:"fans" xml:"fans" mapstructure:"fans"`
	PowerSupply             []HardwareHealthComponentPowerSupply `yaml:"power_supply" json:"power_supply" xml:"power_supply" mapstructure:"power_supply"`
	Temperature             []HardwareHealthComponentTemperature `yaml:"temperature" json:"temperature" xml:"temperature" mapstructure:"temperature"`
	Voltage                 []HardwareHealthComponentVoltage     `yaml:"voltage" json:"voltage" xml:"voltage" mapstructure:"voltage"`
}

// HardwareHealthComponentFan
//
// HardwareHealthComponentFan represents one fan of a device.
//
// swagger:model
type HardwareHealthComponentFan struct {
	Description *string                       `yaml:"description" json:"description" xml:"description" mapstructure:"description"`
	State       *HardwareHealthComponentState `yaml:"state" json:"state" xml:"state" mapstructure:"state"`
}

// HardwareHealthComponentTemperature
//
// HardwareHealthComponentTemperature represents one fan of a device.
//
// swagger:model
type HardwareHealthComponentTemperature struct {
	Description *string                       `yaml:"description" json:"description" xml:"description" mapstructure:"description"`
	Temperature *float64                      `yaml:"temperature" json:"temperature" xml:"temperature" mapstructure:"temperature"`
	State       *HardwareHealthComponentState `yaml:"state" json:"state" xml:"state" mapstructure:"state"`
}

// HardwareHealthComponentVoltage
//
// HardwareHealthComponentVoltage represents the voltage of a device.
//
// swagger:model
type HardwareHealthComponentVoltage struct {
	Description *string                       `yaml:"description" json:"description" xml:"description" mapstructure:"description"`
	Voltage     *float64                      `yaml:"voltage" json:"voltage" xml:"voltage" mapstructure:"voltage"`
	State       *HardwareHealthComponentState `yaml:"state" json:"state" xml:"state" mapstructure:"state"`
}

// HardwareHealthComponentPowerSupply
//
// HardwareHealthComponentPowerSupply represents one power supply of a device.
//
// swagger:model
type HardwareHealthComponentPowerSupply struct {
	Description *string                       `yaml:"description" json:"description" xml:"description" mapstructure:"description"`
	State       *HardwareHealthComponentState `yaml:"state" json:"state" xml:"state" mapstructure:"state"`
}

type HardwareHealthComponentState string

const (
	HardwareHealthComponentStateInitial        HardwareHealthComponentState = "initial"
	HardwareHealthComponentStateNormal         HardwareHealthComponentState = "normal"
	HardwareHealthComponentStateWarning        HardwareHealthComponentState = "warning"
	HardwareHealthComponentStateCritical       HardwareHealthComponentState = "critical"
	HardwareHealthComponentStateShutdown       HardwareHealthComponentState = "shutdown"
	HardwareHealthComponentStateNotPresent     HardwareHealthComponentState = "not_present"
	HardwareHealthComponentStateNotFunctioning HardwareHealthComponentState = "not_functioning"
	HardwareHealthComponentStateUnknown        HardwareHealthComponentState = "unknown"
)

func (h HardwareHealthComponentState) GetInt() (int, error) {
	switch h {
	case HardwareHealthComponentStateInitial:
		return 0, nil
	case HardwareHealthComponentStateNormal:
		return 1, nil
	case HardwareHealthComponentStateWarning:
		return 2, nil
	case HardwareHealthComponentStateCritical:
		return 3, nil
	case HardwareHealthComponentStateShutdown:
		return 4, nil
	case HardwareHealthComponentStateNotPresent:
		return 5, nil
	case HardwareHealthComponentStateNotFunctioning:
		return 6, nil
	case HardwareHealthComponentStateUnknown:
		return 7, nil
	}
	return 7, fmt.Errorf("invalid hardware health state '%s'", h)
}

// HighAvailabilityComponent
//
// HighAvailabilityComponent represents high availability information of a device.
//
// swagger:model
type HighAvailabilityComponent struct {
	State *HighAvailabilityComponentState `yaml:"state" json:"state" xml:"state" mapstructure:"state"`
	Role  *string                         `yaml:"role" json:"role" xml:"role" mapstructure:"role"`
	Nodes *int                            `yaml:"nodes" json:"nodes" xml:"nodes" mapstructure:"nodes"`
}

type HighAvailabilityComponentState string

const (
	HighAvailabilityComponentStateUnsynchronized HighAvailabilityComponentState = "unsynchronized"
	HighAvailabilityComponentStateSynchronized   HighAvailabilityComponentState = "synchronized"
	HighAvailabilityComponentStateStandalone     HighAvailabilityComponentState = "standalone"
)

func (h HighAvailabilityComponentState) GetInt() (int, error) {
	switch h {
	case HighAvailabilityComponentStateUnsynchronized:
		return 0, nil
	case HighAvailabilityComponentStateSynchronized:
		return 1, nil
	case HighAvailabilityComponentStateStandalone:
		return 2, nil
	}
	return 0, fmt.Errorf("invalid high availability state '%s'", h)
}

// Rate
//
// Rate encapsulates values which refer to a time span.
//
// swagger:model
type Rate struct {
	Time  string  `yaml:"time" json:"time" xml:"time" mapstructure:"time"`
	Value float64 `yaml:"value" json:"value" xml:"value" mapstructure:"value"`
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

// GetStatus returns the Status that is encoded by the code integer.
func GetStatus(code int) (Status, error) {
	switch code {
	case 1:
		return StatusUp, nil
	case 2:
		return StatusDown, nil
	case 3:
		return StatusTesting, nil
	case 4:
		return StatusUnknown, nil
	case 5:
		return StatusDormant, nil
	case 6:
		return StatusNotPresent, nil
	case 7:
		return StatusLowerLayerDown, nil
	default:
		return "", errors.New("invalid status code")
	}
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
