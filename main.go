package main

import (
	"fmt"
	"io"
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

type RWC struct {
	io.Reader
	io.Writer
	io.Closer
}

func NewReadWriteCloser(r io.Reader, w io.Writer, c io.Closer) io.ReadWriteCloser {
	return RWC{r, w, c}
}

var opts struct {
	// TODO
	Close		bool		`short:"c" long:"close"         description:"Close connection on EOF from stdin"`
	Detect		bool		`short:"D" long:"detect"        description:"Detect remote shell automatically and try to raise a TTY on the remote (action)"`
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
	Raw			bool		`short:"R" long:"auto-raw"      description:"Put local TTY in Raw mode on connect (action)"`
	// TODO
	Source		string		`short:"s" long:"source"        description:"Local source address (ip or hostname)"`
	TCP			bool		`short:"t" long:"tcp"           description:"TCP mode (default)"`
	// TODO
	Telnet		bool		`short:"T" long:"telnet"        description:"answer using TELNET negotiation"`
	UDP			bool		`short:"u" long:"udp"           description:"UDP mode"`
	// TODO
	Verbose		bool		`short:"v" long:"verbose"       description:"-- Not effective, backwards compatibility"`
	// TODO
	Version		bool		`short:"V" long:"version"       description:"Output version information and exit"`
	HexDump		bool		`short:"x" long:"hexdump"       description:"Hexdump incoming and outgoing traffic"`
	// TODO
	Wait		int			`short:"w" long:"wait"          description:"Timeout for connects and final net reads"`
	Zero		bool		`short:"z" long:"zero"          description:"Zero-I/O mode (used for scanning)"`

	Positional struct {
		Hostname	string	`positional-arg-name:"hostname"`
		Port		string	`positional-arg-name:"port"`
	} `positional-args:"yes"`
}

func main() {
	// Arguments
	// TODO: Break this into groups, in a different small function
	// TODO: Add option name to all arguments (tunnel=<something>)
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

	// InOut
	// TODO: Fix the loggers
	var ioc io.ReadWriteCloser
	var actions []func(service.Server) action.Actor

	if len(opts.Exec) > 0 {
		ioc, err = inout.NewExec(opts.Exec)
		handleErr(err)
	} else if opts.Zero {
		ioc, err = inout.NewZero()
		handleErr(err)
	} else {
		t, err := inout.NewTty()
		handleErr(err)
		in := action.NewIntercept(t.Input())
		// Handle Ctrl-C
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		go func() {
			<-sig
			t.Close()
			os.Exit(0)
		}()

		ioc = NewReadWriteCloser(in, t.Output(), t)

		if opts.Raw {
			a := action.AutoRawActionGetter{t}
			actions = append(actions, a.GetAutoRawAction)
		}
	}

	if opts.HexDump { ioc = &inout.HexDump{ioc} }

	// Service
	var s service.Server
	s = service.NewNet(ioc, protocol, address)

	// Actions
	if opts.Detect { action.NewRaiseTTY(s).Register() }

	for _, a := range actions {
		a(s).Register()
	}

	// Main Loop
	if opts.Listen {
		err = s.Listen()
	} else {
		err = s.Dial()
	}
	handleErr(err)
}
