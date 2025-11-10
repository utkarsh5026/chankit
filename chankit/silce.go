package chankit

import "context"

// SliceToChan converts a slice to a channel, sending each element sequentially.
// By default, uses an unbuffered channel. Use WithBuffer() or WithBufferAuto() to change behavior.
//
// Examples:
//   SliceToChan(ctx, slice)                           // unbuffered
//   SliceToChan(ctx, slice, WithBuffer[int](10))      // buffered with size 10
//   SliceToChan(ctx, slice, WithBufferAuto[int]())    // buffered to slice length
func SliceToChan[T any](ctx context.Context, slice []T, opts ...ChanOption[T]) <-chan T {
	cfg := &chanConfig[T]{bufferSize: 0}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.bufferSize == -1 {
		opts = []ChanOption[T]{WithBuffer[T](len(slice))}
	}

	ch := applyChanOptions(opts...)
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
