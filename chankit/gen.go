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
	ch := applyChanOptions(opts...)
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

// Repeat creates a channel that infinitely repeats the given value.
// The channel will close when the context is cancelled.
//
// Examples:
//   Repeat(ctx, 42)                      // unbuffered, repeats 42 forever
//   Repeat(ctx, "x", WithBuffer[string](5)) // buffered
func Repeat[T any](ctx context.Context, value T, opts ...ChanOption[T]) <-chan T {
	ch := applyChanOptions(opts...)

	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case ch <- value:
			}
		}
	}()

	return ch
}
