package chankit

import (
	"context"
	"time"
)

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
