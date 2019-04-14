package inout

import (
	"golang.org/x/sys/unix"
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

func (this *Tty) EnableRawTty() {
	if this.reset == nil {
		Log.Warningln("[!] Entering RAW mode (Ctrl-c will go to remote) - press Alt-r to go back to normal")
		this.reset, _ = this.Raw()

		// Very targeted fix for broken raw tty for non-tty output
		// Numbers are taken from github.com/mattn/go-tty/tty_linux.go

		// ioctlReadTermios
		termios, _ := unix.IoctlGetTermios(int(this.Input().Fd()), 0x5401)
		termios.Oflag |= unix.OPOST
		// ioctlWriteTermios
		unix.IoctlSetTermios(int(this.Input().Fd()), 0x5402, termios)
	}
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
