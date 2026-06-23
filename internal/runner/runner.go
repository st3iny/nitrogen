package runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/st3iny/nitrogen/internal/config"
)

type Runner struct {
	Services []config.ServiceConfig
}

func (r *Runner) RunServices(ctx context.Context) error {
	slog.Info(fmt.Sprintf("starting %d services", len(r.Services)))

	var wg sync.WaitGroup
	for _, svc := range r.Services {
		wg.Go(func() {
			err := r.runService(ctx, &svc)
			if err != nil {
				slog.Error(fmt.Sprintf("failed to run services: %v", err))
			}
		})
	}

	<-ctx.Done()
	wg.Wait()
	return nil
}

func (r *Runner) runService(ctx context.Context, svc *config.ServiceConfig) error {
	for ctx.Err() == nil {
		cmd := exec.CommandContext(ctx, svc.Command[0], svc.Command[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil && ctx.Err() == nil {
			slog.Error(fmt.Sprintf("failed to run service: %v", err))
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}
