package util

import (
	"os"

	"github.com/fatih/color"
)

const consoleWidth = 79

// StderrDebugger prints debug information out on the console.
type StderrDebugger struct{}

func (s StderrDebugger) writeOut(prefix, str string, width int) {
	indent := "    "
	width -= len(indent)

	os.Stderr.WriteString(prefix + " ")
	var i int
	for i = 1; i*width < len(str); i++ {
		os.Stderr.WriteString(str[(i-1)*width:i*width] + "\n" + indent)
	}
	os.Stderr.WriteString(str[(i-1)*width:] + "\n")
}

// Incoming implements Debugger.Incoming
func (s StderrDebugger) Incoming(b []byte) {
	s.writeOut(color.CyanString("<<<"), string(b), consoleWidth)
}

// Outgoing implements Debugger.Outgoing
func (s StderrDebugger) Outgoing(b []byte) {
	s.writeOut(color.GreenString(">>>"), string(b), consoleWidth)
}

// Error implements Debugger.Error
func (s StderrDebugger) Error(e error) {
	col := color.New(color.FgBlack, color.BgRed)
	s.writeOut(col.SprintfFunc()("ERR"), e.Error(), consoleWidth)
}
