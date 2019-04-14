package inout

import (
	"io"
	"os/exec"
	"errors"
	"strings"
)

type RWC struct {
	io.Reader
	io.Writer
	io.Closer
}

func NewReadWriteCloser(r io.Reader, w io.Writer, c io.Closer) io.ReadWriteCloser {
	return RWC{r, w, c}
}

func NewExecArgs(args []string) (*RWC, error) {
	var c *exec.Cmd

	if len(args) > 1 {
		c = exec.Command(args[0], args[1:]...)
	} else if len(args) == 1 {
		c = exec.Command(args[0])
	} else {
		return nil, errors.New("No command given")
	}

	// TODO: Run it in a TTY
	// TODO: Combine Stderr
	si, _ := c.StdinPipe()
	so, _ := c.StdoutPipe()

	return &RWC{so, si, nil}, c.Start()
}

func NewExec(cmd string) (*RWC, error) {
	return NewExecArgs(strings.Split(cmd, " "))
}
