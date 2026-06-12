package nitrogen

import (
	"context"
	"log/slog"

	"github.com/fsnotify/fsnotify"
	"github.com/st3iny/nitrogen/internal/builder"
	"github.com/st3iny/nitrogen/internal/config"
	"github.com/st3iny/nitrogen/internal/debounce"
	"github.com/st3iny/nitrogen/internal/runner"
	"github.com/st3iny/nitrogen/internal/watcher"
)

var version = "git"

func Run() error {
	slog.Info("starting nitrogen", "version", version)

	cfg, err := config.Parse("config.yaml")
	if err != nil {
		return err
	}

	slog.SetLogLoggerLevel(cfg.LogLevel)

	watcherChannel := make(chan fsnotify.Event, 1)
	reloadEvents := make(chan struct{}, 1)
	rebuildChan := make(chan struct{}, 1)
	restartChan := make(chan struct{}, 1)

	go pipeRebuildEvents(watcherChannel, reloadEvents)
	go debounce.Debounce(reloadEvents, rebuildChan, cfg.ReloadDelay)

	go builder.NewWorker(cfg.Build).Run(rebuildChan, restartChan)
	go runner.NewWorker(cfg.Services).Run(restartChan)

	// Kick off initial build and start services
	rebuildChan <- struct{}{}

	watch, err := watcher.New()
	if err != nil {
		return err
	}
	defer watch.Close()

	watch.SetEventFilter(
		fsnotify.Write,
		fsnotify.Remove,
		fsnotify.Create,
		fsnotify.Rename,
	)
	err = watch.AddDirRecursive(".")
	if err != nil {
		return err
	}

	return watch.Watch(context.Background(), watcherChannel)
}

func pipeRebuildEvents(in <-chan fsnotify.Event, out chan<- struct{}) {
	for {
		<-in
		out <- struct{}{}
	}
}
