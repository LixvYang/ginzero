package main

import (
	"encoding/base64"
	"encoding/binary"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lixvyang/ginzero"
	"github.com/rs/zerolog"
)

func NewReqId() string {
	var pid = uint32(time.Now().UnixNano() % (2 << 31))
	var b [12]byte
	binary.LittleEndian.PutUint32(b[:], pid)
	binary.LittleEndian.PutUint64(b[4:], uint64(time.Now().UnixNano()))
	return base64.URLEncoding.EncodeToString(b[:])
}

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
