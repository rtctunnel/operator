package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/heptiolabs/healthcheck"
	"github.com/rtctunnel/operator"
)

var (
	maxMessageSize = 128 * 1024
	timeout        = time.Second * 30
)

func runHTTP(li net.Listener) error {
	log.WithField("bind-addr", li.Addr()).
		Info("starting server")

	health := healthcheck.NewHandler()

	r := chi.NewRouter()

	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT"},
	})
	r.Use(cors.Handler)

	r.Head("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("RTCTunnel Operator"))
	})
	r.Get("/pub", pub)
	r.Post("/pub", pub)
	r.Get("/sub", sub)
	r.Post("/sub", sub)
	r.Get("/healthz", health.ReadyEndpoint)

	return http.Serve(li, r)
}

var op = operator.New()

func pub(w http.ResponseWriter, r *http.Request) {
	addr := r.FormValue("address")
	data := r.FormValue("data")
	if len(data) > maxMessageSize {
		log.WithField("remote-addr", r.RemoteAddr).
			WithField("addr", addr).
			Warn("data too large")
		http.Error(w, "data too large", http.StatusBadRequest)
		return
	}

	log.WithField("remote-addr", r.RemoteAddr).
		WithField("addr", addr).
		WithField("data", data).
		Info("pub")

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	err := op.Pub(ctx, addr, data)
	if err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusGatewayTimeout)
	}
}

func sub(w http.ResponseWriter, r *http.Request) {
	addr := r.FormValue("address")

	log.WithField("remote-addr", r.RemoteAddr).
		WithField("addr", addr).
		Info("sub")

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	data, err := op.Sub(ctx, addr)
	if err == nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, data)
	} else {
		w.WriteHeader(http.StatusGatewayTimeout)
	}
}
