package inout

import (
	"github.com/mattn/go-tty"
)

// TODO: Windows requires local echo
type Tty struct {
	*tty.TTY
	reset func() error
}

func NewTty() (*Tty, error) {
	t, err := tty.Open()
	tty := &Tty{TTY: t}

	return tty, err
}

func (this *Tty) DisableRawTty() {
	if this.reset != nil {
		Log.Infoln("[i] Exiting RAW mode (Ctrl-c will kill the program)")
		this.reset()
		this.reset = nil
	}
}

func (this *Tty) ToggleRawTTY() {
	if this.reset == nil {
		this.EnableRawTty()
	} else {
		this.DisableRawTty()
	}
}
