package service

import (
	"net"
	"errors"
)

// Make PacketConn a Conn
type PacketConnRW struct {
	net.PacketConn
	remoteAddr net.Addr
}

func (this *PacketConnRW) Read(p []byte) (n int, err error) {
	var addr net.Addr

	if this.remoteAddr == nil {
		return 0, errors.New("RemoteAddr is not set")
	}
	n, addr, err = this.ReadFrom(p)

	if addr == nil || addr.Network() != this.remoteAddr.Network() || addr.String() != this.remoteAddr.String() {
		return 0, err
	}

	return n, err
}

func (this *PacketConnRW) Write(p []byte) (n int, err error) {
	if this.remoteAddr == nil {
		return 0, errors.New("RemoteAddr is not set")
	}
	return this.WriteTo(p, this.remoteAddr)
}

func (this *PacketConnRW) RemoteAddr() net.Addr {
	return this.remoteAddr
}

// Make PacketConn a Listener
type PacketListener struct {
	net.PacketConn
}

func (this *PacketListener) Addr() net.Addr {
	return this.LocalAddr()
}

func (this *PacketListener) Accept() (net.Conn, error) {
	p := make([]byte, 0)

	_, addr, err := this.ReadFrom(p)

	return net.Conn(&PacketConnRW{this, addr}), err
}
