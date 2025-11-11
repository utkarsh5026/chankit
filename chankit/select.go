package chankit

import "context"

// Take emits the first 'count' values from the input channel, then closes.
// This is useful for limiting the number of items processed from a potentially infinite stream.
// The output channel closes when 'count' items are taken, the input closes, or context is cancelled.
//
// Examples:
//
//	Take(ctx, ch, 5)                            // take first 5 values
//	Take(ctx, ch, 10, WithBuffer[int](5))       // with buffered output
//	Take(ctx, infinite, 100)                    // limit infinite stream
func Take[T any](ctx context.Context, in <-chan T, count int, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		taken := 0

		for {
			if taken >= count {
				return
			}

			val, ok := recieve(ctx, in)
			if !ok {
				return
			}

			if !send(ctx, outChan, val) {
				return
			}
			taken++
		}
	}()

	return outChan
}

// Skip discards the first 'count' values from the input channel, then emits all remaining values.
// This is useful for implementing pagination or skipping headers/initial values in a stream.
// The output channel closes when the input closes or context is cancelled.
//
// Examples:
//
//	Skip(ctx, ch, 10)                           // skip first 10 values
//	Skip(ctx, ch, 5, WithBuffer[int](3))        // with buffered output
//	Skip(ctx, ch, 0)                            // skip nothing (pass through)
func Skip[T any](ctx context.Context, in <-chan T, count int, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		skipped := 0

		for {
			val, ok := recieve(ctx, in)
			if !ok {
				return
			}

			if skipped < count {
				skipped++
				continue
			}

			if !send(ctx, outChan, val) {
				return
			}
		}
	}()

	return outChan
}

// TakeWhile emits values from the input channel as long as they satisfy the predicate.
// Once a value fails the predicate test, the output channel closes immediately.
// This is useful for processing streams until a sentinel value or condition is met.
// The output channel closes when predicate returns false, input closes, or context is cancelled.
//
// Examples:
//
//	TakeWhile(ctx, ch, func(x int) bool { return x < 100 })        // take until value >= 100
//	TakeWhile(ctx, ch, func(x int) bool { return x%2 == 0 })       // take consecutive even numbers
//	TakeWhile(ctx, ch, func(s string) bool { return s != "" })     // take until empty string
func TakeWhile[T any](ctx context.Context, in <-chan T, predicate func(T) bool, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		for {
			val, ok := recieve(ctx, in)
			if !ok {
				return
			}

			if !predicate(val) {
				return
			}

			if !send(ctx, outChan, val) {
				return
			}
		}
	}()

	return outChan
}

// SkipWhile discards values from the input channel as long as they satisfy the predicate.
// Once a value fails the predicate test, it and all subsequent values are emitted.
// This is useful for skipping initial values that meet certain criteria (like headers or warm-up data).
// The output channel closes when the input closes or context is cancelled.
//
// Examples:
//
//	SkipWhile(ctx, ch, func(x int) bool { return x < 0 })          // skip negative numbers, emit rest
//	SkipWhile(ctx, ch, func(x int) bool { return x%2 == 0 })       // skip leading even numbers
//	SkipWhile(ctx, ch, func(s string) bool { return s == "" })     // skip empty strings at start
func SkipWhile[T any](ctx context.Context, in <-chan T, predicate func(T) bool, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		skipping := true

		for {
			val, ok := recieve(ctx, in)
			if !ok {
				return
			}

			if skipping {
				if predicate(val) {
					continue
				}
				skipping = false
			}

			if !send(ctx, outChan, val) {
				return
			}
		}
	}()

	return outChan
}
