package event

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
)

func TestMapFSNotifyEvents(t *testing.T) {
	in := make(chan fsnotify.Event)
	out := make(chan Event)

	var wg sync.WaitGroup
	wg.Go(func() { MapFSNotifyEvents(in, out) })

	for i := range 10 {
		e := fsnotify.Event{
			Name: fmt.Sprintf("event %d", i),
		}
		in <- e
		actualEvent := <-out
		assert.Equal(t, e.Name, actualEvent)
	}

	close(in)
	wg.Wait()

	_, ok := <-out
	assert.False(t, ok)
}

func TestTapLogChanges(t *testing.T) {
	in := make(chan Event)
	out := make(chan Event)

	buf := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewTextHandler(buf, nil))

	var wg sync.WaitGroup
	wg.Go(func() { TapLogChanges(logger, in, out) })

	for i := range 10 {
		expectedEvent := fmt.Sprintf("file %d", i)
		in <- expectedEvent
		actualEvent := <-out
		assert.Equal(t, expectedEvent, actualEvent)
		assert.Contains(t, buf.String(), fmt.Sprintf("file %d changed", i))
	}

	close(in)
	wg.Wait()

	_, ok := <-out
	assert.False(t, ok)
}
