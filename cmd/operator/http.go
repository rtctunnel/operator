package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/rs/cors"
	"github.com/rtctunnel/operator"
	"github.com/heptiolabs/healthcheck"
)

var (
	maxMessageSize = 128 * 1024
	timeout        = time.Second * 30
)

func runHTTP(li net.Listener) error {
	log.WithField("bind-addr", li.Addr()).
		Info("starting server")

	health := healthcheck.NewHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/pub", pub)
	mux.HandleFunc("/sub", sub)
	mux.HandleFunc("/healthz", health.ReadyEndpoint)

	handler := cors.Default().Handler(mux)
	return http.Serve(li, handler)
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
