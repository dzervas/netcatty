package inout

import (
	"io"
)

type Zero struct {}

func NewZero() (io.ReadWriteCloser, error) {
	return &Zero{}, nil
}

func (_ *Zero) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

func (_ *Zero) Write(_ []byte) (int, error) {
	return 0, io.EOF
}

func (_ *Zero) Close() error {
	return nil
}
