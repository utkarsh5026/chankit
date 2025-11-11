package chankit

import (
	"context"
	"sync"
	"time"
)

func Delay[T any](ctx context.Context, in <-chan T, delay time.Duration, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		var wg sync.WaitGroup
		defer func() {
			wg.Wait()
			close(outChan)
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case val, ok := <-in:
				if !ok {
					return
				}

				wg.Add(1)
				go func(v T) {
					defer wg.Done()
					timer := time.NewTimer(delay)
					defer timer.Stop()

					select {
					case <-ctx.Done():
						return
					case <-timer.C:
						select {
						case <-ctx.Done():
							return
						case outChan <- v:
							return
						}
					}

				}(val)
			}
		}
	}()

	return outChan
}

func Timeout[T any](ctx context.Context, in <-chan T, timeout time.Duration, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-timer.C:
				return

			case val, ok := <-in:
				if !ok {
					return
				}

				timer.Reset(timeout)
				select {
				case <-ctx.Done():
					return
				case outChan <- val:
				}
			}
		}

	}()
	return outChan
}
