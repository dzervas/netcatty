package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/dzervas/netcatty/netcatty"

	"github.com/jessevdk/go-flags"
	"github.com/mattn/go-tty"
	"github.com/mingrammer/cfmt"
)

var Version = "1.0.0"

var TTY *netcatty.TTYToggle
var resetTTY func() error
var errorString = "netcatty;\r"
var promptFingerprints = map[string]string{
	"cmd": "Microsoft Windows",
	"powershell": "PowerShell",
	// "php": "php",
	"python": "Python",
	"sh": "sh-",
}
var errorFingerprints = map[string]string{
	"powershell": "cmdlet",
	"cmd": "internal or external command",
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

func handleErr(err error) {
	if err != nil {
		cfmt.Errorln(err)
		os.Exit(1)
	}
}

var opts struct {
	// TODO
	Close		bool		`short:"c" long:"close"         description:"Close connection on EOF from stdin"`
	NoDetect	bool		`short:"D" long:"no-detect"     description:"Do not detect remote shell (implies -R)"`
	// TODO
	Exec		string		`short:"e" long:"exec"          description:"Program to exec after connect"`
	// TODO
	Gateway		[]string	`short:"g" long:"gateway"       description:"Source-routing hop point[s], up to 8"`
	// TODO
	Pointer		int			`short:"G" long:"pointer"       description:"Source-routing pointer:4, 8, 12, ..."`
	// TODO
	IdleTimeout	int			`short:"i" long:"interval"      description:"Delay interval for lines sent, ports scanned"`
	Listen		bool		`short:"l" long:"listen"        description:"Listen mode, for inbound connects"`
	// TODO
	Tunnel		string		`short:"L" long:"tunnel"        description:"Forward local port to remote address"`
	// TODO
	DontResolve	bool		`short:"n" long:"dont-resolve"  description:"Numeric-only IP addresses, no DNS"`
	// TODO
	Output		string		`short:"o" long:"output"        description:"Output hexdump traffic to FILE (implies -x)"`
	Port		string		`short:"p" long:"port"          description:"Local port number"`
	// TODO
	Protocol	string		`short:"P" long:"protocol"      description:"Provide protocol in the form of tcp{,4,6}|udp{,4,6}|unix{,gram,packet}|ip{,4,6}[:<protocol-number>|:<protocol-name>]\nFor <protocol-number> check https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml"`
	// TODO
	Randomize	bool		`short:"r" long:"randomize"     description:"Randomize local and remote ports"`
	NoRaw		bool		`short:"R" long:"no-raw"        description:"Do not puy local TTY in Raw mode automatically"`
	// TODO
	Source		string		`short:"s" long:"source"        description:"Local source address (ip or hostname)"`
	TCP			bool		`short:"t" long:"tcp"           description:"TCP mode (default)"`
	// TODO
	Telnet		bool		`short:"T" long:"telnet"        description:"answer using TELNET negotiation"`
	// TODO
	UDP			bool		`short:"u" long:"udp"           description:"UDP mode"`
	// TODO
	Verbose		bool		`short:"v" long:"verbose"       description:"-- Not effective, backwards compatibility"`
	// TODO
	Version		bool		`short:"V" long:"version"       description:"Output version information and exit"`
	// TODO
	Hexdump		bool		`short:"x" long:"hexdump"       description:"Hexdump incoming and outgoing traffic"`
	// TODO
	Wait		int			`short:"w" long:"wait"          description:"Timeout for connects and final net reads"`
	// TODO
	Zero		bool		`short:"z" long:"zero"          description:"Zero-I/O mode (used for scanning)"`

	Positional struct {
		Hostname	string	`positional-arg-name:"hostname"`
		APort		string	`positional-arg-name:"port"`
	} `positional-args:"yes"`
}

func main() {
	// Arguments
	parser := flags.NewNamedParser("netcatty", flags.Default)
	parser.AddGroup("Application Options", "", &opts)
	_, err := parser.Parse()
	if err != nil { return }
	// handleErr(err)
	// if len(args) > 0 { opts.Hostname = args[0] }
	// if len(args) > 1 { opts.APort = args[1] }

	address := strings.Join([]string{opts.Positional.Hostname, opts.Port}, ":")
	protocol := "tcp"
	if opts.UDP {
		protocol = "udp"
	}
	if len(opts.Protocol) > 0 {
		protocol = opts.Protocol
	}

	// Logo & Help
	fmt.Printf("NetCaTTY %s - by DZervas <dzervas@dzervas.gr>\n\n", Version)
	fmt.Println("How to get TTY on remote (automatically executed unless you pass -m):")
	for shell, cmds := range shellInit {
		fmt.Printf("%s:\n", shell)
		for _, cmd := range cmds {
			fmt.Println(cmd)
		}
		fmt.Println()
	}
	fmt.Println()

	// Network Stuff
	var listen net.Listener
	var conn *netcatty.NetProxy
	if opts.Listen {
		ln, err := net.Listen(protocol, address)
		listen = ln
		handleErr(err)
		cfmt.Infof("[i] Listening for %s on %s\n", protocol, listen.Addr())
	}

	// Open a TTY and get its file descriptors
	t, err := tty.Open()
	TTY = &netcatty.TTYToggle{TTY: t}
	handleErr(err)
	out := TTY.Output()
	in := TTY.Input()
	defer TTY.Close()  // Make sure that the TTY will close

	// Handle Ctrl-C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		TTY.Close()
		os.Exit(0)
	}()

	// Main Loop
	for {
		cfmt.Infoln("[i] Waiting for connection...")
		if opts.Listen {
			c, err := listen.Accept()
			handleErr(err)
			conn = &netcatty.NetProxy{c}
		} else {
			c, err := net.Dial(protocol, address)
			conn = &netcatty.NetProxy{c}
			handleErr(err)
		}

		cfmt.Successln("[+] New client connection:", conn.RemoteAddr())
		fmt.Println("Press Ctrl-] to close connection")

		if !opts.NoDetect {
			shell := detectShell(conn)
			cfmt.Infof("[i] Detected %s shell!\n", shell)

			for _, cmd := range shellInit[shell] {
				conn.Write([]byte(cmd + "\n"))
			}

			if !opts.NoRaw { TTY.EnableRawTTY() }
		}

		conn.ProxyFiles(in, out)

		TTY.DisableRawTTY()
		conn.Close()
	}
}
