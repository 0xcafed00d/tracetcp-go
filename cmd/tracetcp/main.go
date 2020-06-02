package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/0xcafed00d/tracetcp-go/tracetcp"
)

type Config struct {
	Help         bool
	Timeout      time.Duration
	NoLookups    bool
	StartHop     int
	EndHop       int
	Queries      int
	Verbose      bool
	OutputWriter string
}

var config Config

func init() {
	flag.BoolVar(&config.Help, "?", false, "display help")
	flag.DurationVar(&config.Timeout, "t", time.Second, "probe reply timeout")
	flag.BoolVar(&config.NoLookups, "n", false, "no reverse DNS lookups")
	flag.IntVar(&config.StartHop, "h", 1, "start hop")
	flag.IntVar(&config.EndHop, "m", 30, "max hops")
	flag.IntVar(&config.Queries, "p", 3, "pings per hop")
	flag.BoolVar(&config.Verbose, "v", false, "verbose output")
	flag.StringVar(&config.OutputWriter, "o", "std", "output format: [std|json]")

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

// Linux to open raw sockets without running as root: sudo setcap cap_net_raw=ep tracetcp
func main() {
	flag.Parse()

	if len(flag.Args()) == 0 && config.Help {
		flag.Usage()
		os.Exit(1)
	}

	if len(flag.Args()) != 1 {
		fmt.Fprintln(os.Stderr, "Host not suplied")
		fmt.Fprintln(os.Stderr, "")
		flag.Usage()
		os.Exit(1)
	}

	host, port, err := tracetcp.SplitHostAndPort(flag.Args()[0], 80)
	exitOnError(err)

	ip, err := tracetcp.LookupAddress(host)
	exitOnError(err)

	trace := tracetcp.NewTrace()
	trace.BeginTrace(ip, port, config.StartHop, config.EndHop, config.Queries, config.Timeout)

	if !config.Verbose {
		log.SetOutput(ioutil.Discard)
	}

	writer, err := tracetcp.GetOutputWriter(config.OutputWriter)
	exitOnError(err)

	writer.Init(port, config.StartHop, config.EndHop, config.Queries, config.NoLookups, os.Stdout)

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
