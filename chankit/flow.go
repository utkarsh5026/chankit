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

func Batch[T any](ctx context.Context, in <-chan T, batchSize int, timeout time.Duration, opts ...ChanOption[[]T]) <-chan []T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		var batch []T
		var timer *time.Timer
		var timerCh <-chan time.Time

		sendBatch := func() {
			if len(batch) > 0 {
				outChan <- batch
				batch = nil
			}
			if timer != nil {
				timer.Stop()
				timerCh = nil
			}
		}

		for {
			select {
			case <-ctx.Done():
				sendBatch()
				return

			case val, ok := <-in:
				if !ok {
					sendBatch()
					return
				}

				if len(batch) == 0 {
					if timer == nil {
						timer = time.NewTimer(timeout)
					} else {
						timer.Reset(timeout)
					}
					timerCh = timer.C
				}

				batch = append(batch, val)

				if len(batch) >= batchSize {
					select {
					case outChan <- batch:
						batch = nil
						timer.Stop()
						timerCh = nil
					case <-ctx.Done():
						sendBatch()
						return
					}
				}

			case <-timerCh:
				sendBatch()
			}
		}
	}()

	return outChan
}

// Debounce emits values from input only after the specified duration has elapsed
// without any new values arriving. If a new value arrives before the duration
// elapses, the timer is reset. This is useful for handling rapid bursts of events
// where you only want to process the final value after activity stops.
//
// Example:
//
//	Input:  [1, 2, 3] (arrive within 100ms of each other)
//	Duration: 100ms
//	Output: [3] (only after 100ms of silence after receiving 3)
func Debounce[T any](ctx context.Context, in <-chan T, d time.Duration, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)

		var timer *time.Timer
		var timerCh <-chan time.Time
		var pending *T

		for {
			select {
			case <-ctx.Done():
				if timer != nil {
					timer.Stop()
				}
				return

			case val, ok := <-in:
				if !ok {
					if pending != nil {
						select {
						case outChan <- *pending:
						case <-ctx.Done():
						}
					}
					if timer != nil {
						timer.Stop()
					}
					return
				}

				pending = &val

				if timer == nil {
					timer = time.NewTimer(d)
					timerCh = timer.C
				} else {
					timer.Stop()
					timer.Reset(d)
				}

			case <-timerCh:
				if pending != nil {
					select {
					case outChan <- *pending:
						pending = nil
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return outChan
}
