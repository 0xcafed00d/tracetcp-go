package tracetcp

import (
	"fmt"
	"log"
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
	return fmt.Sprintf("connectEvent:{type: %v, time: %v, local: %v:%d, remote: %v:%d, ttl: %d, query: %d, err: %v}",
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

	log.Printf("try Connect dest: %v port: %v ttl: %v query: %v timeout: %v",
		dest, port, ttl, query, timeout)

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
	log.Printf(".... try Connect local endpoint: %v : %v", event.localAddr, event.localPort)

	state, err := waitWithTimeout(sock, timeout)
	switch state {
	case SocketError:
		result = makeErrorEvent(&event, err)
	case SocketConnected:
		result = makeEvent(&event, connectConnected)
	case SocketNotReached:
		result = makeEvent(&event, connectFailed)
	case SocketPortClosed:
		result = makeEvent(&event, connectFailed)
	case SocketTimedOut:
		result = makeEvent(&event, connectTimedOut)
	}
	return
}
