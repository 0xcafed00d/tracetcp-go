package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
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
	log.Printf("%s", p)
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

func execHandler(w http.ResponseWriter, r *http.Request) {
	fw := flushWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}

	host := r.FormValue("host")
	port := r.FormValue("port")

	if r.FormValue("source") == "ok" {
		host = r.RemoteAddr[:strings.Index(r.RemoteAddr, ":")]
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

	cmd.Args = append(cmd.Args, host)
	cmd.Args = append(cmd.Args, port)

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(w, "%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	fw := flushWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}

	fmt.Fprint(w, "[[[[[output start]]]]]\n")
	for n := 0; n < 30; n++ {
		fmt.Fprintf(&fw, "**************  OUTPUT LINE: %d ******************************************************************\n", n)
		time.Sleep(1000 * time.Millisecond)
	}
	fmt.Fprint(w, "[[[[[output complete]]]]]\n")
}

func main() {
	http.HandleFunc("/editcmd/", editCommandHandler)
	http.HandleFunc("/exec/", execHandler)
	http.HandleFunc("/test/", testHandler)
	http.ListenAndServe(":8080", nil)
}
