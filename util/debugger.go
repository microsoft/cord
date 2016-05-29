package util

import (
	"os"

	"github.com/fatih/color"
)

const defaultConsoleWidth = 79

// StderrDebugger prints debug information out on the console.
type StderrDebugger struct {
	Truncate bool
	Width    int
}

func (s StderrDebugger) writeOut(prefix, str string) {
	width := s.Width
	if s.Width == 0 {
		width = defaultConsoleWidth
	}

	indent := "    "
	width -= len(indent)

	os.Stderr.WriteString(prefix + " ")
	if s.Truncate {
		if len(str) > width {
			str = str[0:width-3] + "..."
		}
		os.Stderr.WriteString(str + "\n")
		return
	}

	var i int
	for i = 1; i*width < len(str); i++ {
		os.Stderr.WriteString(str[(i-1)*width:i*width] + "\n" + indent)
	}
	os.Stderr.WriteString(str[(i-1)*width:] + "\n")
}

// Incoming implements Debugger.Incoming
func (s StderrDebugger) Incoming(b []byte) {
	s.writeOut(color.CyanString("<<<"), string(b))
}

// Outgoing implements Debugger.Outgoing
func (s StderrDebugger) Outgoing(b []byte) {
	s.writeOut(color.GreenString(">>>"), string(b))
}

// Connecting implements Debugger.Connecting
func (s StderrDebugger) Connecting(endpoint string) {
	s.writeOut(color.YellowString("CNX"), endpoint)
}

// Error implements Debugger.Error
func (s StderrDebugger) Error(e error) {
	col := color.New(color.FgBlack, color.BgRed)
	s.writeOut(col.SprintfFunc()("ERR"), e.Error())
}
