package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
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

func doTrace(w http.ResponseWriter, host, port string) {
	fw := flushWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}

	if !validate(host) {
		fmt.Fprint(w, "Invalid Host Name")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !validate(port) {
		fmt.Fprint(w, "Invalid Port Number")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cmd := exec.Command("tracetcp")
	cmd.Stdout = &fw
	cmd.Stderr = &fw

	cmd.Args = append(cmd.Args, host+":"+port)

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(w, "%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func doTraceHandler(w http.ResponseWriter, r *http.Request) {

}

func execHandler(w http.ResponseWriter, r *http.Request) {

	host := r.FormValue("host")
	port := r.FormValue("port")

	if r.FormValue("source") == "ok" {
		host = r.RemoteAddr[:strings.Index(r.RemoteAddr, ":")]
	}

	doTrace(w, host, port)
}

func main() {
	http.HandleFunc("/editcmd/", editCommandHandler)
	http.HandleFunc("/exec/", execHandler)
	http.HandleFunc("/dotrace/", dotraceHandler)
	http.ListenAndServe(":8080", nil)
}
