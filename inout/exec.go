package inout

import (
	"os"
	"os/exec"
	"errors"
	"strings"

	"github.com/kr/pty"
)

func NewExecArgs(args []string) (*os.File, error) {
	var c *exec.Cmd

	if len(args) > 1 {
		c = exec.Command(args[0], args[1:]...)
	} else if len(args) == 1 {
		c = exec.Command(args[0])
	} else {
		return nil, errors.New("No command given")
	}

	return pty.Start(c)
}

func NewExec(cmd string) (*os.File, error) {
	return NewExecArgs(strings.Split(cmd, " "))
}
