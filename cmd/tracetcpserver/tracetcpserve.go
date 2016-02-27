package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type flushWriter struct {
	f http.Flusher
	w io.Writer
}

func (fw *flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	//log.Printf("%s", p)
	if fw.f != nil {
		fw.f.Flush()
	}
	return
}

func editCommandHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "editcmd.html")
}

func validate(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) && r != '.' {
			return false
		}
	}
	return true
}

type traceConfig struct {
	host     string
	port     string
	starthop int
	endhop   int
	timeout  time.Duration
	queries  int
	nolookup bool
}

var defaultTraceConfig = traceConfig{
	host:     "",
	port:     "http",
	starthop: 1,
	endhop:   30,
	timeout:  1 * time.Second,
	queries:  3,
	nolookup: false,
}

func doTrace(w http.ResponseWriter, config *traceConfig) {
	fw := flushWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}

	err := execWithTimeout("tracetcp", makeCommandLine(config), &fw, mainConfig.TraceTimeout)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s\n", err)
	}
}

func makeCommandLine(config *traceConfig) []string {
	args := []string{}

	if config.nolookup {
		args = append(args, "-n")
	}

	args = append(args, "-h", fmt.Sprint(config.starthop))
	args = append(args, "-m", fmt.Sprint(config.endhop))
	args = append(args, "-p", fmt.Sprint(config.queries))
	args = append(args, "-t", fmt.Sprint(config.timeout))
	args = append(args, config.host+":"+config.port)

	return args
}

func validateConfig(config *traceConfig) error {

	if !validate(config.host) {
		return fmt.Errorf("Invalid Host Name")
	}

	if !validate(config.port) {
		return fmt.Errorf("Invalid Port Number")
	}

	if config.starthop < 1 {
		return fmt.Errorf("starthop must be > 1")
	}

	if config.endhop > 128 {
		return fmt.Errorf("endhop must be < 127")
	}

	if config.endhop < config.starthop {
		return fmt.Errorf("endhop must be >= starthop")
	}

	if config.queries < 1 {
		return fmt.Errorf("queries must be > 1")
	}

	if config.queries > 5 {
		return fmt.Errorf("queries must be <= 5")
	}

	if config.timeout > time.Second*3 {
		return fmt.Errorf("timeout must be <= 3s")
	}

	return nil
}

func parseRequest(config traceConfig, reader func(name string) (string, bool)) (*traceConfig, error) {

	if v, ok := reader("host"); ok {
		config.host = v
	}

	if v, ok := reader("port"); ok {
		config.port = v
	}

	var err error

	if v, ok := reader("starthop"); ok {
		config.starthop, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("Invalid Start Hop: %v", err)
		}
	}

	if v, ok := reader("endhop"); ok {
		config.endhop, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("Invalid End Hop: %v", err)
		}
	}

	if v, ok := reader("timeout"); ok {
		config.timeout, err = time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("Invalid Timeout Duration: %v", err)
		}
	}

	if v, ok := reader("queries"); ok {
		config.queries, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("Invalid Query Count: %v", err)
		}
	}

	return &config, nil
}

func doTraceHandler(w http.ResponseWriter, r *http.Request) {

	config, err := parseRequest(defaultTraceConfig, func(name string) (string, bool) {
		if v, ok := r.URL.Query()[name]; ok {
			return v[0], ok
		}
		return "", false
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	err = validateConfig(config)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: ", err)
	}

	doTrace(w, config)
}

func execHandler(w http.ResponseWriter, r *http.Request) {

	config := defaultTraceConfig

	config.host = r.FormValue("host")
	config.port = r.FormValue("port")

	if r.FormValue("source") == "ok" {
		config.host = r.RemoteAddr[:strings.Index(r.RemoteAddr, ":")]
	}

	doTrace(w, &config)
}
