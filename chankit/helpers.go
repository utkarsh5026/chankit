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

// drain consumes all remaining values from a channel without processing them.
// This is used to prevent goroutine leaks when context is cancelled but the
// input channel still has pending values. By draining in a separate goroutine,
// we allow the producer to complete without blocking.
func drain[T any](in <-chan T) {
	for range in {
		// just drain
	}
}

// forwardSimple forwards values from the input channel to the output channel
// with context cancellation support. It performs a simple pass-through operation
// without any transformation or side effects.
func forwardSimple[T any](ctx context.Context, out chan<- T, in <-chan T) {
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

// forwardWithTransform forwards values from the input channel to the output channel
// after applying a transformation function to each value. This allows changing
// the type or content of values as they flow through the channel.
//
// The transform function is called synchronously for each value received.
// If the transform function panics, the panic will propagate and stop the forwarding.
//
// The function respects context cancellation at two points:
// 1. Before receiving from the input channel
// 2. Before sending to the output channel (after transformation)
func forwardWithTransform[T, R any](ctx context.Context, out chan<- R, in <-chan T, transform func(T) R) {
	for {
		select {
		case <-ctx.Done():
			go drain(in)
			return

		case val, ok := <-in:
			if !ok {
				return
			}

			transformedVal := transform(val)

			select {
			case <-ctx.Done():
				go drain(in)
				return
			case out <- transformedVal:
			}
		}
	}
}

// recieve reads a value from the input channel with context cancellation support.
// It returns the received value and a boolean indicating success.
//
// Parameters:
//   - ctx: Context for cancellation control
//   - in: Input channel to receive from
//
// Returns:
//   - T: The received value, or zero value if context is cancelled or channel is closed
//   - bool: true if value was successfully received, false if context was cancelled or channel is closed
//
// The function respects context cancellation and will immediately return if the context is done.
func recieve[T any](ctx context.Context, in <-chan T) (T, bool) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, false

	case val, ok := <-in:
		return val, ok
	}
}

// send writes a value to the output channel with context cancellation support.
// It returns a boolean indicating whether the send was successful.
//
// Parameters:
//   - ctx: Context for cancellation control
//   - out: Output channel to send to
//   - val: Value to send
//
// Returns:
//   - bool: true if value was successfully sent, false if context was cancelled
//
// The function respects context cancellation and will immediately return false if the context is done,
// preventing blocking sends when the operation should be cancelled.
func send[T any](ctx context.Context, out chan<- T, val T) bool {
	select {
	case <-ctx.Done():
		return false

	case out <- val:
		return true
	}
}
