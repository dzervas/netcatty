// +build windows


package inout

func (this *Tty) EnableRawTty() {
	if this.reset == nil {
		Log.Warningln("[!] Entering RAW mode (Ctrl-c will go to remote) - press Alt-r to go back to normal")
		this.reset, _ = this.Raw()
	}
}
