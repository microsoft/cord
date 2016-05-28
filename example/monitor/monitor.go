package main

import (
	"fmt"
	"os"

	"github.com/WatchBeam/cord"
	"github.com/WatchBeam/cord/events"
	"github.com/WatchBeam/cord/model"
	"github.com/WatchBeam/cord/util"
)

func main() {
	c := cord.New(os.Args[1], &cord.WsOptions{
		Debugger: util.StderrDebugger{},
	})

	c.On(events.Ready(func(r *model.Ready) {
		fmt.Printf("%+v\n", r)
	}))

	c.On(events.PresenceUpdate(func(r *model.PresenceUpdate) {
		fmt.Printf("%+v\n", r)
	}))

	select {}
}
