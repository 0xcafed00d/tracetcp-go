package tracetcp

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

type implTraceEventType int

const (
	timedOut implTraceEventType = iota
	ttlExpired
	connected
	connectFailed
	errored
)

type implTraceEvent struct {
	evtype    implTraceEventType
	timeStamp time.Time

	localPort  int
	remotePort int
	remoteAddr net.IPAddr
	localAddr  net.IPAddr
	ttl        int
	query      int
	err        error
}

func tryConnect(dest net.IPAddr, port, ttl, query int,
	timeout time.Duration, result chan implTraceEvent) {

	// fill in the event with as much info as we have so far
	event := implTraceEvent{
		remoteAddr: dest,
		remotePort: port,
		ttl:        ttl,
		query:      query,
	}

	returnError := func(err error) {
		event.err = err
		event.evtype = errored
		event.timeStamp = time.Now()
		result <- event
	}

	sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		returnError(err)
		return
	}
	defer syscall.Close(sock)

	err = syscall.SetsockoptInt(sock, 0x0, syscall.IP_TTL, ttl)
	if err != nil {
		returnError(err)
		return
	}

	err = syscall.SetNonblock(sock, true)
	if err != nil {
		returnError(err)
		return
	}

	// ignore error from connect in non-blocking mode. as it will always return an
	// in progress error
	_ = syscall.Connect(sock, ToSockaddrInet4(dest, port))

	// get the local ip address and port number
	local, err := syscall.Getsockname(sock)
	if err != nil {
		returnError(err)
		return
	}

	event.localAddr, event.localPort, err = ToIPAddrAndPort(local)
	if err != nil {
		returnError(err)
		return
	}
}

func connect(host string, port, ttl int, timeout time.Duration) error {
	sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return err
	}
	defer syscall.Close(sock)

	err = syscall.SetsockoptInt(sock, 0x0, syscall.IP_TTL, ttl)
	if err != nil {
		return err
	}

	err = syscall.SetNonblock(sock, true)
	if err != nil {
		return err
	}

	addr, err := LookupAddress(host)
	if err != nil {
		return nil
	}

	// ignore error from connect in non-blocking mode. as it will always return a
	// in progress error
	_ = syscall.Connect(sock, &syscall.SockaddrInet4{Port: 80, Addr: addr})

	name, err := syscall.Getsockname(sock)
	fmt.Println(err, name)

	fdset := &syscall.FdSet{}
	timeoutVal := &syscall.Timeval{}
	timeoutVal.Sec = int64(timeout / time.Second)
	timeoutVal.Usec = int64(timeout-time.Duration(timeoutVal.Sec)*time.Second) / 1000

	fmt.Println(timeoutVal)

	FD_ZERO(fdset)
	FD_SET(fdset, sock)

	start := time.Now()
	x, err := syscall.Select(sock+1, nil, fdset, nil, timeoutVal)
	elapsed := time.Since(start)

	fmt.Println(x, elapsed)
	if err != nil {
		return err
	}

	if FD_ISSET(fdset, sock) {
		fmt.Println("conencted")
	} else {
		fmt.Println("timedout")
	}

	return nil
}
