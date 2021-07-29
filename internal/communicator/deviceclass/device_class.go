// Package deviceclass contains the logic for interacting with device classes.
// It contains methods that read out the .yaml files representing device classes.
package deviceclass

import (
	"context"
	"fmt"
	"github.com/inexio/thola/config"
	"github.com/inexio/thola/config/codecommunicator"
	"github.com/inexio/thola/internal/communicator/communicator"
	"github.com/inexio/thola/internal/communicator/component"
	"github.com/inexio/thola/internal/communicator/hierarchy"
	"github.com/inexio/thola/internal/mapping"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/inexio/thola/internal/value"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// deviceClass represents a device class.
type deviceClass struct {
	name           string
	match          condition
	config         deviceClassConfig
	identify       deviceClassIdentify
	components     deviceClassComponents
	tryToMatchLast bool
}

// deviceClassIdentify represents the identify part of a device class.
type deviceClassIdentify struct {
	properties deviceClassIdentifyProperties
}

// deviceClassIdentifyProperties represents the identify properties part of a device class.
type deviceClassIdentifyProperties struct {
	vendor       propertyReader
	model        propertyReader
	modelSeries  propertyReader
	serialNumber propertyReader
	osVersion    propertyReader
}

// deviceClassComponents represents the components part of a device class.
type deviceClassComponents struct {
	interfaces     *deviceClassComponentsInterfaces
	ups            *deviceClassComponentsUPS
	cpu            *deviceClassComponentsCPU
	memory         *deviceClassComponentsMemory
	sbc            *deviceClassComponentsSBC
	server         *deviceClassComponentsServer
	disk           *deviceClassComponentsDisk
	hardwareHealth *deviceClassComponentsHardwareHealth
}

// deviceClassComponentsUPS represents the ups components part of a device class.
type deviceClassComponentsUPS struct {
	alarmLowVoltageDisconnect propertyReader
	batteryAmperage           propertyReader
	batteryCapacity           propertyReader
	batteryCurrent            propertyReader
	batteryRemainingTime      propertyReader
	batteryTemperature        propertyReader
	batteryVoltage            propertyReader
	currentLoad               propertyReader
	mainsVoltageApplied       propertyReader
	rectifierCurrent          propertyReader
	systemVoltage             propertyReader
}

// deviceClassComponentsCPU represents the cpu components part of a device class.
type deviceClassComponentsCPU struct {
	load        propertyReader
	temperature propertyReader
}

// deviceClassComponentsMemory represents the memory components part of a device class.
type deviceClassComponentsMemory struct {
	usage propertyReader
}

// deviceClassComponentsSBC represents the sbc components part of a device class.
type deviceClassComponentsSBC struct {
	agents                   groupPropertyReader
	realms                   groupPropertyReader
	globalCallPerSecond      propertyReader
	globalConcurrentSessions propertyReader
	activeLocalContacts      propertyReader
	transcodingCapacity      propertyReader
	licenseCapacity          propertyReader
	systemRedundancy         propertyReader
	systemHealthScore        propertyReader
}

// deviceClassComponentsServer represents the server components part of a device class.
type deviceClassComponentsServer struct {
	procs propertyReader
	users propertyReader
}

// deviceClassComponentsDisk represents the disk component part of a device class.
type deviceClassComponentsDisk struct {
	storages groupPropertyReader
}

// deviceClassComponentsHardwareHealth represents the sbc components part of a device class.
type deviceClassComponentsHardwareHealth struct {
	environmentMonitorState propertyReader
	fans                    groupPropertyReader
	powerSupply             groupPropertyReader
}

// deviceClassConfig represents the config part of a device class.
type deviceClassConfig struct {
	snmp       deviceClassSNMP
	components map[component.Component]bool
}

// deviceClassComponentsInterfaces represents the interface properties part of a device class.
type deviceClassComponentsInterfaces struct {
	Count  string
	Values groupPropertyReader
}

// deviceClassSNMP represents the snmp config part of a device class.
type deviceClassSNMP struct {
	MaxRepetitions uint32 `yaml:"max_repetitions"`
}

// logicalOperator represents a logical operator (OR or AND).
type logicalOperator string

// matchMode represents a match mode that is used to match a condition.
type matchMode string

// yamlDeviceClass represents the structure and the parts of a yaml device class.
type yamlDeviceClass struct {
	Name       string                    `yaml:"name"`
	Match      interface{}               `yaml:"match"`
	Identify   yamlDeviceClassIdentify   `yaml:"identify"`
	Config     yamlDeviceClassConfig     `yaml:"config"`
	Components yamlDeviceClassComponents `yaml:"components"`
}

// yamlDeviceClassIdentify represents the identify part of a yaml device class.
type yamlDeviceClassIdentify struct {
	Properties *yamlDeviceClassIdentifyProperties `yaml:"properties"`
}

// yamlDeviceClassComponents represents the components part of a yaml device class.
type yamlDeviceClassComponents struct {
	Interfaces     *yamlComponentsInterfaces               `yaml:"interfaces"`
	UPS            *yamlComponentsUPSProperties            `yaml:"ups"`
	CPU            *yamlComponentsCPUProperties            `yaml:"cpu"`
	Memory         *yamlComponentsMemoryProperties         `yaml:"memory"`
	SBC            *yamlComponentsSBCProperties            `yaml:"sbc"`
	Server         *yamlComponentsServerProperties         `yaml:"server"`
	Disk           *yamlComponentsDiskProperties           `yaml:"disk"`
	HardwareHealth *yamlComponentsHardwareHealthProperties `yaml:"hardware_health"`
}

// yamlDeviceClassConfig represents the config part of a yaml device class.
type yamlDeviceClassConfig struct {
	SNMP       deviceClassSNMP `yaml:"snmp"`
	Components map[string]bool `yaml:"components"`
}

type yamlConditionSet struct {
	LogicalOperator logicalOperator `yaml:"logical_operator" mapstructure:"logical_operator"`
	Conditions      []interface{}
}

// yamlDeviceClassIdentifyProperties represents the identify properties of a yaml device class.
type yamlDeviceClassIdentifyProperties struct {
	Vendor       []interface{} `yaml:"vendor"`
	Model        []interface{} `yaml:"model"`
	ModelSeries  []interface{} `yaml:"model_series"`
	SerialNumber []interface{} `yaml:"serial_number"`
	OSVersion    []interface{} `yaml:"os_version"`
}

//
// Here are definitions of components of yaml device classes.
//

// yamlComponentsUPSProperties represents the specific properties of ups components of a yaml device class.
type yamlComponentsUPSProperties struct {
	AlarmLowVoltageDisconnect []interface{} `yaml:"alarm_low_voltage_disconnect"`
	BatteryAmperage           []interface{} `yaml:"battery_amperage"`
	BatteryCapacity           []interface{} `yaml:"battery_capacity"`
	BatteryCurrent            []interface{} `yaml:"battery_current"`
	BatteryRemainingTime      []interface{} `yaml:"battery_remaining_time"`
	BatteryTemperature        []interface{} `yaml:"battery_temperature"`
	BatteryVoltage            []interface{} `yaml:"battery_voltage"`
	CurrentLoad               []interface{} `yaml:"current_load"`
	MainsVoltageApplied       []interface{} `yaml:"mains_voltage_applied"`
	RectifierCurrent          []interface{} `yaml:"rectifier_current"`
	SystemVoltage             []interface{} `yaml:"system_voltage"`
}

// yamlComponentsCPUProperties represents the specific properties of cpu components of a yaml device class.
type yamlComponentsCPUProperties struct {
	Load        []interface{} `yaml:"load"`
	Temperature []interface{} `yaml:"temperature"`
}

// yamlComponentsMemoryProperties represents the specific properties of memory components of a yaml device class.
type yamlComponentsMemoryProperties struct {
	Usage []interface{} `yaml:"usage"`
}

// yamlComponentsSBCProperties represents the specific properties of sbc components of a yaml device class.
type yamlComponentsSBCProperties struct {
	Agents                   interface{}   `yaml:"agents"`
	Realms                   interface{}   `yaml:"realms"`
	GlobalCallPerSecond      []interface{} `yaml:"global_call_per_second"`
	GlobalConcurrentSessions []interface{} `yaml:"global_concurrent_sessions"`
	ActiveLocalContacts      []interface{} `yaml:"active_local_contacts"`
	TranscodingCapacity      []interface{} `yaml:"transcoding_capacity"`
	LicenseCapacity          []interface{} `yaml:"license_capacity"`
	SystemRedundancy         []interface{} `yaml:"system_redundancy"`
	SystemHealthScore        []interface{} `yaml:"system_health_score"`
}

// yamlComponentsServerProperties represents the specific properties of server components of a yaml device class.
type yamlComponentsServerProperties struct {
	Procs []interface{} `yaml:"procs"`
	Users []interface{} `yaml:"users"`
}

// yamlComponentsDiskProperties represents the specific properties of disk components of a yaml device class.
type yamlComponentsDiskProperties struct {
	Storages interface{} `yaml:"storages"`
}

// yamlComponentsHardwareHealthProperties represents the specific properties of hardware health components of a yaml device class.
type yamlComponentsHardwareHealthProperties struct {
	EnvironmentMonitorState []interface{} `yaml:"environment_monitor_state"`
	Fans                    interface{}   `yaml:"fans"`
	PowerSupply             interface{}   `yaml:"power_supply"`
}

//
// Here are definitions of interfaces of yaml device classes.
//

type yamlComponentsInterfaces struct {
	Count      string      `yaml:"count"`
	Properties interface{} `yaml:"properties"`
}

type yamlComponentsOID struct {
	network.SNMPGetConfiguration `yaml:",inline" mapstructure:",squash"`
	Operators                    []interface{}      `yaml:"operators"`
	IndicesMapping               *yamlComponentsOID `yaml:"indices_mapping" mapstructure:"indices_mapping"`
}

// GetHierarchy returns the hierarchy of device classes merged with their corresponding code communicator.
func GetHierarchy() (hierarchy.Hierarchy, error) {
	genericDeviceClassDir := "device-classes"
	genericDeviceClassFile, err := config.FileSystem.Open(filepath.Join(genericDeviceClassDir, "generic.yaml"))
	if err != nil {
		return hierarchy.Hierarchy{}, errors.Wrap(err, "failed to open generic device class file")
	}
	hier, err := yamlFile2Hierarchy(genericDeviceClassFile, genericDeviceClassDir, nil, nil)
	if err != nil {
		return hierarchy.Hierarchy{}, errors.Wrap(err, "failed to read in generic device class")
	}
	return hier, nil
}

func yamlFile2Hierarchy(file fs.File, directory string, parentDeviceClass *deviceClass, parentCommunicator communicator.Communicator) (hierarchy.Hierarchy, error) {
	//get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return hierarchy.Hierarchy{}, errors.Wrap(err, "failed to get stat for file")
	}

	if !strings.HasSuffix(fileInfo.Name(), ".yaml") {
		return hierarchy.Hierarchy{}, errors.New("only yaml files are allowed for this function")
	}

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return hierarchy.Hierarchy{}, errors.Wrap(err, "failed to read file")
	}
	var deviceClassYaml yamlDeviceClass
	err = yaml.Unmarshal(contents, &deviceClassYaml)
	if err != nil {
		return hierarchy.Hierarchy{}, errors.Wrap(err, "failed to unmarshal config file")
	}

	devClass, err := deviceClassYaml.convert(parentDeviceClass)
	if err != nil {
		return hierarchy.Hierarchy{}, errors.Wrapf(err, "failed to convert yamlData to deviceClass for device class '%s'", deviceClassYaml.Name)
	}

	networkDeviceCommunicator, err := createNetworkDeviceCommunicator(&devClass, parentCommunicator)
	if err != nil {
		return hierarchy.Hierarchy{}, errors.Wrap(err, "failed to create network device communicator")
	}

	hier := hierarchy.Hierarchy{
		NetworkDeviceCommunicator: networkDeviceCommunicator,
		TryToMatchLast:            devClass.tryToMatchLast,
	}

	// check for sub device classes
	subDirPath := filepath.Join(directory, devClass.name)
	subDir, err := config.FileSystem.ReadDir(subDirPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return hierarchy.Hierarchy{}, errors.Wrap(err, "an unexpected error occurred while trying to open sub device class directory")
		}
	} else {
		subHierarchies, err := readDeviceClassDirectory(subDir, subDirPath, &devClass, networkDeviceCommunicator)
		if err != nil {
			return hierarchy.Hierarchy{}, errors.Wrap(err, "failed to read sub device classes")
		}
		hier.Children = subHierarchies
	}

	return hier, nil
}

func createNetworkDeviceCommunicator(devClass *deviceClass, parentCommunicator communicator.Communicator) (communicator.Communicator, error) {
	devClassCommunicator := &(deviceClassCommunicator{devClass})
	codeCommunicator, err := codecommunicator.GetCodeCommunicator(devClassCommunicator, parentCommunicator)
	if err != nil && !tholaerr.IsNotFoundError(err) {
		return nil, errors.Wrap(err, "failed to get code communicator")
	}
	return communicator.CreateNetworkDeviceCommunicator(&(deviceClassCommunicator{devClass}), codeCommunicator), nil
}

func readDeviceClassDirectory(dir []fs.DirEntry, directory string, parentDeviceClass *deviceClass, parentCommunicator communicator.Communicator) (map[string]hierarchy.Hierarchy, error) {
	deviceClasses := make(map[string]hierarchy.Hierarchy)
	for _, dirEntry := range dir {
		// directories will be ignored here, sub device classes dirs will be called when
		// their parent device class is processed
		if dirEntry.IsDir() {
			continue
		}
		fileInfo, err := dirEntry.Info()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get info for file")
		}

		if !strings.HasSuffix(fileInfo.Name(), ".yaml") {
			// all non directory files need to be yaml file and end with ".yaml"
			return nil, errors.New("only yaml config files are allowed in device class directories")
		}
		fullPathToFile := filepath.Join(directory, fileInfo.Name())
		file, err := config.FileSystem.Open(fullPathToFile)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open file "+fullPathToFile)
		}
		hier, err := yamlFile2Hierarchy(file, directory, parentDeviceClass, parentCommunicator)
		if err != nil {
			return nil, errors.Wrapf(err, "an error occurred while trying to read in yaml config file %s", fileInfo.Name())
		}
		deviceClasses[hier.NetworkDeviceCommunicator.GetIdentifier()] = hier
	}

	return deviceClasses, nil
}

// getName returns the name of the device class.
func (d *deviceClass) getName() string {
	return d.name
}

// match checks if data in context matches the device class.
func (d *deviceClass) matchDevice(ctx context.Context) (bool, error) {
	return d.match.check(ctx)
}

// getAvailableComponents returns the available components.
func (d *deviceClass) getAvailableComponents() map[component.Component]bool {
	return d.config.components
}

func (y *yamlDeviceClass) convert(parent *deviceClass) (deviceClass, error) {
	err := y.validate()
	if err != nil {
		return deviceClass{}, errors.Wrap(err, "invalid yaml device class")
	}
	var devClass deviceClass
	if parent != nil {
		devClass = *parent
		if devClass.name != "generic" {
			devClass.name += "/"
		} else {
			devClass.name = ""
		}
	}
	devClass.name += y.Name
	if y.Name == "generic" {
		devClass.match = &alwaysTrueCondition{}
		devClass.identify = deviceClassIdentify{
			properties: deviceClassIdentifyProperties{},
		}
	} else {
		cond, err := interface2condition(y.Match, classifyDevice)
		if err != nil {
			return deviceClass{}, errors.Wrap(err, "failed to convert device class condition")
		}
		devClass.match = cond
		devClass.tryToMatchLast = conditionContainsUniqueRequest(cond)
		identify, err := y.Identify.convert(devClass.identify)
		if err != nil {
			return deviceClass{}, errors.Wrap(err, "failed to convert identify")
		}
		devClass.identify = identify
	}

	devClass.components, err = y.Components.convert(devClass.components)
	if err != nil {
		return deviceClass{}, errors.Wrap(err, "failed to convert components")
	}

	devClass.config, err = y.Config.convert(devClass.config)
	if err != nil {
		return deviceClass{}, errors.Wrap(err, "failed to convert components")
	}

	return devClass, nil
}

func (y *yamlDeviceClass) validate() error {
	if y.Name == "" {
		return errors.New("device class name is empty")
	}
	if y.Name != "generic" && y.Match == nil {
		return errors.New("device class conditions are missing")
	}

	if strings.Contains(y.Name, "/") {
		return errors.New("device class name cannot contain '/'")
	}
	return nil
}

func (y *yamlDeviceClassIdentify) convert(parentIdentify deviceClassIdentify) (deviceClassIdentify, error) {
	err := y.validate()
	if err != nil {
		return deviceClassIdentify{}, errors.Wrap(err, "identify is invalid")
	}
	var identify deviceClassIdentify
	properties, err := y.Properties.convert(parentIdentify.properties)
	if err != nil {
		return deviceClassIdentify{}, errors.Wrap(err, "failed to read yaml identify properties")
	}
	identify.properties = properties

	return identify, nil
}

func (y *yamlDeviceClassComponents) convert(parentComponents deviceClassComponents) (deviceClassComponents, error) {
	components := parentComponents
	var err error

	if y.Interfaces != nil {
		components.interfaces, err = y.Interfaces.convert(parentComponents.interfaces)
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml interface properties")
		}
	}

	if y.UPS != nil {
		ups, err := y.UPS.convert(parentComponents.ups)
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml UPS properties")
		}
		components.ups = &ups
	}

	if y.CPU != nil {
		cpu, err := y.CPU.convert(parentComponents.cpu)
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml CPU properties")
		}
		components.cpu = &cpu
	}

	if y.Memory != nil {
		memory, err := y.Memory.convert(parentComponents.memory)
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml memory properties")
		}
		components.memory = &memory
	}

	if y.SBC != nil {
		sbc, err := y.SBC.convert(parentComponents.sbc)
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml sbc properties")
		}
		components.sbc = &sbc
	}

	if y.Server != nil {
		server, err := y.Server.convert(parentComponents.server)
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml server properties")
		}
		components.server = &server
	}

	if y.Disk != nil {
		disk, err := y.Disk.convert(parentComponents.disk)
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml disk properties")
		}
		components.disk = &disk
	}

	if y.HardwareHealth != nil {
		hardwareHealth, err := y.HardwareHealth.convert(parentComponents.hardwareHealth)
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml hardware health properties")
		}
		components.hardwareHealth = &hardwareHealth
	}

	return components, nil
}

func (y *yamlComponentsInterfaces) convert(parentComponentsInterfaces *deviceClassComponentsInterfaces) (*deviceClassComponentsInterfaces, error) {
	var interfaceComponent deviceClassComponentsInterfaces
	var err error

	if parentComponentsInterfaces != nil {
		interfaceComponent = *parentComponentsInterfaces
	}

	if y.Properties != nil {
		interfaceComponent.Values, err = interface2GroupPropertyReader(y.Properties, interfaceComponent.Values)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert interface properties")
		}
	}

	if y.Count != "" {
		interfaceComponent.Count = y.Count
	}

	return &interfaceComponent, nil
}

func (y *yamlComponentsOID) convert() (deviceClassOID, error) {
	var idxMappings *deviceClassOID
	if y.IndicesMapping != nil {
		mappings, err := y.IndicesMapping.convert()
		if err != nil {
			return deviceClassOID{}, errors.New("failed to convert indices mappings")
		}
		idxMappings = &mappings
	}

	if y.Operators != nil {
		operators, err := interfaceSlice2propertyOperators(y.Operators, propertyDefault)
		if err != nil {
			return deviceClassOID{}, errors.Wrap(err, "failed to read yaml oids operators")
		}
		return deviceClassOID{
			SNMPGetConfiguration: network.SNMPGetConfiguration{
				OID:          y.OID,
				UseRawResult: y.UseRawResult,
			},
			operators:      operators,
			indicesMapping: idxMappings,
		}, nil
	}

	return deviceClassOID{
		SNMPGetConfiguration: network.SNMPGetConfiguration{
			OID:          y.OID,
			UseRawResult: y.UseRawResult,
		},
		operators:      nil,
		indicesMapping: idxMappings,
	}, nil
}

func (y *yamlComponentsOID) validate() error {
	if err := y.OID.Validate(); err != nil {
		return errors.Wrap(err, "oid is invalid")
	}
	return nil
}

func conditionContainsUniqueRequest(c condition) bool {
	switch c := c.(type) {
	case *SnmpCondition:
		if c.Type == "snmpget" {
			return true
		}
	case *HTTPCondition:
		return true
	case *ConditionSet:
		for _, con := range c.Conditions {
			if conditionContainsUniqueRequest(con) {
				return true
			}
		}
	}
	return false
}

func (y *yamlDeviceClassIdentify) validate() error {
	if y.Properties == nil {
		y.Properties = &yamlDeviceClassIdentifyProperties{}
	}
	return nil
}

func (y *yamlDeviceClassIdentifyProperties) convert(parentProperties deviceClassIdentifyProperties) (deviceClassIdentifyProperties, error) {
	properties := parentProperties
	var err error

	if y.Vendor != nil {
		properties.vendor, err = convertYamlProperty(y.Vendor, propertyVendor, properties.vendor)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert vendor property to property reader")
		}
	}
	if y.Model != nil {
		properties.model, err = convertYamlProperty(y.Model, propertyModel, properties.model)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert model property to property reader")
		}
	}
	if y.ModelSeries != nil {
		properties.modelSeries, err = convertYamlProperty(y.ModelSeries, propertyModelSeries, properties.modelSeries)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert model series property to property reader")
		}
	}
	if y.SerialNumber != nil {
		properties.serialNumber, err = convertYamlProperty(y.SerialNumber, propertyDefault, properties.serialNumber)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert serial number property to property reader")
		}
	}
	if y.OSVersion != nil {
		properties.osVersion, err = convertYamlProperty(y.OSVersion, propertyDefault, properties.osVersion)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert osVersion property to property reader")
		}
	}
	return properties, nil
}

func (y *yamlDeviceClassConfig) convert(parentConfig deviceClassConfig) (deviceClassConfig, error) {
	var cfg deviceClassConfig

	if y.SNMP.MaxRepetitions != 0 {
		cfg.snmp.MaxRepetitions = y.SNMP.MaxRepetitions
	} else {
		cfg.snmp.MaxRepetitions = parentConfig.snmp.MaxRepetitions
	}

	components := make(map[component.Component]bool)
	for k, v := range parentConfig.components {
		components[k] = v
	}

	for k, v := range y.Components {
		comp, err := component.CreateComponent(k)
		if err != nil {
			return deviceClassConfig{}, err
		}
		components[comp] = v
	}

	cfg.components = components

	return cfg, nil
}

func (y *yamlComponentsUPSProperties) convert(parentComponent *deviceClassComponentsUPS) (deviceClassComponentsUPS, error) {
	var properties deviceClassComponentsUPS
	var err error
	if parentComponent != nil {
		properties = *parentComponent
	}

	if y.AlarmLowVoltageDisconnect != nil {
		properties.alarmLowVoltageDisconnect, err = convertYamlProperty(y.AlarmLowVoltageDisconnect, propertyDefault, properties.alarmLowVoltageDisconnect)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert alarm low voltage disconnect property to property reader")
		}
	}
	if y.BatteryAmperage != nil {
		properties.batteryAmperage, err = convertYamlProperty(y.BatteryAmperage, propertyDefault, properties.batteryAmperage)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery amperage property to property reader")
		}
	}
	if y.BatteryCapacity != nil {
		properties.batteryCapacity, err = convertYamlProperty(y.BatteryCapacity, propertyDefault, properties.batteryCapacity)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery capacity property to property reader")
		}
	}
	if y.BatteryCurrent != nil {
		properties.batteryCurrent, err = convertYamlProperty(y.BatteryCurrent, propertyDefault, properties.batteryCurrent)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery current property to property reader")
		}
	}
	if y.BatteryRemainingTime != nil {
		properties.batteryRemainingTime, err = convertYamlProperty(y.BatteryRemainingTime, propertyDefault, properties.batteryRemainingTime)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery remaining time property to property reader")
		}
	}
	if y.BatteryTemperature != nil {
		properties.batteryTemperature, err = convertYamlProperty(y.BatteryTemperature, propertyDefault, properties.batteryTemperature)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery temperature property to property reader")
		}
	}
	if y.BatteryVoltage != nil {
		properties.batteryVoltage, err = convertYamlProperty(y.BatteryVoltage, propertyDefault, properties.batteryVoltage)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery voltage property to property reader")
		}
	}
	if y.CurrentLoad != nil {
		properties.currentLoad, err = convertYamlProperty(y.CurrentLoad, propertyDefault, properties.currentLoad)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert current load property to property reader")
		}
	}
	if y.MainsVoltageApplied != nil {
		properties.mainsVoltageApplied, err = convertYamlProperty(y.MainsVoltageApplied, propertyDefault, properties.mainsVoltageApplied)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert mains voltage applied property to property reader")
		}
	}
	if y.RectifierCurrent != nil {
		properties.rectifierCurrent, err = convertYamlProperty(y.RectifierCurrent, propertyDefault, properties.rectifierCurrent)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert rectifier current property to property reader")
		}
	}
	if y.SystemVoltage != nil {
		properties.systemVoltage, err = convertYamlProperty(y.SystemVoltage, propertyDefault, properties.systemVoltage)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert system voltage property to property reader")
		}
	}
	return properties, nil
}

func (y *yamlComponentsCPUProperties) convert(parentComponent *deviceClassComponentsCPU) (deviceClassComponentsCPU, error) {
	var properties deviceClassComponentsCPU
	var err error

	if parentComponent != nil {
		properties = *parentComponent
	}

	if y.Load != nil {
		properties.load, err = convertYamlProperty(y.Load, propertyDefault, properties.load)
		if err != nil {
			return deviceClassComponentsCPU{}, errors.Wrap(err, "failed to convert load property to property reader")
		}
	}
	if y.Temperature != nil {
		properties.temperature, err = convertYamlProperty(y.Temperature, propertyDefault, properties.temperature)
		if err != nil {
			return deviceClassComponentsCPU{}, errors.Wrap(err, "failed to convert temperature property to property reader")
		}
	}
	return properties, nil
}

func (y *yamlComponentsMemoryProperties) convert(parentComponent *deviceClassComponentsMemory) (deviceClassComponentsMemory, error) {
	var properties deviceClassComponentsMemory
	var err error

	if parentComponent != nil {
		properties = *parentComponent
	}

	if y.Usage != nil {
		properties.usage, err = convertYamlProperty(y.Usage, propertyDefault, properties.usage)
		if err != nil {
			return deviceClassComponentsMemory{}, errors.Wrap(err, "failed to convert memory usage property to property reader")
		}
	}
	return properties, nil
}

func (y *yamlComponentsServerProperties) convert(parentComponent *deviceClassComponentsServer) (deviceClassComponentsServer, error) {
	var properties deviceClassComponentsServer
	var err error

	if parentComponent != nil {
		properties = *parentComponent
	}

	if y.Procs != nil {
		properties.procs, err = convertYamlProperty(y.Procs, propertyDefault, properties.procs)
		if err != nil {
			return deviceClassComponentsServer{}, errors.Wrap(err, "failed to convert procs property to property reader")
		}
	}
	if y.Users != nil {
		properties.users, err = convertYamlProperty(y.Users, propertyDefault, properties.procs)
		if err != nil {
			return deviceClassComponentsServer{}, errors.Wrap(err, "failed to convert users property to property reader")
		}
	}
	return properties, nil
}

func (y *yamlComponentsDiskProperties) convert(parentDisk *deviceClassComponentsDisk) (deviceClassComponentsDisk, error) {
	var properties deviceClassComponentsDisk
	var err error

	if parentDisk != nil {
		properties = *parentDisk
	}

	if y.Storages != nil {
		properties.storages, err = interface2GroupPropertyReader(y.Storages, properties.storages)
		if err != nil {
			return deviceClassComponentsDisk{}, errors.Wrap(err, "failed to convert storages property to group property reader")
		}
	}
	return properties, nil
}

func (y *yamlComponentsSBCProperties) convert(parentComponentsSBC *deviceClassComponentsSBC) (deviceClassComponentsSBC, error) {
	var properties deviceClassComponentsSBC
	var err error

	if parentComponentsSBC != nil {
		properties = *parentComponentsSBC
	}

	if y.Agents != nil {
		properties.agents, err = interface2GroupPropertyReader(y.Agents, properties.agents)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert agents property to group property reader")
		}
	}
	if y.Realms != nil {
		properties.realms, err = interface2GroupPropertyReader(y.Realms, properties.realms)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert realms property to group property reader")
		}
	}
	if y.ActiveLocalContacts != nil {
		properties.activeLocalContacts, err = convertYamlProperty(y.ActiveLocalContacts, propertyDefault, properties.activeLocalContacts)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert active local contacts property to property reader")
		}
	}
	if y.GlobalCallPerSecond != nil {
		properties.globalCallPerSecond, err = convertYamlProperty(y.GlobalCallPerSecond, propertyDefault, properties.globalCallPerSecond)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert global call per second property to property reader")
		}
	}
	if y.GlobalConcurrentSessions != nil {
		properties.globalConcurrentSessions, err = convertYamlProperty(y.GlobalConcurrentSessions, propertyDefault, properties.globalConcurrentSessions)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert global concurrent sessions property to property reader")
		}
	}
	if y.LicenseCapacity != nil {
		properties.licenseCapacity, err = convertYamlProperty(y.LicenseCapacity, propertyDefault, properties.licenseCapacity)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert license capacity property to property reader")
		}
	}
	if y.TranscodingCapacity != nil {
		properties.transcodingCapacity, err = convertYamlProperty(y.TranscodingCapacity, propertyDefault, properties.transcodingCapacity)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert transcoding capacity property to property reader")
		}
	}
	if y.SystemRedundancy != nil {
		properties.systemRedundancy, err = convertYamlProperty(y.SystemRedundancy, propertyDefault, properties.systemRedundancy)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert system redundancy property to property reader")
		}
	}

	if y.SystemHealthScore != nil {
		properties.systemHealthScore, err = convertYamlProperty(y.SystemHealthScore, propertyDefault, properties.systemHealthScore)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert system health score property to property reader")
		}
	}
	return properties, nil
}

func (y *yamlComponentsHardwareHealthProperties) convert(parentHardwareHealth *deviceClassComponentsHardwareHealth) (deviceClassComponentsHardwareHealth, error) {
	var properties deviceClassComponentsHardwareHealth
	var err error

	if parentHardwareHealth != nil {
		properties = *parentHardwareHealth
	}

	if y.Fans != nil {
		properties.fans, err = interface2GroupPropertyReader(y.Fans, properties.fans)
		if err != nil {
			return deviceClassComponentsHardwareHealth{}, errors.Wrap(err, "failed to convert fans property to group property reader")
		}
	}
	if y.PowerSupply != nil {
		properties.powerSupply, err = interface2GroupPropertyReader(y.PowerSupply, properties.powerSupply)
		if err != nil {
			return deviceClassComponentsHardwareHealth{}, errors.Wrap(err, "failed to convert power supply property to group property reader")
		}
	}
	if y.EnvironmentMonitorState != nil {
		properties.environmentMonitorState, err = convertYamlProperty(y.EnvironmentMonitorState, propertyDefault, properties.environmentMonitorState)
		if err != nil {
			return deviceClassComponentsHardwareHealth{}, errors.Wrap(err, "failed to convert environment monitor state property to property reader")
		}
	}

	return properties, nil
}

func (y *yamlConditionSet) convert() (condition, error) {
	err := y.validate()
	if err != nil {
		return nil, errors.Wrap(err, "invalid yaml condition set")
	}
	var conditionSet ConditionSet
	for _, condition := range y.Conditions {
		matcher, err := interface2condition(condition, classifyDevice)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert interface to condition")
		}
		conditionSet.Conditions = append(conditionSet.Conditions, matcher)
	}
	conditionSet.LogicalOperator = y.LogicalOperator
	return &conditionSet, nil
}

func (y *yamlConditionSet) validate() error {
	if len(y.Conditions) == 0 {
		return errors.New("empty condition array")
	}
	err := y.LogicalOperator.validate()
	if err != nil {
		if y.LogicalOperator == "" {
			y.LogicalOperator = "OR" // default logical operator is always OR
		}
		return errors.Wrap(err, "invalid logical operator")
	}
	return nil
}

type relatedTask int

const (
	classifyDevice relatedTask = iota + 1
	propertyVendor
	propertyModel
	propertyModelSeries
	propertyDefault
)

func interface2condition(i interface{}, task relatedTask) (condition, error) {
	m, ok := i.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("failed to convert interface to map[interface{}]interface{}")
	}

	var stringType string
	if _, ok := m["type"]; ok {
		stringType, ok = m["type"].(string)
		if !ok {
			return nil, errors.New("condition type needs to be a string")
		}
	} else {
		// if condition type is empty, and it has conditions and optionally a logical operator,
		// and no other attributes, then it will be considered as a conditionSet per default
		if _, ok = m["conditions"]; ok {
			// if there is only "conditions" in the map or only "conditions" and "logical_operator", nothing else
			if _, ok = m["logical_operator"]; (ok && len(m) == 2) || len(m) == 1 {
				stringType = "conditionSet"
			} else {
				return nil, errors.New("no condition type set and attributes do not match conditionSet")
			}
		} else {
			return nil, errors.New("no condition type set and attributes do not match conditionSet")
		}
	}

	if stringType == "conditionSet" {
		var yamlConditionSet yamlConditionSet
		err := mapstructure.Decode(i, &yamlConditionSet)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode conditionSet")
		}
		return yamlConditionSet.convert()
	}
	//SNMP SnmpCondition Types
	if stringType == "SysObjectID" || stringType == "SysDescription" || stringType == "snmpget" {
		var condition SnmpCondition
		err := mapstructure.Decode(i, &condition)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode Condition")
		}
		err = condition.validate()
		if err != nil {
			return nil, errors.Wrap(err, "invalid snmp condition")
		}
		return &condition, nil
	}
	//HTTP
	if stringType == "HttpGetBody" {
		var condition HTTPCondition
		err := mapstructure.Decode(i, &condition)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode condition")
		}
		err = condition.validate()
		if err != nil {
			return nil, errors.Wrap(err, "invalid http condition")
		}
		return &condition, nil
	}

	if stringType == "Vendor" {
		if task <= propertyVendor {
			return nil, errors.New("cannot use vendor condition, vendor is not available here yet")
		}
		var condition VendorCondition
		err := mapstructure.Decode(i, &condition)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode condition")
		}
		return &condition, nil
	}

	if stringType == "Model" {
		if task <= propertyModel {
			return nil, errors.New("cannot use model condition, model is not available here yet")
		}
		var condition ModelCondition
		err := mapstructure.Decode(i, &condition)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode condition")
		}
		return &condition, nil
	}

	if stringType == "ModelSeries" {
		if task <= propertyModelSeries {
			return nil, errors.New("cannot use model series condition, model series is not available here yet")
		}
		var condition ModelSeriesCondition
		err := mapstructure.Decode(i, &condition)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode condition")
		}
		return &condition, nil
	}
	return nil, fmt.Errorf("invalid condition type '%s'", stringType)
}

func convertYamlProperty(i []interface{}, task relatedTask, parentProperty propertyReader) (propertyReader, error) {
	var readerSet propertyReaderSet
	for _, i := range i {
		reader, err := interface2propertyReader(i, task)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert yaml identify property")
		}
		readerSet = append(readerSet, reader)
	}
	if parentProperty != nil {
		readerSet = append(readerSet, parentProperty)
	}
	return &readerSet, nil
}

func interface2propertyReader(i interface{}, task relatedTask) (propertyReader, error) {
	m, ok := i.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("failed to convert interface to map[interface{}]interface{}")
	}
	if _, ok := m["detection"]; !ok {
		return nil, errors.New("detection is missing in property")
	}
	stringDetection, ok := m["detection"].(string)
	if !ok {
		return nil, errors.New("property detection needs to be a string")
	}
	var basePropReader basePropertyReader
	switch stringDetection {
	case "snmpget":
		var pr snmpGetPropertyReader
		err := mapstructure.Decode(i, &pr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode constant propertyReader")
		}
		basePropReader.propertyReader = &pr
	case "constant":
		v, ok := m["value"]
		if !ok {
			return nil, errors.New("value is missing in constant property reader")
		}
		var pr constantPropertyReader
		if _, ok := v.(map[interface{}]interface{}); ok {
			return nil, errors.New("value must not be a map")
		}
		if _, ok := v.([]interface{}); ok {
			return nil, errors.New("value must not be an array")
		}
		pr.Value = value.New(v)
		basePropReader.propertyReader = &pr
	case "SysObjectID":
		var pr sysObjectIDPropertyReader
		err := mapstructure.Decode(i, &pr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode sysObjectIDPropertyReader")
		}
		basePropReader.propertyReader = &pr
	case "SysDescription":
		var pr sysDescriptionPropertyReader
		err := mapstructure.Decode(i, &pr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode sysDescriptionPropertyReader")
		}
		basePropReader.propertyReader = &pr
	case "Vendor":
		if task <= propertyVendor {
			return nil, errors.New("cannot use vendor property, model series is not available here yet")
		}
		var pr vendorPropertyReader
		err := mapstructure.Decode(i, &pr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode vendor PropertyReader")
		}
		basePropReader.propertyReader = &pr
	case "Model":
		if task <= propertyModel {
			return nil, errors.New("cannot use model property, model series is not available here yet")
		}
		var pr modelPropertyReader
		err := mapstructure.Decode(i, &pr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode model PropertyReader")
		}
		basePropReader.propertyReader = &pr
	case "ModelSeries":
		if task <= propertyModelSeries {
			return nil, errors.New("cannot use model series property, model series is not available here yet")
		}
		var pr modelSeriesPropertyReader
		err := mapstructure.Decode(i, &pr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode model series PropertyReader")
		}
		basePropReader.propertyReader = &pr

	default:
		return nil, errors.New("invalid detection type " + stringDetection)
	}
	if operators, ok := m["operators"]; ok {
		operatorSlice, ok := operators.([]interface{})
		if !ok {
			return nil, errors.New("operators has to be an array")
		}
		operators, err := interfaceSlice2propertyOperators(operatorSlice, task)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert interface slice to string operators")
		}
		basePropReader.operators = operators
	}
	if preConditionInterface, ok := m["pre_condition"]; ok {
		preCondition, err := interface2condition(preConditionInterface, task)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert pre condition interface 2 condition")
		}
		basePropReader.preCondition = preCondition
	}
	return &basePropReader, nil
}

func interfaceSlice2propertyOperators(i []interface{}, task relatedTask) (propertyOperators, error) {
	var propertyOperators propertyOperators
	for _, opInterface := range i {
		m, ok := opInterface.(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("failed to convert interface to map[interface{}]interface{}")
		}
		if _, ok := m["type"]; !ok {
			return nil, errors.New("operator type is missing!")
		}
		stringType, ok := m["type"].(string)
		if !ok {
			return nil, errors.New("operator type needs to be a string")
		}

		switch stringType {
		case "filter":
			var adapter filterOperatorAdapter
			var filter baseStringFilter
			filterMethod, ok := m["filter_method"]
			if ok {
				if filterMethodString, ok := filterMethod.(string); ok {
					filter.FilterMethod = matchMode(filterMethodString)
				} else {
					return nil, errors.New("filter method needs to be a string")
				}
				err := filter.FilterMethod.validate()
				if err != nil {
					return nil, errors.Wrap(err, "invalid filter method")
				}
			} else {
				filter.FilterMethod = "contains"
			}
			val, ok := m["value"]
			if !ok {
				return nil, errors.New("value is missing")
			}
			if valueString, ok := val.(string); ok {
				filter.Value = valueString
			}
			if returnOnMismatchInt, ok := m["return_on_mismatch"]; ok {
				if returnOnMismatch, ok := returnOnMismatchInt.(bool); ok {
					filter.returnOnMismatch = returnOnMismatch
				} else {
					return nil, errors.New("return_on_mismatch needs to be a boolean")
				}
			}
			adapter.operator = &filter
			propertyOperators = append(propertyOperators, &adapter)
		case "modify":
			var modifier modifyOperatorAdapter
			modifyMethod, ok := m["modify_method"]
			if !ok {
				return nil, errors.New("modify method is missing in modify operator")
			}
			modifyMethodString, ok := modifyMethod.(string)
			if !ok {
				return nil, errors.New("modify method isn't a string")
			}
			switch modifyMethodString {
			case "regexSubmatch":
				format, ok := m["format"]
				if !ok {
					return nil, errors.New("format is missing")
				}
				formatString, ok := format.(string)
				if !ok {
					return nil, errors.New("format has to be a string")
				}
				regex, ok := m["regex"]
				if !ok {
					return nil, errors.New("regex is missing")
				}
				regexString, ok := regex.(string)
				if !ok {
					return nil, errors.New("regex has to be a string")
				}
				mod, err := newRegexSubmatchModifier(regexString, formatString)
				if err != nil {
					return nil, errors.Wrap(err, "failed to create new regex submatch modifier")
				}
				modifier.operator = mod
			case "regexReplace":
				replace, ok := m["replace"]
				if !ok {
					return nil, errors.New("replace is missing")
				}
				replaceString, ok := replace.(string)
				if !ok {
					return nil, errors.New("replace has to be a string")
				}
				regex, ok := m["regex"]
				if !ok {
					return nil, errors.New("regex is missing")
				}
				regexString, ok := regex.(string)
				if !ok {
					return nil, errors.New("regex has to be a string")
				}
				mod, err := newRegexReplaceModifier(regexString, replaceString)
				if err != nil {
					return nil, errors.Wrap(err, "failed to create new regex replace modifier")
				}
				modifier.operator = mod
			case "toUpperCase":
				var toUpperCaseModifier toUpperCaseModifier
				modifier.operator = &toUpperCaseModifier
			case "toLowerCase":
				var toLowerCaseModifier toLowerCaseModifier
				modifier.operator = &toLowerCaseModifier
			case "overwrite":
				overwriteString, ok := m["value"].(string)
				if !ok {
					return nil, errors.New("value is missing in overwrite operator, or is not of type string")
				}
				var overwriteModifier overwriteModifier
				overwriteModifier.overwriteString = overwriteString
				modifier.operator = &overwriteModifier
			case "addPrefix":
				prefix, ok := m["value"].(string)
				if !ok {
					return nil, errors.New("value is missing in addPrefix operator, or is not of type string")
				}
				var prefixModifier addPrefixModifier
				prefixModifier.prefix = prefix
				modifier.operator = &prefixModifier
			case "addSuffix":
				suffix, ok := m["value"].(string)
				if !ok {
					return nil, errors.New("value is missing in addSuffix operator, or is not of type string")
				}
				var suffixModifier addSuffixModifier
				suffixModifier.suffix = suffix
				modifier.operator = &suffixModifier
			case "insertReadValue":
				format, ok := m["format"].(string)
				if !ok {
					return nil, errors.New("format is missing in insertReadValue operator, or is not of type string")
				}
				valueReaderInterface, ok := m["read_value"]
				if !ok {
					return nil, errors.New("read value is missing in insertReadValue operator")
				}
				valueReader, err := interface2propertyReader(valueReaderInterface, task)
				if err != nil {
					return nil, errors.Wrap(err, "failed to convert read_value to propertyReader in insertReadValue operator")
				}
				var irvModifier insertReadValueModifier
				irvModifier.format = format
				irvModifier.readValueReader = valueReader
				modifier.operator = &irvModifier
			case "map":
				mappingsInterface, ok := m["mappings"]
				if !ok {
					return nil, errors.New("mappings is missing in map string modifier")
				}
				var ignoreOnMismatch bool
				ignoreOnMismatchInterface, ok := m["ignore_on_mismatch"]
				if ok {
					ignoreOnMismatchBool, ok := ignoreOnMismatchInterface.(bool)
					if !ok {
						return nil, errors.New("ignore_on_mismatch in map modifier needs to be boolean")
					}
					ignoreOnMismatch = ignoreOnMismatchBool
				}

				var mapModifier mapModifier
				mapModifier.ignoreOnMismatch = ignoreOnMismatch

				mappings, ok := mappingsInterface.(map[interface{}]interface{})
				if !ok {
					file, ok := mappingsInterface.(string)
					if !ok {
						return nil, errors.New("mappings needs to be a map[string]string or string in map string modifier")
					}
					mappingsFile, err := mapping.GetMapping(file)
					if err != nil {
						return nil, errors.Wrap(err, "can't get specified mapping")
					}
					mapModifier.mappings = mappingsFile
				} else {
					mapModifier.mappings = make(map[string]string)
					for k, val := range mappings {
						key := fmt.Sprint(k)
						valString := fmt.Sprint(val)

						mapModifier.mappings[key] = valString
					}
				}
				if len(mapModifier.mappings) == 0 {
					return nil, errors.New("mappings is empty")
				}
				modifier.operator = &mapModifier
			case "multiply":
				valueReaderInterface := m["value"]
				if !ok {
					return nil, errors.New("value is missing in multiply")
				}
				valueReader, err := interface2propertyReader(valueReaderInterface, task)
				if err != nil {
					return nil, errors.New("value is missing in multiply modify operator, or is not of type float64")
				}
				var multiplyModifier multiplyNumberModifier
				multiplyModifier.value = valueReader
				modifier.operator = &multiplyModifier
			case "divide":
				valueReaderInterface := m["value"]
				if !ok {
					return nil, errors.New("value is missing in divide")
				}
				valueReader, err := interface2propertyReader(valueReaderInterface, task)
				if err != nil {
					return nil, errors.New("value is missing in divide modify operator, or is not of type float64")
				}
				var divideModifier divideNumberModifier
				divideModifier.value = valueReader
				modifier.operator = &divideModifier
			default:
				return nil, fmt.Errorf("invalid modify method '%s'", modifyMethod)
			}
			propertyOperators = append(propertyOperators, &modifier)
		case "switch":
			var sw switchOperatorAdapter
			var switcher genericStringSwitch
			var switchValue string

			// get switch mode, default = equals
			switchMode, ok := m["switch_mode"]
			if ok {
				if switchModeString, ok := switchMode.(string); ok {
					switcher.switchMode = matchMode(switchModeString)
				} else {
					return nil, errors.New("filter method needs to be a string")
				}
				err := switcher.switchMode.validate()
				if err != nil {
					return nil, errors.Wrap(err, "invalid filter method")
				}
			} else {
				switcher.switchMode = "equals"
			}

			// get switch value, default = "default"
			switchValueInterface, ok := m["switch_value"]
			if ok {
				if switchValue, ok = switchValueInterface.(string); !ok {
					return nil, errors.New("switch value needs to be a string")
				}
			} else {
				switchValue = "default"
			}

			// get switchValueGetter
			switch switchValue {
			case "default":
				switcher.switchValueGetter = &defaultStringSwitchValueGetter{}
			case "snmpwalkCount":
				switchValueGetter := snmpwalkCountStringSwitchValueGetter{}
				oid, ok := m["oid"].(string)
				if !ok {
					return nil, errors.New("oid in snmpwalkCount switch operator is missing, or is not a string")
				}
				switchValueGetter.oid = oid
				if filter, ok := m["snmp_result_filter"]; ok {
					var bStrFilter baseStringFilter
					err := mapstructure.Decode(filter, &bStrFilter)
					if err != nil {
						return nil, errors.Wrap(err, "failed to decode snmp_result_filter")
					}
					err = bStrFilter.FilterMethod.validate()
					if err != nil {
						return nil, errors.Wrap(err, "invalid filter method")
					}
					switchValueGetter.filter = &bStrFilter

					if useOidForFilter, ok := m["use_oid_for_filter"].(bool); ok {
						switchValueGetter.useOidForFilter = useOidForFilter
					}
				}
				switcher.switchValueGetter = &switchValueGetter
			}

			// following operators
			cases, ok := m["cases"].([]interface{})
			if !ok {
				return nil, errors.New("cases are missing in switch operator, or it is not an array")
			}

			for _, cInterface := range cases {
				c, ok := cInterface.(map[interface{}]interface{})
				if !ok {
					return nil, errors.New("switch case needs to be a map")
				}
				caseString, ok := c["case"].(string)
				if !ok {
					caseInt, ok := c["case"].(int)
					if !ok {
						return nil, errors.New("case string is missing in switch operator case, or is not a string or int")
					}
					caseString = strconv.Itoa(caseInt)
				}
				subOperatorsInterface, ok := c["operators"].([]interface{})
				if !ok {
					return nil, fmt.Errorf("operators are missing in switch operator case, or it is not an array, in switch case '%s'", caseString)
				}
				subOperators, err := interfaceSlice2propertyOperators(subOperatorsInterface, task)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to convert []interface{} to propertyOperators in switch case '%s'", caseString)
				}
				switchCase := stringSwitchCase{
					caseString: caseString,
					operators:  subOperators,
				}
				switcher.cases = append(switcher.cases, switchCase)
			}

			sw.operator = &switcher
			propertyOperators = append(propertyOperators, &sw)
		default:
			return nil, fmt.Errorf("invalid operator type '%s'", stringType)
		}
	}
	return propertyOperators, nil
}

func interface2GroupPropertyReader(i interface{}, parentGroupPropertyReader groupPropertyReader) (groupPropertyReader, error) {
	m, ok := i.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("failed to convert group properties to map[interface{}]interface{}")
	}
	if _, ok := m["detection"]; !ok {
		return nil, errors.New("detection is missing in group properties")
	}
	stringDetection, ok := m["detection"].(string)
	if !ok {
		return nil, errors.New("property detection needs to be a string")
	}
	switch stringDetection {
	case "snmpwalk":
		if _, ok := m["values"]; !ok {
			return nil, errors.New("values are missing")
		}
		reader, err := interface2oidReader(m["values"])
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse oid reader")
		}

		devClassOIDs, ok := reader.(*deviceClassOIDs)
		if !ok {
			return nil, errors.New("oid reader is no list of oids")
		}

		inheritValuesFromParent := true
		if b, ok := m["inherit_values"]; ok {
			bb, ok := b.(bool)
			if !ok {
				return nil, errors.New("inherit_values needs to be a boolean")
			}
			inheritValuesFromParent = bb
		}

		//overwrite parent
		if inheritValuesFromParent && parentGroupPropertyReader != nil {
			parentSNMPGroupPropertyReader, ok := parentGroupPropertyReader.(*snmpGroupPropertyReader)
			if !ok {
				return nil, errors.New("can't merge SNMP group property reader with property reader of different type")
			}

			devClassOIDsMerged := parentSNMPGroupPropertyReader.oids.merge(*devClassOIDs)
			devClassOIDs = &devClassOIDsMerged
		}

		return &snmpGroupPropertyReader{*devClassOIDs}, nil
	default:
		return nil, fmt.Errorf("unknown detection type '%s'", stringDetection)
	}
}

func interface2oidReader(i interface{}) (oidReader, error) {
	values, ok := i.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("values needs to be a map")
	}

	result := make(deviceClassOIDs)

	for val, data := range values {
		dataMap, ok := data.(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("value data needs to be a map")
		}

		valString, ok := val.(string)
		if !ok {
			return nil, errors.New("key of snmp property reader must be a string")
		}

		if v, ok := dataMap["values"]; ok {
			if len(dataMap) != 1 {
				return nil, errors.New("value with subvalues has to many keys")
			}
			reader, err := interface2oidReader(v)
			if err != nil {
				return nil, err
			}
			result[valString] = reader
			continue
		}

		if ignore, ok := dataMap["ignore"]; ok {
			if b, ok := ignore.(bool); ok && b {
				result[valString] = &emptyOIDReader{}
				continue
			}
		}

		var oid yamlComponentsOID
		err := mapstructure.Decode(data, &oid)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode values map to yamlComponentsOIDs")
		}
		err = oid.validate()
		if err != nil {
			return nil, errors.Wrapf(err, "oid reader for %s is invalid", valString)
		}
		devClassOID, err := oid.convert()
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert yaml OID to device class OID")
		}
		result[valString] = &devClassOID
	}
	return &result, nil
}

func (m *matchMode) validate() error {
	if *m != "contains" && *m != "!contains" && *m != "startsWith" && *m != "!startsWith" && *m != "regex" && *m != "!regex" && *m != "equals" && *m != "!equals" {
		return errors.New(string("unknown matchmode \"" + *m + "\""))
	}
	return nil
}

func (l *logicalOperator) validate() error {
	if *l != "AND" && *l != "OR" {
		return errors.New(string("unknown logical operator \"" + *l + "\""))
	}
	return nil
}
