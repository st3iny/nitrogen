package runner

import (
	"context"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/st3iny/nitrogen/internal/config"
	"github.com/st3iny/nitrogen/internal/event"
	"github.com/stretchr/testify/assert"
)

func TestWorker_Run(t *testing.T) {
	t.Run("waits for an initial event", func(t *testing.T) {
		testFile1 := path.Join(t.TempDir(), "file1")
		testFile2 := path.Join(t.TempDir(), "file2")
		worker := NewWorker([]config.ServiceConfig{
			{[]string{"touch", testFile1}},
			{[]string{"touch", testFile2}},
		})

		in := make(chan event.Event)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		var wg sync.WaitGroup
		wg.Go(func() { worker.Run(ctx, in) })

		in <- event.InitialBuildEvent

		time.Sleep(1 * time.Second)
		cancel()
		wg.Wait()

		assert.FileExists(t, testFile1)
		assert.FileExists(t, testFile2)
	})
	t.Run("cancels with context", func(t *testing.T) {
		testFile1 := path.Join(t.TempDir(), "file1")
		worker := NewWorker([]config.ServiceConfig{
			{[]string{"touch", testFile1}},
		})

		in := make(chan event.Event)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		worker.Run(ctx, in)

		assert.NoFileExists(t, testFile1)
	})
}
