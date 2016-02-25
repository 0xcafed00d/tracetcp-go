package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Help             bool
	TraceTimeout     time.Duration
	ListenPort       int
	ConcurrentTraces int
}

var config Config

func init() {
	flag.BoolVar(&config.Help, "?", false, "display help")
	flag.DurationVar(&config.TraceTimeout, "t", time.Second*30, "max time allowed for a trace")
	flag.IntVar(&config.ListenPort, "p", 80, "http listen port")
	flag.IntVar(&config.ConcurrentTraces, "c", 30, "max concurrent traces")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tracetcpserver [options]")
		flag.PrintDefaults()
	}
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Linux: to listen on ports <1024: sudo setcap cap_net_bind_service=+ep tracetcpserver
func main() {
	flag.Parse()

	if config.Help {
		flag.Usage()
		os.Exit(1)
	}

	http.HandleFunc("/editcmd/", editCommandHandler)
	http.HandleFunc("/exec/", execHandler)
	http.HandleFunc("/dotrace/", doTraceHandler)

	log.Printf("Listening on port %d", config.ListenPort)

	err := http.ListenAndServe(fmt.Sprintf(":%d", config.ListenPort), nil)
	exitOnError(err)
}
