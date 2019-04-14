package action

import (
	"io"
)

var CloseConnection byte = 0x1D  // Ctrl-]

type Intercept struct{
	io.Reader
	channels map[byte][]chan<- byte
}

func NewIntercept(r io.Reader) *Intercept {
	Log.Infoln("Press Ctrl-] to end the current stream")
	return &Intercept{Reader: r}
}

func (this *Intercept) Notify(c chan<- byte, chars ...byte) {
	if this.channels == nil {
		this.channels = map[byte][]chan<- byte{}
	}

	for _, char := range chars {
		this.channels[char] = append(this.channels[char], c)
	}
}

// Extend the TTY reader to catch shortcuts that we handle locally
func (this *Intercept) Read(b []byte) (n int, err error) {
	n, err = this.Reader.Read(b)

	if n <= 0 {
		return n, err
	}

	for i := 0; i < n; i++ {
		for _, c := range this.channels[b[i]] {
			c <- b[i]
		}
	}

	if b[0] == CloseConnection {
		n = 0
		err = io.EOF
	}

	return n, err
}
