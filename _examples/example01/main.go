package main

import (
	"encoding/base64"
	"encoding/binary"
	"math/rand"
	"os"
	"time"

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

var pid = uint32(os.Getpid())

func GenerateRandnum() int {
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(2 << 31)
	return randNum
}

func defaultGenReqId() string {
	var b [16]byte

	binary.LittleEndian.PutUint32(b[:], pid)
	binary.LittleEndian.PutUint64(b[4:], uint64(time.Now().UnixNano()))
	binary.LittleEndian.PutUint32(b[12:], uint32(GenerateRandnum()))
	return base64.URLEncoding.EncodeToString(b[:])
}

func main() {
	logger := NewLogger()
	r := gin.New()
	r.Use(ginzero.Ginzero(&logger, ginzero.SkipPaths([]string{"/hello"}), ginzero.Genxid(defaultGenReqId)), ginzero.RecoveryWithZero(&logger, true))

	r.GET("/hello", func(c *gin.Context) {
		c.String(200, "hello1")
	})

	r.GET("/hello/2", func(c *gin.Context) {
		c.String(200, "hello2")
	})

	r.GET("/panic", func(c *gin.Context) {
		panic("panic msg.")
	})

	r.Run(":8002")
}
