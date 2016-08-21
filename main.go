package main

import (
	"log"
	"net/http"
	"os"

	"goji.io/pat"

	"goji.io"
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
	mux.HandleFuncC(pat.Get("/websocket-open/:id"), op.webSocketOpen)
	http.Handle("/", mux)
}

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:5000"
	}
	log.SetFlags(0)
	log.Println("listening on", addr)
	http.ListenAndServe(addr, nil)
}
