package chankit

import (
	"context"
	"time"
)

// Pipeline provides a fluent, chainable API for channel operations.
// It wraps a channel and provides methods that return new Pipeline instances,
// allowing for elegant method chaining instead of deeply nested function calls.
//
// Example:
//
//	result := chankit.NewPipeline[int](ctx).
//	    Range(1, 100, 1).
//	    Map(func(x int) int { return x * x }).
//	    Filter(func(x int) bool { return x%2 == 0 }).
//	    Take(10).
//	    ToSlice()
type Pipeline[T any] struct {
	ctx context.Context
	ch  <-chan T
}

// NewPipeline creates a new empty Pipeline with the given context.
// This is typically used as a starting point, followed by a generator method
// like Range, Repeat, or Generate.
//
// Example:
//
//	pipeline := chankit.NewPipeline[int](ctx).Range(1, 10, 1)
func NewPipeline[T any](ctx context.Context) *Pipeline[T] {
	return &Pipeline[T]{
		ctx: ctx,
		ch:  nil,
	}
}

// From creates a Pipeline from an existing channel.
// This allows you to use pipeline operations on channels created elsewhere.
//
// Example:
//
//	ch := make(chan int)
//	pipeline := chankit.From(ctx, ch).Map(func(x int) int { return x * 2 })
func From[T any](ctx context.Context, ch <-chan T) *Pipeline[T] {
	return &Pipeline[T]{
		ctx: ctx,
		ch:  ch,
	}
}

// FromSlice creates a Pipeline from a slice.
//
// Example:
//
//	pipeline := chankit.FromSlice(ctx, []int{1, 2, 3, 4, 5}).
//	    Map(func(x int) int { return x * 2 })
func FromSlice[T any](ctx context.Context, slice []T) *Pipeline[T] {
	ch := SliceToChan(ctx, slice)
	return From(ctx, ch)
}

// ============================================================================
// Generator Methods
// ============================================================================

// RangePipeline generates a sequence of numbers from start to end (exclusive) with the given step.
// Note: This is only available for numeric types (int, float64, etc.)
//
// Example:
//
//	pipeline := chankit.NewPipeline[int](ctx).RangePipeline(1, 10, 1)  // 1, 2, 3, ..., 9
func RangePipeline[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](ctx context.Context, start, end, step T) *Pipeline[T] {
	ch := Range(ctx, start, end, step)
	return From(ctx, ch)
}

// Repeat generates an infinite stream of the same value.
// Use Take or TakeWhile to limit the output.
//
// Example:
//
//	pipeline := chankit.NewPipeline[string](ctx).
//	    Repeat("ping").
//	    Take(5)  // ["ping", "ping", "ping", "ping", "ping"]
func (p *Pipeline[T]) Repeat(value T) *Pipeline[T] {
	ch := Repeat(p.ctx, value)
	return From(p.ctx, ch)
}

// Generate creates values using a generator function.
// The generator should return (value, true) to emit a value, or (zero, false) to stop.
//
// Example:
//
//	i := 0
//	pipeline := chankit.NewPipeline[int](ctx).Generate(func() (int, bool) {
//	    i++
//	    return i, i <= 10
//	})
func (p *Pipeline[T]) Generate(genFunc func() (T, bool), opts ...ChanOption[T]) *Pipeline[T] {
	ch := Generate(p.ctx, genFunc, opts...)
	return From(p.ctx, ch)
}

// ============================================================================
// Transformation Methods
// ============================================================================

// Map transforms each value using the provided function.
// The function can change the type of the values.
//
// Example:
//
//	pipeline.Map(func(x int) int { return x * 2 })
//	pipeline.Map(func(x int) string { return fmt.Sprintf("num_%d", x) })
func (p *Pipeline[T]) Map(fn func(T) any) *Pipeline[any] {
	ch := Map(p.ctx, p.ch, fn)
	return From(p.ctx, ch)
}

// MapTo is a type-safe version of Map that explicitly specifies the output type.
//
// Example:
//
//	pipeline.MapTo(func(x int) string { return fmt.Sprint(x) })
func MapTo[T, R any](p *Pipeline[T], fn func(T) R) *Pipeline[R] {
	ch := Map(p.ctx, p.ch, fn)
	return From(p.ctx, ch)
}

// Filter keeps only values that satisfy the predicate.
//
// Example:
//
//	pipeline.Filter(func(x int) bool { return x%2 == 0 })  // even numbers only
//	pipeline.Filter(func(x int) bool { return x > 10 })    // numbers > 10
func (p *Pipeline[T]) Filter(fn func(T) bool) *Pipeline[T] {
	ch := Filter(p.ctx, p.ch, fn)
	return From(p.ctx, ch)
}

// FlatMap transforms each value into a channel and flattens the results.
//
// Example:
//
//	pipeline.FlatMap(func(x int) <-chan int {
//	    ch := make(chan int, 2)
//	    go func() {
//	        defer close(ch)
//	        ch <- x
//	        ch <- x * 2
//	    }()
//	    return ch
//	})
func (p *Pipeline[T]) FlatMap(fn func(T) <-chan T) *Pipeline[T] {
	ch := FlatMap(p.ctx, p.ch, fn)
	return From(p.ctx, ch)
}

// ============================================================================
// Selection Methods
// ============================================================================

// Take emits only the first n values.
//
// Example:
//
//	pipeline.Take(5)  // first 5 values only
func (p *Pipeline[T]) Take(n int) *Pipeline[T] {
	ch := Take(p.ctx, p.ch, n)
	return From(p.ctx, ch)
}

// Skip discards the first n values and emits the rest.
//
// Example:
//
//	pipeline.Skip(5)  // skip first 5 values
func (p *Pipeline[T]) Skip(n int) *Pipeline[T] {
	ch := Skip(p.ctx, p.ch, n)
	return From(p.ctx, ch)
}

// TakeWhile emits values as long as the predicate is true.
// Stops at the first false value.
//
// Example:
//
//	pipeline.TakeWhile(func(x int) bool { return x < 10 })
func (p *Pipeline[T]) TakeWhile(fn func(T) bool) *Pipeline[T] {
	ch := TakeWhile(p.ctx, p.ch, fn)
	return From(p.ctx, ch)
}

// SkipWhile discards values as long as the predicate is true.
// Starts emitting after the first false value.
//
// Example:
//
//	pipeline.SkipWhile(func(x int) bool { return x < 10 })
func (p *Pipeline[T]) SkipWhile(fn func(T) bool) *Pipeline[T] {
	ch := SkipWhile(p.ctx, p.ch, fn)
	return From(p.ctx, ch)
}

// ============================================================================
// Flow Control Methods
// ============================================================================

// Throttle limits the rate by emitting only the most recent value per interval.
// Intermediate values are dropped.
//
// Example:
//
//	pipeline.Throttle(100 * time.Millisecond)  // at most 1 value per 100ms
func (p *Pipeline[T]) Throttle(d time.Duration) *Pipeline[T] {
	ch := Throttle(p.ctx, p.ch, d)
	return From(p.ctx, ch)
}

// Debounce emits values only after a period of silence.
// The timer resets on each new value.
//
// Example:
//
//	pipeline.Debounce(300 * time.Millisecond)  // wait 300ms of silence
func (p *Pipeline[T]) Debounce(d time.Duration) *Pipeline[T] {
	ch := Debounce(p.ctx, p.ch, d)
	return From(p.ctx, ch)
}

// FixedInterval emits values at a fixed rate, queueing them without dropping.
//
// Example:
//
//	pipeline.FixedInterval(100 * time.Millisecond)  // 1 value every 100ms
func (p *Pipeline[T]) FixedInterval(d time.Duration) *Pipeline[T] {
	ch := FixedInterval(p.ctx, p.ch, d)
	return From(p.ctx, ch)
}

// Batch groups values into slices based on size or timeout.
// Returns a channel of slices instead of a Pipeline to avoid type complexity.
//
// Example:
//
//	batches := pipeline.Batch(10, 1*time.Second)
//	for batch := range batches {
//	    fmt.Printf("Got batch of %d items\n", len(batch))
//	}
func (p *Pipeline[T]) Batch(size int, timeout time.Duration) <-chan []T {
	return Batch(p.ctx, p.ch, size, timeout)
}

// ============================================================================
// Side Effect Methods
// ============================================================================

// Tap allows you to observe values without modifying them.
// This is useful for logging, debugging, or side effects.
//
// Example:
//
//	pipeline.Tap(func(x int) { fmt.Printf("Value: %d\n", x) })
func (p *Pipeline[T]) Tap(fn func(T)) *Pipeline[T] {
	ch := Tap(p.ctx, p.ch, fn)
	return From(p.ctx, ch)
}

// ============================================================================
// Combining Methods
// ============================================================================

// Merge combines this pipeline with other channels into a single stream.
//
// Example:
//
//	ch1 := chankit.NewPipeline[int](ctx).Range(1, 5, 1)
//	ch2 := chankit.NewPipeline[int](ctx).Range(10, 15, 1)
//	merged := ch1.Merge(ch2.Chan())
func (p *Pipeline[T]) Merge(channels ...<-chan T) *Pipeline[T] {
	allChannels := append([]<-chan T{p.ch}, channels...)
	ch := Merge(p.ctx, allChannels...)
	return From(p.ctx, ch)
}

// ZipWith combines this pipeline with another channel into pairs.
// Returns a pipeline of structs containing First and Second fields.
//
// Example:
//
//	numbers := chankit.RangePipeline[int](ctx, 1, 5, 1)
//	letters := chankit.FromSlice(ctx, []string{"a", "b", "c", "d"})
//	zipped := ZipWith(numbers, letters.Chan())
func ZipWith[T, R any](p *Pipeline[T], other <-chan R) *Pipeline[struct {
	First  T
	Second R
}] {
	ch := Zip(p.ctx, p.ch, other)
	return From(p.ctx, ch)
}

// ============================================================================
// Terminal Operations (these consume the pipeline and return results)
// ============================================================================

// ToSlice collects all values into a slice.
// This is a blocking operation that waits for the channel to close.
//
// Example:
//
//	result := pipeline.ToSlice()
func (p *Pipeline[T]) ToSlice() []T {
	return ChanToSlice(p.ctx, p.ch)
}

// Reduce aggregates all values into a single result.
// This is a blocking operation.
//
// Example:
//
//	sum := pipeline.Reduce(func(acc, x int) int { return acc + x }, 0)
func (p *Pipeline[T]) Reduce(fn func(acc, val T) T, initial T) T {
	return Reduce(p.ctx, p.ch, fn, initial)
}

// ReduceTo is a type-safe version of Reduce that can change the accumulator type.
//
// Example:
//
//	sum := ReduceTo(pipeline, func(acc int, x int) int { return acc + x }, 0)
func ReduceTo[T, R any](p *Pipeline[T], fn func(acc R, val T) R, initial R) R {
	return Reduce(p.ctx, p.ch, fn, initial)
}

// ForEach executes a function for each value in the pipeline.
// This is a blocking operation.
//
// Example:
//
//	pipeline.ForEach(func(x int) { fmt.Println(x) })
func (p *Pipeline[T]) ForEach(fn func(T)) {
	for {
		val, ok := recieve(p.ctx, p.ch)
		if !ok {
			return
		}
		fn(val)
	}
}

// Count returns the number of values in the pipeline.
// This is a blocking operation.
//
// Example:
//
//	count := pipeline.Count()
func (p *Pipeline[T]) Count() int {
	count := 0
	for {
		_, ok := recieve(p.ctx, p.ch)
		if !ok {
			return count
		}
		count++
	}
}

// Chan returns the underlying channel.
// This allows you to use the pipeline with other channel operations.
//
// Example:
//
//	ch := pipeline.Chan()
//	for val := range ch {
//	    fmt.Println(val)
//	}
func (p *Pipeline[T]) Chan() <-chan T {
	return p.ch
}

// ============================================================================
// Utility Methods
// ============================================================================

// WithBuffer sets the buffer size for subsequent operations.
// This is useful for performance tuning.
//
// Example:
//
//	pipeline.WithBuffer(100).Map(expensiveFunc)
func (p *Pipeline[T]) WithBuffer(size int) *Pipeline[T] {
	// Note: This would require refactoring to pass buffer options through
	// For now, users can use the Chan() method and create a buffered wrapper
	return p
}

// ============================================================================
// Convenience Aliases (LINQ-style)
// ============================================================================

// Where is an alias for Filter (LINQ-style naming).
//
// Example:
//
//	pipeline.Where(func(x int) bool { return x > 10 })
func (p *Pipeline[T]) Where(fn func(T) bool) *Pipeline[T] {
	return p.Filter(fn)
}

// Select is an alias for Map (LINQ-style naming).
//
// Example:
//
//	pipeline.Select(func(x int) int { return x * 2 })
func (p *Pipeline[T]) Select(fn func(T) any) *Pipeline[any] {
	return p.Map(fn)
}

// First returns the first value in the pipeline.
// This is a blocking operation.
//
// Example:
//
//	first := pipeline.First()
func (p *Pipeline[T]) First() (T, bool) {
	return recieve(p.ctx, p.ch)
}

// Last returns the last value in the pipeline.
// This is a blocking operation that consumes the entire pipeline.
//
// Example:
//
//	last := pipeline.Last()
func (p *Pipeline[T]) Last() (T, bool) {
	var last T
	found := false
	for {
		val, ok := recieve(p.ctx, p.ch)
		if !ok {
			return last, found
		}
		last = val
		found = true
	}
}

// Any returns true if any value satisfies the predicate.
// This is a blocking operation that short-circuits on first match.
//
// Example:
//
//	hasEven := pipeline.Any(func(x int) bool { return x%2 == 0 })
func (p *Pipeline[T]) Any(fn func(T) bool) bool {
	for {
		val, ok := recieve(p.ctx, p.ch)
		if !ok {
			return false
		}
		if fn(val) {
			return true
		}
	}
}

// All returns true if all values satisfy the predicate.
// This is a blocking operation that short-circuits on first non-match.
//
// Example:
//
//	allPositive := pipeline.All(func(x int) bool { return x > 0 })
func (p *Pipeline[T]) All(fn func(T) bool) bool {
	for {
		val, ok := recieve(p.ctx, p.ch)
		if !ok {
			return true
		}
		if !fn(val) {
			return false
		}
	}
}
