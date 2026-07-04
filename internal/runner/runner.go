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
	Services     []config.ServiceConfig
	RestartDelay time.Duration
}

func (r *Runner) RunServices(ctx context.Context) {
	slog.Info(fmt.Sprintf("starting %d services", len(r.Services)))

	var wg sync.WaitGroup
	for _, svc := range r.Services {
		wg.Go(func() {
			r.runService(ctx, &svc)
		})
	}

	<-ctx.Done()
	wg.Wait()
}

func (r *Runner) runService(ctx context.Context, svc *config.ServiceConfig) {
	for ctx.Err() == nil {
		cmd := exec.CommandContext(ctx, svc.Command[0], svc.Command[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil && ctx.Err() == nil {
			slog.Error(fmt.Sprintf("failed to run service: %v", err))
		}

		time.Sleep(r.RestartDelay)
	}
}
