package main

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/rtctunnel/operator"
)

const (
	maxMessageSize = 10 * 1024
)

func runHTTP(addr string) error {
	log.WithField("local-addr", addr).
		Info("starting server")

	mux := http.NewServeMux()
	mux.HandleFunc("/pub", pub)
	mux.HandleFunc("/sub", sub)
	return http.ListenAndServe(addr, mux)
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

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
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

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	data, err := op.Sub(ctx, addr)
	if err == nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, data)
	} else {
		w.WriteHeader(http.StatusGatewayTimeout)
	}
}
