package chankit

import "context"

// ChanOption is a functional option for configuring channel behavior
type ChanOption[T any] func(*chanConfig[T])

// chanConfig holds configuration for channel creation
type chanConfig[T any] struct {
	bufferSize int
}

// WithBuffer sets a custom buffer size for the channel
func WithBuffer[T any](size int) ChanOption[T] {
	return func(cfg *chanConfig[T]) {
		cfg.bufferSize = size
	}
}

// WithBufferAuto sets the buffer size to match the input slice length
// This allows the producer goroutine to finish immediately without blocking
func WithBufferAuto[T any]() ChanOption[T] {
	return func(cfg *chanConfig[T]) {
		cfg.bufferSize = -1 // sentinel value for auto-sizing
	}
}

// SliceToChan converts a slice to a channel, sending each element sequentially.
// By default, uses an unbuffered channel. Use WithBuffer() or WithBufferAuto() to change behavior.
//
// Examples:
//   SliceToChan(ctx, slice)                           // unbuffered
//   SliceToChan(ctx, slice, WithBuffer[int](10))      // buffered with size 10
//   SliceToChan(ctx, slice, WithBufferAuto[int]())    // buffered to slice length
func SliceToChan[T any](ctx context.Context, slice []T, opts ...ChanOption[T]) <-chan T {
	cfg := &chanConfig[T]{bufferSize: 0} // unbuffered by default

	for _, opt := range opts {
		opt(cfg)
	}

	bufferSize := cfg.bufferSize
	if bufferSize == -1 {
		bufferSize = len(slice)
	}

	ch := make(chan T, bufferSize)
	go func() {
		defer close(ch)

		for _, item := range slice {
			select {
			case <-ctx.Done():
				return
			case ch <- item:
			}
		}
	}()
	return ch
}

// SliceOption is a functional option for configuring slice behavior
type SliceOption[T any] func(*sliceConfig[T])

// sliceConfig holds configuration for slice creation
type sliceConfig[T any] struct {
	initialCapacity int
}

// WithCapacity sets the initial capacity for the slice
// Use this when you know approximately how many items to expect
func WithCapacity[T any](capacity int) SliceOption[T] {
	return func(cfg *sliceConfig[T]) {
		cfg.initialCapacity = capacity
	}
}

// ChanToSlice converts a channel to a slice, collecting all elements until the channel closes.
// By default, creates a slice with zero initial capacity. Use WithCapacity() for better performance.
//
// Examples:
//   ChanToSlice(ctx, ch)                          // default capacity
//   ChanToSlice(ctx, ch, WithCapacity[int](100))  // pre-allocated capacity
func ChanToSlice[T any](ctx context.Context, ch <-chan T, opts ...SliceOption[T]) []T {
	cfg := &sliceConfig[T]{initialCapacity: 0}

	for _, opt := range opts {
		opt(cfg)
	}

	slice := make([]T, 0, cfg.initialCapacity)

	for {
		select {
		case <-ctx.Done():
			return slice
		case item, ok := <-ch:
			if !ok {
				return slice
			}
			slice = append(slice, item)
		}
	}
}
