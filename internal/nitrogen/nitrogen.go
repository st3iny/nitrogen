package nitrogen

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"

	"github.com/fsnotify/fsnotify"
	"github.com/st3iny/nitrogen/internal/builder"
	"github.com/st3iny/nitrogen/internal/config"
	"github.com/st3iny/nitrogen/internal/debounce"
	"github.com/st3iny/nitrogen/internal/event"
	"github.com/st3iny/nitrogen/internal/runner"
)

var (
	version = "git"

	configFlag *string
)

func Run() error {
	slog.Info("starting nitrogen", "version", version)

	setupFlags()
	flag.Parse()

	cfg, err := config.Parse(*configFlag)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	slog.SetLogLoggerLevel(cfg.LogLevel)

	watcherChannel := make(chan fsnotify.Event, 1)
	reloadChannel := make(chan event.Event, 1)
	debouncedReloadChannel := make(chan event.Event, 1)
	rebuildChan := make(chan event.Event, 1)
	restartChan := make(chan event.Event, 1)

	go event.MapFSNotifyEvents(watcherChannel, reloadChannel)
	go debounce.Debounce(reloadChannel, debouncedReloadChannel, cfg.DebounceDelay)

	go event.TapLogChanges(debouncedReloadChannel, rebuildChan)

	go builder.NewWorker(cfg.Build).Run(rebuildChan, restartChan)
	go runner.NewWorker(cfg.Services).Run(restartChan)

	watch, err := setupWatcher(&cfg.Watch, watcherChannel)
	if err != nil {
		return fmt.Errorf("failed to setup watcher: %w", err)
	}

	// Kick off initial build to start services
	rebuildChan <- event.InitialBuildEvent

	return errors.Join(
		watch.Watch(context.Background(), watcherChannel),
		watch.Close(),
	)
}
