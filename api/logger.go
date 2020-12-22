package api

import (
	"github.com/labstack/echo"
	"github.com/rs/zerolog/log"
	"strconv"
	"time"
)

func LoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var err error
			req := c.Request()
			res := c.Response()
			start := time.Now()

			if err = next(c); err != nil {
				c.Error(err)
			}

			stop := time.Now()

			logger := log.Logger.With().Str("request_id", res.Header().Get(echo.HeaderXRequestID)).Logger()

			evt := logger.Info()
			evt.Str("remote_ip", c.RealIP())
			evt.Str("host", req.Host)
			evt.Str("method", req.Method)
			evt.Str("uri", req.RequestURI)
			evt.Str("user_agent", req.UserAgent())
			evt.Int("status", res.Status)

			if err != nil {
				evt.Err(err)
			}

			evt.Str("latency", stop.Sub(start).String())

			cl := req.Header.Get(echo.HeaderContentLength)
			if cl == "" {
				cl = "0"
			}

			evt.Str("bytes_in", cl)
			evt.Str("bytes_out", strconv.FormatInt(res.Size, 10))
			evt.Msg("")

			return err
		}
	}
}
