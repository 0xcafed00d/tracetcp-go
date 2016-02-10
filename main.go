package main

import (
	"flag"
	"fmt"
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

	fmt.Println(net.LookupSRV("", "", "http"))

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

	for {
		ev := <-trace.Events
		fmt.Println(ev)
		if ev.Type == tracetcp.TraceComplete {
			break
		}
	}
}
