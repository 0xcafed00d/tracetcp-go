package tracetcp

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

func MakeTimeval(t time.Duration) syscall.Timeval {
	return syscall.NsecToTimeval(int64(t))
}

func FD_SET(p *syscall.FdSet, i int) {
	p.Bits[i/64] |= 1 << uint(i) % 64
}

func FD_ISSET(p *syscall.FdSet, i int) bool {
	return (p.Bits[i/64] & (1 << uint(i) % 64)) != 0
}

func FD_ZERO(p *syscall.FdSet) {
	for i := range p.Bits {
		p.Bits[i] = 0
	}
}

func LookupAddress(host string) (*net.IPAddr, error) {
	addresses, err := net.LookupHost(host)
	if err != nil {
		return &net.IPAddr{}, err
	}

	ip, err := net.ResolveIPAddr("ip", addresses[0])
	if err != nil {
		return &net.IPAddr{}, err
	}
	return ip, nil
}

func ToSockaddrInet4(ip net.IPAddr, port int) *syscall.SockaddrInet4 {
	var addr [4]byte
	copy(addr[:], ip.IP.To4())

	return &syscall.SockaddrInet4{Port: port, Addr: addr}
}

func ToIPAddrAndPort(saddr syscall.Sockaddr) (addr net.IPAddr, port int, err error) {

	if sa, ok := saddr.(*syscall.SockaddrInet4); ok {
		port = sa.Port
		addr.IP = append(addr.IP, sa.Addr[:]...)
	} else {
		err = fmt.Errorf("%s", "ToIPAddrAndPort: syscall.Sockaddr not a syscall.SockaddeInet4")
	}

	return
}
