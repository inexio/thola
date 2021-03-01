package api

import (
	"github.com/inexio/thola/core/tholaerr"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"os"
	"strconv"
)

var (
	ipRateLimiter *limiter.Limiter
	store         limiter.Store
)

func ipRateLimit() echo.MiddlewareFunc {

	rate, err := limiter.NewRateFromFormatted(viper.GetString("api.ratelimit"))
	if err != nil {
		log.Error().Msg("Wrong format for ratelimit")
		os.Exit(1)
	}

	store = memory.NewStore()
	ipRateLimiter = limiter.New(store, rate)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			ip := c.RealIP()
			limiterCtx, err := ipRateLimiter.Get(c.Request().Context(), ip)
			if err != nil {
				log.Printf("ipRateLimit - ipRateLimiter.Get - err: %v, %s on %s", err, ip, c.Request().URL)
				return handleError(c, err)
			}

			h := c.Response().Header()
			h.Set("X-RateLimit-Limit", strconv.FormatInt(limiterCtx.Limit, 10))
			h.Set("X-RateLimit-Remaining", strconv.FormatInt(limiterCtx.Remaining, 10))
			h.Set("X-RateLimit-Reset", strconv.FormatInt(limiterCtx.Reset, 10))

			if limiterCtx.Reached {
				log.Printf("Too Many Requests from %s on %s", ip, c.Request().URL)
				return handleError(c, tholaerr.NewTooManyRequestsError("Too Many Requests on "+c.Request().URL.String()))
			}

			return next(c)
		}
	}
}
