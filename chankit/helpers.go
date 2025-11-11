package chankit

import "context"

// ChanOption is a functional option for configuring channel behavior
type ChanOption[T any] func(*chanConfig[T])

// chanConfig holds configuration for channel creation
type chanConfig[T any] struct {
	bufferSize int
}

// applyChanOptions creates a configured channel based on provided options
func applyChanOptions[T any](opts ...ChanOption[T]) chan T {
	cfg := &chanConfig[T]{bufferSize: 0}
	for _, opt := range opts {
		opt(cfg)
	}
	return make(chan T, cfg.bufferSize)
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

func drain[T any](in <-chan T) {
	for range in {
		// just drain
	}
}

func sendTo[T any](ctx context.Context, out chan<- T, in <-chan T) {
	for {
		select {
		case <-ctx.Done():
			go drain(in)
			return
		case val, ok := <-in:
			if !ok {
				return
			}
			select {
			case <-ctx.Done():
				go drain(in)
				return
			case out <- val:
			}
		}
	}
}
