package network

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/gosnmp/gosnmp"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/inexio/thola/core/utility"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/encoding/charmap"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// SNMPClient is used to communicate via snmp.
type SNMPClient struct {
	client    *gosnmp.GoSNMP
	useCache  bool
	getCache  requestCache
	walkCache requestCache
}

type snmpClientCreation struct {
	client  *SNMPClient
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
	v3Data      *SNMPv3ConnectionData
}

// NewSNMPClientByConnectionData tries to create a new snmp client by SNMPConnectionData and returns it.
func NewSNMPClientByConnectionData(ctx context.Context, ipAddress string, data *SNMPConnectionData) (*SNMPClient, error) {
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
			} else {
				for _, community := range data.Communities {
					in <- snmpClientCreationData{
						ipAddress:   ipAddress,
						snmpVersion: version,
						community:   community,
						port:        port,
						timeout:     *data.DiscoverTimeout,
						retries:     *data.DiscoverRetries,
						v3Data:      nil,
					}
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
	var successfulClient *SNMPClient

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
			if successfulClient == nil || successfulClient.client.Version < gosnmp.Version3 {
				successfulClient = res.client
			}
		}
		if res.version == "2c" {
			if !v3Available {
				return res.client, nil
			}
			if successfulClient == nil || successfulClient.client.Version < gosnmp.Version2c {
				successfulClient = res.client
			}
		}
		if res.version == "1" {
			if !v3Available && !v2cAvailable {
				return res.client, nil
			}
			if successfulClient == nil {
				successfulClient = res.client
			}
		}
	}
	if successfulClient != nil {
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
				var client *SNMPClient
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

// NewSNMPv3Client creates a new SNMP v3 Client.
func NewSNMPv3Client(ctx context.Context, ipAddress string, port, timeout, retries int, v3Data *SNMPv3ConnectionData) (*SNMPClient, error) {
	if v3Data.ContextName == nil || v3Data.Level == nil {
		return nil, errors.New("v3 connection requested but no connection data provided")
	}

	client := &gosnmp.GoSNMP{
		Context:       ctx,
		Target:        ipAddress,
		Port:          uint16(port),
		Transport:     "udp",
		Version:       gosnmp.Version3,
		Timeout:       time.Duration(timeout) * time.Second,
		MaxOids:       60,
		Retries:       retries,
		ContextName:   *v3Data.ContextName,
		SecurityModel: gosnmp.UserSecurityModel,
	}

	switch *v3Data.Level {
	case "noAuthNoPriv":
		if v3Data.User == nil {
			return nil, errors.New("no username for snmp v3 provided")
		}
		client.MsgFlags = gosnmp.NoAuthNoPriv
		client.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName: *v3Data.User,
		}
	case "authNoPriv":
		if v3Data.User == nil || v3Data.AuthProtocol == nil || v3Data.AuthKey == nil {
			return nil, errors.New("no username, auth protocol or auth key for snmp v3 provided")
		}
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
		if v3Data.User == nil || v3Data.AuthProtocol == nil || v3Data.AuthKey == nil || v3Data.PrivProtocol == nil || v3Data.PrivKey == nil {
			return nil, errors.New("no username, auth protocol, auth key, priv protocol or priv key for snmp v3 provided")
		}
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
	default:
		return nil, fmt.Errorf("invalid level '%s', only 'noAuthNoPriv', 'authNoPriv' and 'authPriv' are possible", *v3Data.Level)
	}

	var err error

	err = client.ConnectIPv4()
	if err != nil {
		return nil, errors.Wrap(err, "connect ip v4 failed")
	}

	oids := []string{".0.0"}
	_, err = client.GetNext(oids)
	if err != nil {
		return nil, tholaerr.NewSNMPError(err.Error())
	}

	client.Retries = 3
	client.ExponentialTimeout = true

	return &SNMPClient{
		client:    client,
		useCache:  true,
		getCache:  newRequestCache(),
		walkCache: newRequestCache(),
	}, nil
}

// NewSNMPClient creates a new SNMP Client
func NewSNMPClient(ctx context.Context, ipAddress, snmpVersion, community string, port, timeout, retries int) (*SNMPClient, error) {
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
		MaxOids:   60,
		Retries:   retries,
	}

	err = client.ConnectIPv4()
	if err != nil {
		return nil, errors.Wrap(err, "connect ip v4 failed")
	}

	oids := []string{".0.0"}
	_, err = client.GetNext(oids)
	if err != nil {
		return nil, tholaerr.NewSNMPError(err.Error())
	}

	client.Retries = 3
	client.ExponentialTimeout = true

	return &SNMPClient{
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
func (s *SNMPClient) SNMPGet(ctx context.Context, oid ...string) ([]SNMPResponse, error) {
	var snmpResponses []SNMPResponse

	m := make(map[int]SNMPResponse)
	var reqOIDs []string

	if s.useCache {
		for a, o := range oid {
			x, err := s.getCache.get(o)
			if err != nil {
				reqOIDs = append(reqOIDs, o)
			} else {
				res, ok := x.res.(SNMPResponse)
				if !ok {
					return nil, errors.New("cached snmp result is not a snmp response")
				}
				m[a] = res
			}
		}
	} else {
		reqOIDs = oid
	}

	var response *gosnmp.SnmpPacket
	var err error
	s.client.Context = ctx

	if len(reqOIDs) != 0 {
		response, err = s.client.Get(reqOIDs)
		if err != nil {
			return nil, errors.Wrap(err, "error during snmpget")
		}
	}

	successful := false

	var currentResponse gosnmp.SnmpPDU
	for i := 0; i < len(oid); i++ {
		if x, ok := m[i]; ok {
			snmpResponses = append(snmpResponses, x)
			if x.WasSuccessful() {
				successful = true
			}
		} else {
			currentResponse, response.Variables = response.Variables[0], response.Variables[1:]
			snmpResponse := SNMPResponse{}
			snmpResponse.oid = currentResponse.Name
			snmpResponse.value = currentResponse.Value
			snmpResponse.snmpType = currentResponse.Type

			if snmpResponse.WasSuccessful() {
				successful = true
				if s.useCache {
					s.getCache.add(snmpResponse.oid, snmpResponse, nil)
				}
			} else if s.useCache {
				s.getCache.add(snmpResponse.oid, snmpResponse, errors.New("SNMP Request failed"))
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
func (s *SNMPClient) SNMPWalk(ctx context.Context, oid string) ([]SNMPResponse, error) {
	if s.useCache {
		x, err := s.walkCache.get(oid)
		if err == nil {
			res, ok := x.res.([]SNMPResponse)
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
		response, err = s.client.BulkWalkAll(oid)
		if err != nil {
			log.Ctx(ctx).Trace().Err(err).Msg("bulk walk failed")
		}
	}
	if s.client.Version == gosnmp.Version1 || err != nil {
		response, err = s.client.WalkAll(oid)
	}
	if err != nil {
		err = errors.Wrap(err, "snmpwalk failed")
		if s.useCache {
			s.walkCache.add(oid, nil, err)
		}
		return nil, err
	}

	if response == nil {
		err = tholaerr.NewNotFoundError("No Such Object available on this agent at this OID")
		if s.useCache {
			s.walkCache.add(oid, nil, err)
		}
		return nil, err
	}

	var res []SNMPResponse
	for _, currentResponse := range response {
		snmpResponse := SNMPResponse{
			oid:      currentResponse.Name,
			value:    currentResponse.Value,
			snmpType: currentResponse.Type,
		}
		res = append(res, snmpResponse)

		if s.useCache {
			if snmpResponse.WasSuccessful() {
				s.getCache.add(snmpResponse.oid, snmpResponse, nil)
			} else {
				s.getCache.add(snmpResponse.oid, snmpResponse, errors.New("SNMP Request failed"))
			}
		}
	}

	if s.useCache {
		s.walkCache.add(oid, res, nil)
	}

	return res, nil
}

// UseCache configures whether the snmp cache should be used or not
func (s *SNMPClient) UseCache(b bool) {
	s.useCache = b
}

// GetSuccessfulCachedRequests returns all successful cached requests.
func (s *SNMPClient) GetSuccessfulCachedRequests() map[string]cachedRequestResult {
	return s.getCache.getSuccessfulRequests()
}

// Disconnect closes an snmp connection.
func (s *SNMPClient) Disconnect() error {
	return s.client.Conn.Close()
}

// GetCommunity returns the community string
func (s *SNMPClient) GetCommunity() string {
	return s.client.Community
}

// SetCommunity updates the community string. This function is not thread safe!
func (s *SNMPClient) SetCommunity(community string) {
	s.client.Community = community
}

// GetPort returns the port
func (s *SNMPClient) GetPort() int {
	return int(s.client.Port)
}

// GetVersion returns the snmp version.
func (s *SNMPClient) GetVersion() string {
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
func (s *SNMPClient) GetMaxRepetitions() uint8 {
	return utility.IfThenElse(s.client.MaxRepetitions == 0, gosnmp.Default.MaxRepetitions, s.client.MaxRepetitions).(uint8)
}

// SetMaxRepetitions sets the maximum repetitions.
func (s *SNMPClient) SetMaxRepetitions(maxRepetitions uint32) {
	s.client.MaxRepetitions = maxRepetitions
}

// GetV3Level returns the security level of the snmp v3 connection.
// Return value is nil if no snmp v3 is being used.
func (s *SNMPClient) GetV3Level() *string {
	var level string
	level = "authPriv"
	/*switch s.client.MsgFlags {
	case gosnmp.NoAuthNoPriv:
		level = "noAuthNoPriv"
	case gosnmp.AuthNoPriv:
		level = "authNoPriv"
	case gosnmp.AuthPriv:
		level = "authPriv"
	default:
		return nil
	}*/
	return &level
}

// GetV3ContextName returns the context name of the snmp v3 connection.
// Return value is nil if no snmp v3 is being used.
func (s *SNMPClient) GetV3ContextName() *string {
	if s.client.ContextName == "" {
		return nil
	}
	contextName := s.client.ContextName
	return &contextName
}

// GetV3User returns the user of the snmp v3 connection.
// Return value is nil if no snmp v3 is being used.
func (s *SNMPClient) GetV3User() *string {
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
func (s *SNMPClient) GetV3AuthKey() *string {
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
func (s *SNMPClient) GetV3AuthProto() *string {
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
func (s *SNMPClient) GetV3PrivKey() *string {
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
func (s *SNMPClient) GetV3PrivProto() *string {
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
	oid      string
	snmpType gosnmp.Asn1BER
	value    interface{}
}

// WasSuccessful returns if the snmp request was successful.
func (s *SNMPResponse) WasSuccessful() bool {
	return s.snmpType != gosnmp.NoSuchObject && s.snmpType != gosnmp.NoSuchInstance && s.snmpType != gosnmp.Null
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

// GetValue returns the value of the snmp response.
func (s *SNMPResponse) GetValue() (interface{}, error) {
	if !s.WasSuccessful() {
		return "", tholaerr.NewNotFoundError("no such object")
	}
	v, err := s.getValueDecoded()
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode value")
	}
	return v, nil
}

// GetValueString returns the value string of the snmp response.
func (s *SNMPResponse) GetValueString() (string, error) {
	return s.getValueString(true)
}

func (s *SNMPResponse) getValueString(decoded bool) (string, error) {
	if !s.WasSuccessful() {
		return "", tholaerr.NewNotFoundError("no such object")
	}
	v := s.value
	var err error
	var value string

	if decoded {
		v, err = s.getValueDecoded()
		if err != nil {
			return "", errors.Wrap(err, "failed to decode value")
		}
	}

	switch x := v.(type) {
	case string:
		value = x
	case []byte:
		value = string(x)
	case nil:
		return "", tholaerr.NewNotFoundError("value is nil")
	default:
		value = fmt.Sprint(x)
	}
	return value, nil
}

// GetValueStringRaw returns the raw value string of the snmp response.
func (s *SNMPResponse) GetValueStringRaw() (string, error) {
	if !s.WasSuccessful() {
		return "", tholaerr.NewNotFoundError("no such object")
	}
	if s.snmpType == gosnmp.OctetString {
		switch x := s.value.(type) {
		case string:
			return strings.ToUpper(hex.EncodeToString([]byte(x))), nil
		case []byte:
			return strings.ToUpper(hex.EncodeToString(x)), nil
		}
	}

	return s.getValueString(false)
}

// GetOID returns the oid of the response.
func (s *SNMPResponse) GetOID() string {
	return s.oid
}

// GetSNMPType returns the snmp type of the response.
func (s *SNMPResponse) GetSNMPType() gosnmp.Asn1BER {
	return s.snmpType
}

type SNMPGetConfiguration struct {
	OID          OID  `yaml:"oid" mapstructure:"oid"`
	UseRawResult bool `yaml:"use_raw_result" mapstructure:"use_raw_result"`
}

// GetValueBySNMPGetConfiguration returns the value of the snmp response according to the snmpgetConfig
func (s *SNMPResponse) GetValueBySNMPGetConfiguration(snmpgetConfig SNMPGetConfiguration) (string, error) {
	var value string
	var err error
	if snmpgetConfig.UseRawResult {
		value, err = s.GetValueStringRaw()
		if err != nil {
			return "", errors.Wrap(err, "failed to get snmp result raw string")
		}
	} else {
		value, err = s.GetValueString()
		if err != nil {
			return "", errors.Wrap(err, "failed to get snmp result string")
		}
	}
	return value, nil
}

// OID represents an SNMP oid.
type OID string

// Validate checks if the oid is syntactically correct
func (o *OID) Validate() error {
	m, err := regexp.MatchString("^[0-9.]+$", string(*o))
	if err != nil {
		return errors.Wrap(err, "regex match string failed")
	}
	if !m {
		return errors.New("invalid oid")
	}
	return nil
}
