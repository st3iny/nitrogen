package runner

import (
	"context"
	"log/slog"

	"github.com/st3iny/nitrogen/internal/config"
	"github.com/st3iny/nitrogen/internal/event"
)

type Worker struct {
	runner Runner
}

func NewWorker(services []config.ServiceConfig) *Worker {
	return &Worker{
		runner: Runner{
			Services: services,
		},
	}
}

func (r *Worker) Run(restartChan <-chan event.Event) {
	// Delay initial start until receiving a restart event from the builder
	<-restartChan

	for {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-restartChan
			cancel()
		}()

		err := r.runner.RunServices(ctx)
		if err != nil {
			slog.Error(err.Error(), "component", "service-runner")
		}
	}
}
