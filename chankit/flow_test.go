package chankit

import (
	"context"
	"testing"
	"time"
)

// TestThrottle tests the Throttle function
func TestThrottle(t *testing.T) {
	t.Run("basic throttle behavior", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		throttleDuration := 100 * time.Millisecond

		out := Throttle(ctx, in, throttleDuration)

		// Send values rapidly
		go func() {
			for i := 1; i <= 10; i++ {
				in <- i
				time.Sleep(10 * time.Millisecond)
			}
			time.Sleep(150 * time.Millisecond) // Wait for at least one more tick
			close(in)
		}()

		// Collect output values with timestamps
		var results []int
		start := time.Now()
		var timestamps []time.Duration

		for val := range out {
			results = append(results, val)
			timestamps = append(timestamps, time.Since(start))
		}

		// Should receive fewer values than sent due to throttling
		if len(results) == 0 {
			t.Fatal("expected at least one value, got none")
		}

		// Check that values are spaced by approximately throttleDuration
		for i := 1; i < len(timestamps); i++ {
			interval := timestamps[i] - timestamps[i-1]
			if interval < throttleDuration-20*time.Millisecond {
				t.Errorf("interval %d: expected ~%v, got %v", i, throttleDuration, interval)
			}
		}
	})

	t.Run("drops intermediate values", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)
		throttleDuration := 100 * time.Millisecond

		// Send multiple values before first tick
		for i := 1; i <= 5; i++ {
			in <- i
		}

		out := Throttle(ctx, in, throttleDuration)

		// Wait for first throttled value
		time.Sleep(throttleDuration + 20*time.Millisecond)

		select {
		case val := <-out:
			// Should receive the last value (5)
			if val != 5 {
				t.Errorf("expected last value 5, got %d", val)
			}
		case <-time.After(50 * time.Millisecond):
			t.Fatal("timeout waiting for throttled value")
		}

		close(in)
	})

	t.Run("handles slow input", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		throttleDuration := 50 * time.Millisecond

		out := Throttle(ctx, in, throttleDuration)

		go func() {
			in <- 1
			time.Sleep(150 * time.Millisecond) // Wait longer than throttle duration
			in <- 2
			time.Sleep(150 * time.Millisecond)
			in <- 3
			time.Sleep(100 * time.Millisecond) // Ensure last value is emitted
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive all values since they arrive slower than throttle rate
		expected := []int{1, 2, 3}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d: %v", len(expected), len(results), results)
		}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("context cancellation stops throttle", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in := make(chan int)
		throttleDuration := 100 * time.Millisecond

		out := Throttle(ctx, in, throttleDuration)

		// Send some values
		go func() {
			for i := 1; i <= 10; i++ {
				select {
				case in <- i:
					time.Sleep(10 * time.Millisecond)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()

		// Cancel context after short delay
		time.Sleep(150 * time.Millisecond)
		cancel()

		// Channel should close
		timeout := time.After(200 * time.Millisecond)
		for {
			select {
			case _, ok := <-out:
				if !ok {
					return // Channel closed as expected
				}
			case <-timeout:
				t.Fatal("channel did not close after context cancellation")
			}
		}
	})

	t.Run("input channel closure", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		throttleDuration := 50 * time.Millisecond

		out := Throttle(ctx, in, throttleDuration)

		// Send values and close input
		go func() {
			in <- 1
			in <- 2
			in <- 3
			time.Sleep(100 * time.Millisecond) // Wait for throttle to emit
			close(in)
		}()

		// Collect all values
		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive at least one value and channel should close
		if len(results) == 0 {
			t.Fatal("expected at least one value before closure")
		}
	})

	t.Run("with buffered output channel", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		throttleDuration := 50 * time.Millisecond

		out := Throttle(ctx, in, throttleDuration, WithBuffer[int](3))

		// Send values rapidly
		go func() {
			for i := 1; i <= 10; i++ {
				in <- i
				time.Sleep(10 * time.Millisecond)
			}
			close(in)
		}()

		// Collect all output
		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should have received some throttled values
		if len(results) == 0 {
			t.Fatal("expected at least one value")
		}
	})

	t.Run("no pending value on ticker", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		throttleDuration := 50 * time.Millisecond

		out := Throttle(ctx, in, throttleDuration)

		// Don't send any values, just wait for ticker to fire
		time.Sleep(throttleDuration + 20*time.Millisecond)

		// Should not receive anything
		select {
		case val := <-out:
			t.Fatalf("unexpected value received: %d", val)
		case <-time.After(20 * time.Millisecond):
			// Expected: no value
		}

		close(in)

		// Channel should close
		_, ok := <-out
		if ok {
			t.Fatal("channel should be closed")
		}
	})

	t.Run("rapid fire then wait", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 100)
		throttleDuration := 100 * time.Millisecond

		// Send 10 values rapidly
		for i := 1; i <= 10; i++ {
			in <- i
		}

		out := Throttle(ctx, in, throttleDuration)

		// Wait for first throttled output (should be value 10, the last one)
		time.Sleep(throttleDuration + 20*time.Millisecond)

		select {
		case val := <-out:
			if val != 10 {
				t.Errorf("expected last value 10, got %d", val)
			}
		case <-time.After(50 * time.Millisecond):
			t.Fatal("timeout waiting for first throttled value")
		}

		// Send more values
		for i := 11; i <= 20; i++ {
			in <- i
		}

		// Wait for next throttled output
		time.Sleep(throttleDuration + 20*time.Millisecond)

		select {
		case val := <-out:
			if val != 20 {
				t.Errorf("expected last value 20, got %d", val)
			}
		case <-time.After(50 * time.Millisecond):
			t.Fatal("timeout waiting for second throttled value")
		}

		close(in)
	})

	t.Run("exact timing validation", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		throttleDuration := 100 * time.Millisecond

		out := Throttle(ctx, in, throttleDuration)

		start := time.Now()

		// Send values continuously
		go func() {
			for i := 1; i <= 100; i++ {
				in <- i
				time.Sleep(5 * time.Millisecond)
			}
			close(in)
		}()

		count := 0
		for range out {
			count++
		}

		elapsed := time.Since(start)

		// Calculate expected count (approximately elapsed / throttleDuration)
		expectedCount := int(elapsed / throttleDuration)

		// Allow some tolerance
		if count < expectedCount-1 || count > expectedCount+1 {
			t.Errorf("expected approximately %d values, got %d (elapsed: %v)", expectedCount, count, elapsed)
		}
	})
}

// TestFixedInterval tests the FixedInterval function
func TestFixedInterval(t *testing.T) {
	t.Run("processes all values at fixed rate", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)
		intervalDuration := 50 * time.Millisecond

		// Send 5 values rapidly
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := FixedInterval(ctx, in, intervalDuration)

		var results []int
		start := time.Now()
		var timestamps []time.Duration

		for val := range out {
			results = append(results, val)
			timestamps = append(timestamps, time.Since(start))
		}

		// Should receive ALL 5 values
		expected := []int{1, 2, 3, 4, 5}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(results))
		}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}

		// Check timing - each value should be spaced by intervalDuration
		for i := 1; i < len(timestamps); i++ {
			interval := timestamps[i] - timestamps[i-1]
			if interval < intervalDuration-10*time.Millisecond {
				t.Errorf("interval %d: expected ~%v, got %v", i, intervalDuration, interval)
			}
		}
	})

	t.Run("queues values and drains on close", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 20)
		intervalDuration := 30 * time.Millisecond

		// Send 10 values
		for i := 1; i <= 10; i++ {
			in <- i
		}
		close(in)

		out := FixedInterval(ctx, in, intervalDuration)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive all 10 values
		if len(results) != 10 {
			t.Fatalf("expected 10 values, got %d", len(results))
		}

		// Verify values are in order
		for i, v := range results {
			if v != i+1 {
				t.Errorf("at index %d: expected %d, got %d", i, i+1, v)
			}
		}
	})

	t.Run("handles slow input", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		intervalDuration := 50 * time.Millisecond

		out := FixedInterval(ctx, in, intervalDuration)

		go func() {
			in <- 1
			time.Sleep(200 * time.Millisecond)
			in <- 2
			time.Sleep(200 * time.Millisecond)
			in <- 3
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		expected := []int{1, 2, 3}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(results))
		}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in := make(chan int, 100)
		intervalDuration := 50 * time.Millisecond

		// Queue many values
		for i := 1; i <= 20; i++ {
			in <- i
		}
		close(in)

		out := FixedInterval(ctx, in, intervalDuration)

		// Receive a few values then cancel
		count := 0
		for range out {
			count++
			if count == 3 {
				cancel()
			}
		}

		// Should have stopped after cancellation
		if count > 5 {
			t.Errorf("expected ~3-4 values after cancellation, got %d", count)
		}
	})

	t.Run("with buffered output channel", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)
		intervalDuration := 30 * time.Millisecond

		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := FixedInterval(ctx, in, intervalDuration, WithBuffer[int](3))

		var results []int
		for val := range out {
			results = append(results, val)
		}

		if len(results) != 5 {
			t.Fatalf("expected 5 values, got %d", len(results))
		}
	})

	t.Run("empty queue on ticker", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		intervalDuration := 50 * time.Millisecond

		out := FixedInterval(ctx, in, intervalDuration)

		// Wait for ticker to fire with empty queue
		time.Sleep(intervalDuration + 20*time.Millisecond)

		// Should not receive anything
		select {
		case val := <-out:
			t.Fatalf("unexpected value received: %d", val)
		case <-time.After(20 * time.Millisecond):
			// Expected: no value
		}

		close(in)

		// Channel should close
		_, ok := <-out
		if ok {
			t.Fatal("channel should be closed")
		}
	})

	t.Run("precise interval timing", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 100)
		intervalDuration := 50 * time.Millisecond

		// Send 10 values
		for i := 1; i <= 10; i++ {
			in <- i
		}
		close(in)

		out := FixedInterval(ctx, in, intervalDuration)

		start := time.Now()
		count := 0
		for range out {
			count++
		}
		elapsed := time.Since(start)

		// Should take approximately 10 * intervalDuration
		expectedDuration := 10 * intervalDuration
		tolerance := 100 * time.Millisecond

		if elapsed < expectedDuration-tolerance || elapsed > expectedDuration+tolerance {
			t.Errorf("expected duration ~%v, got %v", expectedDuration, elapsed)
		}

		if count != 10 {
			t.Errorf("expected 10 values, got %d", count)
		}
	})
}
