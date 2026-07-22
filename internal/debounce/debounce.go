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
		var ok bool
		select {
		case event, ok = <-in:
			if !ok {
				return
			}

			timer = time.After(delay)
		case <-timer:
			out <- event
		}
	}
}
