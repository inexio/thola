package test

import (
	"context"
	"errors"
	"fmt"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/parser"
	"github.com/inexio/thola/core/request"
	"github.com/rs/zerolog/log"
	"os/exec"
	"strconv"
	"time"
)

// DeviceTestData represents a test_data.yaml file.
type DeviceTestData struct {
	Type         string                     `json:"type"`
	Expectations DeviceTestDataExpectations `json:"expectations"`
	Connection   network.ConnectionData     `json:"-"`
}

// DeviceTestDataExpectations represents the expectations part of an test data file.
type DeviceTestDataExpectations struct {
	Identify              *request.IdentifyResponse            `json:"identify" mapstructure:"identify"`
	ReadCountInterfaces   *request.ReadCountInterfacesResponse `json:"readCountInterfaces" mapstructure:"readCountInterfaces"`
	CheckInterfaceMetrics *request.CheckResponse               `json:"checkInterfaceMetrics" mapstructure:"checkInterfaceMetrics"`
	CheckUPS              *request.CheckResponse               `json:"checkUPS" mapstructure:"checkUPS"`
	CheckCPULoad          *request.CheckResponse               `json:"checkCPULoad" mapstructure:"checkCPULoad"`
	CheckMemoryUsage      *request.CheckResponse               `json:"checkMemoryUsage" mapstructure:"checkMemoryUsage"`
	CheckSBC              *request.CheckResponse               `json:"checkSBC" mapstructure:"checkSBC"`
}

// GetAvailableRequestTypes returns all available request types
func (d *DeviceTestData) GetAvailableRequestTypes() []string {
	var res []string

	if d.Expectations.Identify != nil {
		res = append(res, "identify")
	}

	if d.Expectations.ReadCountInterfaces != nil {
		res = append(res, "read count-interfaces")
	}

	if d.Expectations.CheckInterfaceMetrics != nil {
		res = append(res, "check interface-metrics")
	}

	if d.Expectations.CheckUPS != nil {
		res = append(res, "check ups")
	}

	if d.Expectations.CheckCPULoad != nil {
		res = append(res, "check cpu-load")
	}

	if d.Expectations.CheckMemoryUsage != nil {
		res = append(res, "check memory-usage")
	}

	if d.Expectations.CheckSBC != nil {
		res = append(res, "check sbc")
	}

	return res
}

// BuildupTestEnvironment build up the test environment
func BuildupTestEnvironment(dockerPath string) {
	dockerUp := exec.Command("docker-compose", "up", "--detach", "--build")
	dockerUp.Dir = dockerPath
	err := dockerUp.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run docker-compose up")
	}
}

// CleanupTestEnvironment cleans up the test environment
func CleanupTestEnvironment(dockerPath string) {
	dockerDown := exec.Command("docker-compose", "down")
	dockerDown.Dir = dockerPath
	err := dockerDown.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run docker-compose down")
	}
}

// ProcessRequest sends the request to the thola api and returns the response
func ProcessRequest(r request.Request, port int) (request.Response, error) {
	req, err := parser.Parse(r, "json")
	if err != nil {
		return nil, err
	}

	client, err := network.NewHTTPClient("http://localhost:" + strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	var requestEndpoint string
	var response request.Response

	switch r.(type) {
	case *request.IdentifyRequest:
		requestEndpoint = "identify"
		response = &request.IdentifyResponse{}
	case *request.ReadCountInterfacesRequest:
		requestEndpoint = "read/count-interfaces"
		response = &request.ReadCountInterfacesResponse{}
	case *request.CheckInterfaceMetricsRequest:
		requestEndpoint = "check/interface-metrics"
		response = &request.CheckResponse{}
	case *request.CheckUPSRequest:
		requestEndpoint = "check/ups"
		response = &request.CheckResponse{}
	case *request.CheckCPULoadRequest:
		requestEndpoint = "check/cpu-load"
		response = &request.CheckResponse{}
	case *request.CheckMemoryUsageRequest:
		requestEndpoint = "check/memory-usage"
		response = &request.CheckResponse{}
	case *request.CheckSBCRequest:
		requestEndpoint = "check/sbc"
		response = &request.CheckResponse{}
	default:
		return nil, errors.New("unknown request type")
	}

	res, err := client.Request(context.TODO(), "POST", requestEndpoint, string(req), nil, nil)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, fmt.Errorf("%v", string(res.Body()))
	}

	err = parser.ToStruct(res.Body(), "json", response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// WaitForSNMPSim waits until the snmp sim initialized all devices
func WaitForSNMPSim(community string, port int) error {
	r := request.IdentifyRequest{
		BaseRequest: request.BaseRequest{
			DeviceData: request.DeviceData{
				IPAddress: "172.20.0.8",
				ConnectionData: network.ConnectionData{
					SNMP: &network.SNMPConnectionData{
						Communities:              []string{community},
						Versions:                 []string{"2c"},
						Ports:                    []int{161},
						DiscoverParallelRequests: nil,
						DiscoverTimeout:          nil,
						DiscoverRetries:          nil,
					},
					HTTP: nil,
				},
			},
			Timeout: nil,
		},
	}

	timeout := make(chan bool)
	go func() {
		time.Sleep(180 * time.Second)
		timeout <- true
	}()

	for {
		select {
		case <-timeout:
			return errors.New("timeout exceeded")
		default:
			_, err := ProcessRequest(&r, port)
			if err == nil {
				return nil
			}
			time.Sleep(time.Second)
		}
	}
}
