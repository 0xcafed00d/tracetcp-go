package main

import (
	"fmt"
	"time"

	"github.com/simulatedsimian/tracetcp-go/tracetcp"
)

func main() {
	ip, err := tracetcp.LookupAddress("www.google.com")
	fmt.Println(ip, err)

	trace := tracetcp.NewTrace()

	trace.BeginTrace(ip, 89, 1, 25, 3, 1*time.Second)

	for {
		ev := <-trace.Events
		fmt.Println(ev)
		if ev.Type == tracetcp.TraceComplete {
			break
		}
	}
}
