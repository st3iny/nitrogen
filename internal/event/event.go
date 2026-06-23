package event

import (
	"fmt"
	"log/slog"

	"github.com/fsnotify/fsnotify"
)

type Event = string

const InitialBuildEvent Event = "<initial build>"

func MapFSNotifyEvents(in <-chan fsnotify.Event, out chan<- Event) {
	for {
		e := <-in
		out <- e.Name
	}
}

func TapLogChanges(in <-chan Event, out chan<- Event) {
	for {
		e := <-in
		slog.Info(fmt.Sprintf("%v changed", e))
		out <- e
	}
}
