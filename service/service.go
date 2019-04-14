package service

import (
	"io"
	"github.com/amoghe/distillog"
)

var Log = distillog.NewStdoutLogger("service")

type Event int

const (
	EStart		Event = 0
	EStop		Event = 1
	EConnect	Event = 2
	EDisconnect	Event = 3
)

type Service struct {
	Input io.Reader
	Output io.Writer
	channels map[Event][]chan<- Event
}

func (this *Service) Notify(c chan<- Event, events ...Event) {
	if this.channels == nil {
		this.channels = map[Event][]chan<- Event{}
	}

	for _, e := range events {
		this.channels[e] = append(this.channels[e], c)
	}
}

func (this *Service) fireEvent(e Event) {
	for _, c := range this.channels[e] {
		c <- e
	}
}

func (this *Service) ProxyLoop(in io.Reader, out io.Writer) {
	done := make(chan bool)
	cp := func(dst io.Writer, src io.Reader) {
		io.Copy(dst, src)
		done <- true
	}

	go cp(this.Output, in)
	go cp(out, this.Input)

	// Wait until the socket is closed (or you exit)
	<-done
}
