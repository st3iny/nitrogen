package nitrogen

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/goccy/go-yaml"
	"github.com/st3iny/nitrogen/internal/builder"
	"github.com/st3iny/nitrogen/internal/config"
	"github.com/st3iny/nitrogen/internal/debounce"
	"github.com/st3iny/nitrogen/internal/event"
	"github.com/st3iny/nitrogen/internal/runner"
)

var (
	printConfigFlag *bool
	configFlag      *string
)

func Run() error {
	setupFlags()
	flag.Parse()

	cfg, err := config.Parse(*configFlag)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	slog.SetLogLoggerLevel(cfg.LogLevel)

	if *printConfigFlag {
		return runPrintConfigCommand(cfg)
	}

	return runDefaultCommand(cfg)
}

func runPrintConfigCommand(cfg *config.Config) error {
	yamlEncoder := yaml.NewEncoder(os.Stdout)
	return errors.Join(
		yamlEncoder.Encode(cfg),
		yamlEncoder.Close(),
	)
}

func runDefaultCommand(cfg *config.Config) error {
	ctx := context.Background()

	watcherChannel := make(chan fsnotify.Event, 1)
	reloadChannel := make(chan event.Event, 1)
	debouncedReloadChannel := make(chan event.Event, 1)
	rebuildChan := make(chan event.Event, 1)
	restartChan := make(chan event.Event, 1)

	go event.MapFSNotifyEvents(watcherChannel, reloadChannel)
	go debounce.Debounce(reloadChannel, debouncedReloadChannel, cfg.DebounceDelay)

	go event.TapLogChanges(slog.Default(), debouncedReloadChannel, rebuildChan)

	go builder.NewWorker(cfg.Build).Run(ctx, rebuildChan, restartChan)
	go runner.NewWorker(cfg.Services).Run(ctx, restartChan)

	watch, err := setupWatcher(cfg)
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
