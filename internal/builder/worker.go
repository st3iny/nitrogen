package builder

import (
	"fmt"
	"log/slog"
)

type Worker struct {
	steps [][]string
}

func NewWorker(steps [][]string) *Worker {
	return &Worker{
		steps: steps,
	}
}

func (w *Worker) Run(rebuildChan <-chan struct{}, restartChan chan<- struct{}) {
	for {
		<-rebuildChan

		err := Build(w.steps)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to build: %v", err))
			continue
		}

		restartChan <- struct{}{}
	}
}
