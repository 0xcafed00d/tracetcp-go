package tracetcp

import (
	"fmt"
	"io"
)

type TraceOutputWriter interface {
	Init(port int, hopsFrom, hopsTo, queriesPerHop int, noLookups bool, out io.Writer)
	Event(e TraceEvent) error
}

var outputWriters = map[string]TraceOutputWriter{
	"std":  &StdTraceWriter{},
	"json": &JSONTraceWriter{},
}

func GetOutputWriter(name string) (TraceOutputWriter, error) {
	if writer, ok := outputWriters[name]; ok {
		return writer, nil
	}
	return nil, fmt.Errorf("Invalid output format name: %v", name)
}
