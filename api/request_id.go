package api

import (
	"github.com/labstack/echo"
	"github.com/rs/xid"
)

func requestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			rid := req.Header.Get(echo.HeaderXRequestID)
			if rid == "" {
				rid = xid.New().String()
				req.Header.Set(echo.HeaderXRequestID, rid)
			}
			res.Header().Set(echo.HeaderXRequestID, rid)

			return next(c)
		}
	}
}
