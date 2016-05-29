package main

import (
	"os"

	"github.com/WatchBeam/cord"
	"github.com/WatchBeam/cord/events"
	"github.com/WatchBeam/cord/model"
	"github.com/WatchBeam/cord/util"
)

func main() {
	c := cord.New(os.Args[1], &cord.WsOptions{
		Debugger: util.StderrDebugger{Truncate: true},
	})

	c.On(events.Ready(func(r *model.Ready) {
		// fmt.Printf("%+v\n", r)
	}))

	c.On(events.PresenceUpdate(func(r *model.PresenceUpdate) {
		// fmt.Printf("%+v\n", r)
	}))

	for err := range c.Errs() {
		// fmt.Printf("Got an error: %s", err)

		if _, isFatal := err.(cord.FatalError); isFatal {
			os.Exit(1)
		}
	}
}
