package builder

import (
	"fmt"
	"log/slog"

	"github.com/st3iny/nitrogen/internal/event"
)

type Worker struct {
	steps [][]string
}

func NewWorker(steps [][]string) *Worker {
	return &Worker{
		steps: steps,
	}
}

func (w *Worker) Run(rebuildChan <-chan event.Event, restartChan chan<- event.Event) {
	for {
		e := <-rebuildChan

		err := Build(w.steps)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to build: %v", err))
			continue
		}

		restartChan <- e
	}
}
