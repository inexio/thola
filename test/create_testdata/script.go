package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/parser"
	"github.com/inexio/thola/internal/request"
	"github.com/inexio/thola/test"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var (
	port       int
	snmpRecDir string
	ignore     = flag.String("ignore", "", "ignore snmprecs whose filepath matches this regex")
)

func init() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	flag.Parse()

	if portString := os.Getenv("THOLA_TEST_APIPORT"); portString != "" {
		p, err := strconv.Atoi(portString)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse port")
		}
		port = p
	} else {
		port = 8237
	}

	if dir := os.Getenv("THOLA_TEST_SNMPRECDIR"); dir != "" {
		snmpRecDir = dir
	} else {
		_, currFilename, _, _ := runtime.Caller(0)
		snmpRecDir = filepath.Join(filepath.Dir(filepath.Dir(currFilename)), "testdata/devices")
	}

	if ignore != nil && *ignore == "" {
		if ignoreEnv := os.Getenv("THOLA_TEST_IGNORE"); ignoreEnv != "" {
			ignore = &ignoreEnv
		}
	}
}

func main() {
	_, currFilename, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(filepath.Dir(filepath.Dir(currFilename)), "testdata")

	test.BuildupTestEnvironment(testdataDir)
	defer test.CleanupTestEnvironment(testdataDir)

	fileInfo, err := os.Stat(snmpRecDir)
	if err != nil {
		log.Error().Err(err).Msg("error during Os.stat")
		return
	}
	if !fileInfo.IsDir() {
		log.Error().Err(err).Msg("snmp rec path must be a directory")
		return
	}

	testDevices, err := buildTestDevicesRecursive("")
	if err != nil {
		log.Error().Err(err).Msg("error while building test devices")
		return
	}

	if len(testDevices) == 0 {
		log.Error().Err(err).Msg("no test devices found")
		return
	}

	err = test.WaitForSNMPSim(testDevices[0], port)
	if err != nil {
		log.Error().Err(err).Msg("error while waiting for snmp sim")
		return
	}

	err = createTestdata(testDevices)
	if err != nil {
		log.Error().AnErr("Error", err).Msg("Can't create testdata")
		return
	}

	fmt.Println("Generating testdata was successful")
}

func buildTestDevicesRecursive(relativePath string) ([]string, error) {
	fileDir, err := ioutil.ReadDir(filepath.Join(snmpRecDir, relativePath))
	if err != nil {
		return nil, errors.Wrap(err, "error during read dir")
	}

	var recordings []string
	var subDirs []string

	regex, err := regexp.Compile(`^\..*`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build regex")
	}

	for _, f := range fileDir {
		if !regex.MatchString(f.Name()) {
			if f.IsDir() {
				subDirs = append(subDirs, f.Name())
			} else if strings.HasSuffix(f.Name(), ".snmprec") {
				recordings = append(recordings, strings.TrimSuffix(f.Name(), ".snmprec"))
			}
		}
	}

	var testDevices []string
	for _, d := range subDirs {
		devices, err := buildTestDevicesRecursive(filepath.Join(relativePath, d))
		if err != nil {
			return nil, err
		}
		testDevices = append(testDevices, devices...)
	}

	for _, rec := range recordings {
		community := filepath.Join(relativePath, rec)

		if ignore != nil && *ignore != "" {
			if ok, err := regexp.MatchString(*ignore, community); err == nil && !ok {
				testDevices = append(testDevices, community)
			}
		} else {
			testDevices = append(testDevices, community)
		}
	}

	return testDevices, nil
}

func createTestdata(testDevices []string) error {
	for _, device := range testDevices {
		deviceTestData, err := getDeviceTestData(device)
		if err != nil {
			return errors.Wrap(err, "get device testdata failed for device "+device)
		}

		deviceDataJson, err := json.MarshalIndent(deviceTestData, "", "\t")
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(filepath.Join(snmpRecDir, device+".testdata"), deviceDataJson, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func getDeviceTestData(device string) (test.DeviceTestData, error) {
	connectionData := network.ConnectionData{
		SNMP: &network.SNMPConnectionData{
			Communities: []string{device},
			Versions:    []string{"2c"},
			Ports:       []int{161},
		},
	}

	baseRequest := request.BaseRequest{
		DeviceData: request.DeviceData{
			IPAddress:      "172.20.0.8",
			ConnectionData: connectionData,
		},
	}

	var identifyResponse *request.IdentifyResponse
	res, err := test.ProcessRequest(
		&request.IdentifyRequest{
			BaseRequest: baseRequest,
		},
		port,
	)
	if err != nil {
		log.Info().Err(err).Msg("identify for device " + device + " failed")
	} else {
		identifyResponse = res.(*request.IdentifyResponse)
	}

	var readCountInterfacesResponse *request.ReadCountInterfacesResponse
	res, err = test.ProcessRequest(
		&request.ReadCountInterfacesRequest{
			ReadRequest: request.ReadRequest{
				BaseRequest: baseRequest,
			},
		},
		port,
	)
	if err != nil {
		log.Info().Err(err).Msg("read count interfaces for device " + device + " failed")
	} else {
		readCountInterfacesResponse = res.(*request.ReadCountInterfacesResponse)
	}

	var checkInterfaceMetricsResponse *request.CheckResponse
	res, err = test.ProcessRequest(
		&request.CheckInterfaceMetricsRequest{
			CheckDeviceRequest: request.CheckDeviceRequest{
				BaseRequest:  baseRequest,
				CheckRequest: request.CheckRequest{},
			},
			PrintInterfaces: true,
		},
		port,
	)
	if err != nil {
		log.Info().Err(err).Msg("check interface metrics for device " + device + " failed")
	} else if res.GetExitCode() == 3 {
		errString, err := parser.Parse(res, "")
		if err != nil {
			return test.DeviceTestData{}, errors.New("failed to parse error message")
		}
		log.Info().Err(errors.New(string(errString))).Msg("check interface metrics for device " + device + " failed")
	} else {
		checkInterfaceMetricsResponse = res.(*request.CheckResponse)
		checkInterfaceMetricsResponse.RawOutput = ""
	}

	var checkUPSResponse *request.CheckResponse
	res, err = test.ProcessRequest(
		&request.CheckUPSRequest{
			CheckDeviceRequest: request.CheckDeviceRequest{
				BaseRequest:  baseRequest,
				CheckRequest: request.CheckRequest{},
			},
		},
		port,
	)
	if err != nil {
		log.Info().Err(err).Msg("check ups for device " + device + " failed")
	} else if res.GetExitCode() == 3 {
		errString, err := parser.Parse(res, "")
		if err != nil {
			return test.DeviceTestData{}, errors.New("failed to parse error message")
		}
		log.Info().Err(errors.New(string(errString))).Msg("check ups for device " + device + " failed")
	} else {
		checkUPSResponse = res.(*request.CheckResponse)
		checkUPSResponse.RawOutput = ""
	}

	var checkCPULoadResponse *request.CheckResponse
	res, err = test.ProcessRequest(
		&request.CheckCPULoadRequest{
			CheckDeviceRequest: request.CheckDeviceRequest{
				BaseRequest:  baseRequest,
				CheckRequest: request.CheckRequest{},
			},
		},
		port,
	)
	if err != nil {
		log.Info().Err(err).Msg("check cpu load for device " + device + " failed")
	} else if res.GetExitCode() == 3 {
		errString, err := parser.Parse(res, "")
		if err != nil {
			return test.DeviceTestData{}, errors.New("failed to parse error message")
		}
		log.Info().Err(errors.New(string(errString))).Msg("check cpu load for device " + device + " failed")
	} else {
		checkCPULoadResponse = res.(*request.CheckResponse)
		checkCPULoadResponse.RawOutput = ""
	}

	var checkMemoryUsageResponse *request.CheckResponse
	res, err = test.ProcessRequest(
		&request.CheckMemoryUsageRequest{
			CheckDeviceRequest: request.CheckDeviceRequest{
				BaseRequest:  baseRequest,
				CheckRequest: request.CheckRequest{},
			},
		},
		port,
	)
	if err != nil {
		log.Info().Err(err).Msg("check memory usage for device " + device + " failed")
	} else if res.GetExitCode() == 3 {
		errString, err := parser.Parse(res, "")
		if err != nil {
			return test.DeviceTestData{}, errors.New("failed to parse error message")
		}
		log.Info().Err(errors.New(string(errString))).Msg("check memory usage for device " + device + " failed")
	} else {
		checkMemoryUsageResponse = res.(*request.CheckResponse)
		checkMemoryUsageResponse.RawOutput = ""
	}

	var checkDiskResponse *request.CheckResponse
	res, err = test.ProcessRequest(
		&request.CheckDiskRequest{
			CheckDeviceRequest: request.CheckDeviceRequest{
				BaseRequest:  baseRequest,
				CheckRequest: request.CheckRequest{},
			},
		},
		port,
	)
	if err != nil {
		log.Info().Err(err).Msg("check disk for device " + device + " failed")
	} else if res.GetExitCode() == 3 {
		errString, err := parser.Parse(res, "")
		if err != nil {
			return test.DeviceTestData{}, errors.New("failed to parse error message")
		}
		log.Info().Err(errors.New(string(errString))).Msg("check disk for device " + device + " failed")
	} else {
		checkDiskResponse = res.(*request.CheckResponse)
		checkDiskResponse.RawOutput = ""
	}

	var checkServerResponse *request.CheckResponse
	res, err = test.ProcessRequest(
		&request.CheckServerRequest{
			CheckDeviceRequest: request.CheckDeviceRequest{
				BaseRequest:  baseRequest,
				CheckRequest: request.CheckRequest{},
			},
		},
		port,
	)
	if err != nil {
		log.Info().Err(err).Msg("check server for device " + device + " failed")
	} else if res.GetExitCode() == 3 {
		errString, err := parser.Parse(res, "")
		if err != nil {
			return test.DeviceTestData{}, errors.New("failed to parse error message")
		}
		log.Info().Err(errors.New(string(errString))).Msg("check server for device " + device + " failed")
	} else {
		checkServerResponse = res.(*request.CheckResponse)
		checkServerResponse.RawOutput = ""
	}

	var checkSBCResponse *request.CheckResponse
	res, err = test.ProcessRequest(
		&request.CheckSBCRequest{
			CheckDeviceRequest: request.CheckDeviceRequest{
				BaseRequest:  baseRequest,
				CheckRequest: request.CheckRequest{},
			},
		},
		port,
	)
	if err != nil {
		log.Info().Err(err).Msg("check sbc for device " + device + " failed")
	} else if res.GetExitCode() == 3 {
		errString, err := parser.Parse(res, "")
		if err != nil {
			return test.DeviceTestData{}, errors.New("failed to parse error message")
		}
		log.Info().Err(errors.New(string(errString))).Msg("check sbc for device " + device + " failed")
	} else {
		checkSBCResponse = res.(*request.CheckResponse)
		checkSBCResponse.RawOutput = ""
	}

	return test.DeviceTestData{
		Type: "snmpsim",
		Expectations: test.DeviceTestDataExpectations{
			Identify:              identifyResponse,
			ReadCountInterfaces:   readCountInterfacesResponse,
			CheckInterfaceMetrics: checkInterfaceMetricsResponse,
			CheckUPS:              checkUPSResponse,
			CheckCPULoad:          checkCPULoadResponse,
			CheckMemoryUsage:      checkMemoryUsageResponse,
			CheckDisk:             checkDiskResponse,
			CheckServer:           checkServerResponse,
			CheckSBC:              checkSBCResponse,
		},
		Connection: connectionData,
	}, nil
}
