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

// TestBatch tests the Batch function
func TestBatch(t *testing.T) {
	t.Run("batches by size", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)
		batchSize := 3
		timeout := 1 * time.Second

		// Send 9 values (should create 3 batches)
		for i := 1; i <= 9; i++ {
			in <- i
		}
		close(in)

		out := Batch(ctx, in, batchSize, timeout)

		var batches [][]int
		for batch := range out {
			batches = append(batches, batch)
		}

		// Should have 3 batches of size 3
		if len(batches) != 3 {
			t.Fatalf("expected 3 batches, got %d", len(batches))
		}

		expected := [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
		for i, batch := range batches {
			if len(batch) != len(expected[i]) {
				t.Errorf("batch %d: expected length %d, got %d", i, len(expected[i]), len(batch))
			}
			for j, v := range batch {
				if v != expected[i][j] {
					t.Errorf("batch %d, index %d: expected %d, got %d", i, j, expected[i][j], v)
				}
			}
		}
	})

	t.Run("batches by timeout", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		batchSize := 10
		timeout := 100 * time.Millisecond

		out := Batch(ctx, in, batchSize, timeout)

		// Send a few values slowly
		go func() {
			in <- 1
			in <- 2
			time.Sleep(timeout + 50*time.Millisecond) // Wait for timeout
			in <- 3
			in <- 4
			time.Sleep(timeout + 50*time.Millisecond) // Wait for timeout
			close(in)
		}()

		var batches [][]int
		for batch := range out {
			batches = append(batches, batch)
		}

		// Should have 2 batches (triggered by timeout)
		if len(batches) != 2 {
			t.Fatalf("expected 2 batches, got %d", len(batches))
		}

		expected := [][]int{{1, 2}, {3, 4}}
		for i, batch := range batches {
			if len(batch) != len(expected[i]) {
				t.Errorf("batch %d: expected length %d, got %d", i, len(expected[i]), len(batch))
			}
			for j, v := range batch {
				if v != expected[i][j] {
					t.Errorf("batch %d, index %d: expected %d, got %d", i, j, expected[i][j], v)
				}
			}
		}
	})

	t.Run("partial batch on channel close", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)
		batchSize := 5
		timeout := 1 * time.Second

		// Send 7 values (1 full batch + 2 remaining)
		for i := 1; i <= 7; i++ {
			in <- i
		}
		close(in)

		out := Batch(ctx, in, batchSize, timeout)

		var batches [][]int
		for batch := range out {
			batches = append(batches, batch)
		}

		// Should have 2 batches
		if len(batches) != 2 {
			t.Fatalf("expected 2 batches, got %d", len(batches))
		}

		// First batch should be full
		if len(batches[0]) != 5 {
			t.Errorf("first batch: expected length 5, got %d", len(batches[0]))
		}

		// Second batch should be partial
		if len(batches[1]) != 2 {
			t.Errorf("second batch: expected length 2, got %d", len(batches[1]))
		}

		expected := [][]int{{1, 2, 3, 4, 5}, {6, 7}}
		for i, batch := range batches {
			for j, v := range batch {
				if v != expected[i][j] {
					t.Errorf("batch %d, index %d: expected %d, got %d", i, j, expected[i][j], v)
				}
			}
		}
	})

	t.Run("context cancellation with partial batch", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in := make(chan int)
		batchSize := 10
		timeout := 1 * time.Second

		out := Batch(ctx, in, batchSize, timeout)

		// Send a few values
		go func() {
			in <- 1
			in <- 2
			in <- 3
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		var batches [][]int
		for batch := range out {
			batches = append(batches, batch)
		}

		// Should have received the partial batch on cancellation
		if len(batches) != 1 {
			t.Fatalf("expected 1 batch, got %d", len(batches))
		}

		if len(batches[0]) != 3 {
			t.Errorf("expected batch length 3, got %d", len(batches[0]))
		}

		expected := []int{1, 2, 3}
		for i, v := range batches[0] {
			if v != expected[i] {
				t.Errorf("index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("timer starts on first value", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		batchSize := 10
		timeout := 100 * time.Millisecond

		out := Batch(ctx, in, batchSize, timeout)

		start := time.Now()

		// Wait before sending first value
		time.Sleep(50 * time.Millisecond)

		go func() {
			in <- 1
			in <- 2
			// Don't send more, let timeout trigger
		}()

		// Should receive batch after timeout from first value
		batch := <-out

		elapsed := time.Since(start)

		// Should be approximately 50ms (wait) + 100ms (timeout) = 150ms
		expectedMin := 140 * time.Millisecond
		expectedMax := 180 * time.Millisecond

		if elapsed < expectedMin || elapsed > expectedMax {
			t.Errorf("expected elapsed time ~150ms, got %v", elapsed)
		}

		if len(batch) != 2 {
			t.Errorf("expected batch length 2, got %d", len(batch))
		}

		close(in)
	})

	t.Run("multiple batches mixed size and timeout", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		batchSize := 3
		timeout := 80 * time.Millisecond

		out := Batch(ctx, in, batchSize, timeout)

		go func() {
			// Batch 1: size-triggered
			in <- 1
			in <- 2
			in <- 3

			time.Sleep(20 * time.Millisecond)

			// Batch 2: timeout-triggered
			in <- 4
			in <- 5
			time.Sleep(timeout + 20*time.Millisecond)

			// Batch 3: size-triggered
			in <- 6
			in <- 7
			in <- 8

			close(in)
		}()

		var batches [][]int
		for batch := range out {
			batches = append(batches, batch)
		}

		if len(batches) != 3 {
			t.Fatalf("expected 3 batches, got %d", len(batches))
		}

		expected := [][]int{{1, 2, 3}, {4, 5}, {6, 7, 8}}
		for i, batch := range batches {
			if len(batch) != len(expected[i]) {
				t.Errorf("batch %d: expected length %d, got %d", i, len(expected[i]), len(batch))
			}
			for j, v := range batch {
				if v != expected[i][j] {
					t.Errorf("batch %d, index %d: expected %d, got %d", i, j, expected[i][j], v)
				}
			}
		}
	})

	t.Run("empty channel", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		batchSize := 5
		timeout := 50 * time.Millisecond

		close(in)

		out := Batch(ctx, in, batchSize, timeout)

		var batches [][]int
		for batch := range out {
			batches = append(batches, batch)
		}

		// Should receive no batches
		if len(batches) != 0 {
			t.Errorf("expected 0 batches, got %d", len(batches))
		}
	})

	t.Run("with buffered output channel", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 20)
		batchSize := 3
		timeout := 1 * time.Second

		// Send 9 values
		for i := 1; i <= 9; i++ {
			in <- i
		}
		close(in)

		out := Batch(ctx, in, batchSize, timeout, WithBuffer[[]int](5))

		var batches [][]int
		for batch := range out {
			batches = append(batches, batch)
		}

		if len(batches) != 3 {
			t.Fatalf("expected 3 batches, got %d", len(batches))
		}
	})

	t.Run("single value with timeout", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		batchSize := 5
		timeout := 80 * time.Millisecond

		out := Batch(ctx, in, batchSize, timeout)

		go func() {
			in <- 42
			// Wait for timeout to trigger
			time.Sleep(timeout + 50*time.Millisecond)
			close(in)
		}()

		var batches [][]int
		for batch := range out {
			batches = append(batches, batch)
		}

		if len(batches) != 1 {
			t.Fatalf("expected 1 batch, got %d", len(batches))
		}

		if len(batches[0]) != 1 {
			t.Errorf("expected batch length 1, got %d", len(batches[0]))
		}

		if batches[0][0] != 42 {
			t.Errorf("expected value 42, got %d", batches[0][0])
		}
	})

	t.Run("rapid consecutive batches", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 50)
		batchSize := 5
		timeout := 1 * time.Second

		// Send 25 values rapidly
		for i := 1; i <= 25; i++ {
			in <- i
		}
		close(in)

		out := Batch(ctx, in, batchSize, timeout)

		var batches [][]int
		for batch := range out {
			batches = append(batches, batch)
		}

		// Should have 5 batches
		if len(batches) != 5 {
			t.Fatalf("expected 5 batches, got %d", len(batches))
		}

		// All batches should be full
		for i, batch := range batches {
			if len(batch) != batchSize {
				t.Errorf("batch %d: expected length %d, got %d", i, batchSize, len(batch))
			}
		}
	})

	t.Run("context cancellation before first value", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in := make(chan int)
		batchSize := 5
		timeout := 100 * time.Millisecond

		out := Batch(ctx, in, batchSize, timeout)

		// Cancel immediately
		cancel()

		// Channel should close without any batches
		var batches [][]int
		for batch := range out {
			batches = append(batches, batch)
		}

		if len(batches) != 0 {
			t.Errorf("expected 0 batches, got %d", len(batches))
		}
	})

	t.Run("batch size of 1", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)
		batchSize := 1
		timeout := 1 * time.Second

		// Send 5 values
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := Batch(ctx, in, batchSize, timeout)

		var batches [][]int
		for batch := range out {
			batches = append(batches, batch)
		}

		// Should have 5 batches of size 1
		if len(batches) != 5 {
			t.Fatalf("expected 5 batches, got %d", len(batches))
		}

		for i, batch := range batches {
			if len(batch) != 1 {
				t.Errorf("batch %d: expected length 1, got %d", i, len(batch))
			}
			if batch[0] != i+1 {
				t.Errorf("batch %d: expected value %d, got %d", i, i+1, batch[0])
			}
		}
	})

	t.Run("timer resets after size-triggered batch", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		batchSize := 3
		timeout := 100 * time.Millisecond

		out := Batch(ctx, in, batchSize, timeout)

		go func() {
			// First batch: size-triggered
			in <- 1
			in <- 2
			in <- 3

			// Wait a bit, then start second batch
			time.Sleep(50 * time.Millisecond)

			// Second batch: timeout-triggered
			in <- 4
			in <- 5

			// Should timeout after 100ms from first value (4), not from previous batch
			time.Sleep(timeout + 50*time.Millisecond)
			close(in)
		}()

		var batches [][]int
		start := time.Now()
		for batch := range out {
			batches = append(batches, batch)
		}
		elapsed := time.Since(start)

		if len(batches) != 2 {
			t.Fatalf("expected 2 batches, got %d", len(batches))
		}

		// Total time should be less than 250ms (not 200ms from continuous timing)
		// First batch immediate, wait 50ms, second batch after 100ms timeout
		if elapsed > 300*time.Millisecond {
			t.Errorf("timer may not have reset properly, elapsed: %v", elapsed)
		}
	})
}

// TestDebounce tests the Debounce function
func TestDebounce(t *testing.T) {
	t.Run("basic debounce behavior", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		debounceDuration := 100 * time.Millisecond

		out := Debounce(ctx, in, debounceDuration)

		// Send values rapidly
		go func() {
			for i := 1; i <= 5; i++ {
				in <- i
				time.Sleep(20 * time.Millisecond)
			}
			// Wait for debounce to settle
			time.Sleep(debounceDuration + 50*time.Millisecond)
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive only the last value (5) after silence
		if len(results) != 1 {
			t.Fatalf("expected 1 value, got %d: %v", len(results), results)
		}
		if results[0] != 5 {
			t.Errorf("expected value 5, got %d", results[0])
		}
	})

	t.Run("emits each value after silence period", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		debounceDuration := 80 * time.Millisecond

		out := Debounce(ctx, in, debounceDuration)

		go func() {
			// First burst
			in <- 1
			in <- 2
			in <- 3
			// Wait for debounce
			time.Sleep(debounceDuration + 50*time.Millisecond)

			// Second burst
			in <- 4
			in <- 5
			// Wait for debounce
			time.Sleep(debounceDuration + 50*time.Millisecond)

			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive 2 values (3 and 5)
		expected := []int{3, 5}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d: %v", len(expected), len(results), results)
		}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("single value is emitted after duration", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		debounceDuration := 50 * time.Millisecond

		out := Debounce(ctx, in, debounceDuration)

		start := time.Now()

		go func() {
			in <- 42
			time.Sleep(debounceDuration + 50*time.Millisecond)
			close(in)
		}()

		val := <-out
		elapsed := time.Since(start)

		if val != 42 {
			t.Errorf("expected value 42, got %d", val)
		}

		// Should take at least debounceDuration
		if elapsed < debounceDuration {
			t.Errorf("value emitted too early: %v (expected >= %v)", elapsed, debounceDuration)
		}
	})

	t.Run("timer resets on each new value", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		debounceDuration := 100 * time.Millisecond

		out := Debounce(ctx, in, debounceDuration)

		go func() {
			// Send values every 60ms (within debounce window)
			for i := 1; i <= 5; i++ {
				in <- i
				time.Sleep(60 * time.Millisecond)
			}
			// Final wait for debounce
			time.Sleep(debounceDuration + 50*time.Millisecond)
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive only the last value (timer kept resetting)
		if len(results) != 1 {
			t.Fatalf("expected 1 value (timer should have reset), got %d: %v", len(results), results)
		}
		if results[0] != 5 {
			t.Errorf("expected value 5, got %d", results[0])
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in := make(chan int)
		debounceDuration := 100 * time.Millisecond

		out := Debounce(ctx, in, debounceDuration)

		go func() {
			in <- 1
			in <- 2
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		// Channel should close without emitting (cancelled before debounce)
		timeout := time.After(200 * time.Millisecond)
		count := 0
		for {
			select {
			case _, ok := <-out:
				if !ok {
					// Channel closed
					if count > 0 {
						t.Errorf("expected no values before cancellation, got %d", count)
					}
					return
				}
				count++
			case <-timeout:
				t.Fatal("channel did not close after context cancellation")
			}
		}
	})

	t.Run("emits final value on input channel close", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		debounceDuration := 100 * time.Millisecond

		out := Debounce(ctx, in, debounceDuration)

		go func() {
			in <- 1
			in <- 2
			in <- 3
			// Close immediately without waiting for debounce
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should emit the final pending value (3) immediately on close
		if len(results) != 1 {
			t.Fatalf("expected 1 value, got %d: %v", len(results), results)
		}
		if results[0] != 3 {
			t.Errorf("expected value 3, got %d", results[0])
		}
	})

	t.Run("no values if input closes immediately", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		debounceDuration := 50 * time.Millisecond

		close(in)

		out := Debounce(ctx, in, debounceDuration)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 values, got %d: %v", len(results), results)
		}
	})

	t.Run("with buffered output channel", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		debounceDuration := 50 * time.Millisecond

		out := Debounce(ctx, in, debounceDuration, WithBuffer[int](3))

		go func() {
			// Burst 1
			in <- 1
			in <- 2
			time.Sleep(debounceDuration + 20*time.Millisecond)

			// Burst 2
			in <- 3
			in <- 4
			time.Sleep(debounceDuration + 20*time.Millisecond)

			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		expected := []int{2, 4}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(results))
		}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("slow input values pass through", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		debounceDuration := 50 * time.Millisecond

		out := Debounce(ctx, in, debounceDuration)

		go func() {
			in <- 1
			time.Sleep(debounceDuration + 50*time.Millisecond)
			in <- 2
			time.Sleep(debounceDuration + 50*time.Millisecond)
			in <- 3
			time.Sleep(debounceDuration + 50*time.Millisecond)
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Each value has enough silence after it
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

	t.Run("multiple rapid bursts", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		debounceDuration := 80 * time.Millisecond

		out := Debounce(ctx, in, debounceDuration)

		go func() {
			// Burst 1: values 1-5
			for i := 1; i <= 5; i++ {
				in <- i
				time.Sleep(10 * time.Millisecond)
			}
			time.Sleep(debounceDuration + 30*time.Millisecond)

			// Burst 2: values 6-10
			for i := 6; i <= 10; i++ {
				in <- i
				time.Sleep(10 * time.Millisecond)
			}
			time.Sleep(debounceDuration + 30*time.Millisecond)

			// Burst 3: values 11-15
			for i := 11; i <= 15; i++ {
				in <- i
				time.Sleep(10 * time.Millisecond)
			}
			time.Sleep(debounceDuration + 30*time.Millisecond)

			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive last value from each burst: 5, 10, 15
		expected := []int{5, 10, 15}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d: %v", len(expected), len(results), results)
		}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("very short debounce duration", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		debounceDuration := 10 * time.Millisecond

		out := Debounce(ctx, in, debounceDuration)

		go func() {
			for i := 1; i <= 5; i++ {
				in <- i
				time.Sleep(5 * time.Millisecond)
			}
			time.Sleep(debounceDuration + 20*time.Millisecond)
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// With very short duration, should still debounce
		if len(results) == 0 {
			t.Fatal("expected at least one value")
		}
		// Should receive the last value
		if results[len(results)-1] != 5 {
			t.Errorf("expected last value 5, got %d", results[len(results)-1])
		}
	})

	t.Run("context cancellation with pending value", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in := make(chan int)
		debounceDuration := 200 * time.Millisecond

		out := Debounce(ctx, in, debounceDuration)

		go func() {
			in <- 1
			in <- 2
			in <- 3
			time.Sleep(50 * time.Millisecond)
			cancel() // Cancel before debounce completes
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Pending value should be lost on cancellation
		if len(results) != 0 {
			t.Errorf("expected 0 values after cancellation, got %d: %v", len(results), results)
		}
	})

	t.Run("exact timing validation", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		debounceDuration := 100 * time.Millisecond

		out := Debounce(ctx, in, debounceDuration)

		start := time.Now()

		go func() {
			in <- 1
			time.Sleep(debounceDuration + 50*time.Millisecond)
			close(in)
		}()

		<-out // Receive the value
		elapsed := time.Since(start)

		// Should take approximately debounceDuration
		expectedMin := debounceDuration
		expectedMax := debounceDuration + 50*time.Millisecond

		if elapsed < expectedMin || elapsed > expectedMax {
			t.Errorf("expected elapsed time between %v and %v, got %v", expectedMin, expectedMax, elapsed)
		}
	})

	t.Run("burst then immediate close", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)
		debounceDuration := 100 * time.Millisecond

		// Pre-buffer values
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := Debounce(ctx, in, debounceDuration)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should emit the last buffered value immediately
		if len(results) != 1 {
			t.Fatalf("expected 1 value, got %d: %v", len(results), results)
		}
		if results[0] != 5 {
			t.Errorf("expected value 5, got %d", results[0])
		}
	})
}
