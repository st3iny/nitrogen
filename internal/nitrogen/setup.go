package nitrogen

import (
	"flag"
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/st3iny/nitrogen/internal/config"
	"github.com/st3iny/nitrogen/internal/replace"
	"github.com/st3iny/nitrogen/internal/watcher"
)

func setupFlags() {
	configFlag = flag.String("config", ".nitrogen.yaml", "path to a config file")
}

func setupWatcher(cfg *config.Config) (*watcher.Watcher, error) {
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

	for _, exclude := range cfg.Watch.Exclude {
		err = watch.AddExcludeGlob(exclude)
		if err != nil {
			return nil, fmt.Errorf("failed to add exclude glob: %w", err)
		}
	}

	replaces, err := replace.Parse(cfg.GoModFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse replace directives: %w", err)
	}

	includes := []string{"."}
	includes = append(includes, cfg.Watch.Include...)
	for _, repl := range replaces {
		includes = append(includes, repl.LocalPath)
	}

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
