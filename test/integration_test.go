package test

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/parser"
	"github.com/inexio/thola/core/request"
	"github.com/mitchellh/colorstring"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
)

type testDevice struct {
	info         testDeviceInfo
	expectations DeviceTestDataExpectations
	requestTypes []string
	retries      int
}

type testDeviceInfo interface {
	generateRequest(requestType string) (request.Request, error)
	getIdentifier() string
}

type testDeviceInfoSNMPSim struct {
	requestDeviceData request.DeviceData
}

type testConfig struct {
	ConcurrentRequests int    `yaml:"concurrentRequests"`
	Retries            int    `yaml:"retries"`
	SimpleUI           bool   `yaml:"simpleUI"`
	SNMPRecDir         string `yaml:"snmpRecDir"`
	APIPort            int    `yaml:"apiPort"`
	KeepDockerAlive    bool   `yaml:"keepDockerAlive"`
}

type statistics struct {
	failed  map[string]string
	success map[string]string
}

var (
	testConf       testConfig
	bar            *progressbar.ProgressBar
	requestCounter int32
	snmpSimIPs     chan string
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	_, currFilename, _, _ := runtime.Caller(0)

	viper.AddConfigPath(filepath.Join(path.Dir(currFilename), "testdata"))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.SetEnvPrefix("THOLA_TEST")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read in test config!")
	}

	err = viper.Unmarshal(&testConf)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal test config!")
	}

	snmpSimIPs = make(chan string, 3)
	snmpSimIPs <- "172.20.0.8"
	snmpSimIPs <- "172.20.0.9"
}

func Test(t *testing.T) {
	if !testConf.SimpleUI {
		_, _ = colorstring.Println("[cyan][1/3][reset] Building up test environment...")
	}

	_, currFilename, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(path.Dir(currFilename), "testdata")

	BuildupTestEnvironment(testdataDir)
	if !testConf.KeepDockerAlive {
		defer CleanupTestEnvironment(testdataDir)
	}

	deviceChannel, err := createTestDevices()
	if !assert.NoError(t, err, "an error occurred while creating test devices") {
		return
	}

	deviceAmount := len(deviceChannel)
	assert.True(t, deviceAmount > 0, "no device data found")

	err = waitForDevices(deviceChannel)
	if !assert.NoError(t, err, "an error occurred while waiting for the test devices being ready") {
		return
	}

	bar = progressbar.NewOptions(int(requestCounter),
		progressbar.OptionSetDescription("[cyan][2/3][reset] Running tests..."),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	if !testConf.SimpleUI {
		_ = bar.RenderBlank()
	}

	assert.True(t, testConf.ConcurrentRequests > 0, "invalid amount of concurrent requests")

	statChan := make(chan statistics, testConf.ConcurrentRequests)
	for i := 0; i < testConf.ConcurrentRequests; i++ {
		go sendRequests(deviceChannel, statChan)
	}

	stats := statistics{
		failed:  make(map[string]string),
		success: make(map[string]string),
	}
	for i := 0; i < testConf.ConcurrentRequests; i++ {
		stat := <-statChan
		for k, v := range stat.failed {
			stats.failed[k] = v
		}
		for k, v := range stat.success {
			stats.success[k] = v
		}
	}

	if !testConf.SimpleUI {
		fmt.Print("\n")
		_, _ = colorstring.Println("[cyan][3/3][reset] Finished!")
		fmt.Print("\n")
	}

	assert.True(t, len(stats.failed) == 0, "some devices failed")

	//OUTPUT
	if len(stats.success) > 0 {
		fmt.Println("SUCCESS DEVICES:")
		for testDevicePath, msg := range stats.success {
			fmt.Println(testDevicePath + ": " + msg)
		}
		fmt.Print("\n")
	}
	if len(stats.failed) > 0 {
		fmt.Println("FAILED DEVICES:")
		for testDevicePath, msg := range stats.failed {
			fmt.Println(testDevicePath + ": " + msg)
		}
		fmt.Print("\n")
	}
	fmt.Printf("%d out of %d test devices were successful!\n\n", len(stats.success), deviceAmount)
}

func sendRequests(input chan testDevice, output chan statistics) {
	stats := statistics{
		failed:  make(map[string]string),
		success: make(map[string]string),
	}
	for {
		select {
		case testDevice := <-input:
			r, err := testDevice.info.generateRequest(testDevice.requestTypes[0])
			if err != nil {
				stats.failed[testDevice.info.getIdentifier()] = fmt.Sprintf("generating new request failed: %s", err.Error())
				continue
			}
			response, err := ProcessRequest(r, testConf.APIPort)
			if err != nil {
				if testDevice.checkForRetryRequest() {
					testDevice.retries++
					input <- testDevice
				} else {
					stats.failed[testDevice.info.getIdentifier()] = err.Error()
				}
			} else {
				err := testDevice.expectations.compareExpectations(response, testDevice.requestTypes[0])
				if err == nil {
					// EXPECTATIONS MATCH
					if len(testDevice.requestTypes) > 1 {
						testDevice.requestTypes[0] = testDevice.requestTypes[len(testDevice.requestTypes)-1]
						testDevice.requestTypes = testDevice.requestTypes[:len(testDevice.requestTypes)-1]
						input <- testDevice
					} else {
						stats.success[testDevice.info.getIdentifier()] = "success"
					}
				} else {
					// EXPECTATIONS DONT MATCH
					if testDevice.checkForRetryRequest() {
						testDevice.retries++
						input <- testDevice
					} else {
						stats.failed[testDevice.info.getIdentifier()] = fmt.Sprintf("expectations did not match: %s", err.Error())
					}
				}
			}
			if !testConf.SimpleUI {
				_ = bar.Add(1)
			}
		default:
			output <- stats
			return
		}
	}
}

func (t *testDeviceInfoSNMPSim) generateRequest(requestType string) (request.Request, error) {
	switch requestType {
	case "identify":
		r := request.IdentifyRequest{}
		r.DeviceData = t.requestDeviceData
		return &r, nil
	case "read count-interfaces":
		r := request.ReadCountInterfacesRequest{}
		r.DeviceData = t.requestDeviceData
		return &r, nil
	case "check interface-metrics":
		r := request.CheckInterfaceMetricsRequest{}
		r.DeviceData = t.requestDeviceData
		r.PrintInterfaces = true
		return &r, nil
	case "check ups":
		r := request.CheckUPSRequest{}
		r.DeviceData = t.requestDeviceData
		return &r, nil
	default:
		return nil, errors.New("unknown requestType")
	}
}

func (t *testDeviceInfoSNMPSim) getIdentifier() string {
	return t.requestDeviceData.ConnectionData.SNMP.Communities[0]
}

func (t *testDevice) checkForRetryRequest() bool {
	return t.retries < testConf.Retries
}

func (e *DeviceTestDataExpectations) compareExpectations(response request.Response, requestType string) error {
	switch requestType {
	case "identify":
		if !cmp.Equal(e.Identify, response) {
			return errors.New("difference:" + cmp.Diff(e.Identify, response))
		}
	case "read count-interfaces":
		if !cmp.Equal(e.ReadCountInterfaces, response) {
			return errors.New("difference:\n" + cmp.Diff(e.ReadCountInterfaces, response))
		}
	case "check interface-metrics":
		if !cmp.Equal(e.CheckInterfaceMetrics, response, cmp.FilterValues(isNotEmpty, metricsTransformer()), metricsRawOutputFilter(), cmpopts.EquateEmpty(), performanceDataPointComparer()) {
			return errors.New("difference:\n" + cmp.Diff(e.CheckInterfaceMetrics, response, cmp.FilterValues(isNotEmpty, metricsTransformer()), metricsRawOutputFilter(), cmpopts.EquateEmpty(), performanceDataPointComparer()))
		}
	case "check ups":
		if !cmp.Equal(e.CheckUPS, response, metricsTransformer(), metricsRawOutputFilter()) {
			return errors.New("difference:\n" + cmp.Diff(e.CheckInterfaceMetrics, response, metricsTransformer(), metricsRawOutputFilter()))
		}
	default:
		return errors.New("unknown request type")
	}

	return nil
}

func waitForDevices(deviceChannel chan testDevice) error {
	device := <-deviceChannel
	err := WaitForSNMPSim(device.info.getIdentifier(), testConf.APIPort)
	deviceChannel <- device
	return err
}

func createTestDevices() (chan testDevice, error) {
	res, err := regexp.MatchString("^/.*", testConf.SNMPRecDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to match regex")
	}
	var recDir string

	if res {
		recDir = testConf.SNMPRecDir
	} else {
		_, currFilename, _, _ := runtime.Caller(0)
		recDir = filepath.Join(filepath.Join(path.Dir(currFilename), "testdata"), testConf.SNMPRecDir)
	}

	fileInfo, err := os.Stat(recDir)
	if err != nil {
		return nil, errors.Wrap(err, "error during Os.stat")
	}
	if !fileInfo.IsDir() {
		return nil, errors.New("only directories can be passed to this function")
	}

	testDevices, err := buildRecursiveTestDevices(recDir, "")
	if err != nil {
		return nil, errors.Wrap(err, "error during create recursive test devices")
	}

	if len(testDevices) == 0 {
		return nil, errors.New("no test devices found, ending test")
	}

	deviceChannel := make(chan testDevice, len(testDevices))
	for i := range testDevices {
		deviceChannel <- testDevices[i]
	}

	return deviceChannel, nil
}

func buildRecursiveTestDevices(dir, relativePath string) ([]testDevice, error) {
	fileDir, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrap(err, "error during read dir")
	}
	var subDirs []os.FileInfo
	files := make(map[string]string)
	for _, f := range fileDir {
		res, err := regexp.MatchString("^\\..*", f.Name())
		if err != nil {
			return nil, errors.Wrap(err, "failed to match regex")
		}
		if !res {
			if f.IsDir() {
				subDirs = append(subDirs, f)
			} else {
				files[f.Name()] = f.Name()
			}
		}
	}
	hasFiles := len(files) != 0
	hasSubDirs := len(subDirs) != 0
	if hasFiles && hasSubDirs {
		return nil, errors.New("test devices directory is faulty! there are files and directories in one directory")
	}
	var testDevices []testDevice
	if hasSubDirs {
		for _, f := range subDirs {
			devices, err := buildRecursiveTestDevices(filepath.Join(dir, f.Name()), filepath.Join(relativePath, f.Name()))
			if err != nil {
				return nil, err
			}
			testDevices = append(testDevices, devices...)
		}
	}
	if hasFiles {
		device, err := buildTestDeviceByFiles(files, dir, relativePath)
		if err != nil {
			return nil, err
		}
		testDevices = append(testDevices, device)
	}
	return testDevices, nil
}

func buildTestDeviceByFiles(files map[string]string, dir, relativePath string) (testDevice, error) {
	testDataFile, ok := files["test_data.json"]
	if !ok {
		return testDevice{}, errors.New("test_data.json is missing in " + dir)
	}

	var testData DeviceTestData
	contents, err := ioutil.ReadFile(filepath.Join(dir, testDataFile))
	if err != nil {
		return testDevice{}, errors.Wrap(err, "error during read file")
	}
	err = parser.ToStruct(contents, "json", &testData)
	if err != nil {
		return testDevice{}, errors.Wrap(err, "error during unmarshalling json file")
	}

	testDev := testDevice{
		expectations: testData.Expectations,
		requestTypes: testData.GetAvailableRequestTypes(),
	}

	atomic.AddInt32(&requestCounter, int32(len(testDev.requestTypes)))

	switch testData.Type {
	case "snmpsim":
		deviceInfo, err := buildTestDeviceSNMPSim(files, filepath.Join(relativePath, "public"))
		if err != nil {
			return testDevice{}, errors.Wrap(err, "failed to build snmpsim test device")
		}
		testDev.info = deviceInfo
	default:
		log.Error().Msg("unknown test device type " + testData.Type)
		return testDevice{}, nil
	}

	return testDev, nil
}

func buildTestDeviceSNMPSim(files map[string]string, snmpCommunity string) (*testDeviceInfoSNMPSim, error) {
	_, ok := files["public.snmprec"]
	if !ok {
		return nil, errors.New("snmprec file is missing for test device " + snmpCommunity)
	}

	ip := <-snmpSimIPs
	snmpSimIPs <- ip

	testDeviceInfo := testDeviceInfoSNMPSim{
		request.DeviceData{
			IPAddress: ip,
			ConnectionData: network.ConnectionData{
				SNMP: &network.SNMPConnectionData{
					Communities:              []string{snmpCommunity},
					Versions:                 []string{"2c"},
					Ports:                    []int{161},
					DiscoverParallelRequests: nil,
					DiscoverTimeout:          nil,
					DiscoverRetries:          nil,
				},
			},
		},
	}

	return &testDeviceInfo, nil
}

func isNotEmpty(x, y interface{}) bool {
	vx, vy := reflect.ValueOf(x), reflect.ValueOf(y)
	return !((x != nil && y != nil && vx.Type() == vy.Type()) &&
		(vx.Kind() == reflect.Slice || vx.Kind() == reflect.Map) &&
		(vx.Len() == 0 && vy.Len() == 0))
}

func metricsRawOutputFilter() cmp.Option {
	return cmp.FilterPath(func(path cmp.Path) bool {
		// Skip struct values which match the listed types
		ok, err := regexp.MatchString(".*RawOutput.*", path.GoString())
		if err != nil {
			panic(err)
		}
		return ok
	}, cmp.Ignore())
}

func metricsTransformer() cmp.Option {
	return cmp.Transformer("Sort", func(in []monitoringplugin.PerformanceDataPointInfo) []monitoringplugin.PerformanceDataPointInfo {
		out := make([]monitoringplugin.PerformanceDataPointInfo, len(in))
		copy(out, in)
		sort.Slice(out, func(i, j int) bool {
			return out[i].Label+out[i].Metric < out[j].Label+out[j].Metric
		})
		return out
	})
}

func performanceDataPointComparer() cmp.Option {
	return cmp.Comparer(func(x, y monitoringplugin.PerformanceDataPointInfo) bool {
		return fmt.Sprint(x) == fmt.Sprint(y)
	})
}
