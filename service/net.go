package service

import (
	"io"
	"net"
	"strings"
)

type Net struct {
	*Service
}

func NewNet(ioc io.ReadWriteCloser) *Net {
	return &Net{&Service{ReadWriteCloser: ioc}}
}

func (this *Net) Listen(network, address string) (ln *LoggedListener, err error) {
	var listen net.Listener

	// TODO: This should break into a separate server
	// Check if the protocol is "normal", if it has a net.Listen implementation or not
	normalProto := strings.HasPrefix(network, "tcp") ||
		network == "unix" ||
		network == "unixpacket"

	if normalProto {
		listen, err = net.Listen(network, address)
	} else {
		l, e := net.ListenPacket(network, address)
		listen = &PacketListener{l}
		err = e
	}

	this.fireEvent(EStart, nil)
	Log.Infof("Listening for %s on %s\n", network, listen.Addr())

	return &LoggedListener{this.Service, listen}, err
}

func (this *Net) Dial(network, address string) (conn *LoggedConn, err error) {
	Log.Infof("Connecting to %s over %s...\n", address, network)
	this.fireEvent(EStart, nil)
	c, e := net.Dial(network, address)

	err = e
	conn = &LoggedConn{this.Service, c}
	if conn == nil || err != nil { return conn, err }

	Log.Infof("Connected to %s\n", conn.RemoteAddr())
	this.fireEvent(EConnect, conn)

	return conn, err
}
