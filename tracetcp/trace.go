package tracetcp

import (
	"fmt"
	"net"
	"time"
)

type TraceEventType int

const (
	None TraceEventType = iota
	TimedOut
	TTLExpired
	Connected
	RemoteClosed
	TraceComplete
	TraceAborted
	TraceFailed
)

// implementation of fmt.Stinger interface
func (t TraceEventType) String() string {
	switch t {
	case None:
		return "None"
	case TimedOut:
		return "TimedOut"
	case TTLExpired:
		return "TTLExpired"
	case Connected:
		return "Connected"
	case RemoteClosed:
		return "RemoteClosed"
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

// implementation of fmt.Stinger interface
func (e TraceEvent) String() string {
	return fmt.Sprintf("TraceEvent:{type: %v, addr: %v, timetaken: %v, hop: %d, query %d, err: %v}",
		e.Type, e.Addr, e.Time, e.Hop, e.Query, e.Err)
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
	icmpChan := make(chan icmpEvent)

	go receiveICMP(icmpChan)

	for ttl := beginTTL; ttl <= endTTL; ttl++ {
		for q := 0; q < queries; q++ {
			queryStart := time.Now()
			ev := tryConnect(*addr, port, ttl, q, timeout)
			if t.colate(ev, icmpChan, queryStart) {
				t.Events <- TraceEvent{Type: TraceComplete, Time: time.Since(traceStart)}
				return
			}
		}
	}
	t.Events <- TraceEvent{Type: TraceComplete, Time: time.Since(traceStart)}
}

func (t *Trace) colate(ev connectEvent, icmpChan chan icmpEvent, queryStart time.Time) bool {
	icmpev := icmpEvent{}

	select {
	case ev := <-icmpChan:
		icmpev = ev
	case <-time.After(50 * time.Millisecond):
	}

	traceEvent := TraceEvent{
		Hop:   ev.ttl,
		Query: ev.query,
		Time:  time.Since(queryStart),
	}

	//	fmt.Println(icmpev)
	//	fmt.Println(ev)

	if ev.evtype == connectError {
		traceEvent.Type = TraceFailed
		traceEvent.Err = ev.err
		t.Events <- traceEvent
		return true
	}

	if icmpev.evtype == icmpError {
		traceEvent.Type = TraceFailed
		traceEvent.Err = icmpev.err
		t.Events <- traceEvent
		return true
	}

	if icmpev.evtype == icmpTTLExpired {
		traceEvent.Type = TTLExpired
		traceEvent.Addr = icmpev.remoteAddr
		t.Events <- traceEvent
		return false
	}

	if ev.evtype == connectConnected {
		traceEvent.Type = Connected
		traceEvent.Addr = ev.remoteAddr
		t.Events <- traceEvent
		return true
	}

	if ev.evtype == connectTimedOut {
		traceEvent.Type = TimedOut
		t.Events <- traceEvent
		return false
	}

	if ev.evtype == connectFailed {
		traceEvent.Type = RemoteClosed
		t.Events <- traceEvent
		return false
	}

	panic("should not get here???")

	return false
}
