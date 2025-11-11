package chankit

import "context"

// Generate creates a channel that produces values from a generator function.
// The generator function returns (value, true) to produce a value, or (zero, false) to stop.
// By default, uses an unbuffered channel. Use WithBuffer() to add buffering.
//
// Examples:
//
//	Generate(ctx, genFunc)                      // unbuffered
//	Generate(ctx, genFunc, WithBuffer[int](10)) // buffered with size 10
func Generate[T any](ctx context.Context, genFunc func() (T, bool), opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)
	go func() {
		defer close(outChan)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				val, ok := genFunc()
				if !ok {
					return
				}

				if !send(ctx, outChan, val) {
					return
				}
			}
		}
	}()
	return outChan
}

// Repeat creates a channel that infinitely repeats the given value.
// The channel will close when the context is cancelled.
//
// Examples:
//
//	Repeat(ctx, 42)                      // unbuffered, repeats 42 forever
//	Repeat(ctx, "x", WithBuffer[string](5)) // buffered
func Repeat[T any](ctx context.Context, value T, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		for {
			if !send(ctx, outChan, value) {
				return
			}
		}
	}()

	return outChan
}

// Range creates a channel that produces values from start to end (exclusive) with the given step.
// For positive steps: generates [start, start+step, start+2*step, ...) while i < end
// For negative steps: generates [start, start+step, start+2*step, ...) while i > end
//
// Examples:
//
//	Range(ctx, 0, 10, 1)           // 0, 1, 2, ..., 9
//	Range(ctx, 10, 0, -1)          // 10, 9, 8, ..., 1
//	Range(ctx, 0, 5, 2)            // 0, 2, 4
//	Range(ctx, 0, 10, 1, WithBuffer[int](5))  // buffered
func Range[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](ctx context.Context, start, end, step T, opts ...ChanOption[T]) <-chan T {
	ch := applyChanOptions(opts...)

	go func() {
		defer close(ch)

		if step > 0 {
			for i := start; i < end; i += step {
				if !send(ctx, ch, i) {
					return
				}
			}
			return
		}

		if step < 0 {
			for i := start; i > end; i += step {
				if !send(ctx, ch, i) {
					return
				}
			}
		}
	}()

	return ch
}
