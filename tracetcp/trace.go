package tracetcp

import (
	"fmt"
	"net"
	"time"
)

type TraceEventType int

const (
	TimedOut TraceEventType = iota
	TTLExpired
	Connected
	TraceComplete
	TraceAborted
	TraceFailed
)

func (t TraceEventType) String() string {
	switch t {
	case TimedOut:
		return "TimedOut"
	case TTLExpired:
		return "TTLExpired"
	case Connected:
		return "Connected"
	case TraceComplete:
		return "TraceComplete"
	case TraceAborted:
		return "TraceAborted"
	case TraceFailed:
		return "TraceFailed"
	}
	return "Invalid TraceEventType"
}

type TraceEvent struct {
	Type  TraceEventType
	Addr  net.IPAddr
	Time  time.Duration
	Hop   int
	Query int
	Err   error
}

type Trace struct {
	Events       chan TraceEvent
	errors       chan error
	abortChan    chan bool
	TraceRunning bool
}

func NewTrace() *Trace {
	t := Trace{}

	t.Events = make(chan TraceEvent)
	t.errors = make(chan error)
	t.abortChan = make(chan bool)

	return &t
}

func (t *Trace) BeginTrace(addr *net.IPAddr, port, beginTTL, endTTL, queries int, timeout time.Duration) error {
	if t.TraceRunning {
		return fmt.Errorf("Trace already in progress")
	}
	t.TraceRunning = true
	go t.traceImpl(addr, port, beginTTL, endTTL, queries, timeout)
	return nil
}

func (t *Trace) AbortTrace() {

}

func (t *Trace) traceImpl(addr *net.IPAddr, port, beginTTL, endTTL, queries int, timeout time.Duration) {

	traceStart := time.Now()
	implEvents := make(chan implTraceEvent)

	go tryConnect(*addr, port, beginTTL, 0, timeout, implEvents)
	fmt.Println(<-implEvents)
    fmt.Println(<-implEvents)

	traceTime := time.Since(traceStart)
	t.Events <- TraceEvent{Type: TraceComplete, Time: traceTime}
}
