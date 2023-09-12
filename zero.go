// Package ginzero provides log handing useing zerolog package.
// Code structure based on ginrus package.
package ginzero

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type Fn func(c *gin.Context) []*zerolog.Event

// ZapLogger is the minimal logger interface compatible with zap.Logger
type ZeroLogger interface {
	Info() *zerolog.Event
	Error() *zerolog.Event
}

type Config struct {
	TimeFormat string
	UTC        bool
	SkipPaths  []string
	Events     Fn
}

func GinzeroWithConfig(logger ZeroLogger, conf *Config) gin.HandlerFunc {
	skipPaths := make(map[string]bool, len(conf.SkipPaths))
	for _, path := range conf.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		if _, ok := skipPaths[path]; !ok {
			end := time.Now()
			latency := end.Sub(start)
			if conf.UTC {
				end = end.UTC()
			}

			// l := zerolog.New(io.Discard).With().Logger()
			// l.Info()
			fields := []*zerolog.Event{
				zerolog.Dict().Int("status", c.Writer.Status()),
				zerolog.Dict().Str("method", c.Request.Method),
				zerolog.Dict().Str("path", path),
				zerolog.Dict().Str("query", query),
				zerolog.Dict().Str("ip", c.ClientIP()),
				zerolog.Dict().Str("user-agent", c.Request.UserAgent()),
				zerolog.Dict().Dur("latency", latency),
			}

			if conf.TimeFormat != "" {
				// fields = append(fields, zap.String("time", end.Format(conf.TimeFormat)))
				fields = append(fields, zerolog.Dict().Str("time", end.Format(conf.TimeFormat)))
			}

			if conf.Events != nil {
				fields = append(fields, conf.Events(c)...)
			}

			if len(c.Errors) > 0 {
				// Append error field if this is an erroneous request.
				for _, e := range c.Errors.Errors() {
					logger.Error().Str("error", e).Fields(fields).Send()
				}
			} else {
				logger.Info().Str("path", path).Fields(fields).Send()
			}
		}
	}
}

func defaultHandleRecovery(c *gin.Context, err interface{}) {
	c.AbortWithStatus(http.StatusInternalServerError)
}
func RecoveryWithZero(logger ZeroLogger, stack bool) gin.HandlerFunc {
	return CustomRecoveryWithZero(logger, stack, defaultHandleRecovery)
}

func CustomRecoveryWithZero(logger ZeroLogger, stack bool, recovery gin.RecoveryFunc) gin.HandlerFunc {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Error().
						Str("path", c.Request.URL.Path).
						Any("error", err).
						Str("request", string(httpRequest)).
						Send()
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					logger.Error().
						Stack().
						Str("error", "[Recovery from panic]").
						Time("time", time.Now()).
						Any("error", err).
						Str("request", string(httpRequest)).
						Send()

				} else {
					logger.Error().
						Str("error", "[Recovery from panic]").
						Time("time", time.Now()).
						Any("error", err).
						Str("request", string(httpRequest)).
						Send()
				}
				recovery(c, err)
			}
		}()
		c.Next()
	}
}
