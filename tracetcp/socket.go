package tracetcp

import (
	"fmt"
	"syscall"
	"time"
)

type SocketState int

const (
	SocketConnected SocketState = iota
	SocketNotReached
	SocketTimedOut
	SocketPortClosed
	SocketError
)

func (s SocketState) String() string {
	switch s {
	case SocketConnected:
		return "SocketConnected"
	case SocketNotReached:
		return "SocketNotReached"
	case SocketTimedOut:
		return "SocketTimedOut"
	case SocketPortClosed:
		return "SocketPortClosed"
	case SocketError:
		return "SocketError"
	}
	return "SocketInvlaidState"
}

func waitWithTimeout(socket int, timeout time.Duration) (state SocketState, err error) {
	wfdset := &syscall.FdSet{}

	FD_ZERO(wfdset)
	FD_SET(wfdset, socket)

	timeval := syscall.NsecToTimeval(int64(timeout))

	syscall.Select(socket+1, nil, wfdset, nil, &timeval)

	errcode, err := syscall.GetsockoptInt(socket, syscall.SOL_SOCKET, syscall.SO_ERROR)
	if err != nil {
		state = SocketError
		return
	}

	if errcode == int(syscall.EHOSTUNREACH) {
		state = SocketNotReached
		return
	}

	if errcode == int(syscall.ECONNREFUSED) {
		state = SocketPortClosed
		return
	}

	if errcode != 0 {
		state = SocketError
		err = fmt.Errorf("Connect Error: %v", errcode)
		return
	}

	if FD_ISSET(wfdset, socket) {
		state = SocketConnected
	} else {
		state = SocketTimedOut
	}
	return
}
