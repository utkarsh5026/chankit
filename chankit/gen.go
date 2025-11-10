package chankit

import "context"

// Generate creates a channel that produces values from a generator function.
// The generator function returns (value, true) to produce a value, or (zero, false) to stop.
// By default, uses an unbuffered channel. Use WithBuffer() to add buffering.
//
// Examples:
//   Generate(ctx, genFunc)                      // unbuffered
//   Generate(ctx, genFunc, WithBuffer[int](10)) // buffered with size 10
func Generate[T any](ctx context.Context, genFunc func() (T, bool), opts ...ChanOption[T]) <-chan T {
	cfg := &chanConfig[T]{bufferSize: 0}

	for _, opt := range opts {
		opt(cfg)
	}

	ch := make(chan T, cfg.bufferSize)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if val, ok := genFunc(); ok {
					select {
					case <-ctx.Done():
						return
					case ch <- val:
					}
				} else {
					return
				}
			}
		}
	}()
	return ch
}
