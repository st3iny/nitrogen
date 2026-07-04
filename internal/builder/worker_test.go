package builder

import (
	"context"
	"fmt"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/st3iny/nitrogen/internal/event"
	"github.com/stretchr/testify/assert"
)

func TestWorker_Run(t *testing.T) {
	t.Run("waits for initial event", func(t *testing.T) {
		testFile1 := path.Join(t.TempDir(), "test1")
		testFile2 := path.Join(t.TempDir(), "test2")
		worker := NewWorker([][]string{
			{"touch", testFile1},
			{"touch", testFile2},
		})

		in := make(chan event.Event)
		out := make(chan event.Event)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		var wg sync.WaitGroup
		wg.Go(func() { worker.Run(ctx, in, out) })
		time.Sleep(1 * time.Second)
		cancel()
		wg.Wait()

		assert.NoFileExists(t, testFile1)
		assert.NoFileExists(t, testFile2)
	})
	t.Run("runs build steps after receiving an event", func(t *testing.T) {
		testFile1 := path.Join(t.TempDir(), "test1")
		testFile2 := path.Join(t.TempDir(), "test2")
		worker := NewWorker([][]string{
			{"touch", testFile1},
			{"touch", testFile2},
		})

		in := make(chan event.Event)
		out := make(chan event.Event)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		var wg sync.WaitGroup
		wg.Go(func() { worker.Run(ctx, in, out) })

		in <- event.InitialBuildEvent
		<-out

		cancel()
		wg.Wait()

		assert.FileExists(t, testFile1)
		assert.FileExists(t, testFile2)
	})
	t.Run("cancels with context", func(t *testing.T) {
		testFile1 := path.Join(t.TempDir(), "test1")
		worker := NewWorker([][]string{
			{"touch", testFile1},
		})

		in := make(chan event.Event)
		out := make(chan event.Event)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		worker.Run(ctx, in, out)

		assert.NoFileExists(t, testFile1)
	})
	t.Run("forwards events", func(t *testing.T) {
		testFile1 := path.Join(t.TempDir(), "test1")
		testFile2 := path.Join(t.TempDir(), "test2")
		worker := NewWorker([][]string{
			{"touch", testFile1},
			{"touch", testFile2},
		})

		in := make(chan event.Event)
		out := make(chan event.Event)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		var wg sync.WaitGroup
		wg.Go(func() { worker.Run(ctx, in, out) })

		for i := range 5 {
			expectedEvent := fmt.Sprintf("%x", i)
			in <- expectedEvent
			actualEvent := <-out
			assert.Equal(t, expectedEvent, actualEvent)
		}

		cancel()
		wg.Wait()
	})
	t.Run("does not forward in case of build step error", func(t *testing.T) {
		worker := NewWorker([][]string{
			{"false"},
		})

		in := make(chan event.Event)
		out := make(chan event.Event)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		var wg sync.WaitGroup
		wg.Go(func() { worker.Run(ctx, in, out) })

		in <- event.InitialBuildEvent

		cancel()
		wg.Wait()

		_, ok := <-out
		assert.False(t, ok)
	})
}
