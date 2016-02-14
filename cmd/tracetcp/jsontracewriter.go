package main

import (
	"encoding/json"
	"io"
	"net"

	"github.com/simulatedsimian/tracetcp-go/tracetcp"
)

type JSONTraceWriter struct {
	port          int
	hopsFrom      int
	hopsTo        int
	queriesPerHop int
	noLooups      bool
	out           io.Writer
	currentHop    int
	currentAddr   *net.IPAddr

	jsonData []tracetcp.TraceEvent
}

func (w *JSONTraceWriter) Init(port int, hopsFrom, hopsTo, queriesPerHop int, noLookups bool, out io.Writer) {
	w.port = port
	w.hopsFrom = hopsFrom
	w.hopsTo = hopsTo
	w.queriesPerHop = queriesPerHop
	w.noLooups = noLookups
	w.out = out
	w.currentHop = 0
}

func (w *JSONTraceWriter) Event(e tracetcp.TraceEvent) error {

	w.jsonData = append(w.jsonData, e)

	if e.Type == tracetcp.TraceComplete {
		jsonenc := json.NewEncoder(w.out)
		jsonenc.Encode(w.jsonData)
	}

	return nil
}
