package netcatty

import (
	"os"
	"io"
	"net"
)

type Intercept struct{
	*os.File
}

// Extend the TTY reader to catch shortcuts that we handle locally
func (this Intercept) Read(b []byte) (n int, err error) {
	n, err = this.File.Read(b)

	if n <= 0 {
		return n, err
	}

	switch b[0] {
	case 0x1D:  // Ctrl-]
		n = 0
		err = io.EOF
	// case 0x1B:  // Ctrl-I
	// 	if n > 1 && b[1] == 0x72 {
	// 		this.ToggleRawTTY()
	// 		n = 0
	// 	}
	default:
	}

	return n, err
}

type NetProxy struct {
	net.Conn
}

func (this *NetProxy) ProxyFiles(in *os.File, out *os.File) {
	done := make(chan bool)
	cp := func(dst io.Writer, src io.Reader) {
		io.Copy(dst, src)
		done <- true
	}

	// Proxy the PTY to the socket - and back
	go cp(this, Intercept{in})
	go cp(out, this)

	// Wait until the socket is closed (or you exit)
	<-done
}
