package chankit

import "context"

// Map applies a transformation function to each value from the input channel.
// The output channel closes when the input closes or context is cancelled.
//
// Examples:
//
//	Map(ctx, ch, func(x int) int { return x * 2 })           // double values
//	Map(ctx, ch, func(x int) string { return fmt.Sprint(x) }) // int to string
//	Map(ctx, ch, mapFunc, WithBuffer[string](10))            // with buffering
func Map[T, R any](ctx context.Context, in <-chan T, mapFunc func(T) R, opts ...ChanOption[R]) <-chan R {
	outChan := applyChanOptions(opts...)
	go func() {
		defer close(outChan)
		for val := range in {
			select {
			case <-ctx.Done():
				return
			case outChan <- mapFunc(val):
			}
		}
	}()
	return outChan
}

// Filter creates a channel that only emits values satisfying the predicate function.
// The output channel closes when the input closes or context is cancelled.
//
// Examples:
//
//	Filter(ctx, ch, func(x int) bool { return x%2 == 0 })    // even numbers only
//	Filter(ctx, ch, func(x int) bool { return x > 0 })       // positive numbers
//	Filter(ctx, ch, filterFunc, WithBuffer[int](5))          // with buffering
func Filter[T any](ctx context.Context, in <-chan T, filterFunc func(T) bool, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		for val := range in {
			select {
			case <-ctx.Done():
				return
			default:
				ok := filterFunc(val)
				if ok {
					select {
					case <-ctx.Done():
						return
					case outChan <- val:
					}
				}
			}
		}
	}()
	return outChan
}

// Reduce aggregates all values from the input channel into a single result.
// This is a blocking operation that returns when the channel closes or context is cancelled.
//
// Examples:
//
//	Reduce(ctx, ch, func(sum, x int) int { return sum + x }, 0)     // sum all values
//	Reduce(ctx, ch, func(max, x int) int { ... }, 0)                // find maximum
//	Reduce(ctx, ch, func(acc, x string) string { ... }, "")         // concatenate strings
func Reduce[T, R any](ctx context.Context, in <-chan T, reduceFunc func(R, T) R, initial R) R {
	accumulator := initial
	for {
		select {
		case <-ctx.Done():
			return accumulator
		case val, ok := <-in:
			if !ok {
				return accumulator
			}
			accumulator = reduceFunc(accumulator, val)
		}
	}
}
