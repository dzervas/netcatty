package action

import (
	"io"
	"fmt"

	// "github.com/dzervas/netcatty/service"
)

const (
	CloseConnection	byte = 0x1D  // Ctrl-]
	OpenUI			byte = 0x14  // Ctrl-T
)

type UIComponent struct {
	Run func(io.ReadWriter)
	SetsState []string
	GetsState []string
	Help string
}

var insideUI = false
var UIShortcuts = map[byte]UIComponent{
	byte('s'): {DetectShellRun, []string{"shell"}, []string{}, "Detect the shell that is running on the other side"},
	byte('h'): {UIHelpRun, []string{}, []string{}, "Show this help message"},
}

func UIHelpRun(_ io.ReadWriter) {
	fmt.Println("------ UI Help ------")
	// for short, comm := range UIShortcuts {
		// fmt.Printf("%c\t%s", short, comm.Help)
	// }
}

type Intercept struct {
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

	if b[0] == CloseConnection {
		return 0, io.EOF
	} else if b[0] == OpenUI {
		insideUI = true
		fmt.Println("Inside UI!")
		if n == 1 { return 0, nil }
	}

	for i := 0; i < n; i++ {
		for _, c := range this.channels[b[i]] {
			c <- b[i]
		}

		if !insideUI { break }

		for short, comm := range UIShortcuts {
			if b[i] == short {
				// Fire events on keypress, not functions
				// comm.Run()
				fmt.Println(comm.Help)
			}
		}
	}

	insideUI = false

	return n, err
}
