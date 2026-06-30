package nitrogen

import (
	"flag"
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/st3iny/nitrogen/internal/config"
	"github.com/st3iny/nitrogen/internal/watcher"
)

func setupFlags() {
	configFlag = flag.String("config", ".nitrogen.yaml", "path to a config file")
}

func setupWatcher(cfg *config.WatchConfig, watcherChannel chan<- fsnotify.Event) (*watcher.Watcher, error) {
	watch, err := watcher.New()
	if err != nil {
		return nil, err
	}

	watch.SetEventFilter(
		fsnotify.Write,
		fsnotify.Remove,
		fsnotify.Create,
		fsnotify.Rename,
	)

	for _, exclude := range cfg.Exclude {
		err = watch.AddExcludeGlob(exclude)
		if err != nil {
			return nil, fmt.Errorf("failed to add exclude glob: %w", err)
		}
	}

	includes := append(cfg.Include, ".")
	for _, include := range includes {
		dir, err := isDir(include)
		if err != nil {
			continue
		}

		if dir {
			err = watch.AddDirRecursive(include)
		} else {
			err = watch.Add(include)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to watch: %w", err)
		}
	}

	return watch, nil
}
