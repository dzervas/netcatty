// +build windows


package inout

func (this *Tty) EnableRawTty() {
	if this.reset == nil {
		Log.Warningln("[!] Entering RAW mode (Ctrl-c will go to remote)")
		this.reset, _ = this.Raw()
	}
}
