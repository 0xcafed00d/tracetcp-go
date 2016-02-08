package tracetcp

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"time"
)

type TraceEventType int

const (
	None TraceEventType = iota
	TimedOut
	TTLExpired
	Connected
	RemoteClosed
	TraceStarted
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
	case TraceStarted:
		return "TraceStarted"
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
	Events         chan TraceEvent
	TraceRunning   AtomicBool
	AbortRequested AtomicBool
}

func NewTrace() *Trace {
	t := Trace{}

	t.Events = make(chan TraceEvent, 100)

	return &t
}

func (t *Trace) BeginTrace(addr *net.IPAddr, port, beginTTL, endTTL, queries int, timeout time.Duration) error {
	if t.TraceRunning.Read() {
		return fmt.Errorf("Trace already in progress")
	}
	t.TraceRunning.Write(true)
	go t.traceImpl(addr, port, beginTTL, endTTL, queries, timeout)
	return nil
}

func (t *Trace) AbortTrace() {
	t.AbortRequested.Write(true)
}

func (t *Trace) traceImpl(addr *net.IPAddr, port, beginTTL, endTTL, queries int, timeout time.Duration) {

	icmpChan := make(chan icmpEvent, 100)
	go receiveICMP(icmpChan)

	traceStart := time.Now()
	t.Events <- TraceEvent{Type: TraceStarted, Time: time.Since(traceStart)}
	for ttl := beginTTL; ttl <= endTTL; ttl++ {
		for q := 0; q < queries; q++ {
			log.Printf("Probe query: %v hops: %v", q, ttl)
			queryStart := time.Now()
			ev := tryConnect(*addr, port, ttl, q, timeout)
			if t.correlateEvents(ev, icmpChan, queryStart) {
				t.Events <- TraceEvent{Type: TraceComplete, Time: time.Since(traceStart)}
				return
			}
		}
	}
	t.Events <- TraceEvent{Type: TraceComplete, Time: time.Since(traceStart)}
	t.TraceRunning.Write(false)
}

func (t *Trace) correlateEvents(ev connectEvent, icmpChan chan icmpEvent, queryStart time.Time) bool {

	icmpev := icmpEvent{}

	// collect all pending icmp events
	done := false
	for !done {
		select {
		case iev := <-icmpChan:
			if reflect.DeepEqual(iev.localAddr, ev.localAddr) && iev.localPort == ev.localPort {
				done = true
				icmpev = iev
			}
		case <-time.After(100 * time.Millisecond):
			done = true
		}
	}
	log.Println(ev)
	if icmpev.evtype == icmpNone {
		log.Println("No matching ICMP event")
	} else {
		log.Println("matching icmp event", icmpev)
	}

	traceEvent := TraceEvent{
		Hop:   ev.ttl,
		Query: ev.query,
		Time:  time.Since(queryStart),
	}

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
