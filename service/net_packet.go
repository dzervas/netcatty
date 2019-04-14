package service

import (
	"net"
	"errors"
)

// net.Conn interface implementation
type PacketConn struct {
	net.PacketConn
	remoteAddr net.Addr
}

func (this *PacketConn) Read(p []byte) (n int, err error) {
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

func (this *PacketConn) Write(p []byte) (n int, err error) {
	if this.remoteAddr == nil {
		return 0, errors.New("RemoteAddr is not set")
	}
	return this.WriteTo(p, this.remoteAddr)
}

func (this *PacketConn) RemoteAddr() net.Addr {
	return this.remoteAddr
}


// net.Listener interface implementation
type PacketListener struct {
	net.PacketConn
}

func (this *PacketListener) Addr() net.Addr {
	return this.LocalAddr()
}

func (this *PacketListener) Accept() (net.Conn, error) {
	p := make([]byte, 0)

	// TODO: Accept should invoke a Read and wait for channel
	// (to catch all packets handled from other PacketConn.Read)
	_, addr, err := this.ReadFrom(p)

	return net.Conn(&PacketConn{this, addr}), err
}
