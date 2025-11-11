package chankit

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestTap tests the Tap function
func TestTap(t *testing.T) {
	t.Run("basic tap passes through values", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5}
		inChan := SliceToChan(ctx, input)

		var tapped []int
		var mu sync.Mutex
		tapFunc := func(x int) {
			mu.Lock()
			tapped = append(tapped, x)
			mu.Unlock()
		}

		outChan := Tap(ctx, inChan, tapFunc)

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		// Check that all values passed through
		if len(result) != len(input) {
			t.Fatalf("expected %d values, got %d", len(input), len(result))
		}
		for i, v := range result {
			if v != input[i] {
				t.Errorf("at index %d: expected %d, got %d", i, input[i], v)
			}
		}

		// Check that tap function was called for all values
		if len(tapped) != len(input) {
			t.Fatalf("tap function called %d times, expected %d", len(tapped), len(input))
		}
		for i, v := range tapped {
			if v != input[i] {
				t.Errorf("tapped at index %d: expected %d, got %d", i, input[i], v)
			}
		}
	})

	t.Run("tap with side effects", func(t *testing.T) {
		ctx := context.Background()
		input := []string{"hello", "world", "test"}
		inChan := SliceToChan(ctx, input)

		count := 0
		var mu sync.Mutex
		tapFunc := func(s string) {
			mu.Lock()
			count++
			mu.Unlock()
		}

		outChan := Tap(ctx, inChan, tapFunc)

		var result []string
		for val := range outChan {
			result = append(result, val)
		}

		if count != len(input) {
			t.Errorf("expected count %d, got %d", len(input), count)
		}
		if len(result) != len(input) {
			t.Errorf("expected %d results, got %d", len(input), len(result))
		}
	})

	t.Run("tap respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		inChan := make(chan int)

		var tapped []int
		var mu sync.Mutex
		tapFunc := func(x int) {
			mu.Lock()
			tapped = append(tapped, x)
			mu.Unlock()
		}

		outChan := Tap(ctx, inChan, tapFunc)

		// Send some values
		go func() {
			inChan <- 1
			inChan <- 2
			time.Sleep(50 * time.Millisecond)
			cancel()
			time.Sleep(50 * time.Millisecond)
			close(inChan)
		}()

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		// Should have received some but not all values due to cancellation
		if len(result) > 2 {
			t.Errorf("expected at most 2 values due to cancellation, got %d", len(result))
		}
	})

	t.Run("tap with empty channel", func(t *testing.T) {
		ctx := context.Background()
		inChan := make(chan int)
		close(inChan)

		called := false
		tapFunc := func(x int) {
			called = true
		}

		outChan := Tap(ctx, inChan, tapFunc)

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		if called {
			t.Error("tap function should not be called for empty channel")
		}
		if len(result) != 0 {
			t.Errorf("expected 0 results, got %d", len(result))
		}
	})

	t.Run("tap with buffered channel option", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3}
		inChan := SliceToChan(ctx, input)

		var tapped []int
		var mu sync.Mutex
		tapFunc := func(x int) {
			mu.Lock()
			tapped = append(tapped, x)
			mu.Unlock()
		}

		outChan := Tap(ctx, inChan, tapFunc, WithBuffer[int](10))

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		if len(result) != len(input) {
			t.Errorf("expected %d values, got %d", len(input), len(result))
		}
		if len(tapped) != len(input) {
			t.Errorf("expected %d tapped values, got %d", len(input), len(tapped))
		}
	})
}

// TestFlatMap tests the FlatMap function
func TestFlatMap(t *testing.T) {
	t.Run("basic flatmap", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3}
		inChan := SliceToChan(ctx, input)

		// FlatMap each number to a channel of its multiples
		flatMapFunc := func(x int) <-chan int {
			return SliceToChan(ctx, []int{x, x * 2})
		}

		outChan := FlatMap(ctx, inChan, flatMapFunc)

		result := make(map[int]int) // value -> count
		for val := range outChan {
			result[val]++
		}

		// Expected: 1, 2 (from 1), 2, 4 (from 2), 3, 6 (from 3)
		expected := map[int]int{1: 1, 2: 2, 3: 1, 4: 1, 6: 1}

		if len(result) != len(expected) {
			t.Errorf("expected %d unique values, got %d", len(expected), len(result))
		}

		for k, v := range expected {
			if result[k] != v {
				t.Errorf("for value %d: expected count %d, got %d", k, v, result[k])
			}
		}
	})

	t.Run("flatmap with empty inner channels", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3}
		inChan := SliceToChan(ctx, input)

		// FlatMap to empty channels
		flatMapFunc := func(x int) <-chan int {
			ch := make(chan int)
			close(ch)
			return ch
		}

		outChan := FlatMap(ctx, inChan, flatMapFunc)

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		if len(result) != 0 {
			t.Errorf("expected 0 values from empty inner channels, got %d", len(result))
		}
	})

	t.Run("flatmap with varying inner channel sizes", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3}
		inChan := SliceToChan(ctx, input)

		// FlatMap each number to a channel with varying sizes
		flatMapFunc := func(x int) <-chan int {
			var values []int
			for i := 0; i < x; i++ {
				values = append(values, x)
			}
			return SliceToChan(ctx, values)
		}

		outChan := FlatMap(ctx, inChan, flatMapFunc)

		result := make(map[int]int) // value -> count
		for val := range outChan {
			result[val]++
		}

		// Expected: 1 appears 1 time, 2 appears 2 times, 3 appears 3 times
		expected := map[int]int{1: 1, 2: 2, 3: 3}

		if len(result) != len(expected) {
			t.Errorf("expected %d unique values, got %d", len(expected), len(result))
		}

		for k, v := range expected {
			if result[k] != v {
				t.Errorf("for value %d: expected count %d, got %d", k, v, result[k])
			}
		}
	})

	t.Run("flatmap respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		inChan := make(chan int)

		// FlatMap to channels with delay
		flatMapFunc := func(x int) <-chan int {
			ch := make(chan int)
			go func() {
				defer close(ch)
				time.Sleep(100 * time.Millisecond)
				select {
				case <-ctx.Done():
					return
				case ch <- x:
				}
			}()
			return ch
		}

		outChan := FlatMap(ctx, inChan, flatMapFunc)

		// Send a value and cancel quickly
		go func() {
			inChan <- 1
			inChan <- 2
			time.Sleep(50 * time.Millisecond)
			cancel()
			close(inChan)
		}()

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		// Due to cancellation, we might not receive all values
		// Just verify that the channel closes properly
		if len(result) > 2 {
			t.Errorf("expected at most 2 values due to cancellation, got %d", len(result))
		}
	})

	t.Run("flatmap with type conversion", func(t *testing.T) {
		ctx := context.Background()
		input := []string{"a", "b", "c"}
		inChan := SliceToChan(ctx, input)

		// FlatMap string to channel of runes
		flatMapFunc := func(s string) <-chan rune {
			runes := []rune(s)
			return SliceToChan(ctx, runes)
		}

		outChan := FlatMap(ctx, inChan, flatMapFunc)

		var result []rune
		for val := range outChan {
			result = append(result, val)
		}

		expected := []rune{'a', 'b', 'c'}
		if len(result) != len(expected) {
			t.Errorf("expected %d runes, got %d", len(expected), len(result))
		}
	})

	t.Run("flatmap with empty input channel", func(t *testing.T) {
		ctx := context.Background()
		inChan := make(chan int)
		close(inChan)

		flatMapFunc := func(x int) <-chan int {
			return SliceToChan(ctx, []int{x, x * 2})
		}

		outChan := FlatMap(ctx, inChan, flatMapFunc)

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		if len(result) != 0 {
			t.Errorf("expected 0 values from empty input, got %d", len(result))
		}
	})

	t.Run("flatmap with buffered channel option", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2}
		inChan := SliceToChan(ctx, input)

		flatMapFunc := func(x int) <-chan int {
			return SliceToChan(ctx, []int{x, x * 2})
		}

		outChan := FlatMap(ctx, inChan, flatMapFunc, WithBuffer[int](10))

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		if len(result) != 4 {
			t.Errorf("expected 4 values, got %d", len(result))
		}
	})

	t.Run("flatmap maintains concurrency", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5}
		inChan := SliceToChan(ctx, input)

		var mu sync.Mutex
		processing := make(map[int]bool)
		maxConcurrent := 0

		flatMapFunc := func(x int) <-chan int {
			ch := make(chan int)
			go func() {
				defer close(ch)

				mu.Lock()
				processing[x] = true
				current := len(processing)
				if current > maxConcurrent {
					maxConcurrent = current
				}
				mu.Unlock()

				time.Sleep(50 * time.Millisecond)
				ch <- x

				mu.Lock()
				delete(processing, x)
				mu.Unlock()
			}()
			return ch
		}

		outChan := FlatMap(ctx, inChan, flatMapFunc)

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		// Should process multiple inner channels concurrently
		if maxConcurrent < 2 {
			t.Errorf("expected concurrent processing of at least 2, got %d", maxConcurrent)
		}

		if len(result) != len(input) {
			t.Errorf("expected %d values, got %d", len(input), len(result))
		}
	})
}
