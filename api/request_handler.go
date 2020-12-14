package api

import (
	"crypto/subtle"
	"fmt"
	"github.com/inexio/thola/api/statistics"
	"github.com/inexio/thola/core/request"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
	"net/http"
	"sync"
)

var deviceLocks struct {
	sync.RWMutex

	locks map[string]*sync.Mutex
}

// StartAPI starts the API.
func StartAPI() {
	deviceLocks.locks = make(map[string]*sync.Mutex)
	e := echo.New()

	e.HideBanner = true
	fmt.Print(" ______   __  __     ______     __         ______   \n" +
		"/\\__  _\\ /\\ \\_\\ \\   /\\  __ \\   /\\ \\       /\\  __ \\  \n" +
		"\\/_/\\ \\/ \\ \\  __ \\  \\ \\ \\/\\ \\  \\ \\ \\____  \\ \\  __ \\ \n" +
		"   \\ \\_\\  \\ \\_\\ \\_\\  \\ \\_____\\  \\ \\_____\\  \\ \\_\\ \\_\\\n" +
		"    \\/_/   \\/_/\\/_/   \\/_____/   \\/_____/   \\/_/\\/_/\n\n")

	if (viper.GetString("api.username") != "") && (viper.GetString("api.password") != "") {
		e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
			// Be careful to use constant time comparison to prevent timing attacks
			if subtle.ConstantTimeCompare([]byte(username), []byte(viper.GetString("restapi.username"))) == 1 &&
				subtle.ConstantTimeCompare([]byte(password), []byte(viper.GetString("restapi.password"))) == 1 {
				return true, nil
			}
			return false, nil
		}))
	}

	if viper.GetString("api.ratelimit") != "" {
		e.Use(ipRateLimit())
	}

	e.Use(statistics.Middleware())

	// swagger:operation POST /identify identify identify
	// ---
	// summary: Identifies a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json<
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/IdentifyRequest'
	// responses:
	//   200:
	//     description: Returns the device which was found.
	//     schema:
	//       $ref: '#/definitions/IdentifyResponse'
	//   400:
	//     description: Returns a string that the request was formatted wrong.
	//   404:
	//     description: Returns a string that no device was found.
	e.POST("/identify", identify)

	// swagger:operation POST /check/identify check checkIdentify
	// ---
	// summary: Checks if identify matches the expectations.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckIdentifyRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/identify", checkIdentify)

	// swagger:operation POST /check/snmp check checkSNMP
	// ---
	// summary: Checks SNMP availability.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckSNMPRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/snmp", checkSNMP)

	// swagger:operation POST /check/interface-metrics check checkInterfaceMetrics
	// ---
	// summary: Check to read out interface metrics.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckInterfaceMetricsRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/interface-metrics", checkInterfaceMetrics)

	// swagger:operation POST /check/thola-server check checkTholaServer
	// ---
	// summary: Check existence of thola servers.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckTholaServerRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/thola-server", checkTholaServer)

	// swagger:operation POST /check/ups check checkUPS
	// ---
	// summary: Checks whether a UPS device has its main voltage applied.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckUPSRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/ups", checkUPS)

	e.POST("/check/memory-usage", checkMemoryUsage)

	// swagger:operation POST /check/metrics check checkMetrics
	// ---
	// summary: Prints all available metrics for a device as performance data.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckMetricsRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/metrics", checkMetrics)

	// swagger:operation POST /read/interfaces read readInterfaces
	// ---
	// summary: Reads out data of the interfaces of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadInterfacesRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadInterfacesResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/interfaces", readInterfaces)

	// swagger:operation POST /read/count-interfaces read readCountInterfaces
	// ---
	// summary: Counts the interfaces of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadCountInterfacesRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadCountInterfacesResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/count-interfaces", readCountInterfaces)

	// swagger:operation POST /read/cpu-load read readCPULoad
	// ---
	// summary: Read out the CPU load of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadCPULoadRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadCPULoadResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/cpu-load", readCPULoad)

	// swagger:operation POST /read/memory-usage read readMemoryUsage
	// ---
	// summary: Read out the memory usage of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadMemoryUsageRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadMemoryUsageResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/memory-usage", readMemoryUsage)

	// swagger:operation POST /read/ups read readUPS
	// ---
	// summary: Reads out UPS data of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadUPSRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadUPSResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/ups", readUPS)

	// swagger:operation POST /read/available-components read readAvailableComponents
	// ---
	// summary: Returns the available components for the device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadAvailableComponentsRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadAvailableComponentsResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/available-components", readAvailableComponents)

	if viper.GetString("api.certfile") != "" && viper.GetString("api.keyfile") != "" {
		e.Logger.Fatal(e.StartTLS(":"+viper.GetString("api.port"), viper.GetString("api.certfile"), viper.GetString("api.keyfile")))
	} else {
		e.Logger.Fatal(e.Start(":" + viper.GetString("api.port")))
	}
}

func identify(ctx echo.Context) error {
	r := request.IdentifyRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(&r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkIdentify(ctx echo.Context) error {
	r := request.CheckIdentifyRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(&r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkSNMP(ctx echo.Context) error {
	r := request.CheckSNMPRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(&r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkInterfaceMetrics(ctx echo.Context) error {
	r := request.CheckInterfaceMetricsRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(&r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkTholaServer(ctx echo.Context) error {
	r := request.CheckTholaServerRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(&r, nil)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkUPS(ctx echo.Context) error {
	r := request.CheckUPSRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(&r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkMemoryUsage(ctx echo.Context) error {
	r := request.CheckMemoryUsageRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(&r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkMetrics(ctx echo.Context) error {
	r := request.CheckMetricsRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(&r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readInterfaces(ctx echo.Context) error {
	r := request.ReadInterfacesRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(&r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readCountInterfaces(ctx echo.Context) error {
	r := request.ReadCountInterfacesRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(&r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readCPULoad(ctx echo.Context) error {
	r := request.ReadCPULoadRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(&r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readMemoryUsage(ctx echo.Context) error {
	r := request.ReadMemoryUsageRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(&r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readUPS(ctx echo.Context) error {
	r := request.ReadUPSRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(&r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readAvailableComponents(ctx echo.Context) error {
	r := request.ReadAvailableComponentsRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(&r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func handleError(ctx echo.Context, err error) error {
	if tholaerr.IsNetworkError(err) {
		return returnInFormat(ctx, http.StatusBadRequest, tholaerr.OutputError{Error: "Network error: " + err.Error()})
	}
	if tholaerr.IsNotImplementedError(err) {
		return returnInFormat(ctx, http.StatusInternalServerError, tholaerr.OutputError{Error: "Function not implemented: " + err.Error()})
	}
	if tholaerr.IsNotFoundError(err) {
		return returnInFormat(ctx, http.StatusNotAcceptable, tholaerr.OutputError{Error: "Not found: " + err.Error()})
	}
	if tholaerr.IsTooManyRequestsError(err) {
		return returnInFormat(ctx, http.StatusTooManyRequests, tholaerr.OutputError{Error: "Too many requests: " + err.Error()})
	}
	return returnInFormat(ctx, http.StatusBadRequest, tholaerr.OutputError{Error: "Request failed: " + err.Error()})
}

func returnInFormat(ctx echo.Context, statusCode int, resp interface{}) error {
	if viper.GetString("api.format") == "json" {
		return ctx.JSON(statusCode, resp)
	} else if viper.GetString("api.format") == "xml" {
		return ctx.XML(statusCode, resp)
	}
	return ctx.String(http.StatusInternalServerError, "Invalid output format set")
}

func getDeviceLock(ip string) *sync.Mutex {
	deviceLocks.RLock()
	lock, ok := deviceLocks.locks[ip]
	deviceLocks.RUnlock()
	if !ok {
		deviceLocks.Lock()
		if lock, ok = deviceLocks.locks[ip]; !ok {
			lock = &sync.Mutex{}
			deviceLocks.locks[ip] = lock
		}
		deviceLocks.Unlock()
	}
	return lock
}

func handleAPIRequest(r request.Request, ip *string) (request.Response, error) {
	if ip != nil && !viper.GetBool("request.no-ip-lock") {
		lock := getDeviceLock(*ip)
		lock.Lock()
		defer lock.Unlock()
	}
	return request.ProcessRequest(r)
}
