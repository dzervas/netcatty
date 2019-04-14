package action

import (
	"github.com/dzervas/netcatty/inout"
	"github.com/dzervas/netcatty/service"
)

type AutoRawActionGetter struct {
	*inout.Tty
}

func (this *AutoRawActionGetter) GetAutoRawAction(ser service.Server) Actor {
	action := &Action{
		service: ser,
		events: []service.Event{service.EConnect, service.EDisconnect},
	}
	return &AutoRaw{action, this.Tty}
}

type AutoRaw struct {
	*Action
	tty *inout.Tty
}

func (this *AutoRaw) Register() {
	this.Action.Register()
	go this.handleConnections()
}

func (this *AutoRaw) handleConnections() {
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
