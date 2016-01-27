package tracetcp

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

type icmpEventType int

const (
	icmpNone icmpEventType = iota
	icmpTTLExpired
	icmpNoRoute
	icmpError
)

// implementation of fmt.Stinger interface
func (t icmpEventType) String() string {
	switch t {
	case icmpNone:
		return "none"
	case icmpTTLExpired:
		return "ttlExpired"
	case icmpNoRoute:
		return "noRoute"
	case icmpError:
		return "error"
	}
	return "Invalid implTraceEventType"
}

type icmpEvent struct {
	evtype    icmpEventType
	timeStamp time.Time

	localAddr  net.IPAddr
	localPort  int
	remoteAddr net.IPAddr
	remotePort int
	err        error
}

// implementation of fmt.Stinger interface
func (e icmpEvent) String() string {
	return fmt.Sprintf("{type: %v, time: %v, local: %v:%d, remote: %v:%d, err: %v}",
		e.evtype.String(), e.timeStamp, e.localAddr, e.localPort, e.remoteAddr, e.remotePort, e.err)
}

func makeICMPErrorEvent(event *icmpEvent, err error) icmpEvent {
	event.err = err
	event.evtype = icmpError
	event.timeStamp = time.Now()
	return *event
}
func makeICMPEvent(event *icmpEvent, evtype icmpEventType) icmpEvent {
	event.evtype = evtype
	event.timeStamp = time.Now()
	return *event
}
func receiveICMP(result chan icmpEvent) {
	event := icmpEvent{}

	// Set up the socket to receive inbound packets
	sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if err != nil {
		result <- makeICMPErrorEvent(&event, err)
		return
	}

	err = syscall.Bind(sock, &syscall.SockaddrInet4{})
	if err != nil {
		result <- makeICMPErrorEvent(&event, err)
		return
	}

	var pkt = make([]byte, 1024)
	for {
		_, from, err := syscall.Recvfrom(sock, pkt, 0)
		if err != nil {
			result <- makeICMPErrorEvent(&event, err)
			return
		}

		// fill in the remote endpoint deatils on the event struct
		event.remoteAddr, _, _ = ToIPAddrAndPort(from)
		result <- makeICMPEvent(&event, icmpTTLExpired)
	}
}
