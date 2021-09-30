package network

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/gosnmp/gosnmp"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/inexio/thola/internal/utility"
	"github.com/inexio/thola/internal/value"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/encoding/charmap"
	"regexp"
	"strings"
	"time"
	"unicode"
)

//go:generate go run github.com/vektra/mockery/v2 --name=SNMPClient --inpackage

// SNMPClient is used to communicate via snmp.
type SNMPClient interface {
	Disconnect() error

	SNMPGet(ctx context.Context, oid ...OID) ([]SNMPResponse, error)
	SNMPWalk(ctx context.Context, oid OID) ([]SNMPResponse, error)

	UseCache(b bool)
	HasSuccessfulCachedRequest() bool

	GetCommunity() string
	SetCommunity(community string)
	GetPort() int
	GetVersion() string
	GetMaxRepetitions() uint32

	SetMaxRepetitions(maxRepetitions uint32)
	SetMaxOIDs(maxOIDs int) error

	GetV3Level() *string
	GetV3ContextName() *string
	GetV3User() *string
	GetV3AuthKey() *string
	GetV3AuthProto() *string
	GetV3PrivKey() *string
	GetV3PrivProto() *string
}

type snmpClient struct {
	client    *gosnmp.GoSNMP
	useCache  bool
	getCache  requestCache
	walkCache requestCache
}

type snmpClientCreation struct {
	client  SNMPClient
	version string
	err     error
}

type snmpClientCreationData struct {
	ipAddress   string
	snmpVersion string
	community   string
	port        int
	timeout     int
	retries     int
	v3Data      SNMPv3ConnectionData
}

// NewSNMPClientByConnectionData tries to create a new snmp client by SNMPConnectionData and returns it.
func NewSNMPClientByConnectionData(ctx context.Context, ipAddress string, data *SNMPConnectionData) (SNMPClient, error) {
	if data == nil {
		return nil, errors.New("snmp connection data is nil")
	}

	if data.DiscoverParallelRequests == nil || data.DiscoverTimeout == nil || data.DiscoverRetries == nil || *data.DiscoverParallelRequests <= 0 || *data.DiscoverTimeout <= 0 {
		return nil, tholaerr.NewPreConditionError("invalid connection preferences")
	}

	var v2cAvailable, v3Available bool

	// validate snmp version
	for _, v := range data.Versions {
		version, err := getGoSNMPVersion(v)
		if err != nil {
			return nil, err
		}
		if version == gosnmp.Version3 {
			v3Available = true
		}
		if version == gosnmp.Version2c {
			v2cAvailable = true
		}
	}

	amount := len(data.Ports) * len(data.Versions) * len(data.Communities)
	in := make(chan snmpClientCreationData, amount)
	out := make(chan snmpClientCreation, amount)
	amount = 0

	for _, port := range data.Ports {
		for _, version := range data.Versions {
			// v3 has no community set
			if version == "3" {
				in <- snmpClientCreationData{
					ipAddress:   ipAddress,
					snmpVersion: version,
					port:        port,
					timeout:     *data.DiscoverTimeout,
					retries:     *data.DiscoverRetries,
					v3Data:      data.V3Data,
				}
				amount++
			} else {
				for _, community := range data.Communities {
					in <- snmpClientCreationData{
						ipAddress:   ipAddress,
						snmpVersion: version,
						community:   community,
						port:        port,
						timeout:     *data.DiscoverTimeout,
						retries:     *data.DiscoverRetries,
					}
					amount++
				}
			}
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i := 0; i < *data.DiscoverParallelRequests; i++ {
		go createNewSNMPClientConcurrent(ctx, in, out)
	}

	var criticalError error
	var successfulClient SNMPClient

	for i := 0; i < amount; i++ {
		res := <-out
		if res.err != nil {
			if !tholaerr.IsNetworkError(res.err) {
				s := "non network error occurred during NewSNMPClient"
				log.Ctx(ctx).Error().Err(res.err).Msg(s)
				if criticalError == nil {
					criticalError = errors.Wrap(res.err, s)
				}
			}
			continue
		}
		if res.version == "3" {
			successfulClient = res.client
			break
		}
		if res.version == "2c" {
			if !v3Available {
				successfulClient = res.client
				break
			}
			if successfulClient == nil || successfulClient.GetVersion() == "1" {
				successfulClient = res.client
			}
		}
		if res.version == "1" {
			if !v3Available && !v2cAvailable {
				successfulClient = res.client
				break
			}
			if successfulClient == nil {
				successfulClient = res.client
			}
		}
	}
	if successfulClient != nil {
		if data.MaxRepetitions != nil {
			log.Ctx(ctx).Debug().Msg("set snmp max repetitions of connection data")
			successfulClient.SetMaxRepetitions(*data.MaxRepetitions)
		}
		return successfulClient, nil
	}
	if criticalError != nil {
		return nil, criticalError
	}
	return nil, tholaerr.NewSNMPError("cannot connect with any of the given connection data")
}

func createNewSNMPClientConcurrent(ctx context.Context, in chan snmpClientCreationData, out chan snmpClientCreation) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			select {
			case data := <-in:
				var client SNMPClient
				var err error
				if data.snmpVersion == "3" {
					client, err = NewSNMPv3Client(ctx, data.ipAddress, data.port, data.timeout, data.retries, data.v3Data)
				} else {
					client, err = NewSNMPClient(ctx, data.ipAddress, data.snmpVersion, data.community, data.port, data.timeout, data.retries)
				}
				out <- snmpClientCreation{client, data.snmpVersion, err}
			default:
				return
			}
		}
	}
}

// NewSNMPClient creates a new SNMP Client
func NewSNMPClient(ctx context.Context, ipAddress, snmpVersion, community string, port, timeout, retries int) (SNMPClient, error) {
	version, err := getGoSNMPVersion(snmpVersion)
	if err != nil {
		return nil, err
	}

	client := &gosnmp.GoSNMP{
		Context:   ctx,
		Target:    ipAddress,
		Port:      uint16(port),
		Transport: "udp",
		Community: community,
		Version:   version,
		Timeout:   time.Duration(timeout) * time.Second,
		MaxOids:   utility.IfThenElseInt(version == gosnmp.Version1, 1, gosnmp.MaxOids),
		Retries:   retries,
	}

	return newSNMPClientTestConnection(client)
}

// NewSNMPv3Client creates a new SNMP v3 Client.
func NewSNMPv3Client(ctx context.Context, ipAddress string, port, timeout, retries int, v3Data SNMPv3ConnectionData) (SNMPClient, error) {
	client := &gosnmp.GoSNMP{
		Context:       ctx,
		Target:        ipAddress,
		Port:          uint16(port),
		Transport:     "udp",
		Version:       gosnmp.Version3,
		Timeout:       time.Duration(timeout) * time.Second,
		MaxOids:       gosnmp.MaxOids,
		Retries:       retries,
		SecurityModel: gosnmp.UserSecurityModel,
	}

	if v3Data.ContextName != nil {
		client.ContextName = *v3Data.ContextName
	}

	switch *v3Data.Level {
	case "noAuthNoPriv":
		client.MsgFlags = gosnmp.NoAuthNoPriv
		client.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName: *v3Data.User,
		}
	case "authNoPriv":
		authProtocol, err := getGoSNMPV3AuthProtocol(*v3Data.AuthProtocol)
		if err != nil {
			return nil, err
		}

		client.MsgFlags = gosnmp.AuthNoPriv
		client.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 *v3Data.User,
			AuthenticationProtocol:   authProtocol,
			AuthenticationPassphrase: *v3Data.AuthKey,
		}
	case "authPriv":
		authProtocol, err := getGoSNMPV3AuthProtocol(*v3Data.AuthProtocol)
		if err != nil {
			return nil, err
		}

		privProtocol, err := getGoSNMPV3PrivProtocol(*v3Data.PrivProtocol)
		if err != nil {
			return nil, err
		}

		client.MsgFlags = gosnmp.AuthPriv
		client.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 *v3Data.User,
			AuthenticationProtocol:   authProtocol,
			AuthenticationPassphrase: *v3Data.AuthKey,
			PrivacyProtocol:          privProtocol,
			PrivacyPassphrase:        *v3Data.PrivKey,
		}
	}

	return newSNMPClientTestConnection(client)
}

func newSNMPClientTestConnection(client *gosnmp.GoSNMP) (*snmpClient, error) {
	err := client.ConnectIPv4()
	if err != nil {
		return nil, errors.Wrap(err, "connect ip v4 failed")
	}

	oids := []string{".0.0"}
	_, err = client.GetNext(oids)
	if err != nil {
		return nil, tholaerr.NewSNMPError(err.Error())
	}

	client.Retries = gosnmp.Default.Retries
	client.Timeout = gosnmp.Default.Timeout

	return &snmpClient{
		client:    client,
		useCache:  true,
		getCache:  newRequestCache(),
		walkCache: newRequestCache(),
	}, nil
}

func getGoSNMPVersion(version string) (gosnmp.SnmpVersion, error) {
	switch version {
	case "1":
		return gosnmp.Version1, nil
	case "2c":
		return gosnmp.Version2c, nil
	case "3":
		return gosnmp.Version3, nil
	default:
		return 0, fmt.Errorf("invalid snmp version '%s', only '1', '2c' and '3' are possible", version)
	}
}

func ValidateSNMPv3AuthProtocol(protocol string) error {
	_, err := getGoSNMPV3AuthProtocol(protocol)
	return err
}

func getGoSNMPV3AuthProtocol(protocol string) (gosnmp.SnmpV3AuthProtocol, error) {
	switch protocol {
	case "noAuth":
		return gosnmp.NoAuth, nil
	case "md5", "MD5":
		return gosnmp.MD5, nil
	case "sha", "SHA":
		return gosnmp.SHA, nil
	case "sha224", "SHA224":
		return gosnmp.SHA224, nil
	case "sha256", "SHA256":
		return gosnmp.SHA256, nil
	case "sha384", "SHA384":
		return gosnmp.SHA384, nil
	case "sha512", "SHA512":
		return gosnmp.SHA512, nil
	default:
		return 0, fmt.Errorf("invalid authentication protocol '%s'", protocol)
	}
}

func ValidateSNMPv3PrivProtocol(protocol string) error {
	_, err := getGoSNMPV3PrivProtocol(protocol)
	return err
}

func getGoSNMPV3PrivProtocol(protocol string) (gosnmp.SnmpV3PrivProtocol, error) {
	switch protocol {
	case "noAuth":
		return gosnmp.NoPriv, nil
	case "des", "DES":
		return gosnmp.DES, nil
	case "aes", "AES":
		return gosnmp.AES, nil
	case "aes192", "AES192":
		return gosnmp.AES192, nil
	case "aes256", "AES256":
		return gosnmp.AES256, nil
	case "aes192c", "AES192C":
		return gosnmp.AES192C, nil
	case "aes256c", "AES256C":
		return gosnmp.AES256C, nil
	default:
		return 0, fmt.Errorf("invalid privacy protocol '%s'", protocol)
	}
}

// SNMPGet sends one or more simple snmpget requests to the target host and returns the result.
func (s *snmpClient) SNMPGet(ctx context.Context, oid ...OID) ([]SNMPResponse, error) {
	var snmpResponses []SNMPResponse
	var successful bool
	var reqOIDs []OID

	if s.useCache {
		for _, o := range oid {
			cacheEntry, err := s.getCache.get(o.String())
			if err != nil {
				reqOIDs = append(reqOIDs, o)
			} else {
				res, ok := cacheEntry.res.(SNMPResponse)
				if !ok {
					return nil, errors.New("cached SNMP Get result is not a SNMP response")
				}
				log.Ctx(ctx).Trace().Str("network_request", "snmpget").Str("oid", o.String()).Msg("used cached SNMP Get result")
				snmpResponses = append(snmpResponses, res)
				if res.WasSuccessful() {
					successful = true
				}
			}
		}
	} else {
		reqOIDs = oid
	}

	var batch []OID
	s.client.Context = ctx

	for len(reqOIDs) > 0 {
		var batchSize int
		if s.client.MaxOids >= len(reqOIDs) {
			batchSize = len(reqOIDs)
		} else {
			batchSize = s.client.MaxOids
		}
		batch, reqOIDs = reqOIDs[:batchSize], reqOIDs[batchSize:]

		var batchString []string
		for _, elem := range batch {
			batchString = append(batchString, elem.String())
		}
		response, err := s.client.Get(batchString)
		if err != nil {
			log.Ctx(ctx).Trace().Str("network_request", "snmpget").Strs("oid", batchString).Err(err).Msg("SNMP Get failed")
			return nil, errors.Wrap(err, "error during snmpget")
		}

		for _, currentResponse := range response.Variables {
			snmpResponse := NewSNMPResponse(OID(currentResponse.Name), currentResponse.Type, currentResponse.Value)

			if snmpResponse.WasSuccessful() {
				log.Ctx(ctx).Trace().Str("network_request", "snmpget").Str("oid", snmpResponse.oid.String()).Msg("SNMP Get was successful")
				successful = true
				if s.useCache {
					s.getCache.add(snmpResponse.oid.String(), snmpResponse, nil)
				}
			} else {
				log.Ctx(ctx).Trace().Str("network_request", "snmpget").Str("oid", snmpResponse.oid.String()).Msg("No Such Object available on this agent at this OID")
				if s.useCache {
					s.getCache.add(snmpResponse.oid.String(), snmpResponse, errors.New("SNMP Request failed"))
				}
			}

			snmpResponses = append(snmpResponses, snmpResponse)
		}
	}

	if !successful {
		return nil, tholaerr.NewNotFoundError("No Such Object available on this agent at this OID")
	}

	return snmpResponses, nil
}

// SNMPWalk sends a snmpwalk request to the specified oid.
func (s *snmpClient) SNMPWalk(ctx context.Context, oid OID) ([]SNMPResponse, error) {
	if s.useCache {
		cacheEntry, err := s.walkCache.get(oid.String())
		if err == nil {
			log.Ctx(ctx).Trace().Str("network_request", "snmpwalk").Str("oid", oid.String()).Msg("used cached snmp walk result")
			if cacheEntry.err != nil {
				return nil, cacheEntry.err
			}
			res, ok := cacheEntry.res.([]SNMPResponse)
			if !ok {
				return nil, errors.New("cached snmp result is not a snmp response")
			}
			return res, nil
		}
	}

	s.client.Context = ctx

	var response []gosnmp.SnmpPDU
	var err error
	if s.client.Version != gosnmp.Version1 {
		response, err = s.client.BulkWalkAll(oid.String())
		if err != nil {
			log.Ctx(ctx).Trace().Str("network_request", "snmpwalk").Str("oid", oid.String()).Err(err).Msg("snmp bulk walk failed")
		}
	}
	if s.client.Version == gosnmp.Version1 || err != nil {
		response, err = s.client.WalkAll(oid.String())
	}
	if err != nil {
		log.Ctx(ctx).Trace().Str("network_request", "snmpwalk").Str("oid", oid.String()).Err(err).Msg("snmp walk failed")
		err = errors.Wrap(err, "snmpwalk failed")
		if s.useCache {
			s.walkCache.add(oid.String(), nil, err)
		}
		return nil, err
	}

	if response == nil {
		log.Ctx(ctx).Trace().Str("network_request", "snmpwalk").Str("oid", oid.String()).Msg("No Such Object available on this agent at this OID")
		err = tholaerr.NewNotFoundError("No Such Object available on this agent at this OID")
		if s.useCache {
			s.walkCache.add(oid.String(), nil, err)
		}
		return nil, err
	}

	var res []SNMPResponse
	for _, currentResponse := range response {
		snmpResponse := NewSNMPResponse(OID(currentResponse.Name), currentResponse.Type, currentResponse.Value)
		res = append(res, snmpResponse)

		if s.useCache {
			if snmpResponse.WasSuccessful() {
				s.getCache.add(snmpResponse.oid.String(), snmpResponse, nil)
			} else {
				s.getCache.add(snmpResponse.oid.String(), snmpResponse, errors.New("SNMP Request failed"))
			}
		}
	}

	if s.useCache {
		s.walkCache.add(oid.String(), res, nil)
	}

	log.Ctx(ctx).Trace().Str("network_request", "snmpwalk").Str("oid", oid.String()).Msg("snmp walk successful")

	return res, nil
}

// UseCache configures whether the snmp cache should be used or not
func (s *snmpClient) UseCache(b bool) {
	s.useCache = b
}

// HasSuccessfulCachedRequest returns if there was at least one successful cached request.
func (s *snmpClient) HasSuccessfulCachedRequest() bool {
	return len(s.getCache.getSuccessfulRequests()) > 0
}

// Disconnect closes an snmp connection.
func (s *snmpClient) Disconnect() error {
	return s.client.Conn.Close()
}

// GetCommunity returns the community string
func (s *snmpClient) GetCommunity() string {
	return s.client.Community
}

// SetCommunity updates the community string. This function is not thread safe!
func (s *snmpClient) SetCommunity(community string) {
	s.client.Community = community
}

// GetPort returns the port
func (s *snmpClient) GetPort() int {
	return int(s.client.Port)
}

// GetVersion returns the snmp version.
func (s *snmpClient) GetVersion() string {
	switch s.client.Version {
	case gosnmp.Version1:
		return "1"
	case gosnmp.Version2c:
		return "2c"
	case gosnmp.Version3:
		return "3"
	}
	return ""
}

// GetMaxRepetitions returns the max repetitions.
func (s *snmpClient) GetMaxRepetitions() uint32 {
	return s.client.MaxRepetitions
}

// SetMaxRepetitions sets the maximum repetitions.
func (s *snmpClient) SetMaxRepetitions(maxRepetitions uint32) {
	s.client.MaxRepetitions = maxRepetitions
}

// SetMaxOIDs sets the maximum OIDs.
func (s *snmpClient) SetMaxOIDs(maxOIDs int) error {
	if maxOIDs < 1 {
		return errors.New("invalid max oids")
	}
	if s.client.Version == gosnmp.Version1 {
		return errors.New("max oids cannot be changed for snmp v1")
	}
	s.client.MaxOids = maxOIDs
	return nil
}

// GetV3Level returns the security level of the snmp v3 connection.
// Return value is nil if no snmp v3 is being used.
func (s *snmpClient) GetV3Level() *string {
	var level string
	switch s.client.MsgFlags {
	case gosnmp.NoAuthNoPriv | gosnmp.Reportable:
		level = "noAuthNoPriv"
	case gosnmp.AuthNoPriv | gosnmp.Reportable:
		level = "authNoPriv"
	case gosnmp.AuthPriv | gosnmp.Reportable:
		level = "authPriv"
	default:
		return nil
	}
	return &level
}

// GetV3ContextName returns the context name of the snmp v3 connection.
// Return value is nil if no snmp v3 is being used.
func (s *snmpClient) GetV3ContextName() *string {
	if s.client.ContextName == "" {
		return nil
	}
	contextName := s.client.ContextName
	return &contextName
}

// GetV3User returns the user of the snmp v3 connection.
// Return value is nil if no snmp v3 is being used.
func (s *snmpClient) GetV3User() *string {
	r, ok := s.client.SecurityParameters.(*gosnmp.UsmSecurityParameters)
	if !ok {
		return nil
	}
	if r.UserName == "" {
		return nil
	}
	return &r.UserName
}

// GetV3AuthKey returns the auth key of the snmp v3 connection.
// Return value is nil if no snmp v3 is being used.
func (s *snmpClient) GetV3AuthKey() *string {
	r, ok := s.client.SecurityParameters.(*gosnmp.UsmSecurityParameters)
	if !ok {
		return nil
	}
	if r.AuthenticationPassphrase == "" {
		return nil
	}
	return &r.AuthenticationPassphrase
}

// GetV3AuthProto returns the auth protocol of the snmp v3 connection.
// Return value is nil if no snmp v3 is being used.
func (s *snmpClient) GetV3AuthProto() *string {
	r, ok := s.client.SecurityParameters.(*gosnmp.UsmSecurityParameters)
	if !ok {
		return nil
	}
	var proto string
	switch r.AuthenticationProtocol {
	case gosnmp.NoAuth:
		proto = "NoAuth"
	case gosnmp.MD5:
		proto = "MD5"
	case gosnmp.SHA:
		proto = "SHA"
	case gosnmp.SHA224:
		proto = "SHA224"
	case gosnmp.SHA256:
		proto = "SHA256"
	case gosnmp.SHA384:
		proto = "SHA384"
	case gosnmp.SHA512:
		proto = "SHA512"
	default:
		return nil
	}
	return &proto
}

// GetV3PrivKey returns the priv key of the snmp v3 connection.
// Return value is nil if no snmp v3 is being used.
func (s *snmpClient) GetV3PrivKey() *string {
	r, ok := s.client.SecurityParameters.(*gosnmp.UsmSecurityParameters)
	if !ok {
		return nil
	}
	if r.PrivacyPassphrase == "" {
		return nil
	}
	return &r.PrivacyPassphrase
}

// GetV3PrivProto returns the priv protocol of the snmp v3 connection.
// Return value is nil if no snmp v3 is being used.
func (s *snmpClient) GetV3PrivProto() *string {
	r, ok := s.client.SecurityParameters.(*gosnmp.UsmSecurityParameters)
	if !ok {
		return nil
	}
	var proto string
	switch r.PrivacyProtocol {
	case gosnmp.NoPriv:
		proto = "NoPriv"
	case gosnmp.DES:
		proto = "DES"
	case gosnmp.AES:
		proto = "AES"
	case gosnmp.AES192:
		proto = "AES192"
	case gosnmp.AES256:
		proto = "AES256"
	case gosnmp.AES192C:
		proto = "AES192C"
	case gosnmp.AES256C:
		proto = "AES256C"
	default:
		return nil
	}
	return &proto
}

// SNMPResponse is the response returned for a single snmp request.
type SNMPResponse struct {
	oid      OID
	snmpType gosnmp.Asn1BER
	value    interface{}
}

// NewSNMPResponse creates a new SNMP Response
func NewSNMPResponse(oid OID, snmpType gosnmp.Asn1BER, value interface{}) SNMPResponse {
	return SNMPResponse{
		oid:      oid,
		snmpType: snmpType,
		value:    value,
	}
}

// WasSuccessful returns if the snmp request was successful.
func (s *SNMPResponse) WasSuccessful() bool {
	return s.snmpType != gosnmp.NoSuchObject && s.snmpType != gosnmp.NoSuchInstance && s.snmpType != gosnmp.Null
}

// GetValue returns the value of the snmp response.
func (s *SNMPResponse) GetValue() (value.Value, error) {
	if !s.WasSuccessful() {
		return nil, tholaerr.NewNotFoundError("no such object")
	}
	v, err := s.getValueDecoded()
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode value")
	}
	return value.New(v), nil
}

// GetValueRaw returns the raw value string of the snmp response.
func (s *SNMPResponse) GetValueRaw() (value.Value, error) {
	if !s.WasSuccessful() {
		return nil, tholaerr.NewNotFoundError("no such object")
	}
	if s.snmpType == gosnmp.OctetString {
		switch x := s.value.(type) {
		case string:
			return value.New(strings.ToUpper(hex.EncodeToString([]byte(x)))), nil
		case []byte:
			return value.New(strings.ToUpper(hex.EncodeToString(x))), nil
		}
	}
	return value.New(s.value), nil
}

func (s *SNMPResponse) getValueDecoded() (interface{}, error) {
	var err error
	i := s.value
	switch x := s.value.(type) {
	case string:
		i, err = charmap.ISO8859_1.NewDecoder().String(x)
		i = strings.TrimFunc(i.(string), func(r rune) bool {
			return !unicode.IsGraphic(r)
		})
	case []byte:
		i, err = charmap.ISO8859_1.NewDecoder().Bytes(x)
		i = bytes.TrimFunc(i.([]byte), func(r rune) bool {
			return !unicode.IsGraphic(r)
		})
	}
	if err != nil {
		return nil, err
	}
	return i, nil
}

// GetOID returns the oid of the response.
func (s *SNMPResponse) GetOID() OID {
	return s.oid
}

// GetSNMPType returns the snmp type of the response.
func (s *SNMPResponse) GetSNMPType() gosnmp.Asn1BER {
	return s.snmpType
}

// SNMPGetConfiguration represents the configuration needed to get a value.
type SNMPGetConfiguration struct {
	OID          OID  `yaml:"oid" mapstructure:"oid"`
	UseRawResult bool `yaml:"use_raw_result" mapstructure:"use_raw_result"`
}

// GetValueBySNMPGetConfiguration returns the value of the snmp response according to the snmpgetConfig
func (s *SNMPResponse) GetValueBySNMPGetConfiguration(snmpGetConfig SNMPGetConfiguration) (value.Value, error) {
	var val value.Value
	var err error
	if snmpGetConfig.UseRawResult {
		val, err = s.GetValueRaw()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get snmp result raw string")
		}
	} else {
		val, err = s.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get snmp result string")
		}
	}
	return val, nil
}

// OID represents an SNMP OID.
type OID string

func (o OID) String() string {
	return string(o)
}

// Validate checks if the OID is syntactically correct
func (o OID) Validate() error {
	m, err := regexp.MatchString("^[0-9.]+$", string(o))
	if err != nil {
		return errors.Wrap(err, "regex match string failed")
	}
	if !m {
		return errors.New("invalid oid")
	}
	return nil
}

// GetIndex returns the last index of the OID.
func (o OID) GetIndex() string {
	x := strings.Split(o.String(), ".")
	return x[len(x)-1]
}

// AddSuffix returns a OID with the specified suffix attached.
func (o OID) AddSuffix(suffix string) OID {
	return OID(o.String() + suffix)
}
