package chankit

import (
	"context"
	"sort"
	"testing"
	"time"
)

// TestMerge tests the Merge function
func TestMerge(t *testing.T) {
	t.Run("merges values from multiple channels", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan int, 3)
		ch2 := make(chan int, 3)
		ch3 := make(chan int, 3)

		// Send values to all channels
		ch1 <- 1
		ch1 <- 2
		ch1 <- 3
		close(ch1)

		ch2 <- 10
		ch2 <- 20
		ch2 <- 30
		close(ch2)

		ch3 <- 100
		ch3 <- 200
		ch3 <- 300
		close(ch3)

		out := Merge(ctx, ch1, ch2, ch3)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive all 9 values
		if len(results) != 9 {
			t.Fatalf("expected 9 values, got %d", len(results))
		}

		// Sort to verify all values are present
		sort.Ints(results)
		expected := []int{1, 2, 3, 10, 20, 30, 100, 200, 300}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("handles single channel", func(t *testing.T) {
		ctx := context.Background()
		ch := make(chan int, 3)

		ch <- 1
		ch <- 2
		ch <- 3
		close(ch)

		out := Merge(ctx, ch)

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

	t.Run("handles empty channels", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan int)
		ch2 := make(chan int)

		close(ch1)
		close(ch2)

		out := Merge(ctx, ch1, ch2)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 values, got %d", len(results))
		}
	})

	t.Run("merges channels with different speeds", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan int)
		ch2 := make(chan int)

		go func() {
			for i := 1; i <= 3; i++ {
				ch1 <- i
				time.Sleep(10 * time.Millisecond)
			}
			close(ch1)
		}()

		go func() {
			for i := 10; i <= 13; i++ {
				ch2 <- i
				time.Sleep(5 * time.Millisecond)
			}
			close(ch2)
		}()

		out := Merge(ctx, ch1, ch2)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive all 7 values (3 from ch1, 4 from ch2)
		if len(results) != 7 {
			t.Fatalf("expected 7 values, got %d", len(results))
		}

		// Verify all values are present (order may vary)
		sort.Ints(results)
		expected := []int{1, 2, 3, 10, 11, 12, 13}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("context cancellation stops merge", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ch1 := make(chan int)
		ch2 := make(chan int)

		go func() {
			for i := 1; i <= 100; i++ {
				select {
				case ch1 <- i:
					time.Sleep(5 * time.Millisecond)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()

		go func() {
			for i := 100; i <= 200; i++ {
				select {
				case ch2 <- i:
					time.Sleep(5 * time.Millisecond)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()

		out := Merge(ctx, ch1, ch2)

		// Collect a few values then cancel
		count := 0
		for range out {
			count++
			if count == 10 {
				cancel()
			}
		}

		// Should have stopped after cancellation
		if count > 20 {
			t.Errorf("expected ~10-15 values after cancellation, got %d", count)
		}
	})

	t.Run("merges many channels", func(t *testing.T) {
		ctx := context.Background()

		channels := make([]<-chan int, 10)
		var expected []int
		for i := 0; i < 10; i++ {
			ch := make(chan int, 5)
			base := i * 10
			for j := 0; j < 5; j++ {
				val := base + j
				ch <- val
				expected = append(expected, val)
			}
			close(ch)
			channels[i] = ch
		}

		out := Merge(ctx, channels...)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive all 50 values (10 channels * 5 values)
		if len(results) != 50 {
			t.Fatalf("expected 50 values, got %d", len(results))
		}

		// Verify all values are present
		sort.Ints(results)
		sort.Ints(expected)
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("no channels provided", func(t *testing.T) {
		ctx := context.Background()

		out := Merge[int](ctx)

		// Channel should close immediately
		timeout := time.After(100 * time.Millisecond)
		select {
		case _, ok := <-out:
			if ok {
				t.Fatal("expected channel to be closed")
			}
		case <-timeout:
			t.Fatal("channel did not close")
		}
	})

	t.Run("one channel closes early", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan int, 2)
		ch2 := make(chan int)

		ch1 <- 1
		ch1 <- 2
		close(ch1)

		go func() {
			time.Sleep(50 * time.Millisecond)
			ch2 <- 10
			ch2 <- 20
			ch2 <- 30
			close(ch2)
		}()

		out := Merge(ctx, ch1, ch2)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive all values from both channels
		if len(results) != 5 {
			t.Fatalf("expected 5 values, got %d", len(results))
		}

		sort.Ints(results)
		expected := []int{1, 2, 10, 20, 30}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("concurrent writes don't block", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan int)
		ch2 := make(chan int)
		ch3 := make(chan int)

		out := Merge(ctx, ch1, ch2, ch3)

		// Start goroutines that write concurrently
		done := make(chan bool)
		go func() {
			for i := 0; i < 10; i++ {
				ch1 <- i
			}
			close(ch1)
			done <- true
		}()

		go func() {
			for i := 10; i < 20; i++ {
				ch2 <- i
			}
			close(ch2)
			done <- true
		}()

		go func() {
			for i := 20; i < 30; i++ {
				ch3 <- i
			}
			close(ch3)
			done <- true
		}()

		// Collect results
		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Wait for all goroutines to finish
		<-done
		<-done
		<-done

		// Should receive all 30 values
		if len(results) != 30 {
			t.Fatalf("expected 30 values, got %d", len(results))
		}
	})

	t.Run("merges string channels", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan string, 2)
		ch2 := make(chan string, 2)

		ch1 <- "hello"
		ch1 <- "world"
		close(ch1)

		ch2 <- "foo"
		ch2 <- "bar"
		close(ch2)

		out := Merge(ctx, ch1, ch2)

		var results []string
		for val := range out {
			results = append(results, val)
		}

		if len(results) != 4 {
			t.Fatalf("expected 4 values, got %d", len(results))
		}

		// Verify all strings are present
		sort.Strings(results)
		expected := []string{"bar", "foo", "hello", "world"}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %s, got %s", i, expected[i], v)
			}
		}
	})

	t.Run("context done before channels close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		ch1 := make(chan int)
		ch2 := make(chan int)

		// These channels will never close
		go func() {
			for i := 0; ; i++ {
				select {
				case ch1 <- i:
					time.Sleep(10 * time.Millisecond)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()

		go func() {
			for i := 100; ; i++ {
				select {
				case ch2 <- i:
					time.Sleep(10 * time.Millisecond)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()

		out := Merge(ctx, ch1, ch2)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should have stopped when context was done
		// We expect some values but not many (context timeout is 50ms)
		if len(results) > 20 {
			t.Errorf("expected fewer than 20 values with 50ms timeout, got %d", len(results))
		}
	})
}

// TestZip tests the Zip function
func TestZip(t *testing.T) {
	t.Run("zips values from two channels", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan int, 3)
		ch2 := make(chan string, 3)

		ch1 <- 1
		ch1 <- 2
		ch1 <- 3
		close(ch1)

		ch2 <- "a"
		ch2 <- "b"
		ch2 <- "c"
		close(ch2)

		out := Zip(ctx, ch1, ch2)

		var results []struct {
			First  int
			Second string
		}
		for val := range out {
			results = append(results, val)
		}

		// Should receive 3 pairs
		if len(results) != 3 {
			t.Fatalf("expected 3 pairs, got %d", len(results))
		}

		expected := []struct {
			First  int
			Second string
		}{
			{1, "a"},
			{2, "b"},
			{3, "c"},
		}

		for i, pair := range results {
			if pair.First != expected[i].First || pair.Second != expected[i].Second {
				t.Errorf("at index %d: expected (%d, %s), got (%d, %s)",
					i, expected[i].First, expected[i].Second, pair.First, pair.Second)
			}
		}
	})

	t.Run("stops when first channel closes", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan int, 2)
		ch2 := make(chan string, 5)

		ch1 <- 1
		ch1 <- 2
		close(ch1)

		ch2 <- "a"
		ch2 <- "b"
		ch2 <- "c"
		ch2 <- "d"
		ch2 <- "e"
		close(ch2)

		out := Zip(ctx, ch1, ch2)

		var results []struct {
			First  int
			Second string
		}
		for val := range out {
			results = append(results, val)
		}

		// Should only receive 2 pairs (limited by ch1)
		if len(results) != 2 {
			t.Fatalf("expected 2 pairs, got %d", len(results))
		}

		if results[0].First != 1 || results[0].Second != "a" {
			t.Errorf("pair 0: expected (1, a), got (%d, %s)", results[0].First, results[0].Second)
		}
		if results[1].First != 2 || results[1].Second != "b" {
			t.Errorf("pair 1: expected (2, b), got (%d, %s)", results[1].First, results[1].Second)
		}
	})

	t.Run("stops when second channel closes", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan int, 5)
		ch2 := make(chan string, 2)

		ch1 <- 1
		ch1 <- 2
		ch1 <- 3
		ch1 <- 4
		ch1 <- 5
		close(ch1)

		ch2 <- "a"
		ch2 <- "b"
		close(ch2)

		out := Zip(ctx, ch1, ch2)

		var results []struct {
			First  int
			Second string
		}
		for val := range out {
			results = append(results, val)
		}

		// Should only receive 2 pairs (limited by ch2)
		if len(results) != 2 {
			t.Fatalf("expected 2 pairs, got %d", len(results))
		}
	})

	t.Run("handles empty channels", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan int)
		ch2 := make(chan string)

		close(ch1)
		close(ch2)

		out := Zip(ctx, ch1, ch2)

		var results []struct {
			First  int
			Second string
		}
		for val := range out {
			results = append(results, val)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 pairs, got %d", len(results))
		}
	})

	t.Run("context cancellation stops zip", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ch1 := make(chan int)
		ch2 := make(chan string)

		go func() {
			for i := 1; i <= 100; i++ {
				select {
				case ch1 <- i:
					time.Sleep(5 * time.Millisecond)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()

		go func() {
			for i := 'a'; i <= 'z'; i++ {
				select {
				case ch2 <- string(i):
					time.Sleep(5 * time.Millisecond)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()

		out := Zip(ctx, ch1, ch2)

		// Collect a few pairs then cancel
		count := 0
		for range out {
			count++
			if count == 5 {
				cancel()
			}
		}

		// Should have stopped after cancellation
		if count > 10 {
			t.Errorf("expected ~5-7 pairs after cancellation, got %d", count)
		}
	})

	t.Run("handles slow first channel", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan int)
		ch2 := make(chan string, 10)

		// Pre-fill ch2
		for i := 'a'; i <= 'e'; i++ {
			ch2 <- string(i)
		}
		close(ch2)

		go func() {
			for i := 1; i <= 5; i++ {
				time.Sleep(20 * time.Millisecond)
				ch1 <- i
			}
			close(ch1)
		}()

		out := Zip(ctx, ch1, ch2)

		start := time.Now()
		var results []struct {
			First  int
			Second string
		}
		for val := range out {
			results = append(results, val)
		}
		elapsed := time.Since(start)

		// Should receive 5 pairs
		if len(results) != 5 {
			t.Fatalf("expected 5 pairs, got %d", len(results))
		}

		// Should take at least 100ms (5 * 20ms) due to slow first channel
		if elapsed < 100*time.Millisecond {
			t.Errorf("expected at least 100ms due to slow first channel, got %v", elapsed)
		}
	})

	t.Run("handles slow second channel", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan int, 10)
		ch2 := make(chan string)

		// Pre-fill ch1
		for i := 1; i <= 5; i++ {
			ch1 <- i
		}
		close(ch1)

		go func() {
			for i := 'a'; i <= 'e'; i++ {
				time.Sleep(20 * time.Millisecond)
				ch2 <- string(i)
			}
			close(ch2)
		}()

		out := Zip(ctx, ch1, ch2)

		start := time.Now()
		var results []struct {
			First  int
			Second string
		}
		for val := range out {
			results = append(results, val)
		}
		elapsed := time.Since(start)

		// Should receive 5 pairs
		if len(results) != 5 {
			t.Fatalf("expected 5 pairs, got %d", len(results))
		}

		// Should take at least 100ms (5 * 20ms) due to slow second channel
		if elapsed < 100*time.Millisecond {
			t.Errorf("expected at least 100ms due to slow second channel, got %v", elapsed)
		}
	})

	t.Run("zips different types", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan bool, 3)
		ch2 := make(chan float64, 3)

		ch1 <- true
		ch1 <- false
		ch1 <- true
		close(ch1)

		ch2 <- 1.5
		ch2 <- 2.5
		ch2 <- 3.5
		close(ch2)

		out := Zip(ctx, ch1, ch2)

		var results []struct {
			First  bool
			Second float64
		}
		for val := range out {
			results = append(results, val)
		}

		if len(results) != 3 {
			t.Fatalf("expected 3 pairs, got %d", len(results))
		}

		expected := []struct {
			First  bool
			Second float64
		}{
			{true, 1.5},
			{false, 2.5},
			{true, 3.5},
		}

		for i, pair := range results {
			if pair.First != expected[i].First || pair.Second != expected[i].Second {
				t.Errorf("at index %d: expected (%v, %.1f), got (%v, %.1f)",
					i, expected[i].First, expected[i].Second, pair.First, pair.Second)
			}
		}
	})

	t.Run("context cancelled while waiting for first channel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ch1 := make(chan int)
		ch2 := make(chan string, 5)

		// Fill ch2
		for i := 'a'; i <= 'e'; i++ {
			ch2 <- string(i)
		}
		close(ch2)

		out := Zip(ctx, ch1, ch2)

		// Cancel while waiting for first value from ch1
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		var results []struct {
			First  int
			Second string
		}
		for val := range out {
			results = append(results, val)
		}

		// Should receive no pairs
		if len(results) != 0 {
			t.Errorf("expected 0 pairs, got %d", len(results))
		}
	})

	t.Run("context cancelled while waiting for second channel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ch1 := make(chan int, 5)
		ch2 := make(chan string)

		// Fill ch1
		for i := 1; i <= 5; i++ {
			ch1 <- i
		}
		close(ch1)

		out := Zip(ctx, ch1, ch2)

		// Cancel while waiting for first value from ch2
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		var results []struct {
			First  int
			Second string
		}
		for val := range out {
			results = append(results, val)
		}

		// Should receive no pairs
		if len(results) != 0 {
			t.Errorf("expected 0 pairs, got %d", len(results))
		}
	})

	t.Run("single pair", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan int, 1)
		ch2 := make(chan string, 1)

		ch1 <- 42
		close(ch1)

		ch2 <- "answer"
		close(ch2)

		out := Zip(ctx, ch1, ch2)

		var results []struct {
			First  int
			Second string
		}
		for val := range out {
			results = append(results, val)
		}

		if len(results) != 1 {
			t.Fatalf("expected 1 pair, got %d", len(results))
		}

		if results[0].First != 42 || results[0].Second != "answer" {
			t.Errorf("expected (42, answer), got (%d, %s)", results[0].First, results[0].Second)
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		ch1 := make(chan int)
		ch2 := make(chan string)

		// Slow producers
		go func() {
			for i := 1; ; i++ {
				select {
				case ch1 <- i:
					time.Sleep(30 * time.Millisecond)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()

		go func() {
			for i := 'a'; ; i++ {
				select {
				case ch2 <- string(i):
					time.Sleep(30 * time.Millisecond)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()

		out := Zip(ctx, ch1, ch2)

		var results []struct {
			First  int
			Second string
		}
		for val := range out {
			results = append(results, val)
		}

		// Should have stopped due to context timeout
		// With 30ms per pair and 100ms timeout, expect ~3 pairs
		if len(results) > 5 {
			t.Errorf("expected ~3 pairs with 100ms timeout, got %d", len(results))
		}
	})

	t.Run("alternating speed channels", func(t *testing.T) {
		ctx := context.Background()
		ch1 := make(chan int)
		ch2 := make(chan string)

		go func() {
			for i := 1; i <= 5; i++ {
				ch1 <- i
				if i%2 == 0 {
					time.Sleep(20 * time.Millisecond)
				} else {
					time.Sleep(5 * time.Millisecond)
				}
			}
			close(ch1)
		}()

		go func() {
			for i := 'a'; i <= 'e'; i++ {
				ch2 <- string(i)
				if i%2 == 0 {
					time.Sleep(5 * time.Millisecond)
				} else {
					time.Sleep(20 * time.Millisecond)
				}
			}
			close(ch2)
		}()

		out := Zip(ctx, ch1, ch2)

		var results []struct {
			First  int
			Second string
		}
		for val := range out {
			results = append(results, val)
		}

		if len(results) != 5 {
			t.Fatalf("expected 5 pairs, got %d", len(results))
		}
	})
}
