package nitrogen

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/fsnotify/fsnotify"
	"github.com/st3iny/nitrogen/internal/builder"
	"github.com/st3iny/nitrogen/internal/config"
	"github.com/st3iny/nitrogen/internal/debounce"
	"github.com/st3iny/nitrogen/internal/event"
	"github.com/st3iny/nitrogen/internal/runner"
	"github.com/st3iny/nitrogen/internal/watcher"
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
		return err
	}

	slog.SetLogLoggerLevel(cfg.LogLevel)

	watcherChannel := make(chan fsnotify.Event, 1)
	reloadChannel := make(chan event.Event, 1)
	debouncedReloadChannel := make(chan event.Event, 1)
	rebuildChan := make(chan event.Event, 1)
	restartChan := make(chan event.Event, 1)

	go event.MapFSNotifyEvents(watcherChannel, reloadChannel)
	go debounce.Debounce(reloadChannel, debouncedReloadChannel, cfg.ReloadDelay)

	go event.TapLogChanges(debouncedReloadChannel, rebuildChan)

	go builder.NewWorker(cfg.Build).Run(rebuildChan, restartChan)
	go runner.NewWorker(cfg.Services).Run(restartChan)

	// Kick off initial build and start services
	rebuildChan <- event.InitialBuildEvent

	return runWatcher(&cfg.Watch, watcherChannel)
}

func setupFlags() {
	configFlag = flag.String("config", ".nitrogen.yaml", "path to a config file")
}

func runWatcher(cfg *config.WatchConfig, watcherChannel chan<- fsnotify.Event) error {
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

	for _, exclude := range cfg.Exclude {
		err = watch.AddExcludeGlob(exclude)
		if err != nil {
			return fmt.Errorf("failed to add exclude glob: %w", err)
		}
	}

	includes := append(cfg.Include, ".")
	for _, include := range includes {
		err = watch.AddDirRecursive(include)
		if err != nil {
			return fmt.Errorf("failed to watch dir: %w", err)
		}
	}

	return watch.Watch(context.Background(), watcherChannel)
}
