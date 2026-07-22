package debounce

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDebounce(t *testing.T) {
	t.Run("forwards events", func(t *testing.T) {
		in := make(chan string)
		out := make(chan string)

		var wg sync.WaitGroup
		wg.Go(func() { Debounce(in, out, 100*time.Millisecond) })

		expectedEvent := "event"

		now := time.Now()
		in <- expectedEvent
		actualEvent := <-out

		close(in)
		wg.Wait()

		delta := time.Now().Sub(now)
		assert.Greater(t, delta, 90*time.Millisecond)
		assert.Less(t, delta, 110*time.Millisecond)
		assert.Equal(t, expectedEvent, actualEvent)
	})
	t.Run("debounces multiple events", func(t *testing.T) {
		in := make(chan string)
		out := make(chan string)

		var wg sync.WaitGroup
		wg.Go(func() { Debounce(in, out, 100*time.Millisecond) })

		expectedEvent := "event"

		for i := range 5 {
			in <- expectedEvent
			if i < 4 {
				time.Sleep(50 * time.Millisecond)
			}
		}

		now := time.Now()
		actualEvent := <-out

		close(in)
		wg.Wait()

		delta := time.Now().Sub(now)
		assert.Greater(t, delta, 90*time.Millisecond)
		assert.Less(t, delta, 110*time.Millisecond)
		assert.Equal(t, expectedEvent, actualEvent)
	})
}
