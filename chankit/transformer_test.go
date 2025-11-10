package chankit

import (
	"context"
	"testing"
	"time"
)

// TestMap tests the Map function
func TestMap(t *testing.T) {
	t.Run("basic map", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5}
		inChan := SliceToChan(ctx, input)

		// Map to double the values
		mapFunc := func(x int) int { return x * 2 }
		outChan := Map(ctx, inChan, mapFunc)

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		expected := []int{2, 4, 6, 8, 10}
		if len(result) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("map with type conversion", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5}
		inChan := SliceToChan(ctx, input)

		// Map int to string
		mapFunc := func(x int) string {
			return string(rune('A' + x - 1))
		}
		outChan := Map(ctx, inChan, mapFunc)

		var result []string
		for val := range outChan {
			result = append(result, val)
		}

		expected := []string{"A", "B", "C", "D", "E"}
		if len(result) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %s, got %s", i, expected[i], v)
			}
		}
	})

	t.Run("map with buffer", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5}
		inChan := SliceToChan(ctx, input)

		mapFunc := func(x int) int { return x * x }
		outChan := Map(ctx, inChan, mapFunc, WithBuffer[int](3))

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		expected := []int{1, 4, 9, 16, 25}
		if len(result) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		inChan := SliceToChan(ctx, input)

		mapFunc := func(x int) int { return x * 2 }
		outChan := Map(ctx, inChan, mapFunc)

		// Read a few values
		<-outChan
		<-outChan

		// Cancel context
		cancel()

		// Channel should close
		timeout := time.After(100 * time.Millisecond)
		for {
			select {
			case _, ok := <-outChan:
				if !ok {
					return // success
				}
			case <-timeout:
				t.Fatal("channel did not close after context cancellation")
			}
		}
	})

	t.Run("empty input", func(t *testing.T) {
		ctx := context.Background()
		var input []int
		inChan := SliceToChan(ctx, input)

		mapFunc := func(x int) int { return x * 2 }
		outChan := Map(ctx, inChan, mapFunc)

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		if len(result) != 0 {
			t.Fatalf("expected 0 values, got %d", len(result))
		}
	})

	t.Run("channel closes properly", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3}
		inChan := SliceToChan(ctx, input)

		mapFunc := func(x int) int { return x * 2 }
		outChan := Map(ctx, inChan, mapFunc)

		// Consume all values
		for range outChan {
		}

		// Verify channel is closed
		_, ok := <-outChan
		if ok {
			t.Fatal("expected channel to be closed")
		}
	})
}

// TestFilter tests the Filter function
func TestFilter(t *testing.T) {
	t.Run("basic filter", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		inChan := SliceToChan(ctx, input)

		// Filter even numbers
		filterFunc := func(x int) bool { return x%2 == 0 }
		outChan := Filter(ctx, inChan, filterFunc)

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		expected := []int{2, 4, 6, 8, 10}
		if len(result) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("filter all false", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 3, 5, 7, 9}
		inChan := SliceToChan(ctx, input)

		// Filter even numbers (all should be filtered out)
		filterFunc := func(x int) bool { return x%2 == 0 }
		outChan := Filter(ctx, inChan, filterFunc)

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		if len(result) != 0 {
			t.Fatalf("expected 0 values, got %d", len(result))
		}
	})

	t.Run("filter all true", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5}
		inChan := SliceToChan(ctx, input)

		// Filter that accepts everything
		filterFunc := func(x int) bool { return true }
		outChan := Filter(ctx, inChan, filterFunc)

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		if len(result) != len(input) {
			t.Fatalf("expected %d values, got %d", len(input), len(result))
		}
		for i, v := range result {
			if v != input[i] {
				t.Errorf("at index %d: expected %d, got %d", i, input[i], v)
			}
		}
	})

	t.Run("filter with buffer", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		inChan := SliceToChan(ctx, input)

		filterFunc := func(x int) bool { return x > 5 }
		outChan := Filter(ctx, inChan, filterFunc, WithBuffer[int](3))

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		expected := []int{6, 7, 8, 9, 10}
		if len(result) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("context cancellation with filter all false", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Create a channel that sends many values
		inChan := make(chan int)
		go func() {
			defer close(inChan)
			for i := 0; i < 1000; i++ {
				select {
				case <-ctx.Done():
					return
				case inChan <- i:
				}
			}
		}()

		// Filter that always returns false
		filterFunc := func(x int) bool { return false }
		outChan := Filter(ctx, inChan, filterFunc)

		// Let some values be processed
		time.Sleep(10 * time.Millisecond)

		// Cancel context
		cancel()

		// Channel should close even though filter always returns false
		timeout := time.After(100 * time.Millisecond)
		for {
			select {
			case _, ok := <-outChan:
				if !ok {
					return // success
				}
			case <-timeout:
				t.Fatal("channel did not close after context cancellation")
			}
		}
	})

	t.Run("context cancellation during filtering", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		inChan := SliceToChan(ctx, input)

		filterFunc := func(x int) bool { return x%2 == 0 }
		outChan := Filter(ctx, inChan, filterFunc)

		// Read a value
		<-outChan

		// Cancel context
		cancel()

		// Channel should close
		timeout := time.After(100 * time.Millisecond)
		for {
			select {
			case _, ok := <-outChan:
				if !ok {
					return // success
				}
			case <-timeout:
				t.Fatal("channel did not close after context cancellation")
			}
		}
	})

	t.Run("empty input", func(t *testing.T) {
		ctx := context.Background()
		var input []int
		inChan := SliceToChan(ctx, input)

		filterFunc := func(x int) bool { return x%2 == 0 }
		outChan := Filter(ctx, inChan, filterFunc)

		var result []int
		for val := range outChan {
			result = append(result, val)
		}

		if len(result) != 0 {
			t.Fatalf("expected 0 values, got %d", len(result))
		}
	})

	t.Run("channel closes properly", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5}
		inChan := SliceToChan(ctx, input)

		filterFunc := func(x int) bool { return x%2 == 0 }
		outChan := Filter(ctx, inChan, filterFunc)

		// Consume all values
		for range outChan {
		}

		// Verify channel is closed
		_, ok := <-outChan
		if ok {
			t.Fatal("expected channel to be closed")
		}
	})

	t.Run("filter strings", func(t *testing.T) {
		ctx := context.Background()
		input := []string{"apple", "banana", "apricot", "cherry", "avocado"}
		inChan := SliceToChan(ctx, input)

		// Filter strings starting with 'a'
		filterFunc := func(s string) bool { return len(s) > 0 && s[0] == 'a' }
		outChan := Filter(ctx, inChan, filterFunc)

		var result []string
		for val := range outChan {
			result = append(result, val)
		}

		expected := []string{"apple", "apricot", "avocado"}
		if len(result) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %s, got %s", i, expected[i], v)
			}
		}
	})
}

// TestReduce tests the Reduce function
func TestReduce(t *testing.T) {
	t.Run("basic reduce - sum", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5}
		inChan := SliceToChan(ctx, input)

		reduceFunc := func(acc, val int) int { return acc + val }
		result := Reduce(ctx, inChan, reduceFunc, 0)

		expected := 15
		if result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}
	})

	t.Run("reduce - product", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5}
		inChan := SliceToChan(ctx, input)

		reduceFunc := func(acc, val int) int { return acc * val }
		result := Reduce(ctx, inChan, reduceFunc, 1)

		expected := 120
		if result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}
	})

	t.Run("reduce - find max", func(t *testing.T) {
		ctx := context.Background()
		input := []int{3, 7, 2, 9, 1, 5}
		inChan := SliceToChan(ctx, input)

		reduceFunc := func(acc, val int) int {
			if val > acc {
				return val
			}
			return acc
		}
		result := Reduce(ctx, inChan, reduceFunc, input[0])

		expected := 9
		if result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}
	})

	t.Run("reduce - string concatenation", func(t *testing.T) {
		ctx := context.Background()
		input := []string{"hello", "world", "from", "chankit"}
		inChan := SliceToChan(ctx, input)

		reduceFunc := func(acc, val string) string {
			if acc == "" {
				return val
			}
			return acc + " " + val
		}
		result := Reduce(ctx, inChan, reduceFunc, "")

		expected := "hello world from chankit"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("reduce with type conversion", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5}
		inChan := SliceToChan(ctx, input)

		// Reduce to count of elements
		reduceFunc := func(acc int, val int) int { return acc + 1 }
		result := Reduce(ctx, inChan, reduceFunc, 0)

		expected := 5
		if result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Create a channel that sends many values slowly
		inChan := make(chan int)
		go func() {
			defer close(inChan)
			for i := 0; i < 1000; i++ {
				select {
				case <-ctx.Done():
					return
				case inChan <- i:
					time.Sleep(1 * time.Millisecond)
				}
			}
		}()

		// Start reducing in background
		resultChan := make(chan int)
		go func() {
			reduceFunc := func(acc, val int) int { return acc + val }
			result := Reduce(ctx, inChan, reduceFunc, 0)
			resultChan <- result
		}()

		// Cancel after a short time
		time.Sleep(20 * time.Millisecond)
		cancel()

		// Should get partial result
		select {
		case result := <-resultChan:
			// Result should be less than full sum (1+2+...+999)
			if result >= 499500 {
				t.Errorf("expected partial result, got full result: %d", result)
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatal("reduce did not complete after context cancellation")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		ctx := context.Background()
		var input []int
		inChan := SliceToChan(ctx, input)

		reduceFunc := func(acc, val int) int { return acc + val }
		result := Reduce(ctx, inChan, reduceFunc, 42)

		// Should return initial value
		expected := 42
		if result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}
	})

	t.Run("single value", func(t *testing.T) {
		ctx := context.Background()
		input := []int{5}
		inChan := SliceToChan(ctx, input)

		reduceFunc := func(acc, val int) int { return acc + val }
		result := Reduce(ctx, inChan, reduceFunc, 10)

		expected := 15
		if result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}
	})
}

// TestMapFilterReduce tests combining Map, Filter, and Reduce
func TestMapFilterReduce(t *testing.T) {
	t.Run("map then filter", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5}
		inChan := SliceToChan(ctx, input)

		// Map to double
		mapFunc := func(x int) int { return x * 2 }
		mapped := Map(ctx, inChan, mapFunc)

		// Filter even numbers (all should pass since we doubled)
		filterFunc := func(x int) bool { return x > 5 }
		filtered := Filter(ctx, mapped, filterFunc)

		var result []int
		for val := range filtered {
			result = append(result, val)
		}

		expected := []int{6, 8, 10}
		if len(result) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("filter then map then reduce", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		inChan := SliceToChan(ctx, input)

		// Filter even numbers
		filterFunc := func(x int) bool { return x%2 == 0 }
		filtered := Filter(ctx, inChan, filterFunc)

		// Map to square
		mapFunc := func(x int) int { return x * x }
		mapped := Map(ctx, filtered, mapFunc)

		// Reduce to sum
		reduceFunc := func(acc, val int) int { return acc + val }
		result := Reduce(ctx, mapped, reduceFunc, 0)

		// 2^2 + 4^2 + 6^2 + 8^2 + 10^2 = 4 + 16 + 36 + 64 + 100 = 220
		expected := 220
		if result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}
	})
}
