package action

import (
	"github.com/dzervas/netcatty/inout"
	"github.com/dzervas/netcatty/service"
)

type LocalRawTTY struct {
	*Action
	tty *inout.Tty
}

func NewLocalRawTTY(ser service.Server, tty *inout.Tty) *LocalRawTTY {
	action := &Action{
		service: ser,
		events: []service.Event{service.EConnect, service.EDisconnect},
	}
	return &LocalRawTTY{action, tty}
}

func (this *LocalRawTTY) Register() {
	this.Action.Register()
	go this.handleConnections()
}

func (this *LocalRawTTY) handleConnections() {
	loop: for {
		switch e := <-this.channel; e.Event {
		case service.EConnect:
			this.tty.EnableRawTty()
		case service.EDisconnect:
			this.tty.DisableRawTty()
		case service.EUnregister:
			this.tty.DisableRawTty()
			break loop
		}
	}
}
