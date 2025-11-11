package chankit

import (
	"context"
	"sync"
)

// Tap creates a channel that passes through all values from the input channel,
// calling tapFunc on each value without modifying it. This is useful for side effects
// like logging or metrics collection.
//
// Example:
//
//	input := make(chan int)
//	go func() {
//		for i := 1; i <= 3; i++ {
//			input <- i
//		}
//		close(input)
//	}()
//
//	output := Tap(ctx, input, func(v int) {
//		fmt.Printf("Logging: %d\n", v)
//	})
//
//	for val := range output {
//		fmt.Println(val) // Prints: 1, 2, 3
//	}
//	// Also logs: "Logging: 1", "Logging: 2", "Logging: 3"
func Tap[T any](ctx context.Context, in <-chan T, tapFunc func(T), opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		for {
			val, ok := recieve(ctx, in)
			if !ok {
				return
			}

			tapFunc(val)

			if !send(ctx, outChan, val) {
				return
			}
		}
	}()

	return outChan
}

// FlatMap transforms each value from the input channel into a channel of values using flatMapFunc,
// then flattens all resulting channels into a single output channel. This is useful for operations
// where each input value needs to be expanded into multiple output values concurrently.
// Each inner channel is processed in its own goroutine, allowing parallel processing of multiple streams.
//
// Example:
//
//	input := make(chan int)
//	go func() {
//		for i := 1; i <= 3; i++ {
//			input <- i
//		}
//		close(input)
//	}()
//
//	// Expand each number into a channel of its multiples
//	output := FlatMap(ctx, input, func(n int) <-chan int {
//		ch := make(chan int)
//		go func() {
//			defer close(ch)
//			for j := 1; j <= 2; j++ {
//				ch <- n * j
//			}
//		}()
//		return ch
//	})
//
//	for val := range output {
//		fmt.Println(val) // Prints: 1, 2, 2, 4, 3, 6 (order may vary due to concurrency)
//	}
func FlatMap[T, R any](ctx context.Context, in <-chan T, flatMapFunc func(T) <-chan R, opts ...ChanOption[R]) <-chan R {
	outChan := applyChanOptions(opts...)

	go func() {
		var wg sync.WaitGroup

		defer func() {
			wg.Wait()
			close(outChan)
		}()

		for {
			val, ok := recieve(ctx, in)
			if !ok {
				return
			}

			innerChan := flatMapFunc(val)
			wg.Add(1)
			go func() {
				defer wg.Done()
				forwardSimple(ctx, outChan, innerChan)
			}()
		}
	}()

	return outChan
}
