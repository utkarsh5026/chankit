package chankit

import (
	"context"
	"time"
)

// Throttle limits the rate of values emitted from a channel by dropping intermediate values.
// Only the most recent value received within each time interval is emitted.
// This is useful for UI updates, event debouncing, or reducing high-frequency data streams.
//
// Example:
//
//	Input:  [1, 2, 3, 4, 5] (all arrive at time 0)
//	Duration: 100ms
//	Output: [5] (at 100ms) - values 1-4 were dropped
func Throttle[T any](ctx context.Context, in <-chan T, d time.Duration, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		ticker := time.NewTicker(d)
		defer ticker.Stop()

		var pending *T

		for {
			select {
			case <-ctx.Done():
				return

			case val, ok := <-in:
				if !ok {
					return
				}
				pending = &val

			case <-ticker.C:
				if pending != nil {
					select {
					case <-ctx.Done():
						return
					case outChan <- *pending:
						pending = nil
					}
				}
			}
		}
	}()

	return outChan
}

// FixedInterval processes every value from the input channel at a fixed interval.
// Unlike Throttle, this function does NOT drop values - it queues them and
// emits one value per time interval until all values are processed.
// This provides fixed-rate pacing rather than true rate limiting with bursts.
// Useful for API calls, downloads, or any operation requiring consistent spacing.
//
// Example:
//
//	Input:  [1, 2, 3, 4, 5] (all arrive at time 0)
//	Duration: 100ms
//	Output: [1] (at 100ms), [2] (at 200ms), [3] (at 300ms), [4] (at 400ms), [5] (at 500ms)
func FixedInterval[T any](ctx context.Context, in <-chan T, d time.Duration, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		ticker := time.NewTicker(d)
		defer ticker.Stop()

		var queue []T

		for {
			select {
			case <-ctx.Done():
				return

			case val, ok := <-in:
				if !ok {
					for len(queue) > 0 {
						select {
						case <-ctx.Done():
							return
						case <-ticker.C:
							select {
							case <-ctx.Done():
								return
							case outChan <- queue[0]:
								queue = queue[1:]
							}
						}
					}
					return
				}
				queue = append(queue, val)

			case <-ticker.C:
				if len(queue) > 0 {
					select {
					case <-ctx.Done():
						return
					case outChan <- queue[0]:
						queue = queue[1:]
					}
				}
			}
		}
	}()

	return outChan
}
