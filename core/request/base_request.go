package request

import (
	"context"
	"github.com/inexio/thola/core/database"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/inexio/thola/core/utility"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/thoas/go-funk"
	"net"
	"strconv"
	"time"
)

// BaseRequest is a generic request that is processed by thola
type BaseRequest struct {
	// Date of the Device
	DeviceData DeviceData `json:"device_data" xml:"device_data"`

	// Timeout for the request (0 => no timeout)
	Timeout *int `json:"timeout" xml:"timeout"`
}

// DeviceData
//
// DeviceData includes all data that can be used to contact a device
//
// swagger:model
type DeviceData struct {
	// The IP of the device
	//
	// example: 203.0.113.195
	IPAddress string `json:"ip_address" xml:"ip_address"`
	// Data of the connection to the device
	ConnectionData network.ConnectionData `json:"connection_data" xml:"connection_data"`
}

// GetDeviceData returns the device data of the request
func (r *BaseRequest) GetDeviceData() *DeviceData {
	return &r.DeviceData
}

func (r *BaseRequest) validate(ctx context.Context) error {
	configData := getConfigConnectionData()

	if configData.SNMP == nil {
		configData.SNMP = &network.SNMPConnectionData{}
	}
	if configData.HTTP == nil {
		configData.HTTP = &network.HTTPConnectionData{}
	}

	db, err := database.GetDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get DB")
	}

	cacheData, err := db.GetConnectionData(ctx, r.DeviceData.IPAddress)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) {
			return err
		}
		cacheData = network.ConnectionData{}
	}
	if cacheData.SNMP == nil {
		cacheData.SNMP = &network.SNMPConnectionData{}
	}
	if cacheData.HTTP == nil {
		cacheData.HTTP = &network.HTTPConnectionData{}
	}

	mergedData := network.ConnectionData{
		SNMP: &network.SNMPConnectionData{
			Communities:              funk.UniqString(append(cacheData.SNMP.Communities, configData.SNMP.Communities...)),
			Versions:                 funk.UniqString(append(cacheData.SNMP.Versions, configData.SNMP.Versions...)),
			Ports:                    funk.UniqInt(append(cacheData.SNMP.Ports, configData.SNMP.Ports...)),
			DiscoverParallelRequests: configData.SNMP.DiscoverParallelRequests,
			DiscoverTimeout:          configData.SNMP.DiscoverTimeout,
			DiscoverRetries:          configData.SNMP.DiscoverRetries,
		},
		HTTP: &network.HTTPConnectionData{
			HTTPPorts:    funk.UniqInt(append(cacheData.HTTP.HTTPPorts, configData.HTTP.HTTPPorts...)),
			HTTPSPorts:   funk.UniqInt(append(cacheData.HTTP.HTTPSPorts, configData.HTTP.HTTPSPorts...)),
			AuthUsername: utility.IfThenElse(cacheData.HTTP.AuthUsername == nil, configData.HTTP.AuthUsername, cacheData.HTTP.AuthUsername).(*string),
			AuthPassword: utility.IfThenElse(cacheData.HTTP.AuthPassword == nil, configData.HTTP.AuthPassword, cacheData.HTTP.AuthPassword).(*string),
		},
	}

	if r.DeviceData.ConnectionData.SNMP == nil {
		r.DeviceData.ConnectionData.SNMP = mergedData.SNMP
	}

	if net.ParseIP(r.DeviceData.IPAddress) == nil {
		return errors.New("IP formatted wrong")
	}

	if len(r.DeviceData.ConnectionData.SNMP.Communities) == 0 {
		r.DeviceData.ConnectionData.SNMP.Communities = mergedData.SNMP.Communities
	}

	if len(r.DeviceData.ConnectionData.SNMP.Versions) == 0 {
		r.DeviceData.ConnectionData.SNMP.Versions = mergedData.SNMP.Versions
	}
	for _, version := range r.DeviceData.ConnectionData.SNMP.Versions {
		if !(version == "1" || version == "2c" || version == "3") {
			return errors.New("invalid SNMP version")
		}
	}

	if len(r.DeviceData.ConnectionData.SNMP.Ports) == 0 {
		r.DeviceData.ConnectionData.SNMP.Ports = mergedData.SNMP.Ports
	}
	for _, port := range r.DeviceData.ConnectionData.SNMP.Ports {
		if port <= 0 {
			return errors.New("invalid SNMP port")
		}
	}

	if r.DeviceData.ConnectionData.HTTP == nil {
		r.DeviceData.ConnectionData.HTTP = mergedData.HTTP
	}

	if len(r.DeviceData.ConnectionData.HTTP.HTTPPorts) == 0 {
		r.DeviceData.ConnectionData.HTTP.HTTPPorts = mergedData.HTTP.HTTPPorts
	}
	for _, port := range r.DeviceData.ConnectionData.HTTP.HTTPPorts {
		if port <= 0 {
			return errors.New("invalid HTTP port")
		}
	}

	if len(r.DeviceData.ConnectionData.HTTP.HTTPPorts) == 0 {
		r.DeviceData.ConnectionData.HTTP.HTTPSPorts = mergedData.HTTP.HTTPSPorts
	}
	for _, port := range r.DeviceData.ConnectionData.HTTP.HTTPSPorts {
		if port <= 0 {
			return errors.New("invalid HTTPS port")
		}
	}

	if r.DeviceData.ConnectionData.HTTP.AuthUsername == nil {
		r.DeviceData.ConnectionData.HTTP.AuthUsername = mergedData.HTTP.AuthUsername
	}

	if r.DeviceData.ConnectionData.HTTP.AuthPassword == nil {
		r.DeviceData.ConnectionData.HTTP.AuthPassword = mergedData.HTTP.AuthPassword
	}

	if r.DeviceData.ConnectionData.SNMP.DiscoverParallelRequests == nil {
		r.DeviceData.ConnectionData.SNMP.DiscoverParallelRequests = mergedData.SNMP.DiscoverParallelRequests
	}

	if r.DeviceData.ConnectionData.SNMP.DiscoverTimeout == nil {
		r.DeviceData.ConnectionData.SNMP.DiscoverTimeout = mergedData.SNMP.DiscoverTimeout
	}

	if r.DeviceData.ConnectionData.SNMP.DiscoverRetries == nil {
		r.DeviceData.ConnectionData.SNMP.DiscoverRetries = mergedData.SNMP.DiscoverRetries
	}

	if *r.DeviceData.ConnectionData.SNMP.DiscoverParallelRequests <= 0 || *r.DeviceData.ConnectionData.SNMP.DiscoverTimeout <= 0 {
		return errors.New("invalid snmp connection preferences")
	}

	if r.Timeout == nil {
		timeout := viper.GetInt("request.timeout")
		r.Timeout = &timeout
	}

	return nil
}

func (r *BaseRequest) getTimeout() *int {
	return r.Timeout
}

func (r *BaseRequest) handlePreProcessError(err error) (Response, error) {
	return nil, err
}

func getConfigConnectionData() network.ConnectionData {
	parallelRequests := viper.GetInt("device.snmp-discover-par-requests")
	timeout := viper.GetInt("device.snmp-discover-timeout")
	retries := viper.GetInt("device.snmp-discover-retries")
	authUsername := viper.GetString("device.http-username")
	authPassword := viper.GetString("device.http-password")
	return network.ConnectionData{
		SNMP: &network.SNMPConnectionData{
			Communities:              viper.GetStringSlice("device.snmp-communities"),
			Versions:                 viper.GetStringSlice("device.snmp-versions"),
			Ports:                    viper.GetIntSlice("device.snmp-ports"),
			DiscoverParallelRequests: &parallelRequests,
			DiscoverTimeout:          &timeout,
			DiscoverRetries:          &retries,
		},
		HTTP: &network.HTTPConnectionData{
			HTTPPorts:    viper.GetIntSlice("device.http-ports"),
			HTTPSPorts:   viper.GetIntSlice("device.https-ports"),
			AuthUsername: &authUsername,
			AuthPassword: &authPassword,
		},
	}
}

func (r *BaseRequest) setupConnection(ctx context.Context) (*network.RequestDeviceConnection, error) {
	var con network.RequestDeviceConnection
	con.RawConnectionData = r.DeviceData.ConnectionData
	createdData := false
	if r.DeviceData.ConnectionData.SNMP != nil {
		snmpCon, err := r.setupSNMPConnection(ctx)
		if err != nil {
			log.Ctx(ctx).Trace().Err(err).Msg("failed to setup snmp connection data")
		} else {
			log.Ctx(ctx).Trace().Err(err).Msg("successfully setup snmp connection data")
			con.SNMP = snmpCon
			createdData = true
		}
	}

	if r.DeviceData.ConnectionData.HTTP != nil && (len(r.DeviceData.ConnectionData.HTTP.HTTPSPorts) != 0 || len(r.DeviceData.ConnectionData.HTTP.HTTPPorts) != 0) {
		httpCon, err := r.setupHTTPConnection()
		if err != nil {
			log.Ctx(ctx).Trace().Err(err).Msg("failed to setup http connection data")
		} else {
			log.Ctx(ctx).Trace().Err(err).Msg("successfully setup http connection data")
			con.HTTP = httpCon
			createdData = true
		}
	}
	if !createdData {
		return nil, errors.New("cannot create any connection to the device")
	}
	return &con, nil
}

func (r *BaseRequest) setupSNMPConnection(ctx context.Context) (*network.RequestDeviceConnectionSNMP, error) {
	snmpClient, err := network.NewSNMPClientByConnectionData(ctx, r.DeviceData.IPAddress, r.DeviceData.ConnectionData.SNMP)
	if err != nil {
		return nil, errors.Wrap(err, "error during NewSNMPClientByConnectionData")
	}

	var con network.RequestDeviceConnectionSNMP
	con.SnmpClient = snmpClient

	return &con, nil
}

func (r *BaseRequest) setupHTTPConnection() (*network.RequestDeviceConnectionHTTP, error) {
	if r.DeviceData.ConnectionData.HTTP == nil {
		return nil, errors.New("no http connection data available")
	}

	var httpClient *network.HTTPClient
	var err error
	for _, port := range r.DeviceData.ConnectionData.HTTP.HTTPSPorts {
		httpClient, err = network.NewHTTPClient("https://" + r.DeviceData.IPAddress + ":" + strconv.Itoa(port))
		if err == nil {
			break
		}
	}
	if r.DeviceData.ConnectionData.HTTP.HTTPSPorts == nil || err != nil {
		for _, port := range r.DeviceData.ConnectionData.HTTP.HTTPPorts {
			httpClient, err = network.NewHTTPClient("http://" + r.DeviceData.IPAddress + ":" + strconv.Itoa(port))
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to create http client")
	}
	if r.DeviceData.ConnectionData.HTTP.AuthPassword != nil && r.DeviceData.ConnectionData.HTTP.AuthUsername != nil {
		err = httpClient.SetUsernameAndPassword(*r.DeviceData.ConnectionData.HTTP.AuthUsername, *r.DeviceData.ConnectionData.HTTP.AuthPassword)
		if err != nil {
			return nil, errors.Wrap(err, "error during set username and password")
		}
	}
	httpClient.SetTimeout(15 * time.Second)
	con := &network.RequestDeviceConnectionHTTP{}
	con.HTTPClient = httpClient
	con.ConnectionData = r.DeviceData.ConnectionData.HTTP
	return con, nil
}

// BaseResponse
//
// BaseResponse defines attributes every response has.
//
// swagger:model
type BaseResponse struct {
}

// GetExitCode returns the exit code of the response.
func (b *BaseResponse) GetExitCode() int {
	return 0
}
