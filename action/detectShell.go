package action

import (
	"io"
	"strings"
)

var errorString = "netcatty;\r"
var promptFingerprints = map[string]string{
	"cmd": "Microsoft Windows",
	"powershell": "PowerShell",
	"php": "php",
	"python": "Python",
	"sh": "sh-",
}
var errorFingerprints = map[string]string{
	"powershell": "cmdlet",
	"cmd": "internal or external command",
	"node": "ReferenceError: netcatty is not defined",
	"php": "php",
	"python": "NameError",
	"sh": "netcatty: command not found",
}

func DetectShell(conn io.ReadWriter) {
	var answer string
	buf := make([]byte, 512)
	shell := ""

	p, _ := conn.Read(buf)
	if p > 0 {
		answer = string(buf[:p])
	}

	for k, v := range promptFingerprints {
		if strings.Contains(answer, v) {
			shell = k
			break
		}
	}

	if len(shell) == 0 {
		Log.Warningln("Could not detect shell from prompt, doing error-based detection")
		conn.Write([]byte(errorString))

		n, _ := conn.Read(buf[p:])
		if n > 0 {
			answer = string(buf[:p+n])
		}

		for k, v := range errorFingerprints {
			if strings.Contains(answer, v) {
				shell = k
				break
			}
		}
	}

	Log.Infoln(answer)
	State["shell"] = shell
}
