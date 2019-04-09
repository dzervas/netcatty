package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/mattn/go-tty"
	"github.com/mingrammer/cfmt"
)

var TTY *tty.TTY
var resetTTY func() error
var errorString = "netcatty;\r"
var promptFingerprints = map[string]string{
	// "cmd": "Microsoft Windows",
	// "powershell": "PowerShell",
	// "php": "php",
	"python": "Python",
	"sh": "sh-",
}
var errorFingerprints = map[string]string{
	// "powershell": "cmdlet",
	// "cmd": "internal or external command",
	// "node": "ReferenceError: netcatty is not defined",
	// "php": "php",
	"python": "NameError",
	"sh": "netcatty: command not found",
}
var shellInit = map[string][]string{
	"python": { "import pty; pty.spawn('python')" },
	"sh": {
		"script -qefc '/bin/sh' /dev/null",
		"TERM=xterm",
	},
}

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
	case 0x1B:  // Ctrl-I
		if n > 1 && b[1] == 0x72 {
			toggleRawTTY()
			n = 0
		}
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

func enableRawTTY() {
	if resetTTY == nil {
		cfmt.Warningln("[!] Entering RAW mode (Ctrl-c will go to remote) - press Alt-r to go back to normal")
		resetTTY, _ = TTY.Raw()
	}
}

func disableRawTTY() {
	if resetTTY != nil {
		cfmt.Infoln("[i] Exiting RAW mode (Ctrl-c will kill the program)")
		resetTTY()
		resetTTY = nil
	}
}

func toggleRawTTY() {
	if resetTTY == nil {
		enableRawTTY()
	} else {
		disableRawTTY()
	}
}

func detectShell(conn net.Conn) string {
	var answer string
	buf := make([]byte, 512)
	shell := ""

	p, _ := conn.Read(buf)
	if p > 0 {
		answer = string(buf[:p])
	}

	for k, v := range promptFingerprints {
		if strings.Contains(answer, v) {
			shell = k
			break
		}
	}

	if len(shell) == 0 {
		fmt.Println("Could not detect shell from prompt, doing error-based detection")
		conn.Write([]byte(errorString))

		n, _ := conn.Read(buf[p:])
		if n > 0 {
			answer = string(buf[:p+n])
		}

		for k, v := range errorFingerprints {
			if strings.Contains(answer, v) {
				shell = k
				break
			}
		}
	}

	fmt.Println(answer)

	if len(shell) == 0 {
		cfmt.Warningln("Could not detect shell, falling back to sh :(")
		shell = "sh"
	}

	return shell
}

func main() {
	// Arguments
	isListen := flag.Bool("l", false, "Enable listening mode")
	address := flag.String("a", ":4444", "Listen/Connect address in the form of 'ip:port'.\nDomains, IPv6 as ip and Service as port ('localhost:http') also work.")
	flag.Parse()

	// Logo & Help
	fmt.Println("NetCaTTY - by DZervas <dzervas@dzervas.gr>")
	fmt.Println()

	// Network Stuff
	var listen net.Listener
	var conn *NetProxy
	if *isListen {
		ln, err := net.Listen("tcp", *address)
		listen = ln
		handleErr(err)
	}

	// Open a TTY and get its file descriptors
	t, err := tty.Open()
	TTY = t
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
			conn = &NetProxy{c}
			handleErr(err)
		} else {
			c, err := net.Dial("tcp", *address)
			conn = &NetProxy{c}
			handleErr(err)
		}

		cfmt.Successln("[+] New client connection:", conn.RemoteAddr())
		fmt.Println("Press Ctrl-] to close connection")

		shell := detectShell(conn)
		cfmt.Infof("[i] Detected %s shell!\n", shell)

		for _, cmd := range shellInit[shell] {
			conn.Write([]byte(cmd + "\n"))
		}
		enableRawTTY()

		conn.ProxyFiles(in, out)

		disableRawTTY()
		conn.Close()
	}
}
