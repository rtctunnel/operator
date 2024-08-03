package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/heptiolabs/healthcheck"
	"github.com/rtctunnel/operator"
)

var (
	maxMessageSize = 128 * 1024
	timeout        = time.Second * 30
)

func runHTTP(addr, key, cert string) error {
	log.WithField("bind-addr", addr).
		WithField("key", key).
		WithField("cert", cert).
		Info("starting server")

	health := healthcheck.NewHandler()
	useHttps := len(key) != 0 && len(cert) != 0

	gin.SetMode(gin.ReleaseMode)
	gins := gin.New()

	gins.GET("/", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/plain")
		_, _ = c.Writer.Write([]byte("RTCTunnel Operator"))
	})

	gins.POST("/pub", pub)
	gins.GET("/pub", pub)
	gins.POST("/sub", sub)
	gins.GET("/sub", sub)
	gins.GET("/healthz",
		func(c *gin.Context) {
			health.ReadyEndpoint(c.Writer, c.Request)
		})

	gins.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT"},
		// AllowHeaders:     []string{"Origin"},
		// ExposeHeaders:    []string{"Content-Length"},
		// AllowCredentials: true,
		// AllowOriginFunc: func(origin string) bool {
		//   return origin == "https://github.com"
		// },
		MaxAge: 1 * time.Minute,
	}))

	var err error
	if useHttps {
		err = gins.RunTLS(
			addr,
			cert,
			key)
	} else {
		err = gins.Run(addr)
	}

	if errors.Is(err, http.ErrServerClosed) {
	} else if err != nil {
		return err
	}
	return nil

}

var op = operator.New()

func pub(c *gin.Context) {
	addr := c.Request.FormValue("address")
	data := c.Request.FormValue("data")
	if len(data) > maxMessageSize {
		log.WithField("remote-addr", c.Request.RemoteAddr).
			WithField("addr", addr).
			Warn("data too large")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "data too large",
		})
		return
	}
}

func sub(c *gin.Context) {
	addr := c.Request.FormValue("address")

	log.WithField("remote-addr", c.Request.RemoteAddr).
		WithField("addr", addr).
		Info("sub")

	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	data, err := op.Sub(ctx, addr)
	if err == nil {
		c.Writer.Header().Set("Content-Type", "text/plain")
		io.WriteString(c.Writer, data)
	} else {
		c.Writer.WriteHeader(http.StatusGatewayTimeout)
	}
}
