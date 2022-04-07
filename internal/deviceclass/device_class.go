// Package deviceclass contains the logic for interacting with device classes.
// It contains methods that read out the .yaml files representing device classes.
package deviceclass

import (
	"context"
	"github.com/inexio/thola/config"
	"github.com/inexio/thola/config/codecommunicator"
	"github.com/inexio/thola/internal/communicator"
	"github.com/inexio/thola/internal/communicator/hierarchy"
	"github.com/inexio/thola/internal/component"
	"github.com/inexio/thola/internal/deviceclass/condition"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/inexio/thola/internal/deviceclass/property"
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
	interfaces       *deviceClassComponentsInterfaces
	ups              *deviceClassComponentsUPS
	cpu              *deviceClassComponentsCPU
	memory           *deviceClassComponentsMemory
	sbc              *deviceClassComponentsSBC
	server           *deviceClassComponentsServer
	disk             *deviceClassComponentsDisk
	hardwareHealth   *deviceClassComponentsHardwareHealth
	highAvailability *deviceClassComponentsHighAvailability
	siem             *deviceClassComponentsSIEM
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
	usage groupproperty.Reader
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

// deviceClassComponentsHardwareHealth represents the hardware health part of a device class.
type deviceClassComponentsHardwareHealth struct {
	environmentMonitorState property.Reader
	fans                    groupproperty.Reader
	powerSupply             groupproperty.Reader
	temperature             groupproperty.Reader
	voltage                 groupproperty.Reader
}

// deviceClassComponentsHighAvailability represents the high availability part of a device class.
type deviceClassComponentsHighAvailability struct {
	state property.Reader
	role  property.Reader
	nodes property.Reader
}

// deviceClassComponentsSIEM represents the SIEM part of a device class.
type deviceClassComponentsSIEM struct {
	siem                                         property.Reader
	systemVersion                                property.Reader
	lastRecordedMessagesPerSecondNormalizer      property.Reader
	averageMessagesPerSecondLast5minNormalizer   property.Reader
	lastRecordedMessagesPerSecondStoreHandler    property.Reader
	averageMessagesPerSecondLast5minStoreHandler property.Reader
	servicesCurrentlyDown                        property.Reader

	cpuConsumptionCollection      property.Reader
	cpuConsumptionNormalization   property.Reader
	cpuConsumptionEnrichment      property.Reader
	cpuConsumptionIndexing        property.Reader
	cpuConsumptionDashboardAlerts property.Reader

	memoryConsumptionCollection      property.Reader
	memoryConsumptionNormalization   property.Reader
	memoryConsumptionEnrichment      property.Reader
	memoryConsumptionIndexing        property.Reader
	memoryConsumptionDashboardAlerts property.Reader

	queueCollection      property.Reader
	queueNormalization   property.Reader
	queueEnrichment      property.Reader
	queueIndexing        property.Reader
	queueDashboardAlerts property.Reader

	activeSearchProcesses property.Reader

	diskUsageDashboardAlerts property.Reader

	zfsPools groupproperty.Reader

	repositories groupproperty.Reader

	//director
	fabricServerVersion                       property.Reader
	fabricServerIOWait                        property.Reader
	fabricServerVMSwapiness                   property.Reader
	fabricServerClusterSize                   property.Reader
	fabricServerProxyCpuUsage                 property.Reader
	fabricServerProxyMemoryUsage              property.Reader
	fabricServerProxyNumberOfAliveConnections property.Reader
	fabricServerProxyState                    property.Reader
	fabricServerProxyNodesCount               property.Reader
	fabricServerStorageCpuUsage               property.Reader
	fabricServerStorageMemoryUsage            property.Reader
	fabricServerStorageConfiguredCapacity     property.Reader
	fabricServerStorageAvailableCapacity      property.Reader
	fabricServerStorageDfsUsed                property.Reader
	fabricServerStorageUnderReplicatedBlocks  property.Reader
	fabricServerStorageLiveDataNodes          property.Reader

	fabricServerAuthenticatorCpuUsage           property.Reader
	fabricServerAuthenticatorMemoryUsage        property.Reader
	fabricServerAuthenticatorServiceStatus      property.Reader
	fabricServerAuthenticatorAdminServiceStatus property.Reader
	fabricServerZFSPools                        groupproperty.Reader

	apiServerVersion     property.Reader
	apiServerIOWait      property.Reader
	apiServerVMSwapiness property.Reader
	apiServerCpuUsage    property.Reader
	apiServerMemoryUsage property.Reader
}

// deviceClassConfig represents the config part of a device class.
type deviceClassConfig struct {
	snmp       deviceClassSNMP
	components map[component.Component]bool
}

// deviceClassComponentsInterfaces represents the interface properties part of a device class.
type deviceClassComponentsInterfaces struct {
	count      property.Reader
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
	Interfaces       *yamlComponentsInterfaces               `yaml:"interfaces"`
	UPS              *yamlComponentsUPSProperties            `yaml:"ups"`
	CPU              *yamlComponentsCPUProperties            `yaml:"cpu"`
	Memory           *yamlComponentsMemoryProperties         `yaml:"memory"`
	SBC              *yamlComponentsSBCProperties            `yaml:"sbc"`
	Server           *yamlComponentsServerProperties         `yaml:"server"`
	Disk             *yamlComponentsDiskProperties           `yaml:"disk"`
	HardwareHealth   *yamlComponentsHardwareHealthProperties `yaml:"hardware_health"`
	HighAvailability *yamlComponentsHighAvailability         `yaml:"high_availability"`
	SIEM             *yamlComponentsSIEM                     `yaml:"siem"`
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
	Properties interface{} `yaml:"properties"`
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
	Temperature             interface{}   `yaml:"temperature"`
	Voltage                 interface{}   `yaml:"voltage"`
}

// yamlComponentsHa represents the specific properties of HA components of a yaml device class.
type yamlComponentsHighAvailability struct {
	State []interface{}
	Role  []interface{}
	Nodes []interface{}
}

// yamlComponentsSIEM represents the SIEM part of a device class.
type yamlComponentsSIEM struct {
	SIEM                                         []interface{} `yaml:"siem"`
	SystemVersion                                []interface{} `yaml:"system_version"`
	LastRecordedMessagesPerSecondNormalizer      []interface{} `yaml:"last_recorded_messages_per_second_normalizer"`
	AverageMessagesPerSecondLast5minNormalizer   []interface{} `yaml:"average_messages_per_second_last_5_min_normalizer"`
	LastRecordedMessagesPerSecondStoreHandler    []interface{} `yaml:"last_recorded_messages_per_second_store_handler"`
	AverageMessagesPerSecondLast5minStoreHandler []interface{} `yaml:"average_messages_per_second_last_5_min_store_handler"`
	ServicesCurrentlyDown                        []interface{} `yaml:"services_currently_down"`

	CpuConsumptionCollection      []interface{} `yaml:"cpu_consumption_collection"`
	CpuConsumptionNormalization   []interface{} `yaml:"cpu_consumption_normalization"`
	CpuConsumptionEnrichment      []interface{} `yaml:"cpu_consumption_enrichment"`
	CpuConsumptionIndexing        []interface{} `yaml:"cpu_consumption_indexing"`
	CpuConsumptionDashboardAlerts []interface{} `yaml:"cpu_consumption_dashboard_alerts"`

	MemoryConsumptionCollection      []interface{} `yaml:"memory_consumption_collection"`
	MemoryConsumptionNormalization   []interface{} `yaml:"memory_consumption_normalization"`
	MemoryConsumptionEnrichment      []interface{} `yaml:"memory_consumption_enrichment"`
	MemoryConsumptionIndexing        []interface{} `yaml:"memory_consumption_indexing"`
	MemoryConsumptionDashboardAlerts []interface{} `yaml:"memory_consumption_dashboard_alerts"`

	QueueCollection      []interface{} `yaml:"queue_collection"`
	QueueNormalization   []interface{} `yaml:"queue_normalization"`
	QueueEnrichment      []interface{} `yaml:"queue_enrichment"`
	QueueIndexing        []interface{} `yaml:"queue_indexing"`
	QueueDashboardAlerts []interface{} `yaml:"queue_dashboard_alerts"`

	ActiveSearchProcesses []interface{} `yaml:"active_search_processes"`

	DiskUsageDashboardAlerts []interface{} `yaml:"disk_usage_dashboard_alerts"`

	ZFSPools interface{} `yaml:"zfs_pools"`

	Repositories interface{} `yaml:"repositories"`

	//director
	FabricServerProxyNodesCount               []interface{} `yaml:"fabric_server_proxy_nodes_count"`
	FabricServerVersion                       []interface{} `yaml:"fabric_server_version"`
	FabricServerIOWait                        []interface{} `yaml:"fabric_server_i_o_wait"`
	FabricServerVMSwapiness                   []interface{} `yaml:"fabric_server_vm_swapiness"`
	FabricServerClusterSize                   []interface{} `yaml:"fabric_server_cluster_size"`
	FabricServerProxyCpuUsage                 []interface{} `yaml:"fabric_server_proxy_cpu_usage"`
	FabricServerProxyMemoryUsage              []interface{} `yaml:"fabric_server_proxy_memory_usage"`
	FabricServerProxyNumberOfAliveConnections []interface{} `yaml:"fabric_server_proxy_number_of_alive_connections"`
	FabricServerProxyState                    []interface{} `yaml:"fabric_server_proxy_state"`
	FabricServerStorageCpuUsage               []interface{} `yaml:"fabric_server_storage_cpu_usage"`
	FabricServerStorageMemoryUsage            []interface{} `yaml:"fabric_server_storage_memory_usage"`
	FabricServerStorageConfiguredCapacity     []interface{} `yaml:"fabric_server_storage_configured_capacity"`
	FabricServerStorageAvailableCapacity      []interface{} `yaml:"fabric_server_storage_available_capacity"`
	FabricServerStorageDfsUsed                []interface{} `yaml:"fabric_server_storage_dfs_used"`
	FabricServerStorageUnderReplicatedBlocks  []interface{} `yaml:"fabric_server_storage_under_replicated_blocks"`
	FabricServerStorageLiveDataNodes          []interface{} `yaml:"fabric_server_storage_live_data_nodes"`

	FabricServerAuthenticatorCpuUsage           []interface{} `yaml:"fabric_server_authenticator_cpu_usage"`
	FabricServerAuthenticatorMemoryUsage        []interface{} `yaml:"fabric_server_authenticator_memory_usage"`
	FabricServerAuthenticatorServiceStatus      []interface{} `yaml:"fabric_server_authenticator_service_status"`
	FabricServerAuthenticatorAdminServiceStatus []interface{} `yaml:"fabric_server_authenticator_admin_service_status"`
	FabricServerZFSPools                        interface{}   `yaml:"fabric_server_zfs_pools"`

	ApiServerVersion     []interface{} `yaml:"api_server_version"`
	ApiServerIOWait      []interface{} `yaml:"api_server_i_o_wait"`
	ApiServerVMSwapiness []interface{} `yaml:"api_server_vm_swapiness"`
	ApiServerCpuUsage    []interface{} `yaml:"api_server_cluster_cpu_usage"`
	ApiServerMemoryUsage []interface{} `yaml:"api_server_cluster_memory_usage"`
}

//
// Here are definitions of interfaces of yaml device classes.
//

type yamlComponentsInterfaces struct {
	Count      []interface{} `yaml:"count"`
	Properties interface{}   `yaml:"properties"`
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

	if y.HighAvailability != nil {
		ha, err := y.HighAvailability.convert(parentComponents.highAvailability)
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml high availability properties")
		}
		components.highAvailability = &ha
	}

	if y.SIEM != nil {
		siem, err := y.SIEM.convert(parentComponents.siem)
		if err != nil {
			return deviceClassComponents{}, errors.Wrap(err, "failed to read yaml siem properties")
		}
		components.siem = &siem
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

	if y.Count != nil {
		interfaceComponent.count, err = property.InterfaceSlice2Reader(y.Count, condition.PropertyDefault, interfaceComponent.count)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert interface count")
		}
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

	if y.Properties != nil {
		prop.usage, err = groupproperty.Interface2Reader(y.Properties, prop.usage)
		if err != nil {
			return deviceClassComponentsMemory{}, errors.Wrap(err, "failed to convert memory usage property to group property reader")
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
	if y.Temperature != nil {
		prop.temperature, err = groupproperty.Interface2Reader(y.Temperature, prop.temperature)
		if err != nil {
			return deviceClassComponentsHardwareHealth{}, errors.Wrap(err, "failed to convert temperature property to group property reader")
		}
	}
	if y.Voltage != nil {
		prop.voltage, err = groupproperty.Interface2Reader(y.Voltage, prop.voltage)
		if err != nil {
			return deviceClassComponentsHardwareHealth{}, errors.Wrap(err, "failed to convert voltage property to group property reader")
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

func (y *yamlComponentsHighAvailability) convert(parentHA *deviceClassComponentsHighAvailability) (deviceClassComponentsHighAvailability, error) {
	var prop deviceClassComponentsHighAvailability
	var err error

	if parentHA != nil {
		prop = *parentHA
	}

	if y.State != nil {
		prop.state, err = property.InterfaceSlice2Reader(y.State, condition.PropertyDefault, prop.state)
		if err != nil {
			return deviceClassComponentsHighAvailability{}, errors.Wrap(err, "failed to convert state property to property reader")
		}
	}

	if y.Role != nil {
		prop.role, err = property.InterfaceSlice2Reader(y.Role, condition.PropertyDefault, prop.role)
		if err != nil {
			return deviceClassComponentsHighAvailability{}, errors.Wrap(err, "failed to convert role property to property reader")
		}
	}

	if y.Nodes != nil {
		prop.nodes, err = property.InterfaceSlice2Reader(y.Nodes, condition.PropertyDefault, prop.nodes)
		if err != nil {
			return deviceClassComponentsHighAvailability{}, errors.Wrap(err, "failed to convert nodes property to property reader")
		}
	}

	return prop, nil
}

func (y *yamlComponentsSIEM) convert(parentSIEM *deviceClassComponentsSIEM) (deviceClassComponentsSIEM, error) {
	var prop deviceClassComponentsSIEM
	var err error

	if parentSIEM != nil {
		prop = *parentSIEM
	}

	if y.LastRecordedMessagesPerSecondNormalizer != nil {
		prop.lastRecordedMessagesPerSecondNormalizer, err = property.InterfaceSlice2Reader(y.LastRecordedMessagesPerSecondNormalizer, condition.PropertyDefault, prop.lastRecordedMessagesPerSecondNormalizer)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert LastRecordedMessagesPerSecondNormalizer property to property reader")
		}
	}

	if y.AverageMessagesPerSecondLast5minNormalizer != nil {
		prop.averageMessagesPerSecondLast5minNormalizer, err = property.InterfaceSlice2Reader(y.AverageMessagesPerSecondLast5minNormalizer, condition.PropertyDefault, prop.averageMessagesPerSecondLast5minNormalizer)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert AverageMessagesPerSecondLast5minNormalizer property to property reader")
		}
	}

	if y.LastRecordedMessagesPerSecondStoreHandler != nil {
		prop.lastRecordedMessagesPerSecondStoreHandler, err = property.InterfaceSlice2Reader(y.LastRecordedMessagesPerSecondStoreHandler, condition.PropertyDefault, prop.lastRecordedMessagesPerSecondStoreHandler)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert LastRecordedMessagesPerSecondStoreHandler property to property reader")
		}
	}

	if y.AverageMessagesPerSecondLast5minStoreHandler != nil {
		prop.averageMessagesPerSecondLast5minStoreHandler, err = property.InterfaceSlice2Reader(y.AverageMessagesPerSecondLast5minStoreHandler, condition.PropertyDefault, prop.averageMessagesPerSecondLast5minStoreHandler)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert AverageMessagesPerSecondLast5minStoreHandler property to property reader")
		}
	}

	if y.ServicesCurrentlyDown != nil {
		prop.servicesCurrentlyDown, err = property.InterfaceSlice2Reader(y.ServicesCurrentlyDown, condition.PropertyDefault, prop.servicesCurrentlyDown)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert ServicesCurrentlyDown property to property reader")
		}
	}

	if y.SystemVersion != nil {
		prop.systemVersion, err = property.InterfaceSlice2Reader(y.SystemVersion, condition.PropertyDefault, prop.systemVersion)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert SystemVersion property to property reader")
		}
	}

	if y.SIEM != nil {
		prop.siem, err = property.InterfaceSlice2Reader(y.SIEM, condition.PropertyDefault, prop.siem)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.CpuConsumptionCollection != nil {
		prop.cpuConsumptionCollection, err = property.InterfaceSlice2Reader(y.CpuConsumptionCollection, condition.PropertyDefault, prop.cpuConsumptionCollection)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.CpuConsumptionNormalization != nil {
		prop.cpuConsumptionNormalization, err = property.InterfaceSlice2Reader(y.CpuConsumptionNormalization, condition.PropertyDefault, prop.cpuConsumptionNormalization)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.CpuConsumptionEnrichment != nil {
		prop.cpuConsumptionEnrichment, err = property.InterfaceSlice2Reader(y.CpuConsumptionEnrichment, condition.PropertyDefault, prop.cpuConsumptionEnrichment)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.CpuConsumptionIndexing != nil {
		prop.cpuConsumptionIndexing, err = property.InterfaceSlice2Reader(y.CpuConsumptionIndexing, condition.PropertyDefault, prop.cpuConsumptionIndexing)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.CpuConsumptionDashboardAlerts != nil {
		prop.cpuConsumptionDashboardAlerts, err = property.InterfaceSlice2Reader(y.CpuConsumptionDashboardAlerts, condition.PropertyDefault, prop.cpuConsumptionDashboardAlerts)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.MemoryConsumptionCollection != nil {
		prop.memoryConsumptionCollection, err = property.InterfaceSlice2Reader(y.MemoryConsumptionCollection, condition.PropertyDefault, prop.memoryConsumptionCollection)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.MemoryConsumptionNormalization != nil {
		prop.memoryConsumptionNormalization, err = property.InterfaceSlice2Reader(y.MemoryConsumptionNormalization, condition.PropertyDefault, prop.memoryConsumptionNormalization)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.MemoryConsumptionEnrichment != nil {
		prop.memoryConsumptionEnrichment, err = property.InterfaceSlice2Reader(y.MemoryConsumptionEnrichment, condition.PropertyDefault, prop.memoryConsumptionEnrichment)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.MemoryConsumptionIndexing != nil {
		prop.memoryConsumptionIndexing, err = property.InterfaceSlice2Reader(y.MemoryConsumptionIndexing, condition.PropertyDefault, prop.memoryConsumptionIndexing)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.MemoryConsumptionDashboardAlerts != nil {
		prop.memoryConsumptionDashboardAlerts, err = property.InterfaceSlice2Reader(y.MemoryConsumptionDashboardAlerts, condition.PropertyDefault, prop.memoryConsumptionDashboardAlerts)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.QueueCollection != nil {
		prop.queueCollection, err = property.InterfaceSlice2Reader(y.QueueCollection, condition.PropertyDefault, prop.queueCollection)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.QueueNormalization != nil {
		prop.queueNormalization, err = property.InterfaceSlice2Reader(y.QueueNormalization, condition.PropertyDefault, prop.queueNormalization)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.QueueEnrichment != nil {
		prop.queueEnrichment, err = property.InterfaceSlice2Reader(y.QueueEnrichment, condition.PropertyDefault, prop.queueEnrichment)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.QueueIndexing != nil {
		prop.queueIndexing, err = property.InterfaceSlice2Reader(y.QueueIndexing, condition.PropertyDefault, prop.queueIndexing)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.QueueDashboardAlerts != nil {
		prop.queueDashboardAlerts, err = property.InterfaceSlice2Reader(y.QueueDashboardAlerts, condition.PropertyDefault, prop.queueDashboardAlerts)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.ActiveSearchProcesses != nil {
		prop.activeSearchProcesses, err = property.InterfaceSlice2Reader(y.ActiveSearchProcesses, condition.PropertyDefault, prop.activeSearchProcesses)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.DiskUsageDashboardAlerts != nil {
		prop.diskUsageDashboardAlerts, err = property.InterfaceSlice2Reader(y.DiskUsageDashboardAlerts, condition.PropertyDefault, prop.diskUsageDashboardAlerts)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.ZFSPools != nil {
		prop.zfsPools, err = groupproperty.Interface2Reader(y.ZFSPools, prop.zfsPools)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.Repositories != nil {
		prop.repositories, err = groupproperty.Interface2Reader(y.Repositories, prop.repositories)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	//director
	if y.FabricServerProxyCpuUsage != nil {
		prop.fabricServerProxyCpuUsage, err = property.InterfaceSlice2Reader(y.FabricServerProxyCpuUsage, condition.PropertyDefault, prop.fabricServerProxyCpuUsage)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerProxyMemoryUsage != nil {
		prop.fabricServerProxyMemoryUsage, err = property.InterfaceSlice2Reader(y.FabricServerProxyMemoryUsage, condition.PropertyDefault, prop.fabricServerProxyMemoryUsage)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerProxyNumberOfAliveConnections != nil {
		prop.fabricServerProxyNumberOfAliveConnections, err = property.InterfaceSlice2Reader(y.FabricServerProxyNumberOfAliveConnections, condition.PropertyDefault, prop.fabricServerProxyNumberOfAliveConnections)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerProxyState != nil {
		prop.fabricServerProxyState, err = property.InterfaceSlice2Reader(y.FabricServerProxyState, condition.PropertyDefault, prop.fabricServerProxyState)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerProxyNodesCount != nil {
		prop.fabricServerProxyNodesCount, err = property.InterfaceSlice2Reader(y.FabricServerProxyNodesCount, condition.PropertyDefault, prop.fabricServerProxyNodesCount)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerVersion != nil {
		prop.fabricServerVersion, err = property.InterfaceSlice2Reader(y.FabricServerVersion, condition.PropertyDefault, prop.fabricServerVersion)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerIOWait != nil {
		prop.fabricServerIOWait, err = property.InterfaceSlice2Reader(y.FabricServerIOWait, condition.PropertyDefault, prop.fabricServerIOWait)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerVMSwapiness != nil {
		prop.fabricServerVMSwapiness, err = property.InterfaceSlice2Reader(y.FabricServerVMSwapiness, condition.PropertyDefault, prop.fabricServerVMSwapiness)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerClusterSize != nil {
		prop.fabricServerClusterSize, err = property.InterfaceSlice2Reader(y.FabricServerClusterSize, condition.PropertyDefault, prop.fabricServerClusterSize)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerStorageCpuUsage != nil {
		prop.fabricServerStorageCpuUsage, err = property.InterfaceSlice2Reader(y.FabricServerStorageCpuUsage, condition.PropertyDefault, prop.fabricServerStorageCpuUsage)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerStorageMemoryUsage != nil {
		prop.fabricServerStorageMemoryUsage, err = property.InterfaceSlice2Reader(y.FabricServerStorageMemoryUsage, condition.PropertyDefault, prop.fabricServerStorageMemoryUsage)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerStorageConfiguredCapacity != nil {
		prop.fabricServerStorageConfiguredCapacity, err = property.InterfaceSlice2Reader(y.FabricServerStorageConfiguredCapacity, condition.PropertyDefault, prop.fabricServerStorageConfiguredCapacity)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerStorageAvailableCapacity != nil {
		prop.fabricServerStorageAvailableCapacity, err = property.InterfaceSlice2Reader(y.FabricServerStorageAvailableCapacity, condition.PropertyDefault, prop.fabricServerStorageAvailableCapacity)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerStorageDfsUsed != nil {
		prop.fabricServerStorageDfsUsed, err = property.InterfaceSlice2Reader(y.FabricServerStorageDfsUsed, condition.PropertyDefault, prop.fabricServerStorageDfsUsed)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerStorageUnderReplicatedBlocks != nil {
		prop.fabricServerStorageUnderReplicatedBlocks, err = property.InterfaceSlice2Reader(y.FabricServerStorageUnderReplicatedBlocks, condition.PropertyDefault, prop.fabricServerStorageUnderReplicatedBlocks)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerStorageLiveDataNodes != nil {
		prop.fabricServerStorageLiveDataNodes, err = property.InterfaceSlice2Reader(y.FabricServerStorageLiveDataNodes, condition.PropertyDefault, prop.fabricServerStorageLiveDataNodes)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerAuthenticatorCpuUsage != nil {
		prop.fabricServerAuthenticatorCpuUsage, err = property.InterfaceSlice2Reader(y.FabricServerAuthenticatorCpuUsage, condition.PropertyDefault, prop.fabricServerAuthenticatorCpuUsage)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerAuthenticatorMemoryUsage != nil {
		prop.fabricServerAuthenticatorMemoryUsage, err = property.InterfaceSlice2Reader(y.FabricServerAuthenticatorMemoryUsage, condition.PropertyDefault, prop.fabricServerAuthenticatorMemoryUsage)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerAuthenticatorServiceStatus != nil {
		prop.fabricServerAuthenticatorServiceStatus, err = property.InterfaceSlice2Reader(y.FabricServerAuthenticatorServiceStatus, condition.PropertyDefault, prop.fabricServerAuthenticatorServiceStatus)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerAuthenticatorAdminServiceStatus != nil {
		prop.fabricServerAuthenticatorAdminServiceStatus, err = property.InterfaceSlice2Reader(y.FabricServerAuthenticatorAdminServiceStatus, condition.PropertyDefault, prop.fabricServerAuthenticatorAdminServiceStatus)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.FabricServerZFSPools != nil {
		prop.fabricServerZFSPools, err = groupproperty.Interface2Reader(y.FabricServerZFSPools, prop.fabricServerZFSPools)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.ApiServerVersion != nil {
		prop.apiServerVersion, err = property.InterfaceSlice2Reader(y.ApiServerVersion, condition.PropertyDefault, prop.apiServerVersion)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.ApiServerIOWait != nil {
		prop.apiServerIOWait, err = property.InterfaceSlice2Reader(y.ApiServerIOWait, condition.PropertyDefault, prop.apiServerIOWait)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.ApiServerVMSwapiness != nil {
		prop.apiServerVMSwapiness, err = property.InterfaceSlice2Reader(y.ApiServerVMSwapiness, condition.PropertyDefault, prop.apiServerVMSwapiness)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.ApiServerCpuUsage != nil {
		prop.apiServerCpuUsage, err = property.InterfaceSlice2Reader(y.ApiServerCpuUsage, condition.PropertyDefault, prop.apiServerCpuUsage)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	if y.ApiServerMemoryUsage != nil {
		prop.apiServerMemoryUsage, err = property.InterfaceSlice2Reader(y.ApiServerMemoryUsage, condition.PropertyDefault, prop.apiServerMemoryUsage)
		if err != nil {
			return deviceClassComponentsSIEM{}, errors.Wrap(err, "failed to convert siem property to property reader")
		}
	}

	return prop, nil
}
