package main

import (
	"sync"
	"time"
)

// A Payload is a single message handled by the operator
type Payload struct {
	Destination string    `json:"destination"`
	Source      string    `json:"source"`
	Time        time.Time `json:"time"`
	Message     string    `json:"message"`
}

// A Buffer contains messages and a last update time
type Buffer struct {
	lastUpdate time.Time
	payloads   []Payload
}

// A Relay contains messages for sending and receiving
type Relay struct {
	sync.Mutex
	maxMessages int

	buffers map[string]Buffer
}

// NewRelay creates a new Relay
func NewRelay() *Relay {
	return &Relay{
		maxMessages: 32,
		buffers:     make(map[string]Buffer),
	}
}

// Recv receives any payloads for the destination
func (r *Relay) Recv(dst string) []Payload {
	r.Lock()
	defer r.Unlock()

	buf := r.buffers[dst]
	delete(r.buffers, dst)
	return buf.payloads
}

// Send sends any payloads to the destination
func (r *Relay) Send(payload Payload) {
	r.Lock()
	defer r.Unlock()

	buf := r.buffers[payload.Destination]
	if len(buf.payloads) > r.maxMessages {
		copy(buf.payloads, buf.payloads[1:])
		buf.payloads[len(buf.payloads)-1] = payload
	} else {
		buf.payloads = append(buf.payloads, payload)
	}
	r.buffers[payload.Destination] = Buffer{
		lastUpdate: time.Now(),
		payloads:   buf.payloads,
	}
}

// Clean removes any payload buffers older than the cutoff
func (r *Relay) Clean(cutoff time.Time) {
	r.Lock()
	defer r.Unlock()

	for dst, buf := range r.buffers {
		if buf.lastUpdate.Before(cutoff) {
			delete(r.buffers, dst)
		}
	}
}
