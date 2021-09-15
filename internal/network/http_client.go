package network

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// HTTPClient is used for communication over HTTP(s).
type HTTPClient struct {
	client *resty.Client

	host string
	port *int

	format string

	useHTTPS bool

	useAuth  bool
	username string
	password string

	useCache bool

	cache requestCache
}

// NewHTTPClient returns a new HTTP client.
func NewHTTPClient(URI string) (*HTTPClient, error) {
	u, err := url.ParseRequestURI(URI)
	if err != nil {
		return nil, errors.Wrap(err, "invalid target URI")
	}

	uri := u.Hostname()
	if uri == "" {
		return nil, errors.New("invalid target URI")
	}

	if path := u.Path; path != "" {
		uri += path
	}

	httpClient := HTTPClient{host: uri, client: resty.New(), useAuth: false, useHTTPS: true, useCache: true, cache: newRequestCache(), format: "application/json"}

	if u.Scheme == "http" {
		httpClient.useHTTPS = false
	}
	if p := u.Port(); p != "" {
		port, err := strconv.Atoi(p)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot convert port '%s' to int", p)
		}
		httpClient.port = &port
	}
	return &httpClient, nil
}

// SetUsernameAndPassword sets username and password for an http client.
func (h *HTTPClient) SetUsernameAndPassword(username, password string) error {
	if username == "" {
		return errors.New("invalid username")
	}
	if password == "" {
		return errors.New("invalid password")
	}
	h.username = username
	h.password = password
	h.useAuth = true
	return nil
}

// UseHTTPS turns on HTTPS.
func (h *HTTPClient) UseHTTPS(useHTTPS bool) {
	h.useHTTPS = useHTTPS
}

// SetPort sets HTTP(S) port
func (h *HTTPClient) SetPort(port int) {
	h.port = &port
}

// UseDefaultPort sets HTTP(S) port to its default port.
func (h *HTTPClient) UseDefaultPort() {
	h.port = nil
}

// SetTimeout sets a timeout for the http client.
func (h *HTTPClient) SetTimeout(timeout time.Duration) {
	h.client.SetTimeout(timeout)
}

// InsecureSSLCert defines weather insecure ssl certificates are allowed.
func (h *HTTPClient) InsecureSSLCert(b bool) {
	h.client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: b})
}

// UseCache configures whether the http cache should be used or not.
func (h *HTTPClient) UseCache(b bool) {
	h.useCache = b
}

// SetFormat sets the format header which is used in requests.
func (h *HTTPClient) SetFormat(format string) error {
	switch format {
	case "json":
		h.format = "application/json"
	case "xml":
		h.format = "application/xml"
	default:
		return tholaerr.NewNotFoundError("invalid format")
	}
	return nil
}

// HasSuccessfulCachedRequest returns if there was at least one successful cached request.
func (h *HTTPClient) HasSuccessfulCachedRequest() bool {
	return len(h.cache.getSuccessfulRequests()) > 0
}

// Request sends an http request.
func (h *HTTPClient) Request(ctx context.Context, method, path, body string, header, queryParams map[string]string) (*resty.Response, error) {
	if h.useCache && method == http.MethodGet {
		x, err := h.cache.get(h.getRequestCacheKey(path))
		if err == nil {
			res, ok := x.res.(*resty.Response)
			if !ok {
				return nil, errors.New("cached http result is not a resty response")
			}
			return res, nil
		}
	}

	request := h.client.R()
	request.SetHeader("Content-Type", h.format)
	request.SetContext(ctx)

	if header != nil {
		request.SetHeaders(header)
	}

	if queryParams != nil {
		request.SetQueryParams(queryParams)
	}

	if body != "" {
		request.SetBody(body)
	}

	if h.useAuth {
		request.SetBasicAuth(h.username, h.password)
	}

	var response *resty.Response

	URLStr := "https"
	if !h.useHTTPS {
		URLStr = "http"
	}
	URLStr += "://" + h.host
	if h.port != nil {
		URLStr += ":" + strconv.Itoa(*h.port)
	}
	URLStr += "/"
	URL, err := url.Parse(URLStr)
	if err != nil {
		return nil, errors.Wrap(err, "error while parsing url")
	}
	URL.Path = filepath.Join(URL.Path, URLEscapePath(path))

	switch method {
	case http.MethodGet:
		response, err = request.Get(URL.String())
	case http.MethodPost:
		response, err = request.Post(URL.String())
	case http.MethodPut:
		response, err = request.Put(URL.String())
	case http.MethodDelete:
		response, err = request.Delete(URL.String())
	default:
		return nil, errors.New("invalid http method: " + method)
	}
	// save cache
	if h.useCache && method == http.MethodGet {
		h.cache.add(h.getRequestCacheKey(path), response, err)
	}
	if err != nil {
		return nil, tholaerr.NewHTTPError(err.Error())
	}
	return response, nil
}

func (h *HTTPClient) getRequestCacheKey(path string) string {
	return fmt.Sprintf("%s:%d:%s", h.GetProtocolString(), h.port, path)
}

// GetProtocolString returns the protocol as a string.
func (h *HTTPClient) GetProtocolString() string {
	if h.useHTTPS {
		return "https"
	}
	return "http"
}

// GetHostname returns the hostname.
func (h *HTTPClient) GetHostname() string {
	return h.host
}

// URLEscapePath url-escapes a file path.
func URLEscapePath(unescaped string) string {
	arr := strings.Split(unescaped, "/")
	for i, partString := range strings.Split(unescaped, "/") {
		arr[i] = url.QueryEscape(partString)
	}
	return strings.Join(arr, "/")
}
