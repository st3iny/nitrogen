package event

import (
	"fmt"
	"log/slog"

	"github.com/fsnotify/fsnotify"
)

type Event = string

const InitialBuildEvent Event = "<initial build>"

func MapFSNotifyEvents(in <-chan fsnotify.Event, out chan<- Event) {
	defer close(out)

	for {
		e, ok := <-in
		if !ok {
			return
		}

		out <- e.Name
	}
}

func TapLogChanges(logger *slog.Logger, in <-chan Event, out chan<- Event) {
	defer close(out)

	for {
		e, ok := <-in
		if !ok {
			return
		}

		logger.Info(fmt.Sprintf("%v changed", e))
		out <- e
	}
}
