package action

import (
	"io"
)

var shellInit = map[string][]string{
	"python": { "import pty; pty.spawn('python')" },
	"sh": {
		"script -qefc '/bin/sh' /dev/null",
		"TERM=xterm",
	},
}

func RaiseTTY(conn io.ReadWriter) {
	for _, cmd := range shellInit[State["shell"]] {
		conn.Write([]byte(cmd + "\n"))
	}
}
