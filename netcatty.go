package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mattn/go-tty"
	"github.com/mingrammer/cfmt"
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
	default:
	}

	return n, err
}

func handleErr(err error) {
	if err != nil {
		cfmt.Errorln(err)
		os.Exit(1)
	}
}

func resizeTTY(t *tty.TTY) {
	for ws := range t.SIGWINCH() {
		// While this CAN be sent on the remote directly
		// it's better not to, as in case of being inside
		// vim for example things will break
		fmt.Printf("\nRun 'stty rows %d cols %d' to resize terminal\n", ws.H, ws.W)
	}
}

func receive(conn net.Conn, in *os.File, out *os.File) {
	sig := make(chan os.Signal, 1)
	captureSIG := func() {
		for s := range sig {
			switch s {
			case syscall.SIGINT:
				conn.Write([]byte{3})
			case syscall.SIGTSTP:
				conn.Write([]byte{26})
			default:
			}
		}
	}

	done := make(chan bool)
	cp := func(dst io.Writer, src io.Reader) {
		io.Copy(dst, src)
		done <- true
	}

	// Proxy the PTY to the socket - and back
	go cp(conn, Intercept{in})
	go cp(out, conn)

	cfmt.Warningln("[!] Capturing signals - press Ctrl-] to exit")
	signal.Notify(sig, os.Interrupt)
	go captureSIG()

	<-done

	cfmt.Infoln("[i] Releasing signals...")
	signal.Stop(sig)
}

func main() {
	// Arguments
	isListen := flag.Bool("l", false, "Enable listening mode")
	address := flag.String("a", ":4444", "Listen/Connect address in the form of 'ip:port'.\nDomains, IPv6 as ip and Service as port ('localhost:http') also work.")
	flag.Parse()

	fmt.Println("NetCaTTY - by DZervas <dzervas@dzervas.gr>")

	// Network Stuff
	var listen net.Listener
	var conn net.Conn
	if *isListen {
		ln, err := net.Listen("tcp", *address)
		listen = ln
		handleErr(err)
	}

	// Open a TTY and get its file descriptors
	t, err := tty.Open()
	handleErr(err)
	out := t.Output()
	in := t.Input()
	defer t.Close()  // Make sure that the TTY will close
	go resizeTTY(t)  // Handle resizes

	// Main Loop
	for {
		cfmt.Infoln("[i] Waiting for connection...")
		if *isListen {
			c, err := listen.Accept()
			conn = c
			handleErr(err)
		} else {
			c, err := net.Dial("tcp", *address)
			conn = c
			handleErr(err)
		}

		cfmt.Successln("[+] New client connection:", conn.RemoteAddr())
		receive(conn, in, out)

		conn.Close()
	}
}
