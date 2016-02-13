package main

import (
	"io"

	"github.com/simulatedsimian/tracetcp-go/tracetcp"
)

type TraceOutputWriter interface {
	Init(port int, hopsFrom, hopsTo, queriesPerHop int, noLookups bool, out io.Writer)
	Event(e tracetcp.TraceEvent) error
}
