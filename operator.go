package main

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/gorilla/websocket"
	"goji.io/pat"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type listener struct {
	out map[chan Payload]struct{}
}

type operator struct {
	incoming chan Payload
	opened   chan request
	closed   chan request
}

type request struct {
	id  string
	out chan Payload
}

func (op *operator) run() {
	routes := map[string]*listener{}
	for {
		select {
		case req := <-op.incoming:
			log.Info("[operator] incoming",
				zap.Any("request", req))
			li, ok := routes[req.Destination]
			if !ok {
				li = &listener{
					out: map[chan Payload]struct{}{},
				}
				routes[req.Destination] = li
			}

			for c := range li.out {
				select {
				case c <- req:
				default:
					select {
					case <-c:
					default:
					}
					select {
					case c <- req:
					default:
					}
				}
			}
		case req := <-op.opened:
			log.Info("[operator] opened",
				zap.Any("request", req))
			existing, ok := routes[req.id]
			if !ok {
				existing = &listener{
					out: map[chan Payload]struct{}{},
				}
				routes[req.id] = existing
			}
			existing.out[req.out] = struct{}{}
		case req := <-op.closed:
			log.Info("[operator] closed",
				zap.Any("request", req))
			if existing, ok := routes[req.id]; ok {
				delete(existing.out, req.out)
			}
			close(req.out)
		}
	}
}

func (op *operator) webSocketOpen(w http.ResponseWriter, r *http.Request) {
	id := pat.Param(r, "id")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer conn.Close()

	c := make(chan Payload, 8)
	req := request{
		id:  id,
		out: c,
	}
	op.opened <- req
	defer func() {
		op.closed <- req
	}()

	errs := make(chan error, 2)
	go func() {
		for payload := range c {
			err := conn.WriteJSON(payload)
			if err != nil {
				errs <- err
				return
			}
		}
	}()
	go func() {
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				errs <- err
				return
			}
			if mt == websocket.BinaryMessage || mt == websocket.TextMessage {
				var payload Payload
				err = json.Unmarshal(msg, &payload)
				if err != nil {
					log.Error("invalid message",
						zap.String("message", string(msg)),
						zap.Error(err))
					continue
				}
				payload.Source = id
				payload.Time = time.Now()
				op.incoming <- payload
			}
		}
	}()
	<-errs
}
