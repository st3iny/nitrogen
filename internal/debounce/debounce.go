package debounce

import (
	"time"
)

func Debounce[T any](in <-chan T, out chan<- T, delay time.Duration) {
	var (
		event T
		timer <-chan time.Time
	)

	for {
		select {
		case event = <-in:
			timer = time.After(delay)
		case <-timer:
			out <- event
		}
	}
}
