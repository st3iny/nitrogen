package watcher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
	"github.com/gobwas/glob"
)

var ErrClosed = errors.New("watcher closed")

type Watcher struct {
	inner       *fsnotify.Watcher
	eventFilter []fsnotify.Op
	excludes    []glob.Glob
}

func New() (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		inner: watcher,
	}, nil
}

func (w *Watcher) Close() error {
	return w.inner.Close()
}

func (w *Watcher) Add(path string) error {
	slog.Info(fmt.Sprintf("watching %s", path))
	return w.inner.Add(path)
}

func (w *Watcher) AddDirRecursive(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := path.Join(dir, entry.Name())

		if w.isExcluded(entry.Name()) {
			slog.Info(fmt.Sprintf("ignoring %s", fullPath))
			continue
		}

		err = w.Add(fullPath)
		if err != nil {
			return err
		}

		if entry.IsDir() {
			err = w.AddDirRecursive(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Watcher) AddExcludeGlob(pattern string) error {
	exclude, err := glob.Compile(pattern, '/')
	if err != nil {
		return err
	}

	w.excludes = append(w.excludes, exclude)
	return nil
}

func (w *Watcher) SetEventFilter(ops ...fsnotify.Op) {
	w.eventFilter = ops
}

func (w *Watcher) Watch(ctx context.Context, events chan<- fsnotify.Event) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-w.inner.Events:
			if !ok {
				return ErrClosed
			}

			slog.Debug(event.String())
			if hasOp(event, w.eventFilter) {
				events <- event
			}
		case err, ok := <-w.inner.Errors:
			if !ok {
				return ErrClosed
			}

			slog.Error(err.Error())
		}
	}
}

func (w *Watcher) isExcluded(p string) bool {
	for _, exclude := range w.excludes {
		if exclude.Match(path.Base(p)) {
			return true
		}
	}

	return false
}

func hasOp(event fsnotify.Event, ops []fsnotify.Op) bool {
	for _, op := range ops {
		if event.Has(op) {
			return true
		}
	}

	return false
}
