package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/dzervas/netcatty/service"
	"github.com/dzervas/netcatty/inout"
	"github.com/dzervas/netcatty/action"

	"github.com/jessevdk/go-flags"
	"github.com/mingrammer/cfmt"
)

var Version = "1.0.0"

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
	UDP			bool		`short:"u" long:"udp"           description:"UDP mode (implies -D)"`
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
		Port		string	`positional-arg-name:"port"`
	} `positional-args:"yes"`
}

func main() {
	// Arguments
	parser := flags.NewNamedParser("netcatty", flags.Default)
	parser.AddGroup("Application Options", "", &opts)
	_, err := parser.Parse()
	if err != nil { return }

	// netcat behaviour where to connect you do `nc <ip> <port>`
	// but to listen you do `nc -lp <port>`
	address := strings.Join([]string{opts.Positional.Hostname, opts.Positional.Port}, ":")
	if opts.Listen {
		address = strings.Join([]string{opts.Positional.Hostname, opts.Port}, ":")
	}

	protocol := "tcp"
	if opts.UDP {
		protocol = "udp"
	}
	if len(opts.Protocol) > 0 {
		protocol = opts.Protocol
	}

	// Logo & Help
	fmt.Printf("NetCaTTY %s - by DZervas <dzervas@dzervas.gr>\n\n", Version)
	fmt.Println()

	// InOut
	t, _ := inout.NewTty()
	in := action.NewIntercept(t.Input())
	// Handle Ctrl-C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		t.Close()
		os.Exit(0)
	}()

	// Service
	event := make(chan service.Event, 1)
	n := service.NewNet(in, t.Output(), protocol, address)
	n.Notify(event, service.EDisconnect, service.EConnect)
	go func() {
		for {
			switch <-event {
			case service.EConnect:
				if !opts.NoRaw { t.EnableRawTty() }
			case service.EDisconnect:
				t.DisableRawTty()
			}
		}
	}()

	// Main Loop
	for {
		if !opts.NoDetect {
			// cfmt.Infof("[i] Detected %s shell!\n", shell)
		}

		if opts.Listen {
			err = n.Listen()
		} else {
			err = n.Dial()
		}
		handleErr(err)
	}
}
