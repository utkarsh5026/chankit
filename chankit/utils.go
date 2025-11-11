package chankit

import (
	"context"
	"sync"
)

// Tap creates a channel that passes through all values from the input channel,
// calling tapFunc on each value without modifying it. This is useful for side effects
// like logging or metrics collection.
func Tap[T any](ctx context.Context, in <-chan T, tapFunc func(T), opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		for {
			val, ok := recieve(ctx, in)
			if !ok {
				return
			}

			tapFunc(val)

			if !send(ctx, outChan, val) {
				return
			}
		}
	}()

	return outChan
}

func FlatMap[T, R any](ctx context.Context, in <-chan T, flatMapFunc func(T) <-chan R, opts ...ChanOption[R]) <-chan R {
	outChan := applyChanOptions(opts...)

	go func() {
		var wg sync.WaitGroup

		defer func() {
			wg.Wait()
			close(outChan)
		}()

		for {
			val, ok := recieve(ctx, in)
			if !ok {
				return
			}

			innerChan := flatMapFunc(val)
			wg.Add(1)
			go func() {
				defer wg.Done()
				forwardSimple(ctx, outChan, innerChan)
			}()
		}
	}()

	return outChan
}
