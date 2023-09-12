// Package ginzero provides log handing useing zerolog package.
// Code structure based on ginrus package.
package ginzero

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// ZapLogger is the minimal logger interface compatible with zerolog.Logger
type ZeroLogger interface {
	Info() *zerolog.Event
	Error() *zerolog.Event
}

type Config struct {
	TimeFormat string
	UTC        bool
	SkipPaths  []string
}

type OptionFunc func(*Config)

func WithTimeFormat(timeFormat string) OptionFunc {
	return func(c *Config) {
		c.TimeFormat = timeFormat
	}
}

func WithUTC(utc bool) OptionFunc {
	return func(c *Config) {
		c.UTC = utc
	}
}

func WithSkipPaths(paths []string) OptionFunc {
	return func(c *Config) {
		if len(paths) != 0 {
			c.SkipPaths = append(c.SkipPaths, paths...)
		}
	}
}

func Ginzero(logger ZeroLogger, optFuncs ...OptionFunc) gin.HandlerFunc {
	config := &Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
	}

	for _, of := range optFuncs {
		of(config)
	}

	return GinzeroWithConfig(logger, config)
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

		defer func() {
			if _, ok := skipPaths[path]; !ok {
				end := time.Now()
				latency := end.Sub(start)
				if conf.UTC {
					end = end.UTC()
				}
				l := logger.Info().
					Int("status", c.Writer.Status()).
					Str("method", c.Request.Method).
					Str("path", path).
					Str("query", query).
					Str("ip", c.ClientIP()).
					Str("user-agent", c.Request.UserAgent()).
					Dur("latency", latency)

				if conf.TimeFormat != "" {
					l.Str("time", end.Format(conf.TimeFormat))
				}

				if len(c.Errors) > 0 {
					l = logger.Error().
						Int("status", c.Writer.Status()).
						Str("method", c.Request.Method).
						Str("path", path).
						Str("query", query).
						Str("ip", c.ClientIP()).
						Str("user-agent", c.Request.UserAgent()).
						Dur("latency", latency)

					// Append error field if this is an erroneous request.
					for _, e := range c.Errors.Errors() {
						l.Str("error", e).Send()
					}
				} else {
					l.Send()
				}
			}
		}()

		c.Next()
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
					errors.New(string(debug.Stack()))
					logger.Error().
						Stack().
						Err(errors.New(string(debug.Stack()))).
						Str("error", "[Recovery from panic]").
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
