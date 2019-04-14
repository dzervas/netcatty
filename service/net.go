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

func NewNet(ioc io.ReadWriteCloser, network, address string) *Net {
	s := Service{ReadWriteCloser: ioc}
	return &Net{&s, network, address}
}

func (this *Net) Listen() error {
	var conn net.Conn
	var listen net.Listener
	// TODO: This should break into a separate server
	// Check if the protocol is "normal", if it has a net.Listen implementation or not
	normalProto := strings.HasPrefix(this.Network, "tcp") ||
		this.Network == "unix" ||
		this.Network == "unixpacket"

	Log.Debugf("%+v", this)

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

	this.fireEvent(EStart, nil)

	Log.Infoln("Waiting for connection...")
	c, err := listen.Accept()
	if err != nil { return err }
	conn = c

	Log.Infof("Client %s connected\n", conn.RemoteAddr())
	this.fireEvent(EConnect, conn)

	this.ProxyLoop(conn, conn)

	Log.Infof("Client %s disconnected\n", conn.RemoteAddr())
	this.fireEvent(EDisconnect, conn)
	conn.Close()

	listen.Close()
	this.fireEvent(EStop, nil)
	return nil
}

func (this *Net) Dial() error {
	var conn net.Conn

	Log.Infof("Connecting to %s over %s...\n", this.Address, this.Network)
	this.fireEvent(EStart, nil)
	c, err := net.Dial(this.Network, this.Address)
	if err != nil { return err }
	conn = c

	Log.Infof("Connected to %s\n", conn.RemoteAddr())
	this.fireEvent(EConnect, conn)
	
	this.ProxyLoop(conn, conn)

	Log.Infof("Disconnected from %s\n", conn.RemoteAddr())
	this.fireEvent(EDisconnect, conn)

	conn.Close()
	this.fireEvent(EStop, nil)
	return nil
}
