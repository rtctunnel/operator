package main

import (
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"

	"goji.io"
	"goji.io/pat"
)

func init() {
	op := &operator{
		incoming: make(chan Payload),
		opened:   make(chan request),
		closed:   make(chan request),
	}
	go op.run()

	mux := goji.NewMux()
	mux.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			h.ServeHTTP(w, r)
		})
	})
	mux.HandleFunc(pat.Get("/websocket-open/:id"), op.webSocketOpen)
	http.Handle("/", mux)
}

var log *zap.Logger

func main() {
	var err error

	log, err = zap.NewDevelopment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v", err)
		os.Exit(1)
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:5000"
	}
	log.Info("config",
		zap.String("ADDR", addr))

	log.Info("starting listener",
		zap.String("address", addr))
	http.ListenAndServe(addr, nil)
}
