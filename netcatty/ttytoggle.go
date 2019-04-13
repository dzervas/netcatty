package netcatty

import (
	"golang.org/x/sys/unix"
	"github.com/mattn/go-tty"
)

type TTYToggle struct {
	*tty.TTY
	reset func() error
}

func (this *TTYToggle) EnableRawTTY() {
	if this.reset == nil {
		// cfmt.Warningln("[!] Entering RAW mode (Ctrl-c will go to remote) - press Alt-r to go back to normal")
		this.reset, _ = this.TTY.Raw()

		// Very targeted fix for broken raw tty for non-tty output
		// Numbers are taken from github.com/mattn/go-tty/tty_linux.go

		// ioctlReadTermios
		termios, _ := unix.IoctlGetTermios(int(this.TTY.Input().Fd()), 0x5401)
		termios.Oflag |= unix.OPOST
		// ioctlWriteTermios
		unix.IoctlSetTermios(int(this.TTY.Input().Fd()), 0x5402, termios)
	}
}

func (this *TTYToggle) DisableRawTTY() {
	if this.reset != nil {
		// cfmt.Infoln("[i] Exiting RAW mode (Ctrl-c will kill the program)")
		this.reset()
		this.reset = nil
	}
}

func (this *TTYToggle) ToggleRawTTY() {
	if this.reset == nil {
		this.EnableRawTTY()
	} else {
		this.DisableRawTTY()
	}
}
