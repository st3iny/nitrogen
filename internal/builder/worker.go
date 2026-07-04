package builder

import (
	"context"
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

func (w *Worker) Run(ctx context.Context, rebuildChan <-chan event.Event, restartChan chan<- event.Event) {
	defer close(restartChan)

	for {
		var e event.Event
		select {
		case <-ctx.Done():
			return
		case e = <-rebuildChan:
		}

		err := Build(w.steps)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to build: %v", err))
			continue
		}

		restartChan <- e
	}
}
