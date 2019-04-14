package inout

import (
	"io"
	"time"
)

type Interval struct {
	io.ReadWriteCloser
	Interval int
}

func (this *Interval) Read(p []byte) (n int, err error) {
	time.Sleep(time.Duration(this.Interval) * time.Second)
	n, err = this.ReadWriteCloser.Read(p)
	return n, err
}
