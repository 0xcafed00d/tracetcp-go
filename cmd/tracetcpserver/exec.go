package main

import (
	"errors"
	"io"
	"os/exec"
	"time"
)

var ErrTimeout = errors.New("exec timeout")

func execWithTimeout(proc string, args []string, out io.Writer, timeout time.Duration) error {

	cmd := exec.Command(proc)
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.Args = args

	c := make(chan error)
	go func(c chan error) {
		c <- cmd.Run()
	}(c)

	timeoutChan := time.NewTimer(timeout)

	select {
	case err := <-c:
		return err
	case <-timeoutChan.C:
		cmd.Process.Kill()
		return ErrTimeout
	}
}
