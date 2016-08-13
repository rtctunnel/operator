package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"google.golang.org/appengine/channel"

	"goji.io/pat"
	"golang.org/x/net/context"

	"goji.io"
)

// An Operator handles signalling between users
type Operator struct {
	relays []*Relay
}

// Send sends a message to another user
func (op *Operator) Send(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	dst := pat.Param(ctx, "dst")
	src := pat.Param(ctx, "src")
	msg, err := ioutil.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	log.Println("send", string(msg), "from", src, "to", dst)

	channel.SendJSON

	op.relays[0].Send(Payload{
		Destination: dst,
		Source:      src,
		Time:        time.Now(),
		Message:     string(msg),
	})

	json.NewEncoder(w).Encode("OK")
}

// Recv receives a message from another user
func (op *Operator) Recv(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	dst := pat.Param(ctx, "dst")

	msgs := op.relays[0].Recv(dst)
	json.NewEncoder(w).Encode(msgs)
}

func init() {
	op := &Operator{
		relays: []*Relay{NewRelay()},
	}
	go func() {
		for now := range time.Tick(15 * time.Second) {
			cutoff := now.Add(-time.Minute * 10)
			for _, relay := range op.relays {
				relay.Clean(cutoff)
			}
		}
	}()

	mux := goji.NewMux()
	mux.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			h.ServeHTTP(w, r)
		})
	})
	mux.HandleFuncC(pat.Post("/send/:dst/:src"), op.Send)
	mux.HandleFuncC(pat.Get("/recv/:dst"), op.Recv)
	http.Handle("/", mux)
}
