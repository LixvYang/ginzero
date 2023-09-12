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
r.Use(ginzero.Ginzero(&logger, ginzero.WithSkipPaths([]string{"/hello"})))
```
