package runner

import (
	"context"
	"time"

	"github.com/st3iny/nitrogen/internal/config"
	"github.com/st3iny/nitrogen/internal/event"
)

type Worker struct {
	runner Runner
}

func NewWorker(services []config.ServiceConfig) *Worker {
	return &Worker{
		runner: Runner{
			Services:     services,
			RestartDelay: 1 * time.Second,
		},
	}
}

func (r *Worker) Run(ctx context.Context, restartChan <-chan event.Event) {
	// Delay initial start until receiving a restart event from the builder
	select {
	case <-ctx.Done():
		return
	case <-restartChan:
	}

	for ctx.Err() == nil {
		servicesCtx, cancel := context.WithCancel(ctx)
		go func() {
			<-restartChan
			cancel()
		}()

		r.runner.RunServices(servicesCtx)
	}
}
