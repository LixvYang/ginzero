# ginzero

## Usage

### Start using it

Download and install it:

```sh
go get github.com/lixvyang/ginzero
```

Import it in your code:

```go
import "github.com/lixvyang/ginzero"
```

## Example

See the [example](_examples/example01/main.go).

```go
package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lixvyang/ginzero"
	"github.com/rs/zerolog"
)

func NewLogger() zerolog.Logger {
	logger := zerolog.New(os.Stdout).
		With().
		Caller().
		Timestamp().
		Logger()
	return logger
}

func main() {
	logger := NewLogger()
	r := gin.New()
	r.Use(ginzero.Ginzero(&logger), ginzero.RecoveryWithZero(&logger, true))

	r.GET("/hello", func(c *gin.Context) {
		c.String(200, "hello")
	})

	r.GET("/panic", func(c *gin.Context) {
		panic("panic msg.")
	})

	r.Run(":8002")
}
```

## Skip logging

When you want to skip logging for specific path,
please use GinzapWithConfig

```go

r.Use(GinzapWithConfig(utcLogger, &Config{
  TimeFormat: time.RFC3339,
  UTC: true,
  SkipPaths: []string{"/no_log"},
}))
```

## Custom Zap fields

example for custom log request body, response request ID or log [Open Telemetry](https://opentelemetry.io/) TraceID.

```go
func main() {
  r := gin.New()

  logger, _ := zap.NewProduction()

  r.Use(ginzap.GinzapWithConfig(logger, &ginzap.Config{
    UTC:        true,
    TimeFormat: time.RFC3339,
    Context: ginzap.Fn(func(c *gin.Context) []zapcore.Field {
      fields := []zapcore.Field{}
      // log request ID
      if requestID := c.Writer.Header().Get("X-Request-Id"); requestID != "" {
        fields = append(fields, zap.String("request_id", requestID))
      }

      // log trace and span ID
      if trace.SpanFromContext(c.Request.Context()).SpanContext().IsValid() {
        fields = append(fields, zap.String("trace_id", trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()))
        fields = append(fields, zap.String("span_id", trace.SpanFromContext(c.Request.Context()).SpanContext().SpanID().String()))
      }

      // log request body
      var body []byte
      var buf bytes.Buffer
      tee := io.TeeReader(c.Request.Body, &buf)
      body, _ = io.ReadAll(tee)
      c.Request.Body = io.NopCloser(&buf)
      fields = append(fields, zap.String("body", string(body)))

      return fields
    }),
  }))

  // Example ping request.
  r.GET("/ping", func(c *gin.Context) {
    c.Writer.Header().Add("X-Request-Id", "1234-5678-9012")
    c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
  })

  r.POST("/ping", func(c *gin.Context) {
    c.Writer.Header().Add("X-Request-Id", "9012-5678-1234")
    c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
  })

  // Listen and Server in 0.0.0.0:8080
  r.Run(":8080")
}
```