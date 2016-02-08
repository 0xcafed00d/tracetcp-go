package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/simulatedsimian/tracetcp-go/tracetcp"
)

func main() {
	ip, err := tracetcp.LookupAddress("www.ebay.com")
	fmt.Println(ip, err)

	trace := tracetcp.NewTrace()

	trace.BeginTrace(ip, 80, 1, 25, 1, 1*time.Second)

	log.SetOutput(ioutil.Discard)

	for {
		ev := <-trace.Events
		fmt.Println(ev)
		if ev.Type == tracetcp.TraceComplete {
			break
		}
	}
}
