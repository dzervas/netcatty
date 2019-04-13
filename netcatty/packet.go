package netcatty

import (
	"io"
	"net"
)

type ReadWriterProxy struct {
	net.Conn
}

func (this *ReadWriterProxy) ProxyFiles(in io.Reader, out io.Writer) {
	done := make(chan bool)
	cp := func(dst io.Writer, src io.Reader) {
		io.Copy(dst, src)
		done <- true
	}

	// Proxy the PTY to the socket - and back
	go cp(this, in)
	go cp(out, this)

	// Wait until the socket is closed (or you exit)
	<-done
}

// Make PacketConn a Conn
type PacketConnRW struct {
	net.PacketConn
	remoteAddr net.Addr
}

func (this *PacketConnRW) Read(p []byte) (n int, err error) {
	var addr net.Addr

	this.RemoteAddr()
	n, addr, err = this.ReadFrom(p)

	if addr == nil || addr.Network() != this.remoteAddr.Network() || addr.String() != this.remoteAddr.String() {
		return 0, err
	}

	return n, err
}

func (this *PacketConnRW) Write(p []byte) (n int, err error) {
	this.RemoteAddr()
	return this.WriteTo(p, this.remoteAddr)
}

func (this *PacketConnRW) RemoteAddr() net.Addr {
	if this.remoteAddr == nil {
		panic("remoteAddr not set!")
	}

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
