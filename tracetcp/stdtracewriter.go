package tracetcp

import (
	"fmt"
	"io"
	"net"
	"time"
)

type StdTraceWriter struct {
	port          int
	hopsFrom      int
	hopsTo        int
	queriesPerHop int
	noLooups      bool
	out           io.Writer
	currentHop    int
	currentAddr   *net.IPAddr
}

func (w *StdTraceWriter) Init(port int, hopsFrom, hopsTo, queriesPerHop int, noLookups bool, out io.Writer) {
	w.port = port
	w.hopsFrom = hopsFrom
	w.hopsTo = hopsTo
	w.queriesPerHop = queriesPerHop
	w.noLooups = noLookups
	w.out = out
	w.currentHop = 0
}

func (w *StdTraceWriter) Event(e TraceEvent) error {

	if e.Type == TraceFailed {
		fmt.Fprintf(w.out, "Error: %v\n", e.Err)
		return e.Err
	}

	if e.Hop != 0 && w.currentHop != e.Hop {
		w.currentHop = e.Hop
		fmt.Fprintf(w.out, "\n%-3v", e.Hop)
		w.currentAddr = nil
	}

	switch e.Type {
	case TraceStarted:
		var revhost string
		if !w.noLooups {
			revhost, _ = ReverseLookup(e.Addr)
		}
		if revhost != "" {
			fmt.Fprintf(w.out, "Tracing route to %v (%v) on port %v over a maximum of %v hops:\n",
				e.Addr.IP, revhost, w.port, w.hopsTo)
		} else {
			fmt.Fprintf(w.out, "Tracing route to %v on port %v over a maximum of %v hops:\n",
				e.Addr.IP, w.port, w.hopsTo)
		}

	case TimedOut:
		fmt.Fprintf(w.out, "%8v", "*")
	case TTLExpired:
		w.currentAddr = &e.Addr
		fmt.Fprintf(w.out, "%8v", (e.Time/time.Millisecond)*time.Millisecond)
	case Connected:
		fmt.Fprintf(w.out, "Connected to %v on port %v\n", e.Addr.String(), w.port)
	case RemoteClosed:
		fmt.Fprintf(w.out, "Port %v closed at %v\n", w.port, e.Addr.String())
	}

	if e.Query == w.queriesPerHop-1 && w.currentAddr != nil {
		name, _ := ReverseLookup(*w.currentAddr)
		if name == "" || w.noLooups {
			fmt.Fprintf(w.out, "\t%v", w.currentAddr.String())
		} else {
			fmt.Fprintf(w.out, "\t%v (%v)", name, w.currentAddr.String())
		}
	}

	return nil
}
