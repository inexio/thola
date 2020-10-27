package main

import (
	"encoding/json"
	"fmt"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/parser"
	"github.com/inexio/thola/core/request"
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
)

var (
	port       int
	snmpRecDir string
)

func init() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

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

	testDevices, err := buildRecursiveTestDevices(snmpRecDir, "")
	if err != nil {
		log.Error().Err(err).Msg("error while building test devices")
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

func buildRecursiveTestDevices(dir, relativePath string) ([]string, error) {
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
	var testDevices []string
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
		_, ok := files["public.snmprec"]
		if !ok {
			return nil, errors.New("snmprec file is missing for test device " + relativePath)
		}
		testDevices = append(testDevices, filepath.Join(relativePath, "public"))
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

		err = ioutil.WriteFile(filepath.Join(snmpRecDir, filepath.Dir(device), "test_data.json"), deviceDataJson, 0644)
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
	}

	return test.DeviceTestData{
		Type: "snmpsim",
		Expectations: test.DeviceTestDataExpectations{
			Identify:              identifyResponse,
			ReadCountInterfaces:   readCountInterfacesResponse,
			CheckInterfaceMetrics: checkInterfaceMetricsResponse,
			CheckUPS:              checkUPSResponse,
		},
		Connection: connectionData,
	}, nil
}
