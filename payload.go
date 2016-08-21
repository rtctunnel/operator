package main

import (
	"encoding/json"
	"time"
)

// A Payload is a single message handled by the operator
type Payload struct {
	Destination string    `json:"destination"`
	Source      string    `json:"source"`
	Time        time.Time `json:"time"`
	Message     string    `json:"message"`
}

func (p Payload) String() string {
	bs, _ := json.Marshal(p)
	return string(bs)
}
