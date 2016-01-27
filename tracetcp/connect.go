package tracetcp

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

type connectEventType int

const (
	connectNone connectEventType = iota
	connectTimedOut
	connectConnected
	connectFailed
	connectError
)

// implementation of fmt.Stinger interface
func (t connectEventType) String() string {
	switch t {
	case connectNone:
		return "none"
	case connectTimedOut:
		return "timedOut"
	case connectConnected:
		return "connected"
	case connectFailed:
		return "connectFailed"
	case connectError:
		return "errored"
	}
	return "Invalid implTraceEventType"
}

type connectEvent struct {
	evtype    connectEventType
	timeStamp time.Time

	localAddr  net.IPAddr
	localPort  int
	remoteAddr net.IPAddr
	remotePort int
	ttl        int
	query      int
	err        error
}

// implementation of fmt.Stinger interface
func (e connectEvent) String() string {
	return fmt.Sprintf("{type: %v, time: %v, local: %v:%d, remote: %v:%d, ttl: %d, query: %d, err: %v}",
		e.evtype.String(), e.timeStamp, e.localAddr, e.localPort, e.remoteAddr, e.remotePort, e.ttl, e.query, e.err)
}

func makeErrorEvent(event *connectEvent, err error) connectEvent {
	event.err = err
	event.evtype = connectError
	event.timeStamp = time.Now()
	return *event
}

func makeEvent(event *connectEvent, evtype connectEventType) connectEvent {
	event.evtype = evtype
	event.timeStamp = time.Now()
	return *event
}

func tryConnect(dest net.IPAddr, port, ttl, query int, timeout time.Duration) (result connectEvent) {

	// fill in the event with as much info as we have so far
	event := connectEvent{
		remoteAddr: dest,
		remotePort: port,
		ttl:        ttl,
		query:      query,
	}

	sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		result = makeErrorEvent(&event, err)
		return
	}
	defer syscall.Close(sock)

	err = syscall.SetsockoptInt(sock, 0x0, syscall.IP_TTL, ttl)
	if err != nil {
		result = makeErrorEvent(&event, err)
		return
	}

	err = syscall.SetNonblock(sock, true)
	if err != nil {
		result = makeErrorEvent(&event, err)
		return
	}

	// ignore error from connect in non-blocking mode. as it will always return an
	// in progress error
	_ = syscall.Connect(sock, ToSockaddrInet4(dest, port))

	// get the local ip address and port number
	local, err := syscall.Getsockname(sock)
	if err != nil {
		result = makeErrorEvent(&event, err)
		return
	}

	// fill in the local endpoint deatils on the event struct
	event.localAddr, event.localPort, err = ToIPAddrAndPort(local)
	if err != nil {
		result = makeErrorEvent(&event, err)
		return
	}

	fdset := &syscall.FdSet{}
	timeoutVal := MakeTimeval(timeout)

	FD_ZERO(fdset)
	FD_SET(fdset, sock)

	_, err = syscall.Select(sock+1, nil, fdset, nil, &timeoutVal)
	if err != nil {
		result = makeErrorEvent(&event, err)
		return
	}

	// TODO: test for connect failed?

	if FD_ISSET(fdset, sock) {
		// detect if actually connected as select shows ttl expired as connected
		// so if we try to get the remote address and it fails then ttl has expired
		_, err = syscall.Getpeername(sock)
		if err == nil {
			result = makeEvent(&event, connectConnected)
		} else {
			result = makeEvent(&event, connectFailed)
		}
	} else {
		result = makeEvent(&event, connectTimedOut)
	}
	return
}
