package watcher

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatcher_Close(t *testing.T) {
	watch, err := New()
	require.NoError(t, err)

	err = watch.Close()
	assert.NoError(t, err)
}

func TestWatcher_Watch(t *testing.T) {
	tests := []struct {
		name        string
		eventFilter []fsnotify.Op
		assert      func(t *testing.T, actualEvents []fsnotify.Event, file string)
	}{
		{
			name: "without filters",
			assert: func(t *testing.T, actualEvents []fsnotify.Event, file string) {
				require.Len(t, actualEvents, 3)
				assert.Equal(t, file, actualEvents[0].Name)
				assert.Equal(t, fsnotify.Create, actualEvents[0].Op)
				assert.Equal(t, file, actualEvents[1].Name)
				assert.Equal(t, fsnotify.Write, actualEvents[1].Op)
				assert.Equal(t, file, actualEvents[2].Name)
				assert.Equal(t, fsnotify.Remove, actualEvents[2].Op)
			},
		},
		{
			name:        "with write filter",
			eventFilter: []fsnotify.Op{fsnotify.Write},
			assert: func(t *testing.T, actualEvents []fsnotify.Event, file string) {
				require.Len(t, actualEvents, 1)
				assert.Equal(t, file, actualEvents[0].Name)
				assert.Equal(t, fsnotify.Write, actualEvents[0].Op)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			watch, err := New()
			require.NoError(t, err)
			t.Cleanup(func() { _ = watch.Close() })

			if len(test.eventFilter) > 0 {
				watch.SetEventFilter(test.eventFilter...)
			}

			tempDir := t.TempDir()
			createTree(t, tempDir, map[string]any{
				"watched1": map[string]any{
					"file1": "file1",
					"dir1": map[string]any{
						"file2": "file2",
						"file3": "file3",
					},
					"excluded1": map[string]any{
						"file4": "file4",
						"file5": "file5",
					},
				},
				"unrelated1": map[string]any{
					"file6": "file6",
					"file7": "file7",
				},
			})

			err = watch.AddExcludeGlob("excluded*")
			require.NoError(t, err)

			watchedDir := path.Join(tempDir, "watched1")
			err = watch.AddDirRecursive(watchedDir)
			require.NoError(t, err)

			events := make(chan fsnotify.Event, 10)
			watchError := make(chan error, 1)

			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			t.Cleanup(cancel)

			var wg sync.WaitGroup
			wg.Go(func() {
				watchError <- watch.Watch(ctx, events)
			})

			time.Sleep(50 * time.Millisecond)
			file1Path := path.Join(watchedDir, "newFile")
			file1 := touchTempFile(t, file1Path)

			time.Sleep(50 * time.Millisecond)
			_, err = file1.WriteString("newFile")
			require.NoError(t, err)
			err = file1.Close()
			require.NoError(t, err)

			time.Sleep(50 * time.Millisecond)
			err = os.Remove(file1Path)
			require.NoError(t, err)

			time.Sleep(50 * time.Millisecond)
			touchTempFile(t, path.Join(watchedDir, "excluded1", "newExcludedFile"))

			time.Sleep(50 * time.Millisecond)
			touchTempFile(t, path.Join(tempDir, "unrelated1", "newUnrelatedFile"))

			var actualEvents []fsnotify.Event
			for {
				e, ok := <-events
				if !ok {
					break
				}

				actualEvents = append(actualEvents, e)
			}

			test.assert(t, actualEvents, file1Path)

			cancel()
			wg.Wait()
			assert.ErrorIs(t, <-watchError, context.DeadlineExceeded)
		})
	}
}

func TestWatcher_Watch_Closed(t *testing.T) {
	tests := []struct {
		name  string
		close func(w *Watcher)
	}{
		{
			name: "inner event channel",
			close: func(w *Watcher) {
				close(w.inner.Events)
			},
		},
		{
			name: "inner error channel",
			close: func(w *Watcher) {
				close(w.inner.Errors)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w, err := New()
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			errChan := make(chan error, 1)
			var wg sync.WaitGroup
			wg.Go(func() {
				errChan <- w.Watch(ctx, make(chan<- fsnotify.Event))
			})

			time.Sleep(50 * time.Millisecond)
			test.close(w)

			cancel()
			wg.Wait()
			assert.ErrorIs(t, <-errChan, ErrClosed)
		})
	}
}

func TestWatcher_Watch_LogsInnerErrors(t *testing.T) {
	w, err := New()
	require.NoError(t, err)

	handler := recordingHandler{}
	logger := slog.New(&handler)
	w.SetLogger(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	errChan := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Go(func() {
		errChan <- w.Watch(ctx, make(chan<- fsnotify.Event))
	})

	dummyError := errors.New("dummy error")
	w.inner.Errors <- dummyError

	cancel()
	wg.Wait()
	assert.ErrorIs(t, <-errChan, context.Canceled)

	require.Len(t, handler.records, 1)
	assert.Equal(t, handler.records[0].Level, slog.LevelError)
	assert.Equal(t, handler.records[0].Message, "dummy error")
}

func TestWatcher_AddExcludeGlob(t *testing.T) {
	t.Run("bad pattern", func(t *testing.T) {
		w, err := New()
		require.NoError(t, err)

		err = w.AddExcludeGlob("[")
		assert.ErrorContains(t, err, "unexpected end of input")
	})
}

func touchTempFile(t *testing.T, file string) *os.File {
	t.Helper()

	f, err := os.Create(file)
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	return f
}

func createTree(t *testing.T, base string, tree map[string]any) {
	t.Helper()

	for name, subtree := range tree {
		p := path.Join(base, name)
		switch v := subtree.(type) {
		case map[string]any:
			err := os.Mkdir(p, 0o755)
			require.NoError(t, err)
			createTree(t, p, v)
		case string:
			err := os.WriteFile(p, []byte(v), 0o644)
			require.NoError(t, err)
		default:
			t.FailNow()
		}
	}
}

type recordingHandler struct {
	records []slog.Record
}

func (r *recordingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (r *recordingHandler) Handle(ctx context.Context, record slog.Record) error {
	r.records = append(r.records, record)
	return nil
}

func (r *recordingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	panic("not implemented")
}

func (r *recordingHandler) WithGroup(name string) slog.Handler {
	panic("not implemented")
}
