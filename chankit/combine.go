// Package chankit provides utilities for working with Go channels in a functional style.
package chankit

import (
	"context"
	"reflect"
	"sync"
)

// Merge combines multiple input channels into a single output channel.
// Values from all input channels are forwarded to the output channel.
// The output channel closes when all input channels have closed.
// It respects context cancellation and stops immediately when context is canceled.
func Merge[T any](ctx context.Context, chans ...<-chan T) <-chan T {
	outChan := make(chan T)

	var wg sync.WaitGroup
	for _, ch := range chans {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for val := range ch {
				select {
				case <-ctx.Done():
					return
				case outChan <- val:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(outChan)
	}()

	return outChan
}

// Zip combines two channels into a single channel of paired values.
// It stops when either channel closes or context is canceled.
func Zip[T, R any](ctx context.Context, ch1 <-chan T, ch2 <-chan R) <-chan struct {
	First  T
	Second R
} {
	outChan := make(chan struct {
		First  T
		Second R
	})

	go func() {
		defer close(outChan)
		for {
			var val1 T
			var val2 R
			var ok1, ok2 bool

			select {
			case <-ctx.Done():
				return
			case val1, ok1 = <-ch1:
				if !ok1 {
					return
				}
			}

			select {
			case <-ctx.Done():
				return
			case val2, ok2 = <-ch2:
				if !ok2 {
					return // Second channel closed
				}
			}

			select {
			case <-ctx.Done():
				return
			case outChan <- struct {
				First  T
				Second R
			}{First: val1, Second: val2}:
			}
		}
	}()

	return outChan
}

// ZipN combines multiple channels into a single channel of slices.
// It reads one value from each channel and emits them as a slice.
// It stops when any channel closes or context is canceled.
//
// Example:
//
//	ch1 := chankit.SliceToChan(ctx, []int{1, 2, 3})
//	ch2 := chankit.SliceToChan(ctx, []string{"a", "b", "c"})
//	ch3 := chankit.SliceToChan(ctx, []bool{true, false, true})
//	zipped := chankit.ZipN(ctx, ch1, ch2, ch3)
//	// Output: [][]any{{1, "a", true}, {2, "b", false}, {3, "c", true}}
func ZipN(ctx context.Context, channels ...any) <-chan []any {
	outChan := make(chan []any)

	if len(channels) == 0 {
		close(outChan)
		return outChan
	}

	go func() {
		defer close(outChan)

		for {
			result := make([]any, len(channels))

			for i, ch := range channels {
				select {
				case <-ctx.Done():
					return
				default:
					val, ok := receiveFromChannel(ch)
					if !ok {
						return
					}
					result[i] = val
				}
			}

			select {
			case <-ctx.Done():
				return
			case outChan <- result:
			}
		}
	}()

	return outChan
}

// receiveFromChannel is a helper that uses reflection to receive from any channel type
func receiveFromChannel(ch any) (any, bool) {
	val := reflect.ValueOf(ch)

	if val.Kind() != reflect.Chan {
		return nil, false
	}

	received, ok := val.Recv()
	if !ok {
		return nil, false
	}

	return received.Interface(), true
}
