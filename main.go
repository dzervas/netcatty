package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/dzervas/netcatty/service"
	"github.com/dzervas/netcatty/inout"
	"github.com/dzervas/netcatty/action"
	// "github.com/dzervas/netcatty/ui"

	"github.com/dzervas/mage"
	"github.com/jessevdk/go-flags"
	"github.com/mingrammer/cfmt"
)

var Version = "1.1.0"

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
	// This is already the behaviour. Probably too difficult (and no reason) to implement
	// Close		bool		`short:"c" long:"close"         description:"Close connection on EOF from stdin"`
	// TODO: WTF is this?
	// Gateway		[]string	`short:"g" long:"gateway"       description:"Source-routing hop point[s], up to 8"`
	// TODO: WTF is this?
	// Pointer		int			`short:"G" long:"pointer"       description:"Source-routing pointer:4, 8, 12, ..."`
	Listen		bool		`short:"l" long:"listen"        description:"Listen mode, for inbound connects"`
	// TODO: Needs exposure of listen/dial conns
	// DontResolve	bool		`short:"n" long:"dont-resolve"  description:"Numeric-only IP addresses, no DNS"`
	Port		string		`short:"p" long:"local-port"    description:"Local port number"`
	// Does not have any effect - even on netcat
	Randomize	bool		`short:"r" long:"randomize"     description:"Randomize local and remote ports"`
	NoRaw		bool		`short:"R" long:"no-raw"        description:"Do NOT put TTY in Raw mode"`
	Source		string		`short:"s" long:"source"        description:"Local source address (ip or hostname)"`
	// TODO
	Telnet		bool		`short:"T" long:"telnet"        description:"answer using TELNET negotiation"`
	// TODO
	Verbose		bool		`short:"v" long:"verbose"       description:"-- Not effective, backwards compatibility"`
	// TODO
	Version		bool		`short:"V" long:"version"       description:"Output version information and exit"`
	// TODO: Needs exposure of listen/dial conns
	// Wait		int			`short:"w" long:"wait"          description:"Timeout for connects and final net reads"`

	Positional struct {
		Hostname	string	`positional-arg-name:"hostname"`
		Port		string	`positional-arg-name:"port"`
	} `positional-args:"yes"`
}

var optsService struct {
	Protocol	string		`short:"P" long:"protocol"      description:"Provide protocol in the form of tcp{,4,6}|udp{,4,6}|unix{,gram,packet}|ip{,4,6}:[<protocol-number>|<protocol-name>]\nFor <protocol-number> check https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml"`
	TCP			bool		`short:"t" long:"tcp"           description:"TCP mode (default)"`
	UDP			bool		`short:"u" long:"udp"           description:"UDP mode"`
	Mage		int			`long:"mage"                    description:"Use the mage protocol over the selected service, on the specified channel" default:"-1"`
}

var optsInOut struct {
	// This is already the behaviour. Probably too difficult (and no reason) to implement
	// Close		bool		`short:"c" long:"close"         description:"Close connection on EOF from stdin"`
	Exec		string		`short:"e" long:"exec"          description:"Program to exec after connect"`
	Interval	int			`short:"i" long:"interval"      description:"Delay interval for data sent, ports scanned"`
	// TODO: Needs exposure of listen/dial conns
	// Tunnel		string		`short:"L" long:"tunnel"        description:"Forward local port to remote address"`
	// TODO
	Output		string		`short:"o" long:"output"        description:"Output hexdump traffic to FILE (implies -x)"`
	HexDump		bool		`short:"x" long:"hexdump"       description:"Hexdump incoming and outgoing traffic"`
	Zero		bool		`short:"z" long:"zero"          description:"Zero-I/O mode (used for scanning)"`
}

var optsAction struct {
	Detect		bool		`short:"D" long:"detect"        description:"Detect remote shell automatically and try to raise a TTY on the remote" group:"Actions"`
}

func main() {
	// Arguments
	// TODO: Add option name to all arguments (tunnel=<something>)
	parser := flags.NewNamedParser("netcatty", flags.Default)
	parser.AddGroup("Application Options", "", &opts)
	parser.AddGroup("Service", "", &optsService)
	parser.AddGroup("InOut", "", &optsInOut)
	parser.AddGroup("Action", "", &optsAction)

	_, err := parser.Parse()
	if err != nil { return }

	// netcat behaviour where to connect you do `nc <ip> <port>`
	// but to listen you do `nc -lp <port>`
	hostname := opts.Positional.Hostname
	port := opts.Positional.Port
	if opts.Listen && len(opts.Source) > 0 {
		port = opts.Source
	}
	if opts.Listen && len(opts.Port) > 0 {
		port = opts.Port
	}
	address := strings.Join([]string{hostname, port}, ":")

	protocol := "tcp"
	if optsService.UDP {
		protocol = "udp"
	}
	if len(optsService.Protocol) > 0 {
		protocol = optsService.Protocol
	}

	// Logo
	fmt.Printf("NetCaTTY %s - by DZervas <dzervas@dzervas.gr>\n\n", Version)

	// InOut
	// TODO: Fix the loggers
	var ioc io.ReadWriteCloser

	if len(optsInOut.Exec) > 0 {
		ioc, err = inout.NewExec(optsInOut.Exec)
		handleErr(err)
	} else if optsInOut.Zero {
		ioc, err = inout.NewZero()
		handleErr(err)
	} else {
		t, err := inout.NewTty()
		handleErr(err)

		// Handle exit cases
		defer t.Close()

		// Handle Ctrl-C
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		go func() {
			<-sig
			t.Close()
			os.Exit(0)
		}()

		in := action.NewIntercept(t.Input())
		ioc = NewReadWriteCloser(in, t.Output(), t)

		if !opts.NoRaw {
			t.EnableRawTty()
			// in.Notify(ui.Channel)
		}
	}

	if optsInOut.HexDump { ioc = &inout.HexDump{ioc} }
	if optsInOut.Interval > 0 { ioc = &inout.Interval{ioc, optsInOut.Interval} }

	// Service
	var s service.Server
	s = service.NewNet(ioc)

	// Actions
	if optsAction.Detect { action.NewRaiseTTY(s).Register() }

	// Main Loop
	var conn net.Conn
	if opts.Listen {
		ln, e := s.Listen(protocol, address)
		handleErr(e)
		conn, err = ln.Accept()
	} else {
		conn, err = s.Dial(protocol, address)
	}

	handleErr(err)

	// go ui.UIMain()

	// Mage Protocol
	if optsService.Mage >= 0 {
		stream := &mage.Stream{
			Reader: conn,
			Writer: conn,
		}
		readwriter := stream.GetReadWriter(uint8(optsService.Mage))
		s.ProxyLoop(readwriter, readwriter)
	} else {
		s.ProxyLoop(conn, conn)
	}
}
