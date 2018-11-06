package operator

import (
	"context"
	"sync"
)

type opResult struct {
	data string
	err  error
}
type opWaiter struct {
	addr string
	ch   chan string
}
type op struct {
	ctx    context.Context
	addr   string
	data   string
	result chan opResult
}

// An Operator facilitates communication between two or more peers.
type Operator struct {
	pubc   chan op
	unpubc chan opWaiter
	subc   chan op
	unsubc chan opWaiter

	pendingPubs ChannelCollection
	pendingSubs ChannelCollection

	once   sync.Once
	closer chan struct{}
}

// New creates a new Operator.
func New() *Operator {
	o := &Operator{
		pubc:   make(chan op),
		unpubc: make(chan opWaiter),
		subc:   make(chan op),
		unsubc: make(chan opWaiter),

		pendingPubs: make(ChannelCollection),
		pendingSubs: make(ChannelCollection),

		closer: make(chan struct{}),
	}
	go o.run()
	return o
}

// Close closes the operator.
func (o *Operator) Close() error {
	o.once.Do(func() {
		close(o.closer)
	})
	return nil
}

// Pub publishes a message.
func (o *Operator) Pub(ctx context.Context, addr, data string) error {
	c := make(chan opResult, 1)
	select {
	case o.pubc <- op{ctx: ctx, addr: addr, data: data, result: c}:
	case <-o.closer:
		return context.Canceled
	}

	select {
	case r := <-c:
		return r.err
	case <-o.closer:
		return context.Canceled
	}
}

// Sub subcribes to messages.
func (o *Operator) Sub(ctx context.Context, addr string) (string, error) {
	c := make(chan opResult, 1)
	select {
	case o.subc <- op{ctx: ctx, addr: addr, result: c}:
	case <-o.closer:
		return "", context.Canceled
	}

	select {
	case r := <-c:
		return r.data, r.err
	case <-o.closer:
		return "", context.Canceled
	}
}

func (o *Operator) run() {
	for {
		select {
		case op := <-o.pubc:
			_ = o.pubNow(op.addr, op.data, op.result) || o.pubWait(op.ctx, op.addr, op.data, op.result)
		case opw := <-o.unpubc:
			_ = o.unpubNow(opw.addr, opw.ch)

		case op := <-o.subc:
			_ = o.subNow(op.addr, op.result) || o.subWait(op.ctx, op.addr, op.result)
		case opw := <-o.unsubc:
			_ = o.unsubNow(opw)

		case <-o.closer:
			return
		}
	}
}

func (o *Operator) pubNow(addr, data string, result chan opResult) bool {
	for _, c := range o.pendingSubs.List(addr) {
		select {
		case c <- data:
			result <- opResult{}
			return true
		default:
		}
	}
	return false
}

func (o *Operator) pubWait(ctx context.Context, addr, data string, result chan opResult) bool {
	c := make(chan string)
	o.pendingPubs.Add(addr, c)

	go func() {
		select {
		case <-ctx.Done():
			result <- opResult{err: ctx.Err()}
		case c <- data:
			result <- opResult{}
		}

		select {
		case o.unpubc <- opWaiter{addr: addr, ch: c}:
		case <-o.closer:
		}
	}()

	return true
}

func (o *Operator) unpubNow(addr string, c chan string) bool {
	o.pendingPubs.Remove(addr, c)
	return true
}

func (o *Operator) subNow(addr string, result chan opResult) bool {
	for _, c := range o.pendingPubs.List(addr) {
		select {
		case data := <-c:
			result <- opResult{data: data}
			return true
		default:
		}
	}
	return false
}

func (o *Operator) subWait(ctx context.Context, addr string, result chan opResult) bool {
	c := make(chan string)
	o.pendingSubs.Add(addr, c)
	go func() {
		select {
		case <-ctx.Done():
			result <- opResult{err: ctx.Err()}
		case data := <-c:
			result <- opResult{data: data}
		}

		select {
		case o.unsubc <- opWaiter{addr: addr, ch: c}:
		case <-o.closer:
		}
	}()

	return true
}

func (o *Operator) unsubNow(opw opWaiter) bool {
	o.pendingSubs.Remove(opw.addr, opw.ch)
	return true
}
