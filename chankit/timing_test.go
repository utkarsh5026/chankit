package chankit

import (
	"context"
	"testing"
	"time"
)

// TestDelay tests the Delay function
func TestDelay(t *testing.T) {
	t.Run("basic delay behavior", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		delayDuration := 100 * time.Millisecond

		out := Delay(ctx, in, delayDuration)

		start := time.Now()

		go func() {
			in <- 1
			in <- 2
			in <- 3
			close(in)
		}()

		var results []int
		var timestamps []time.Duration

		for val := range out {
			results = append(results, val)
			timestamps = append(timestamps, time.Since(start))
		}

		// Should receive all 3 values (order may vary due to concurrent processing)
		if len(results) != 3 {
			t.Fatalf("expected 3 values, got %d", len(results))
		}

		// Verify all values are present
		found := make(map[int]bool)
		for _, v := range results {
			found[v] = true
		}
		for i := 1; i <= 3; i++ {
			if !found[i] {
				t.Errorf("missing value %d", i)
			}
		}

		// Each value should be delayed by at least delayDuration
		for i, ts := range timestamps {
			if ts < delayDuration-20*time.Millisecond {
				t.Errorf("value at position %d arrived too early: %v (expected >= %v)", i, ts, delayDuration)
			}
		}
	})

	t.Run("delays each value independently", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		delayDuration := 80 * time.Millisecond

		out := Delay(ctx, in, delayDuration)

		start := time.Now()

		go func() {
			in <- 1
			time.Sleep(30 * time.Millisecond)
			in <- 2
			time.Sleep(30 * time.Millisecond)
			in <- 3
			close(in)
		}()

		var results []int
		var timestamps []time.Duration

		for val := range out {
			results = append(results, val)
			timestamps = append(timestamps, time.Since(start))
		}

		// Should receive all values
		if len(results) != 3 {
			t.Fatalf("expected 3 values, got %d", len(results))
		}

		// Values may arrive out of order or in order depending on timing
		// but all should be present
		found := make(map[int]bool)
		for _, v := range results {
			found[v] = true
		}
		for i := 1; i <= 3; i++ {
			if !found[i] {
				t.Errorf("missing value %d", i)
			}
		}

		// First value should arrive at approximately delayDuration
		if timestamps[0] < delayDuration-20*time.Millisecond {
			t.Errorf("first value arrived too early: %v", timestamps[0])
		}
	})

	t.Run("preserves all values with concurrent delays", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		delayDuration := 50 * time.Millisecond

		out := Delay(ctx, in, delayDuration)

		// Send 10 values rapidly
		go func() {
			for i := 1; i <= 10; i++ {
				in <- i
			}
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive all 10 values
		if len(results) != 10 {
			t.Fatalf("expected 10 values, got %d", len(results))
		}

		// Verify all values are present (order may vary due to concurrency)
		found := make(map[int]bool)
		for _, v := range results {
			found[v] = true
		}
		for i := 1; i <= 10; i++ {
			if !found[i] {
				t.Errorf("missing value %d", i)
			}
		}
	})

	t.Run("context cancellation stops delay", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in := make(chan int)
		delayDuration := 200 * time.Millisecond

		out := Delay(ctx, in, delayDuration)

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

		// Cancel context quickly
		time.Sleep(50 * time.Millisecond)
		cancel()

		// Channel should close without receiving delayed values
		timeout := time.After(300 * time.Millisecond)
		count := 0
		for {
			select {
			case _, ok := <-out:
				if !ok {
					// Channel closed
					// Should have received 0 values (all cancelled before delay completed)
					if count > 2 {
						t.Errorf("expected few or no values after quick cancellation, got %d", count)
					}
					return
				}
				count++
			case <-timeout:
				t.Fatal("channel did not close after context cancellation")
			}
		}
	})

	t.Run("context cancellation during send", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in := make(chan int)
		delayDuration := 50 * time.Millisecond

		out := Delay(ctx, in, delayDuration)

		go func() {
			in <- 1
			in <- 2
			close(in)
		}()

		// Wait for delays to complete
		time.Sleep(delayDuration + 20*time.Millisecond)

		// Cancel before reading
		cancel()

		// Should still handle closure gracefully
		timeout := time.After(200 * time.Millisecond)
		for {
			select {
			case _, ok := <-out:
				if !ok {
					return
				}
			case <-timeout:
				t.Fatal("channel did not close after context cancellation")
			}
		}
	})

	t.Run("input channel closure", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		delayDuration := 50 * time.Millisecond

		out := Delay(ctx, in, delayDuration)

		go func() {
			in <- 1
			in <- 2
			in <- 3
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive all values before output closes
		if len(results) != 3 {
			t.Fatalf("expected 3 values, got %d", len(results))
		}
	})

	t.Run("with buffered output channel", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		delayDuration := 50 * time.Millisecond

		out := Delay(ctx, in, delayDuration, WithBuffer[int](5))

		go func() {
			for i := 1; i <= 5; i++ {
				in <- i
			}
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		if len(results) != 5 {
			t.Fatalf("expected 5 values, got %d", len(results))
		}
	})

	t.Run("empty input channel", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		delayDuration := 50 * time.Millisecond

		close(in)

		out := Delay(ctx, in, delayDuration)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive no values
		if len(results) != 0 {
			t.Errorf("expected 0 values, got %d", len(results))
		}
	})

	t.Run("single value delay", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		delayDuration := 100 * time.Millisecond

		out := Delay(ctx, in, delayDuration)

		start := time.Now()

		go func() {
			in <- 42
			close(in)
		}()

		val := <-out
		elapsed := time.Since(start)

		if val != 42 {
			t.Errorf("expected value 42, got %d", val)
		}

		// Should take at least delayDuration
		if elapsed < delayDuration-20*time.Millisecond {
			t.Errorf("value arrived too early: %v (expected >= %v)", elapsed, delayDuration)
		}
	})

	t.Run("very short delay", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		delayDuration := 10 * time.Millisecond

		out := Delay(ctx, in, delayDuration)

		go func() {
			for i := 1; i <= 5; i++ {
				in <- i
			}
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive all values
		if len(results) != 5 {
			t.Fatalf("expected 5 values, got %d", len(results))
		}
	})

	t.Run("concurrent sends block correctly", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		delayDuration := 100 * time.Millisecond

		// Use unbuffered output to test blocking
		out := Delay(ctx, in, delayDuration)

		sent := make(chan struct{})
		go func() {
			for i := 1; i <= 3; i++ {
				in <- i
			}
			close(in)
			close(sent)
		}()

		// Wait for values to be sent
		<-sent

		// Start receiving after delay should have completed
		time.Sleep(delayDuration + 50*time.Millisecond)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		if len(results) != 3 {
			t.Fatalf("expected 3 values, got %d", len(results))
		}
	})

	t.Run("waits for all goroutines before closing", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)
		delayDuration := 100 * time.Millisecond

		// Send multiple values and close immediately
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := Delay(ctx, in, delayDuration)

		start := time.Now()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		elapsed := time.Since(start)

		// Should receive all 5 values
		if len(results) != 5 {
			t.Fatalf("expected 5 values, got %d", len(results))
		}

		// Should take at least delayDuration (all goroutines must complete)
		if elapsed < delayDuration-20*time.Millisecond {
			t.Errorf("channel closed too early: %v (expected >= %v)", elapsed, delayDuration)
		}
	})

	t.Run("zero delay duration", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		delayDuration := 0 * time.Millisecond

		out := Delay(ctx, in, delayDuration)

		go func() {
			for i := 1; i <= 5; i++ {
				in <- i
			}
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should still receive all values (just no delay)
		if len(results) != 5 {
			t.Fatalf("expected 5 values, got %d", len(results))
		}
	})
}

// TestTimeout tests the Timeout function
func TestTimeout(t *testing.T) {
	t.Run("basic timeout behavior", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		timeoutDuration := 150 * time.Millisecond

		out := Timeout(ctx, in, timeoutDuration)

		go func() {
			in <- 1
			time.Sleep(50 * time.Millisecond)
			in <- 2
			time.Sleep(50 * time.Millisecond)
			in <- 3
			// Don't send more, let it timeout
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive all 3 values before timeout
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

	t.Run("times out after inactivity", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		timeoutDuration := 100 * time.Millisecond

		out := Timeout(ctx, in, timeoutDuration)

		start := time.Now()

		go func() {
			in <- 1
			in <- 2
			// Stop sending, should timeout
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		elapsed := time.Since(start)

		// Should receive 2 values
		if len(results) != 2 {
			t.Fatalf("expected 2 values, got %d", len(results))
		}

		// Should timeout after approximately timeoutDuration from last value
		expectedMin := timeoutDuration
		expectedMax := timeoutDuration + 80*time.Millisecond
		if elapsed < expectedMin || elapsed > expectedMax {
			t.Errorf("expected timeout ~%v, got %v", timeoutDuration, elapsed)
		}
	})

	t.Run("timer resets on each value", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		timeoutDuration := 100 * time.Millisecond

		out := Timeout(ctx, in, timeoutDuration)

		go func() {
			// Send values at 60ms intervals (within timeout window)
			for i := 1; i <= 5; i++ {
				in <- i
				time.Sleep(60 * time.Millisecond)
			}
			// Stop sending, should timeout
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive all 5 values (timer keeps resetting)
		if len(results) != 5 {
			t.Fatalf("expected 5 values, got %d", len(results))
		}
	})

	t.Run("immediate timeout with no values", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		timeoutDuration := 80 * time.Millisecond

		out := Timeout(ctx, in, timeoutDuration)

		start := time.Now()

		// Don't send any values

		var results []int
		for val := range out {
			results = append(results, val)
		}

		elapsed := time.Since(start)

		// Should receive no values
		if len(results) != 0 {
			t.Errorf("expected 0 values, got %d", len(results))
		}

		// Should timeout after approximately timeoutDuration
		if elapsed < timeoutDuration-20*time.Millisecond {
			t.Errorf("timed out too early: %v (expected ~%v)", elapsed, timeoutDuration)
		}
	})

	t.Run("context cancellation stops timeout", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in := make(chan int)
		timeoutDuration := 200 * time.Millisecond

		out := Timeout(ctx, in, timeoutDuration)

		go func() {
			for i := 1; i <= 10; i++ {
				select {
				case in <- i:
					time.Sleep(30 * time.Millisecond)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()

		// Cancel context quickly
		time.Sleep(100 * time.Millisecond)
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
		timeoutDuration := 200 * time.Millisecond

		out := Timeout(ctx, in, timeoutDuration)

		go func() {
			in <- 1
			in <- 2
			in <- 3
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive all values and close immediately
		if len(results) != 3 {
			t.Fatalf("expected 3 values, got %d", len(results))
		}
	})

	t.Run("with buffered output channel", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		timeoutDuration := 100 * time.Millisecond

		out := Timeout(ctx, in, timeoutDuration, WithBuffer[int](5))

		go func() {
			for i := 1; i <= 3; i++ {
				in <- i
				time.Sleep(30 * time.Millisecond)
			}
			// Let it timeout
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		if len(results) != 3 {
			t.Fatalf("expected 3 values, got %d", len(results))
		}
	})

	t.Run("fast continuous values", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		timeoutDuration := 150 * time.Millisecond

		out := Timeout(ctx, in, timeoutDuration)

		go func() {
			// Send 20 values rapidly
			for i := 1; i <= 20; i++ {
				in <- i
				time.Sleep(10 * time.Millisecond)
			}
			// Stop sending, should timeout
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive all 20 values
		if len(results) != 20 {
			t.Fatalf("expected 20 values, got %d", len(results))
		}
	})

	t.Run("slow values trigger timeout", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		timeoutDuration := 80 * time.Millisecond

		out := Timeout(ctx, in, timeoutDuration)

		go func() {
			in <- 1
			time.Sleep(50 * time.Millisecond)
			in <- 2
			// Wait longer than timeout
			time.Sleep(timeoutDuration + 50*time.Millisecond)
			in <- 3 // This won't be received
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive only 2 values (timeout before 3rd)
		if len(results) != 2 {
			t.Fatalf("expected 2 values, got %d: %v", len(results), results)
		}
	})

	t.Run("very short timeout", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		timeoutDuration := 20 * time.Millisecond

		out := Timeout(ctx, in, timeoutDuration)

		go func() {
			in <- 1
			// Even short delay may cause timeout
			time.Sleep(30 * time.Millisecond)
			in <- 2
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive at least 1 value
		if len(results) < 1 {
			t.Fatal("expected at least 1 value")
		}
	})

	t.Run("context cancellation during send", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in := make(chan int)
		timeoutDuration := 100 * time.Millisecond

		out := Timeout(ctx, in, timeoutDuration)

		go func() {
			in <- 1
			in <- 2
			time.Sleep(30 * time.Millisecond)
			cancel()
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should handle cancellation gracefully
		if len(results) < 1 {
			t.Fatal("expected at least some values")
		}
	})

	t.Run("alternating fast and slow values", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		timeoutDuration := 100 * time.Millisecond

		out := Timeout(ctx, in, timeoutDuration)

		go func() {
			in <- 1
			time.Sleep(30 * time.Millisecond)
			in <- 2
			time.Sleep(30 * time.Millisecond)
			in <- 3
			// Timeout
			time.Sleep(timeoutDuration + 50*time.Millisecond)
			in <- 4 // Won't be received
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive 3 values before timeout
		if len(results) != 3 {
			t.Fatalf("expected 3 values, got %d", len(results))
		}
	})

	t.Run("timeout with buffered input", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)
		timeoutDuration := 80 * time.Millisecond

		// Pre-buffer values
		for i := 1; i <= 5; i++ {
			in <- i
		}

		out := Timeout(ctx, in, timeoutDuration)

		var results []int
		for val := range out {
			results = append(results, val)
			// Small delay between reads to test timeout behavior
			time.Sleep(20 * time.Millisecond)
		}

		// Should receive all buffered values (timer resets on each)
		if len(results) != 5 {
			t.Fatalf("expected 5 values, got %d", len(results))
		}
	})

	t.Run("zero timeout duration", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		timeoutDuration := 0 * time.Millisecond

		out := Timeout(ctx, in, timeoutDuration)

		go func() {
			in <- 1
		}()

		// Should timeout almost immediately
		start := time.Now()
		var results []int
		for val := range out {
			results = append(results, val)
		}
		elapsed := time.Since(start)

		// May or may not receive the value depending on goroutine scheduling
		// But should complete very quickly
		if elapsed > 100*time.Millisecond {
			t.Errorf("took too long with zero timeout: %v", elapsed)
		}
	})

	t.Run("multiple timeout cycles", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		timeoutDuration := 80 * time.Millisecond

		out := Timeout(ctx, in, timeoutDuration)

		go func() {
			// First batch - should all arrive
			in <- 1
			time.Sleep(30 * time.Millisecond)
			in <- 2
			time.Sleep(30 * time.Millisecond)
			in <- 3
			// Should timeout here
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive 3 values before timeout
		if len(results) != 3 {
			t.Fatalf("expected 3 values, got %d", len(results))
		}
	})
}
