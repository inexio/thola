package communicator

import (
	"context"
	"fmt"
	"github.com/inexio/thola/core/mapping"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/inexio/thola/core/utility"
	"github.com/inexio/thola/core/value"
	"github.com/inexio/thola/core/vfs"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// deviceClassComponent represents a component with an byte.
type deviceClassComponent byte

// All component enums.
const (
	interfacesComponent deviceClassComponent = iota + 1
	upsComponent
	cpuComponent
	memoryComponent
	sbcComponent
	hardwareHealthComponent
)

// deviceClass represents a device class.
type deviceClass struct {
	name              string
	match             condition
	config            deviceClassConfig
	identify          deviceClassIdentify
	components        deviceClassComponents
	yamlFile          string
	parentDeviceClass *deviceClass
	subDeviceClasses  map[string]*deviceClass
	tryToMatchLast    bool
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

// deviceClassComponentsCPU represents the memory components part of a device class.
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
	components map[deviceClassComponent]bool
}

// deviceClassComponentsInterfaces represents the interface properties part of a device class.
type deviceClassComponentsInterfaces struct {
	Count   string
	IfTable groupPropertyReader
	Types   deviceClassInterfaceTypes
}

// deviceClassOIDs maps labels to OIDs.
type deviceClassOIDs map[string]deviceClassOID

type deviceClassOID struct {
	network.SNMPGetConfiguration
	operators propertyOperators
}

// deviceClassInterfaceTypes maps interface types to TypeDefs.
type deviceClassInterfaceTypes map[string]deviceClassInterfaceTypeDef

// deviceClassInterfaceTypeDef represents a interface type (e.g. "radio" interface).
type deviceClassInterfaceTypeDef struct {
	Detection string
	Values    deviceClassOIDs
}

// deviceClassSNMP represents the snmp config part of a device class.
type deviceClassSNMP struct {
	MaxRepetitions uint8 `yaml:"max_repetitions"`
}

// logicalOperator represents a logical operator (OR or AND)
type logicalOperator string

// matchMode represents a match mode that is used to match a condition.
type matchMode string

type yamlDeviceClass struct {
	Name       string                    `yaml:"name"`
	Match      interface{}               `yaml:"match"`
	Identify   yamlDeviceClassIdentify   `yaml:"identify"`
	Config     yamlDeviceClassConfig     `yaml:"config"`
	Components yamlDeviceClassComponents `yaml:"components"`
}

type yamlDeviceClassIdentify struct {
	Properties *yamlDeviceClassIdentifyProperties `yaml:"properties"`
}

type yamlDeviceClassComponents struct {
	Interfaces     *yamlComponentsInterfaces               `yaml:"interfaces"`
	UPS            *yamlComponentsUPSProperties            `yaml:"ups"`
	CPU            *yamlComponentsCPUProperties            `yaml:"cpu"`
	Memory         *yamlComponentsMemoryProperties         `yaml:"memory"`
	SBC            *yamlComponentsSBCProperties            `yaml:"sbc"`
	HardwareHealth *yamlComponentsHardwareHealthProperties `yaml:"hardware_health"`
}

type yamlDeviceClassConfig struct {
	SNMP       deviceClassSNMP `yaml:"snmp"`
	Components map[string]bool `yaml:"components"`
}

type yamlConditionSet struct {
	LogicalOperator logicalOperator `yaml:"logical_operator" mapstructure:"logical_operator"`
	Conditions      []interface{}
}

type yamlDeviceClassIdentifyProperties struct {
	Vendor       []interface{} `yaml:"vendor"`
	Model        []interface{} `yaml:"model"`
	ModelSeries  []interface{} `yaml:"model_series"`
	SerialNumber []interface{} `yaml:"serial_number"`
	OSVersion    []interface{} `yaml:"os_version"`
}

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

type yamlComponentsCPUProperties struct {
	Load        []interface{} `yaml:"load"`
	Temperature []interface{} `yaml:"temperature"`
}

type yamlComponentsMemoryProperties struct {
	Usage []interface{} `yaml:"usage"`
}

type yamlComponentsSBCProperties struct {
	Agents                   interface{}   `yaml:"agents"`
	Realms                   interface{}   `yaml:"realms"`
	GlobalCallPerSecond      []interface{} `yaml:"global_call_per_second"`
	GlobalConcurrentSessions []interface{} `yaml:"global_concurrent_sessions"`
	ActiveLocalContacts      []interface{} `yaml:"active_local_contacts"`
	TranscodingCapacity      []interface{} `yaml:"transcoding_capacity"`
	LicenseCapacity          []interface{} `yaml:"license_capacity"`
	SystemRedundancy         []interface{} `yaml:"system_redundancy"`
}

type yamlComponentsHardwareHealthProperties struct {
	EnvironmentMonitorState []interface{} `yaml:"environment_monitor_state"`
	Fans                    interface{}   `yaml:"fans"`
	PowerSupply             interface{}   `yaml:"power_supply"`
}

type yamlComponentsInterfaces struct {
	Count   string                       `yaml:"count"`
	IfTable interface{}                  `yaml:"ifTable"`
	Types   yamlComponentsInterfaceTypes `yaml:"types"`
}

type yamlComponentsInterfaceTypes map[string]yamlComponentsInterfaceTypeDef

type yamlComponentsInterfaceTypeDef struct {
	Detection string             `yaml:"detection"`
	Values    yamlComponentsOIDs `yaml:"specific_values"`
}

type yamlComponentsOIDs map[string]yamlComponentsOID

type yamlComponentsOID struct {
	network.SNMPGetConfiguration `yaml:",inline" mapstructure:",squash"`
	Operators                    []interface{} `yaml:"operators"`
}

var genericDeviceClass struct {
	sync.Once
	*deviceClass
}

// identifyDeviceClass identify the device class based on the data in the context.
func identifyDeviceClass(ctx context.Context) (*deviceClass, error) {
	deviceClasses, err := getDeviceClasses()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to load device classes")
		return nil, errors.Wrap(err, "error during getDeviceClasses")
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if ok && con.SNMP != nil {
		con.SNMP.SnmpClient.SetMaxRepetitions(1)
	}

	deviceClass, err := identifyDeviceClassRecursive(ctx, deviceClasses, true)
	if err != nil {
		if tholaerr.IsNotFoundError(err) {
			return genericDeviceClass.deviceClass, nil
		}
		return nil, errors.Wrap(err, "error occurred while identifying device class")
	}
	return deviceClass, err
}

func identifyDeviceClassRecursive(ctx context.Context, devClass map[string]*deviceClass, considerPriority bool) (*deviceClass, error) {
	var tryToMatchLastDeviceClasses map[string]*deviceClass

	for n, devClass := range devClass {
		if considerPriority && devClass.tryToMatchLast {
			if tryToMatchLastDeviceClasses == nil {
				tryToMatchLastDeviceClasses = make(map[string]*deviceClass)
			}
			tryToMatchLastDeviceClasses[n] = devClass
			continue
		}

		logger := log.Ctx(ctx).With().Str("device_class", devClass.getName()).Logger()
		ctx = logger.WithContext(ctx)
		log.Ctx(ctx).Trace().Msgf("starting device class match (%s)", devClass.getName())
		match, err := devClass.matchDevice(ctx)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("error during device class match")
			return nil, errors.Wrap(err, "error while trying to match device class: "+devClass.getName())
		}

		if match {
			log.Ctx(ctx).Trace().Msg("device class matched")
			if devClass.subDeviceClasses != nil {
				subDeviceClass, err := identifyDeviceClassRecursive(ctx, devClass.subDeviceClasses, true)
				if err != nil {
					if tholaerr.IsNotFoundError(err) {
						return devClass, nil
					}
					return nil, errors.Wrapf(err, "error occurred while trying to identify sub device class for device class '%s'", devClass.getName())
				}
				return subDeviceClass, nil
			}
			return devClass, nil
		}
		log.Ctx(ctx).Trace().Msg("device class did not match")
	}
	if tryToMatchLastDeviceClasses != nil {
		deviceClass, err := identifyDeviceClassRecursive(ctx, tryToMatchLastDeviceClasses, false)
		if err != nil {
			if !tholaerr.IsNotFoundError(err) {
				return nil, err
			}
		} else {
			return deviceClass, nil
		}
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok {
		return nil, errors.New("no connection data found in context")
	}

	// return generic device class
	if (con.SNMP == nil || len(con.SNMP.SnmpClient.GetSuccessfulCachedRequests()) == 0) && (con.HTTP == nil || len(con.HTTP.HTTPClient.GetSuccessfulCachedRequests()) == 0) {
		return nil, errors.New("no network requests to device succeeded")
	}
	return nil, tholaerr.NewNotFoundError("no device class matched")
}

// getDeviceClasses returns a list of all device classes. device classes is a singleton that is created
// when the function is called for the first time.
func getDeviceClasses() (map[string]*deviceClass, error) {
	var err error
	genericDeviceClass.Do(func() {
		err = readDeviceClasses()
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to read device classes")
	}
	return genericDeviceClass.subDeviceClasses, nil
}

// getDeviceClass returns a single device class.
func getDeviceClass(identifier string) (*deviceClass, error) {
	devClasses, err := getDeviceClasses()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get device classes")
	}
	configIdentifiers := strings.Split(identifier, "/")
	var deviceClass *deviceClass
	var ok bool
	if configIdentifiers[0] == "generic" {
		deviceClass = genericDeviceClass.deviceClass
	} else {
		deviceClass, ok = devClasses[configIdentifiers[0]]
		if !ok {
			return nil, errors.New("device class does not exist")
		}
	}
	for i, identifier := range configIdentifiers {
		if i == 0 {
			continue
		}
		deviceClass, ok = deviceClass.subDeviceClasses[identifier]
		if !ok {
			return nil, errors.New("device class does not exist")
		}
	}
	return deviceClass, nil
}

func readDeviceClasses() error {
	//read in generic device class
	genericDeviceClassDir := "device-classes"
	genericDeviceClassFile, err := vfs.FileSystem.Open(filepath.Join(genericDeviceClassDir, "generic.yaml"))
	if err != nil {
		return errors.Wrap(err, "failed to open generic device class file")
	}
	genericDeviceClass.deviceClass, err = yamlFileToDeviceClass(genericDeviceClassFile, genericDeviceClassDir, nil)
	if err != nil {
		return errors.Wrap(err, "failed to read in generic device class")
	}

	return nil
}

func yamlFileToDeviceClass(file http.File, directory string, parentDeviceClass *deviceClass) (*deviceClass, error) {
	//get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get stat for file")
	}

	if !strings.HasSuffix(fileInfo.Name(), ".yaml") {
		return nil, errors.New("only yaml files are allowed for this function")
	}

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}
	var deviceClassYaml yamlDeviceClass
	err = yaml.Unmarshal(contents, &deviceClassYaml)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config file")
	}

	devClass, err := deviceClassYaml.convert()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert yamlData to deviceClass for device class '%s'", deviceClassYaml.Name)
	}
	// set parent device class
	devClass.parentDeviceClass = parentDeviceClass

	// check for sub device classes
	subDirPath := filepath.Join(directory, devClass.name)
	subDir, err := vfs.FileSystem.Open(subDirPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "an unexpected error occurred while trying to open sub device class directory")
		}
		return &devClass, nil
	}
	subDeviceClasses, err := readDeviceClassDirectory(subDir, subDirPath, &devClass)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read sub device classes")
	}
	devClass.subDeviceClasses = subDeviceClasses

	return &devClass, nil
}

func readDeviceClassDirectory(dir http.File, directory string, parentDeviceClass *deviceClass) (map[string]*deviceClass, error) {
	//get fileinfo
	dirFileInfo, err := dir.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get stat for dir")
	}
	if !dirFileInfo.IsDir() {
		return nil, errors.New("given file is not a dir")
	}
	files, err := dir.Readdir(0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read dir")
	}
	deviceClasses := make(map[string]*deviceClass)
	for _, fileInfo := range files {
		fullPathToFile := filepath.Join(directory, fileInfo.Name())
		if fileInfo.IsDir() {
			// directories will be ignored here, sub device classes dirs will be called when
			// their parent device class is processed
			continue
		}
		if !strings.HasSuffix(fileInfo.Name(), ".yaml") {
			// all non directory files need to be yaml file and end with ".yaml"
			return nil, errors.New("only yaml config files are allowed in device class directories")
		}
		file, err := vfs.FileSystem.Open(fullPathToFile)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open file "+fullPathToFile)
		}
		deviceClass, err := yamlFileToDeviceClass(file, directory, parentDeviceClass)
		if err != nil {
			return nil, errors.Wrapf(err, "an error occurred while trying to read in yaml config file %s", fileInfo.Name())
		}
		deviceClasses[deviceClass.name] = deviceClass
	}
	return deviceClasses, nil
}

// getName returns the name of the device class.
func (o *deviceClass) getName() string {
	if o.parentDeviceClass != nil {
		pName := o.parentDeviceClass.getName()
		return utility.IfThenElseString(pName == "generic", o.name, fmt.Sprintf("%s/%s", pName, o.name))
	}
	return o.name
}

// getParentDeviceClass returns the parent device class.
func (o *deviceClass) getParentDeviceClass() (*deviceClass, error) {
	if o.parentDeviceClass == nil {
		return nil, tholaerr.NewNotFoundError("no parent device class available")
	}
	return o.parentDeviceClass, nil
}

// match checks if data in context matches the device class.
func (o *deviceClass) matchDevice(ctx context.Context) (bool, error) {
	return o.match.check(ctx)
}

// getSNMPMaxRepetitions returns the maximum snmp repetitions.
func (o *deviceClass) getSNMPMaxRepetitions() (uint8, error) {
	if o.config.snmp.MaxRepetitions != 0 {
		return o.config.snmp.MaxRepetitions, nil
	}
	if o.parentDeviceClass != nil {
		return o.parentDeviceClass.getSNMPMaxRepetitions()
	}
	return 0, tholaerr.NewNotFoundError("max_repetitions not found")
}

// getAvailableComponents returns the available components.
func (o *deviceClass) getAvailableComponents() map[deviceClassComponent]bool {
	haha := make(map[deviceClassComponent]bool)
	if o.parentDeviceClass != nil {
		haha = o.parentDeviceClass.getAvailableComponents()
	}
	for k, v := range o.config.components {
		haha[k] = v
	}
	return haha
}

// hasAvailableComponent checks whether the specified component is available.
func (o *deviceClass) hasAvailableComponent(component deviceClassComponent) bool {
	haha := o.getAvailableComponents()
	if v, ok := haha[component]; ok && v {
		return true
	}
	return false
}

func (y *yamlDeviceClass) convert() (deviceClass, error) {
	err := y.validate()
	if err != nil {
		return deviceClass{}, errors.Wrap(err, "invalid yaml device class")
	}
	var devClass deviceClass
	devClass.name = y.Name
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
		identify, err := y.Identify.convert()
		if err != nil {
			return deviceClass{}, errors.Wrap(err, "failed to convert identify")
		}
		devClass.identify = identify
	}

	components, err := y.Components.convert()
	if err != nil {
		return deviceClass{}, errors.Wrap(err, "failed to convert components")
	}
	devClass.components = components

	config, err := y.Config.convert()
	if err != nil {
		return deviceClass{}, errors.Wrap(err, "failed to convert components")
	}
	devClass.config = config

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

func (y *yamlDeviceClassIdentify) convert() (deviceClassIdentify, error) {
	err := y.validate()
	if err != nil {
		return deviceClassIdentify{}, errors.Wrap(err, "identify is invalid")
	}
	var identify deviceClassIdentify
	properties, err := y.Properties.convert()
	if err != nil {
		return deviceClassIdentify{}, errors.Wrap(err, "failed to read yaml identify properties")
	}
	identify.properties = properties

	return identify, nil
}

func (y *yamlDeviceClassComponents) convert() (deviceClassComponents, error) {
	var components deviceClassComponents

	if y.Interfaces != nil {
		interf, err := y.Interfaces.convert()
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml interface properties")
		}
		components.interfaces = &interf
	}

	if y.UPS != nil {
		ups, err := y.UPS.convert()
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml UPS properties")
		}
		components.ups = &ups
	}

	if y.CPU != nil {
		cpu, err := y.CPU.convert()
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml CPU properties")
		}
		components.cpu = &cpu
	}

	if y.Memory != nil {
		memory, err := y.Memory.convert()
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml memory properties")
		}
		components.memory = &memory
	}

	if y.SBC != nil {
		sbc, err := y.SBC.convert()
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml sbc properties")
		}
		components.sbc = &sbc
	}

	if y.HardwareHealth != nil {
		hardwareHealth, err := y.HardwareHealth.convert()
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml hardware health properties")
		}
		components.hardwareHealth = &hardwareHealth
	}

	return components, nil
}

func (y *yamlComponentsInterfaces) convert() (deviceClassComponentsInterfaces, error) {
	var interfaces deviceClassComponentsInterfaces
	var err error

	if y.IfTable != nil {
		interfaces.IfTable, err = interface2GroupPropertyReader(y.IfTable)
		if err != nil {
			return deviceClassComponentsInterfaces{}, errors.Wrap(err, "failed to convert ifTable")
		}
	}

	if y.Types != nil {
		types, err := y.Types.convert()
		if err != nil {
			return deviceClassComponentsInterfaces{}, errors.Wrap(err, "failed to read yaml interfaces types")
		}
		interfaces.Types = types
	}

	interfaces.Count = y.Count

	return interfaces, nil
}

func (y *yamlComponentsInterfaceTypes) convert() (deviceClassInterfaceTypes, error) {
	interfaceTypes := make(map[string]deviceClassInterfaceTypeDef)

	for k, interfaceType := range *y {
		if interfaceType.Detection == "" {
			return deviceClassInterfaceTypes{}, errors.New("detection information missing for special interface type")
		}
		if interfaceType.Values != nil {
			values, err := interfaceType.Values.convert()
			if err != nil {
				return deviceClassInterfaceTypes{}, errors.Wrap(err, "failed to read yaml interfaces types values")
			}
			interfaceTypes[k] = deviceClassInterfaceTypeDef{
				Detection: interfaceType.Detection,
				Values:    values,
			}
		} else {
			interfaceTypes[k] = deviceClassInterfaceTypeDef{
				Detection: interfaceType.Detection,
				Values:    nil,
			}
		}
	}

	return interfaceTypes, nil
}

func (y *yamlComponentsOIDs) convert() (deviceClassOIDs, error) {
	interfaceOIDs := make(map[string]deviceClassOID)

	for k, property := range *y {
		if property.Operators != nil {
			operators, err := interfaceSlice2propertyOperators(property.Operators, propertyDefault)
			if err != nil {
				return deviceClassOIDs{}, errors.Wrap(err, "failed to read yaml oids operators")
			}
			interfaceOIDs[k] = deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID:          (*y)[k].OID,
					UseRawResult: (*y)[k].UseRawResult,
				},
				operators: operators,
			}
		} else {
			interfaceOIDs[k] = deviceClassOID{
				SNMPGetConfiguration: network.SNMPGetConfiguration{
					OID:          (*y)[k].OID,
					UseRawResult: (*y)[k].UseRawResult,
				},
				operators: nil,
			}
		}
	}

	return interfaceOIDs, nil
}

func (d *deviceClassOIDs) validate() error {
	for label, oid := range *d {
		if err := oid.OID.Validate(); err != nil {
			return errors.Wrapf(err, "oid for %s is invalid", label)
		}
	}

	return nil
}

func conditionContainsUniqueRequest(c condition) bool {
	switch c.(type) {
	case *SnmpCondition:
		if c.(*SnmpCondition).Type == "snmpget" {
			return true
		}
	case *HTTPCondition:
		return true
	case *ConditionSet:
		for _, con := range c.(*ConditionSet).Conditions {
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

func (y *yamlDeviceClassIdentifyProperties) convert() (deviceClassIdentifyProperties, error) {
	var properties deviceClassIdentifyProperties
	var err error

	if y.Vendor != nil {
		properties.vendor, err = convertYamlProperty(y.Vendor, propertyVendor)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert vendor property to property reader")
		}
	}
	if y.Model != nil {
		properties.model, err = convertYamlProperty(y.Model, propertyModel)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert model property to property reader")
		}
	}
	if y.ModelSeries != nil {
		properties.modelSeries, err = convertYamlProperty(y.ModelSeries, propertyModelSeries)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert model series property to property reader")
		}
	}
	if y.SerialNumber != nil {
		properties.serialNumber, err = convertYamlProperty(y.SerialNumber, propertyDefault)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert serial number property to property reader")
		}
	}
	if y.OSVersion != nil {
		properties.osVersion, err = convertYamlProperty(y.OSVersion, propertyDefault)
		if err != nil {
			return deviceClassIdentifyProperties{}, errors.Wrap(err, "failed to convert osVersion property to property reader")
		}
	}
	return properties, nil
}

func (y *yamlDeviceClassConfig) convert() (deviceClassConfig, error) {
	var config deviceClassConfig
	config.snmp = y.SNMP
	config.components = make(map[deviceClassComponent]bool)

	for k, v := range y.Components {
		component, err := createComponent(k)
		if err != nil {
			return deviceClassConfig{}, err
		}
		config.components[component] = v
	}

	return config, nil
}

func (y *yamlComponentsUPSProperties) convert() (deviceClassComponentsUPS, error) {
	var properties deviceClassComponentsUPS
	var err error

	if y.AlarmLowVoltageDisconnect != nil {
		properties.alarmLowVoltageDisconnect, err = convertYamlProperty(y.AlarmLowVoltageDisconnect, propertyDefault)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert alarm low voltage disconnect property to property reader")
		}
	}
	if y.BatteryAmperage != nil {
		properties.batteryAmperage, err = convertYamlProperty(y.BatteryAmperage, propertyDefault)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery amperage property to property reader")
		}
	}
	if y.BatteryCapacity != nil {
		properties.batteryCapacity, err = convertYamlProperty(y.BatteryCapacity, propertyDefault)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery capacity property to property reader")
		}
	}
	if y.BatteryCurrent != nil {
		properties.batteryCurrent, err = convertYamlProperty(y.BatteryCurrent, propertyDefault)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery current property to property reader")
		}
	}
	if y.BatteryRemainingTime != nil {
		properties.batteryRemainingTime, err = convertYamlProperty(y.BatteryRemainingTime, propertyDefault)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery remaining time property to property reader")
		}
	}
	if y.BatteryTemperature != nil {
		properties.batteryTemperature, err = convertYamlProperty(y.BatteryTemperature, propertyDefault)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery temperature property to property reader")
		}
	}
	if y.BatteryVoltage != nil {
		properties.batteryVoltage, err = convertYamlProperty(y.BatteryVoltage, propertyDefault)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert battery voltage property to property reader")
		}
	}
	if y.CurrentLoad != nil {
		properties.currentLoad, err = convertYamlProperty(y.CurrentLoad, propertyDefault)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert current load property to property reader")
		}
	}
	if y.MainsVoltageApplied != nil {
		properties.mainsVoltageApplied, err = convertYamlProperty(y.MainsVoltageApplied, propertyDefault)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert mains voltage applied property to property reader")
		}
	}
	if y.RectifierCurrent != nil {
		properties.rectifierCurrent, err = convertYamlProperty(y.RectifierCurrent, propertyDefault)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert rectifier current property to property reader")
		}
	}
	if y.SystemVoltage != nil {
		properties.systemVoltage, err = convertYamlProperty(y.SystemVoltage, propertyDefault)
		if err != nil {
			return deviceClassComponentsUPS{}, errors.Wrap(err, "failed to convert system voltage property to property reader")
		}
	}
	return properties, nil
}

func (y *yamlComponentsCPUProperties) convert() (deviceClassComponentsCPU, error) {
	var properties deviceClassComponentsCPU
	var err error

	if y.Load != nil {
		properties.load, err = convertYamlProperty(y.Load, propertyDefault)
		if err != nil {
			return deviceClassComponentsCPU{}, errors.Wrap(err, "failed to convert load property to property reader")
		}
	}
	if y.Temperature != nil {
		properties.temperature, err = convertYamlProperty(y.Temperature, propertyDefault)
		if err != nil {
			return deviceClassComponentsCPU{}, errors.Wrap(err, "failed to convert temperature property to property reader")
		}
	}
	return properties, nil
}

func (y *yamlComponentsMemoryProperties) convert() (deviceClassComponentsMemory, error) {
	var properties deviceClassComponentsMemory
	var err error
	if y.Usage != nil {
		properties.usage, err = convertYamlProperty(y.Usage, propertyDefault)
		if err != nil {
			return deviceClassComponentsMemory{}, errors.Wrap(err, "failed to convert memory usage property to property reader")
		}
	}
	return properties, nil
}

func (y *yamlComponentsSBCProperties) convert() (deviceClassComponentsSBC, error) {
	var properties deviceClassComponentsSBC
	var err error

	if y.Agents != nil {
		properties.agents, err = interface2GroupPropertyReader(y.Agents)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert agents property to group property reader")
		}
	}
	if y.Realms != nil {
		properties.realms, err = interface2GroupPropertyReader(y.Realms)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert realms property to group property reader")
		}
	}
	if y.ActiveLocalContacts != nil {
		properties.activeLocalContacts, err = convertYamlProperty(y.ActiveLocalContacts, propertyDefault)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert active local contacts property to property reader")
		}
	}
	if y.GlobalCallPerSecond != nil {
		properties.globalCallPerSecond, err = convertYamlProperty(y.GlobalCallPerSecond, propertyDefault)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert global call per second property to property reader")
		}
	}
	if y.GlobalConcurrentSessions != nil {
		properties.globalConcurrentSessions, err = convertYamlProperty(y.GlobalConcurrentSessions, propertyDefault)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert global concurrent sessions property to property reader")
		}
	}
	if y.LicenseCapacity != nil {
		properties.licenseCapacity, err = convertYamlProperty(y.LicenseCapacity, propertyDefault)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert license capacity property to property reader")
		}
	}
	if y.TranscodingCapacity != nil {
		properties.transcodingCapacity, err = convertYamlProperty(y.TranscodingCapacity, propertyDefault)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert transcoding capacity property to property reader")
		}
	}
	if y.SystemRedundancy != nil {
		properties.systemRedundancy, err = convertYamlProperty(y.SystemRedundancy, propertyDefault)
		if err != nil {
			return deviceClassComponentsSBC{}, errors.Wrap(err, "failed to convert system redundancy property to property reader")
		}
	}
	return properties, nil
}

func (y *yamlComponentsHardwareHealthProperties) convert() (deviceClassComponentsHardwareHealth, error) {
	var properties deviceClassComponentsHardwareHealth
	var err error

	if y.Fans != nil {
		properties.fans, err = interface2GroupPropertyReader(y.Fans)
		if err != nil {
			return deviceClassComponentsHardwareHealth{}, errors.Wrap(err, "failed to convert fans property to group property reader")
		}
	}
	if y.PowerSupply != nil {
		properties.powerSupply, err = interface2GroupPropertyReader(y.PowerSupply)
		if err != nil {
			return deviceClassComponentsHardwareHealth{}, errors.Wrap(err, "failed to convert power supply property to group property reader")
		}
	}
	if y.EnvironmentMonitorState != nil {
		properties.environmentMonitorState, err = convertYamlProperty(y.EnvironmentMonitorState, propertyDefault)
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

func convertYamlProperty(i []interface{}, task relatedTask) (propertyReader, error) {
	var readerSet propertyReaderSet
	for _, i := range i {
		reader, err := interface2propertyReader(i, task)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert yaml identify property")
		}
		readerSet = append(readerSet, reader)
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
		var pr constantPropertyReader
		err := mapstructure.Decode(i, &pr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode constant propertyReader")
		}
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
				var mapModifier mapModifier
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
				v := value.New(m["value"])
				val, err := v.Float64()
				if err != nil {
					return nil, errors.New("value is missing in multiply modify operator, or is not of type float64")
				}
				var multiplyModifier multiplyNumberModifier
				multiplyModifier.value = val
				modifier.operator = &multiplyModifier
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

func interface2GroupPropertyReader(i interface{}) (groupPropertyReader, error) {
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
		var oids yamlComponentsOIDs
		if _, ok := m["values"]; !ok {
			return nil, errors.New("values are missing")
		}
		values, ok := m["values"].(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("values needs to be a map")
		}
		err := mapstructure.Decode(values, &oids)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode values map to yamlComponentsOIDs")
		}
		deviceClassOIDs, err := oids.convert()
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert yaml OIDs to device class OIDs")
		}
		err = deviceClassOIDs.validate()
		if err != nil {
			return nil, errors.Wrap(err, "snmpwalk group property reader is invalid")
		}
		return &snmpGroupPropertyReader{deviceClassOIDs}, nil
	default:
		return nil, fmt.Errorf("unknown detection type '%s'", stringDetection)
	}
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

func createComponent(component string) (deviceClassComponent, error) {
	switch component {
	case "interfaces":
		return interfacesComponent, nil
	case "ups":
		return upsComponent, nil
	case "cpu":
		return cpuComponent, nil
	case "memory":
		return memoryComponent, nil
	case "sbc":
		return sbcComponent, nil
	case "hardware_health":
		return hardwareHealthComponent, nil
	default:
		return 0, fmt.Errorf("invalid component type: %s", component)
	}
}

func (d *deviceClassComponent) toString() (string, error) {
	if d == nil {
		return "", errors.New("component is empty")
	}
	switch *d {
	case interfacesComponent:
		return "interfaces", nil
	case upsComponent:
		return "ups", nil
	case cpuComponent:
		return "cpu", nil
	case memoryComponent:
		return "memory", nil
	case sbcComponent:
		return "sbc", nil
	case hardwareHealthComponent:
		return "hardware_health", nil
	default:
		return "", errors.New("unknown component")
	}
}
