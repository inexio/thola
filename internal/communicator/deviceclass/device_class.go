// Package deviceclass contains the logic for interacting with device classes.
// It contains methods that read out the .yaml files representing device classes.
package deviceclass

import (
	"context"
	"github.com/inexio/thola/config"
	"github.com/inexio/thola/config/codecommunicator"
	"github.com/inexio/thola/internal/communicator/communicator"
	"github.com/inexio/thola/internal/communicator/component"
	"github.com/inexio/thola/internal/communicator/deviceclass/condition"
	"github.com/inexio/thola/internal/communicator/deviceclass/groupproperty"
	"github.com/inexio/thola/internal/communicator/deviceclass/property"
	"github.com/inexio/thola/internal/communicator/hierarchy"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/inexio/thola/internal/utility"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// deviceClass represents a device class.
type deviceClass struct {
	name           string
	match          condition.Condition
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
	vendor       property.Reader
	model        property.Reader
	modelSeries  property.Reader
	serialNumber property.Reader
	osVersion    property.Reader
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
	alarmLowVoltageDisconnect property.Reader
	batteryAmperage           property.Reader
	batteryCapacity           property.Reader
	batteryCurrent            property.Reader
	batteryRemainingTime      property.Reader
	batteryTemperature        property.Reader
	batteryVoltage            property.Reader
	currentLoad               property.Reader
	mainsVoltageApplied       property.Reader
	rectifierCurrent          property.Reader
	systemVoltage             property.Reader
}

// deviceClassComponentsCPU represents the cpu components part of a device class.
type deviceClassComponentsCPU struct {
	properties groupproperty.Reader
}

// deviceClassComponentsMemory represents the memory components part of a device class.
type deviceClassComponentsMemory struct {
	usage property.Reader
}

// deviceClassComponentsSBC represents the sbc components part of a device class.
type deviceClassComponentsSBC struct {
	agents                   groupproperty.Reader
	realms                   groupproperty.Reader
	globalCallPerSecond      property.Reader
	globalConcurrentSessions property.Reader
	activeLocalContacts      property.Reader
	transcodingCapacity      property.Reader
	licenseCapacity          property.Reader
	systemRedundancy         property.Reader
	systemHealthScore        property.Reader
}

// deviceClassComponentsServer represents the server components part of a device class.
type deviceClassComponentsServer struct {
	procs property.Reader
	users property.Reader
}

// deviceClassComponentsDisk represents the disk component part of a device class.
type deviceClassComponentsDisk struct {
	properties groupproperty.Reader
}

// deviceClassComponentsHardwareHealth represents the sbc components part of a device class.
type deviceClassComponentsHardwareHealth struct {
	environmentMonitorState property.Reader
	fans                    groupproperty.Reader
	powerSupply             groupproperty.Reader
}

// deviceClassConfig represents the config part of a device class.
type deviceClassConfig struct {
	snmp       deviceClassSNMP
	components map[component.Component]bool
}

// deviceClassComponentsInterfaces represents the interface properties part of a device class.
type deviceClassComponentsInterfaces struct {
	count      string
	properties groupproperty.Reader
}

// deviceClassSNMP represents the snmp config part of a device class.
type deviceClassSNMP struct {
	MaxRepetitions uint32 `yaml:"max_repetitions"`
	MaxOids        int    `yaml:"max_oids"`
}

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
	Properties interface{} `yaml:"properties"`
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
	Properties interface{} `yaml:"properties"`
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

// GetHierarchy returns the hierarchy of device classes merged with their corresponding code communicator.
func GetHierarchy() (hierarchy.Hierarchy, error) {
	genericDeviceClassDir := "deviceclass"
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
	return d.match.Check(ctx)
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
		devClass.match = condition.GetAlwaysTrueCondition()
		devClass.identify = deviceClassIdentify{
			properties: deviceClassIdentifyProperties{},
		}
	} else {
		cond, err := condition.Interface2Condition(y.Match, condition.ClassifyDevice)
		if err != nil {
			return deviceClass{}, errors.Wrap(err, "failed to convert device class condition")
		}
		devClass.match = cond
		devClass.tryToMatchLast = cond.ContainsUniqueRequest()
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
	prop, err := y.Properties.convert(parentIdentify.properties)
	if err != nil {
		return deviceClassIdentify{}, errors.Wrap(err, "failed to read yaml identify properties")
	}
	identify.properties = prop

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
		interfaceComponent.properties, err = groupproperty.Interface2Reader(y.Properties, interfaceComponent.properties)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert interface properties")
		}
	}

	if y.Count != "" {
		interfaceComponent.count = y.Count
	}

	return &interfaceComponent, nil
}

func (y *yamlDeviceClassIdentify) validate() error {
	if y.Properties == nil {
		y.Properties = &yamlDeviceClassIdentifyProperties{}
	}
	return nil
}

func (y *yamlDeviceClassIdentifyProperties) convert(parentProperties deviceClassIdentifyProperties) (deviceClassIdentifyProperties, error) {
	prop := parentProperties
	var err error

	if y.Vendor != nil {
		prop.vendor, err = property.InterfaceSlice2Reader(y.Vendor, condition.PropertyVendor, prop.vendor)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert vendor property to property reader")
		}
	}
	if y.Model != nil {
		prop.model, err = property.InterfaceSlice2Reader(y.Model, condition.PropertyModel, prop.model)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert model property to property reader")
		}
	}
	if y.ModelSeries != nil {
		prop.modelSeries, err = property.InterfaceSlice2Reader(y.ModelSeries, condition.PropertyModelSeries, prop.modelSeries)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert model series property to property reader")
		}
	}
	if y.SerialNumber != nil {
		prop.serialNumber, err = property.InterfaceSlice2Reader(y.SerialNumber, condition.PropertyDefault, prop.serialNumber)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert serial number property to property reader")
		}
	}
	if y.OSVersion != nil {
		prop.osVersion, err = property.InterfaceSlice2Reader(y.OSVersion, condition.PropertyDefault, prop.osVersion)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert osVersion property to property reader")
		}
	}
	return prop, nil
}

func (y *yamlDeviceClassConfig) convert(parentConfig deviceClassConfig) (deviceClassConfig, error) {
	err := y.validate()
	if err != nil {
		return deviceClassConfig{}, errors.Wrap(err, "config is invalid")
	}
	var cfg deviceClassConfig

	if y.SNMP.MaxRepetitions != 0 {
		cfg.snmp.MaxRepetitions = y.SNMP.MaxRepetitions
	} else {
		cfg.snmp.MaxRepetitions = parentConfig.snmp.MaxRepetitions
	}
	cfg.snmp.MaxOids = utility.IfThenElseInt(y.SNMP.MaxOids != 0, y.SNMP.MaxOids, parentConfig.snmp.MaxOids)

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

func (y *yamlDeviceClassConfig) validate() error {
	if y.SNMP.MaxOids < 0 {
		return errors.New("invalid snmp max oids")
	}
	return nil
}

func (y *yamlComponentsUPSProperties) convert(parentComponent *deviceClassComponentsUPS) (deviceClassComponentsUPS, error) {
	var prop deviceClassComponentsUPS
	var err error
	if parentComponent != nil {
		prop = *parentComponent
	}

	if y.AlarmLowVoltageDisconnect != nil {
		prop.alarmLowVoltageDisconnect, err = property.InterfaceSlice2Reader(y.AlarmLowVoltageDisconnect, condition.PropertyDefault, prop.alarmLowVoltageDisconnect)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert alarm low voltage disconnect property to property reader")
		}
	}
	if y.BatteryAmperage != nil {
		prop.batteryAmperage, err = property.InterfaceSlice2Reader(y.BatteryAmperage, condition.PropertyDefault, prop.batteryAmperage)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery amperage property to property reader")
		}
	}
	if y.BatteryCapacity != nil {
		prop.batteryCapacity, err = property.InterfaceSlice2Reader(y.BatteryCapacity, condition.PropertyDefault, prop.batteryCapacity)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery capacity property to property reader")
		}
	}
	if y.BatteryCurrent != nil {
		prop.batteryCurrent, err = property.InterfaceSlice2Reader(y.BatteryCurrent, condition.PropertyDefault, prop.batteryCurrent)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery current property to property reader")
		}
	}
	if y.BatteryRemainingTime != nil {
		prop.batteryRemainingTime, err = property.InterfaceSlice2Reader(y.BatteryRemainingTime, condition.PropertyDefault, prop.batteryRemainingTime)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery remaining time property to property reader")
		}
	}
	if y.BatteryTemperature != nil {
		prop.batteryTemperature, err = property.InterfaceSlice2Reader(y.BatteryTemperature, condition.PropertyDefault, prop.batteryTemperature)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery temperature property to property reader")
		}
	}
	if y.BatteryVoltage != nil {
		prop.batteryVoltage, err = property.InterfaceSlice2Reader(y.BatteryVoltage, condition.PropertyDefault, prop.batteryVoltage)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery voltage property to property reader")
		}
	}
	if y.CurrentLoad != nil {
		prop.currentLoad, err = property.InterfaceSlice2Reader(y.CurrentLoad, condition.PropertyDefault, prop.currentLoad)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert current load property to property reader")
		}
	}
	if y.MainsVoltageApplied != nil {
		prop.mainsVoltageApplied, err = property.InterfaceSlice2Reader(y.MainsVoltageApplied, condition.PropertyDefault, prop.mainsVoltageApplied)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert mains voltage applied property to property reader")
		}
	}
	if y.RectifierCurrent != nil {
		prop.rectifierCurrent, err = property.InterfaceSlice2Reader(y.RectifierCurrent, condition.PropertyDefault, prop.rectifierCurrent)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert rectifier current property to property reader")
		}
	}
	if y.SystemVoltage != nil {
		prop.systemVoltage, err = property.InterfaceSlice2Reader(y.SystemVoltage, condition.PropertyDefault, prop.systemVoltage)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert system voltage property to property reader")
		}
	}
	return prop, nil
}

func (y *yamlComponentsCPUProperties) convert(parentComponent *deviceClassComponentsCPU) (deviceClassComponentsCPU, error) {
	var prop deviceClassComponentsCPU
	var err error

	if parentComponent != nil {
		prop = *parentComponent
	}

	if y.Properties != nil {
		prop.properties, err = groupproperty.Interface2Reader(y.Properties, prop.properties)
		if err != nil {
			return deviceClassComponentsCPU{}, errors.Wrap(err, "failed to convert load property to property reader")
		}
	}
	return prop, nil
}

func (y *yamlComponentsMemoryProperties) convert(parentComponent *deviceClassComponentsMemory) (deviceClassComponentsMemory, error) {
	var prop deviceClassComponentsMemory
	var err error

	if parentComponent != nil {
		prop = *parentComponent
	}

	if y.Usage != nil {
		prop.usage, err = property.InterfaceSlice2Reader(y.Usage, condition.PropertyDefault, prop.usage)
		if err != nil {
			return deviceClassComponentsMemory{}, errors.Wrap(err, "failed to convert memory usage property to property reader")
		}
	}
	return prop, nil
}

func (y *yamlComponentsServerProperties) convert(parentComponent *deviceClassComponentsServer) (deviceClassComponentsServer, error) {
	var prop deviceClassComponentsServer
	var err error

	if parentComponent != nil {
		prop = *parentComponent
	}

	if y.Procs != nil {
		prop.procs, err = property.InterfaceSlice2Reader(y.Procs, condition.PropertyDefault, prop.procs)
		if err != nil {
			return deviceClassComponentsServer{}, errors.Wrap(err, "failed to convert procs property to property reader")
		}
	}
	if y.Users != nil {
		prop.users, err = property.InterfaceSlice2Reader(y.Users, condition.PropertyDefault, prop.procs)
		if err != nil {
			return deviceClassComponentsServer{}, errors.Wrap(err, "failed to convert users property to property reader")
		}
	}
	return prop, nil
}

func (y *yamlComponentsDiskProperties) convert(parentDisk *deviceClassComponentsDisk) (deviceClassComponentsDisk, error) {
	var prop deviceClassComponentsDisk
	var err error

	if parentDisk != nil {
		prop = *parentDisk
	}

	if y.Properties != nil {
		prop.properties, err = groupproperty.Interface2Reader(y.Properties, prop.properties)
		if err != nil {
			return deviceClassComponentsDisk{}, errors.Wrap(err, "failed to convert storages property to group property reader")
		}
	}
	return prop, nil
}

func (y *yamlComponentsSBCProperties) convert(parentComponentsSBC *deviceClassComponentsSBC) (deviceClassComponentsSBC, error) {
	var prop deviceClassComponentsSBC
	var err error

	if parentComponentsSBC != nil {
		prop = *parentComponentsSBC
	}

	if y.Agents != nil {
		prop.agents, err = groupproperty.Interface2Reader(y.Agents, prop.agents)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert agents property to group property reader")
		}
	}
	if y.Realms != nil {
		prop.realms, err = groupproperty.Interface2Reader(y.Realms, prop.realms)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert realms property to group property reader")
		}
	}
	if y.ActiveLocalContacts != nil {
		prop.activeLocalContacts, err = property.InterfaceSlice2Reader(y.ActiveLocalContacts, condition.PropertyDefault, prop.activeLocalContacts)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert active local contacts property to property reader")
		}
	}
	if y.GlobalCallPerSecond != nil {
		prop.globalCallPerSecond, err = property.InterfaceSlice2Reader(y.GlobalCallPerSecond, condition.PropertyDefault, prop.globalCallPerSecond)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert global call per second property to property reader")
		}
	}
	if y.GlobalConcurrentSessions != nil {
		prop.globalConcurrentSessions, err = property.InterfaceSlice2Reader(y.GlobalConcurrentSessions, condition.PropertyDefault, prop.globalConcurrentSessions)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert global concurrent sessions property to property reader")
		}
	}
	if y.LicenseCapacity != nil {
		prop.licenseCapacity, err = property.InterfaceSlice2Reader(y.LicenseCapacity, condition.PropertyDefault, prop.licenseCapacity)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert license capacity property to property reader")
		}
	}
	if y.TranscodingCapacity != nil {
		prop.transcodingCapacity, err = property.InterfaceSlice2Reader(y.TranscodingCapacity, condition.PropertyDefault, prop.transcodingCapacity)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert transcoding capacity property to property reader")
		}
	}
	if y.SystemRedundancy != nil {
		prop.systemRedundancy, err = property.InterfaceSlice2Reader(y.SystemRedundancy, condition.PropertyDefault, prop.systemRedundancy)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert system redundancy property to property reader")
		}
	}

	if y.SystemHealthScore != nil {
		prop.systemHealthScore, err = property.InterfaceSlice2Reader(y.SystemHealthScore, condition.PropertyDefault, prop.systemHealthScore)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert system health score property to property reader")
		}
	}
	return prop, nil
}

func (y *yamlComponentsHardwareHealthProperties) convert(parentHardwareHealth *deviceClassComponentsHardwareHealth) (deviceClassComponentsHardwareHealth, error) {
	var prop deviceClassComponentsHardwareHealth
	var err error

	if parentHardwareHealth != nil {
		prop = *parentHardwareHealth
	}

	if y.Fans != nil {
		prop.fans, err = groupproperty.Interface2Reader(y.Fans, prop.fans)
		if err != nil {
			return deviceClassComponentsHardwareHealth{}, errors.Wrap(err, "failed to convert fans property to group property reader")
		}
	}
	if y.PowerSupply != nil {
		prop.powerSupply, err = groupproperty.Interface2Reader(y.PowerSupply, prop.powerSupply)
		if err != nil {
			return deviceClassComponentsHardwareHealth{}, errors.Wrap(err, "failed to convert power supply property to group property reader")
		}
	}
	if y.EnvironmentMonitorState != nil {
		prop.environmentMonitorState, err = property.InterfaceSlice2Reader(y.EnvironmentMonitorState, condition.PropertyDefault, prop.environmentMonitorState)
		if err != nil {
			return deviceClassComponentsHardwareHealth{}, errors.Wrap(err, "failed to convert environment monitor state property to property reader")
		}
	}

	return prop, nil
}
