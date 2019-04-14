package inout

import (
	"io"
	"fmt"
	"encoding/hex"
)

type HexDump struct {
	io.ReadWriteCloser
}

func (this *HexDump) Read(p []byte) (n int, err error) {
	n, err = this.ReadWriteCloser.Read(p)
	if n > 0 { fmt.Println(hex.Dump(p[:n])) }
	return n, err
}

func (this *HexDump) Write(p []byte) (n int, err error) {
	n, err = this.ReadWriteCloser.Write(p)
	if n > 0 { fmt.Println(hex.Dump(p[:n])) }
	return n, err
}
