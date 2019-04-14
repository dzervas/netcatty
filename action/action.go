package action

import (
	"github.com/dzervas/netcatty/service"

	"github.com/amoghe/distillog"
)

var Log = distillog.NewStdoutLogger("action")

var State = map[string]string{}

type Actor interface {
	Register()
	Unregister()
}

type Action struct {
	service service.Server
	events []service.Event
	channel chan service.EventRWC
}

func (this *Action) Register() {
	this.channel = make(chan service.EventRWC, 1)
	this.service.Notify(this.channel, this.events...)
}

func (this *Action) Block() {
	this.channel <- service.EventRWC{Event: service.EBlock}
}

func (this *Action) Unblock() {
	this.channel <- service.EventRWC{Event: service.EUnblock}
}

func (this *Action) Unregister() {
	this.channel <- service.EventRWC{service.EUnregister, nil}
	this.service.Stop(this.channel)
}
