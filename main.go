package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/simulatedsimian/tracetcp-go/tracetcp"
)

type Config struct {
	Help      bool
	Timeout   time.Duration
	NoLookups bool
	StartHop  int
	EndHop    int
	Queries   int
	Verbose   bool
}

var config Config

type TraceOutputWriter interface {
	Init(port int, queriesPerHop int, noLookups bool, out io.Writer)
	Event(e tracetcp.TraceEvent) error
}

type StdTraceWriter struct {
	port          int
	queriesPerHop int
	noLooups      bool
	out           io.Writer
	currentHop    int
	currentAddr   net.IPAddr
}

func (w *StdTraceWriter) Init(port int, queriesPerHop int, noLookups bool, out io.Writer) {
	w.port = port
	w.queriesPerHop = queriesPerHop
	w.noLooups = noLookups
	w.out = out
	w.currentHop = 0
}

func (w *StdTraceWriter) Event(e tracetcp.TraceEvent) error {

	if e.Hop != 0 && w.currentHop != e.Hop {
		w.currentHop = e.Hop
		fmt.Fprintf(w.out, "\n%v\t", e.Hop)
	}

	switch e.Type {
	case tracetcp.TraceStarted:
		fmt.Fprintf(w.out, "Tracing route to %v on port %v over a maximum of %v hops:\n", e.Addr.IP, w.port, 10)
	case tracetcp.TimedOut:
		fmt.Fprintf(w.out, "%8v", "*")
	case tracetcp.TTLExpired:
		fmt.Fprintf(w.out, "%8v", (e.Time/time.Millisecond)*time.Millisecond)
	case tracetcp.Connected:
		fmt.Fprintf(w.out, "%8v", "connected")
	case tracetcp.RemoteClosed:
		fmt.Fprintf(w.out, "%8v", "Port closed")

	}

	if e.Query == w.queriesPerHop-1 {
		fmt.Fprintf(w.out, "\t%v", e.Addr)
	}

	return nil

}

func init() {
	flag.BoolVar(&config.Help, "?", false, "display help")
	flag.DurationVar(&config.Timeout, "t", time.Second, "probe reply timeout")
	flag.BoolVar(&config.NoLookups, "n", false, "no reverse DNS lookups")
	flag.IntVar(&config.StartHop, "h", 1, "start hop")
	flag.IntVar(&config.EndHop, "m", 30, "max hops")
	flag.IntVar(&config.Queries, "p", 3, "pings per hop")
	flag.BoolVar(&config.Verbose, "v", false, "verbose output")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tracetcp-go [options] hostname[:port]")
		flag.PrintDefaults()
	}
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 && config.Help {
		flag.Usage()
		os.Exit(1)
	}

	if len(flag.Args()) != 1 {
		fmt.Fprintln(os.Stderr, "Host not suplied\n")
		flag.Usage()
		os.Exit(1)
	}

	host, port, err := tracetcp.SplitHostAndPort(flag.Args()[0], 80)
	fmt.Println(host, port, err)
	exitOnError(err)
	if port < 0 {
		port = 80
	}

	ip, err := tracetcp.LookupAddress(host)
	fmt.Println(ip, err)

	trace := tracetcp.NewTrace()

	trace.BeginTrace(ip, port, config.StartHop, config.EndHop, config.Queries, config.Timeout)

	if !config.Verbose {
		log.SetOutput(ioutil.Discard)
	}

	var writer TraceOutputWriter = &StdTraceWriter{}
	writer.Init(port, config.Queries, config.NoLookups, os.Stdout)

	for {
		ev := <-trace.Events
		writer.Event(ev)

		if config.Verbose {
			fmt.Println(ev)
		}
		if ev.Type == tracetcp.TraceComplete {
			break
		}
	}
}
