package service

import (
	"io"
	"net"

	"github.com/amoghe/distillog"
)

var Log = distillog.NewStdoutLogger("service")

type Event int

const (
	EBlock		Event = 0
	EUnblock	Event = 1
	EStart		Event = 2
	EStop		Event = 3
	EConnect	Event = 4
	EDisconnect	Event = 5

	// Not fired by Service, reserved to unregister an Action
	EUnregister	Event = -1
)

type EventRWC struct {
	Event Event
	ReadWriteCloser io.ReadWriteCloser
}

type Server interface {
	Notify(chan EventRWC, ...Event)
	Stop(chan EventRWC)
	Listen(string, string) (*LoggedListener, error)
	Dial(string, string) (*LoggedConn, error)
	ProxyLoop(io.Reader, io.Writer)
}

type Service struct {
	io.ReadWriteCloser
	channels map[Event][]chan EventRWC
}

func (this *Service) Notify(c chan EventRWC, events ...Event) {
	if this.channels == nil {
		this.channels = map[Event][]chan EventRWC{}
	}

	if len(events) == 0 {
		events = []Event{
			EStart,
			EStop,
			EConnect,
			EDisconnect,
		}
	}

	for _, e := range events {
		this.channels[e] = append(this.channels[e], c)
	}
}

func (this *Service) Stop(c chan EventRWC) {
	for e, channs := range this.channels {
		for i, chann := range channs {
			if chann == c {
				this.channels[e] = append(channs[:i], channs[i+1:]...)
			}
		}
	}
}

// TODO: Maybe instead of rwc, send a "data" struct
func (this *Service) fireEvent(e Event, rwc io.ReadWriteCloser) {
	erwc := EventRWC{e, rwc}

	for _, c := range this.channels[e] {
		c <- erwc

		select {
		case e := <-c:
			if e.Event == EBlock {
				<-c
			}
		default:
		}
	}
}

func (this *Service) ProxyLoop(in io.Reader, out io.Writer) {
	done := make(chan bool)
	cp := func(dst io.Writer, src io.Reader) {
		io.Copy(dst, src)
		done <- true
	}

	go cp(this, in)
	go cp(out, this)

	// Wait until the socket is closed (or you exit)
	<-done
}

type LoggedListener struct {
	*Service
	net.Listener
}

func (this *LoggedListener) Accept() (conn net.Conn, err error) {
	Log.Infoln("Waiting for connection...")

	conn, err = this.Listener.Accept()

	Log.Infof("Client %s connected\n", conn.RemoteAddr())
	this.fireEvent(EConnect, conn)

	return conn, err
}

func (this *LoggedListener) Close() (err error) {
	Log.Infof("Stopped listening on %s\n", this.Addr())
	this.fireEvent(EStop, nil)
	return this.Close()
}

type LoggedConn struct {
	*Service
	net.Conn
}

func (this *LoggedConn) Close() (err error) {
	Log.Infof("Client %s disconnected\n", this.RemoteAddr())
	this.fireEvent(EDisconnect, this)
	return this.Close()
}
