package main

import (
	"fmt"
	"github.com/simulatedsimian/tracetcp-go/tracetcp"
	"time"
)

func main() {
	ip, err := tracetcp.LookupAddress("www.google.com")
	fmt.Println(ip, err)

	trace := tracetcp.NewTrace()

	trace.BeginTrace(ip, 80, 1, 40, 3, 1*time.Second)

	fmt.Println(<-trace.Events)

}
