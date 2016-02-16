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

	if e.Hop != 0 && w.currentHop != e.Hop {
		w.currentHop = e.Hop
		fmt.Fprintf(w.out, "\n%-3v", e.Hop)
		w.currentAddr = nil
	}

	switch e.Type {
	case TraceStarted:
		fmt.Fprintf(w.out, "Tracing route to %v on port %v over a maximum of %v hops:\n",
			e.Addr.IP, w.port, w.hopsTo)
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
		addrstr := w.currentAddr.String()

		var names []string
		var err error
		if !w.noLooups {
			names, err = net.LookupAddr(addrstr)
		}
		if err == nil && len(names) > 0 {
			fmt.Fprintf(w.out, "\t%v (%v)", names[0], addrstr)
		} else {
			fmt.Fprintf(w.out, "\t%v", addrstr)
		}
	}

	return nil
}
