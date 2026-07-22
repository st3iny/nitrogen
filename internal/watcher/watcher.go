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
	logger      *slog.Logger
}

func New() (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil { // coverage-ignore // Not possible to mock fsnotify failure
		return nil, err
	}

	return &Watcher{
		inner:  watcher,
		logger: slog.Default(),
	}, nil
}

func (w *Watcher) SetLogger(logger *slog.Logger) {
	w.logger = logger
}

func (w *Watcher) Close() error {
	return w.inner.Close()
}

func (w *Watcher) Add(path string) error {
	w.logger.Info(fmt.Sprintf("watching %s", path))
	return w.inner.Add(path)
}

func (w *Watcher) AddDirRecursive(dir string) error {
	err := w.Add(dir)
	if err != nil { // coverage-ignore // Impractical to mock specific error cases
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil { // coverage-ignore // Impractical to mock specific error cases
		return err
	}

	for _, entry := range entries {
		fullPath := path.Join(dir, entry.Name())

		if w.isExcluded(entry.Name()) {
			w.logger.Info(fmt.Sprintf("ignoring %s", fullPath))
			continue
		}

		err = w.Add(fullPath)
		if err != nil { // coverage-ignore // Impractical to mock specific error cases
			return err
		}

		if entry.IsDir() {
			err = w.AddDirRecursive(fullPath)
			if err != nil { // coverage-ignore // Impractical to mock specific error cases
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
	defer close(events)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-w.inner.Events:
			if !ok {
				return ErrClosed
			}

			slog.Debug(event.String())
			if len(w.eventFilter) > 0 && !hasOp(event, w.eventFilter) {
				continue
			}

			events <- event
		case err, ok := <-w.inner.Errors:
			if !ok {
				return ErrClosed
			}

			w.logger.Error(err.Error())
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
