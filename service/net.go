package service

import (
	"io"
	"net"
	"strings"
)

type Net struct {
	*Service
	Network string
	Address string
}

func NewNet(in io.Reader, out io.Writer, network, address string) *Net {
	s := Service{Input: in, Output: out}
	return &Net{&s, network, address}
}

func (this *Net) Listen() error {
	var conn net.Conn
	var listen net.Listener
	// Check if the protocol is "normal", if it has a net.Listen implementation or not
	normalProto := strings.HasPrefix(this.Network, "tcp") ||
		this.Network == "unix" ||
		this.Network == "unixpacket"

	Log.Debugf("%+v", this)
	this.fireEvent(EStart)

	if normalProto {
		ln, err := net.Listen(this.Network, this.Address)
		listen = ln
		if err != nil { return err }
		Log.Infof("Listening for %s on %s\n", this.Network, listen.Addr())
	} else {
		ln, err := net.ListenPacket(this.Network, this.Address)
		listen = &PacketListener{ln}
		if err != nil { return err }
		Log.Infof("Listening for %s on %s\n", this.Network, listen.Addr())
	}

	Log.Infoln("Waiting for connection...")
	c, err := listen.Accept()
	if err != nil { return err }
	conn = c

	this.fireEvent(EConnect)
	Log.Infof("Client %s connected\n", conn.RemoteAddr())

	this.ProxyLoop(conn, conn)

	this.fireEvent(EDisconnect)
	Log.Infof("Client %s disconnected\n", conn.RemoteAddr())

	conn.Close()
	listen.Close()
	this.fireEvent(EStop)
	return nil
}

func (this *Net) Dial() error {
	var conn net.Conn
	this.fireEvent(EStart)

	Log.Infof("Connecting to %s over %s...\n", this.Address, this.Network)
	c, err := net.Dial(this.Network, this.Address)
	if err != nil { return err }
	conn = c

	this.fireEvent(EConnect)
	Log.Infof("Connected to %s\n", conn.RemoteAddr())
	
	this.ProxyLoop(conn, conn)

	this.fireEvent(EDisconnect)
	Log.Infof("Disconnected from %s\n", conn.RemoteAddr())

	conn.Close()
	this.fireEvent(EStop)
	return nil
}
