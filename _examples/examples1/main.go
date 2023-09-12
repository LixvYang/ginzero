package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lixvyang/ginzero"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	r := gin.New()
	r.Use(ginzero.RecoveryWithZero(&logger, true))

	r.GET("/hello", func(c *gin.Context) {
		c.String(200, "hello")
	})

	r.GET("/world", func(c *gin.Context) {
		panic("123")
	})

	r.Run(":8002")
}
