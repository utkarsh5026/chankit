package chankit

import (
	"context"
	"testing"
	"time"
)

// TestTake tests the Take function
func TestTake(t *testing.T) {
	t.Run("takes specified count", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)

		// Send 10 values
		for i := 1; i <= 10; i++ {
			in <- i
		}
		close(in)

		out := Take(ctx, in, 5)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should take only first 5 values
		expected := []int{1, 2, 3, 4, 5}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(results))
		}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("takes all if count exceeds available", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 5)

		// Send 5 values
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := Take(ctx, in, 10)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should take all 5 available values
		if len(results) != 5 {
			t.Fatalf("expected 5 values, got %d", len(results))
		}
		for i, v := range results {
			if v != i+1 {
				t.Errorf("at index %d: expected %d, got %d", i, i+1, v)
			}
		}
	})

	t.Run("takes zero values", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 5)

		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := Take(ctx, in, 0)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should take nothing
		if len(results) != 0 {
			t.Errorf("expected 0 values, got %d", len(results))
		}
	})

	t.Run("context cancellation stops early", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in := make(chan int)

		out := Take(ctx, in, 10)

		go func() {
			for i := 1; i <= 20; i++ {
				select {
				case in <- i:
					if i == 3 {
						cancel()
					}
					time.Sleep(10 * time.Millisecond)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should stop after cancellation (around 3 or fewer)
		if len(results) > 5 {
			t.Errorf("expected ~3 values after cancellation, got %d", len(results))
		}
	})

	t.Run("input channel closes before count reached", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)

		out := Take(ctx, in, 10)

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

		// Should receive all values before closure
		if len(results) != 3 {
			t.Fatalf("expected 3 values, got %d", len(results))
		}
	})

	t.Run("with buffered output channel", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)

		for i := 1; i <= 10; i++ {
			in <- i
		}
		close(in)

		out := Take(ctx, in, 5, WithBuffer[int](3))

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
		close(in)

		out := Take(ctx, in, 5)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive nothing
		if len(results) != 0 {
			t.Errorf("expected 0 values, got %d", len(results))
		}
	})

	t.Run("slow producer", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)

		out := Take(ctx, in, 3)

		go func() {
			for i := 1; i <= 10; i++ {
				in <- i
				time.Sleep(20 * time.Millisecond)
			}
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should take first 3 values
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

	t.Run("negative count", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 5)

		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := Take(ctx, in, -1)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Negative count should take nothing
		if len(results) != 0 {
			t.Errorf("expected 0 values, got %d", len(results))
		}
	})
}

// TestSkip tests the Skip function
func TestSkip(t *testing.T) {
	t.Run("skips specified count", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)

		// Send 10 values
		for i := 1; i <= 10; i++ {
			in <- i
		}
		close(in)

		out := Skip(ctx, in, 5)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should skip first 5, get last 5
		expected := []int{6, 7, 8, 9, 10}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(results))
		}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("skips all if count exceeds available", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 5)

		// Send 5 values
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := Skip(ctx, in, 10)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should skip all, get nothing
		if len(results) != 0 {
			t.Errorf("expected 0 values, got %d", len(results))
		}
	})

	t.Run("skips zero values", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 5)

		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := Skip(ctx, in, 0)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should skip nothing, get all
		if len(results) != 5 {
			t.Fatalf("expected 5 values, got %d", len(results))
		}
		for i, v := range results {
			if v != i+1 {
				t.Errorf("at index %d: expected %d, got %d", i, i+1, v)
			}
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in := make(chan int)

		out := Skip(ctx, in, 3)

		go func() {
			for i := 1; i <= 20; i++ {
				select {
				case in <- i:
					if i == 8 {
						cancel()
					}
					time.Sleep(10 * time.Millisecond)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should skip first 3, then get some values before cancellation
		if len(results) == 0 {
			t.Error("expected some values after skipping")
		}
	})

	t.Run("with buffered output channel", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)

		for i := 1; i <= 10; i++ {
			in <- i
		}
		close(in)

		out := Skip(ctx, in, 3, WithBuffer[int](5))

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should skip first 3, get remaining 7
		if len(results) != 7 {
			t.Fatalf("expected 7 values, got %d", len(results))
		}
	})

	t.Run("empty input channel", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		close(in)

		out := Skip(ctx, in, 5)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive nothing
		if len(results) != 0 {
			t.Errorf("expected 0 values, got %d", len(results))
		}
	})

	t.Run("slow producer", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)

		out := Skip(ctx, in, 2)

		go func() {
			for i := 1; i <= 5; i++ {
				in <- i
				time.Sleep(20 * time.Millisecond)
			}
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should skip first 2, get last 3
		expected := []int{3, 4, 5}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(results))
		}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("negative count", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 5)

		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := Skip(ctx, in, -1)

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Negative count should skip nothing, get all
		if len(results) != 5 {
			t.Fatalf("expected 5 values, got %d", len(results))
		}
	})
}

// TestTakeWhile tests the TakeWhile function
func TestTakeWhile(t *testing.T) {
	t.Run("takes while predicate is true", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)

		// Send values 1 to 10
		for i := 1; i <= 10; i++ {
			in <- i
		}
		close(in)

		// Take while value < 6
		out := TakeWhile(ctx, in, func(v int) bool { return v < 6 })

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should take 1, 2, 3, 4, 5 (stop at 6)
		expected := []int{1, 2, 3, 4, 5}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(results))
		}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("takes all if predicate always true", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 5)

		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := TakeWhile(ctx, in, func(v int) bool { return true })

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should take all values
		if len(results) != 5 {
			t.Fatalf("expected 5 values, got %d", len(results))
		}
	})

	t.Run("takes none if predicate immediately false", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 5)

		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := TakeWhile(ctx, in, func(v int) bool { return false })

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should take nothing
		if len(results) != 0 {
			t.Errorf("expected 0 values, got %d", len(results))
		}
	})

	t.Run("with even number predicate", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)

		// Start with even numbers: 2, 4, 6, then odd: 7, 9
		values := []int{2, 4, 6, 7, 9, 10}
		for _, v := range values {
			in <- v
		}
		close(in)

		out := TakeWhile(ctx, in, func(v int) bool { return v%2 == 0 })

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should take 2, 4, 6 (stop at 7)
		expected := []int{2, 4, 6}
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
		in := make(chan int)

		out := TakeWhile(ctx, in, func(v int) bool { return v < 100 })

		go func() {
			for i := 1; i <= 20; i++ {
				select {
				case in <- i:
					if i == 3 {
						cancel()
					}
					time.Sleep(10 * time.Millisecond)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should stop after cancellation
		if len(results) > 5 {
			t.Errorf("expected ~3 values after cancellation, got %d", len(results))
		}
	})

	t.Run("with buffered output channel", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)

		for i := 1; i <= 10; i++ {
			in <- i
		}
		close(in)

		out := TakeWhile(ctx, in, func(v int) bool { return v <= 5 }, WithBuffer[int](3))

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
		close(in)

		out := TakeWhile(ctx, in, func(v int) bool { return true })

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive nothing
		if len(results) != 0 {
			t.Errorf("expected 0 values, got %d", len(results))
		}
	})

	t.Run("slow producer", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)

		out := TakeWhile(ctx, in, func(v int) bool { return v < 4 })

		go func() {
			for i := 1; i <= 10; i++ {
				in <- i
				time.Sleep(20 * time.Millisecond)
			}
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should take 1, 2, 3 (stop at 4)
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

	t.Run("with string type", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan string, 10)

		words := []string{"apple", "banana", "cat", "dog", "elephant"}
		for _, w := range words {
			in <- w
		}
		close(in)

		// Take while length < 4
		out := TakeWhile(ctx, in, func(s string) bool { return len(s) < 4 })

		var results []string
		for val := range out {
			results = append(results, val)
		}

		// Should take only "apple" and "cat" (stop at "banana" which is too long)
		// Wait, "apple" is 5 chars, so it stops immediately
		expected := []string{}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d: %v", len(expected), len(results), results)
		}
	})
}

// TestSkipWhile tests the SkipWhile function
func TestSkipWhile(t *testing.T) {
	t.Run("skips while predicate is true", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)

		// Send values 1 to 10
		for i := 1; i <= 10; i++ {
			in <- i
		}
		close(in)

		// Skip while value < 6
		out := SkipWhile(ctx, in, func(v int) bool { return v < 6 })

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should skip 1-5, get 6-10
		expected := []int{6, 7, 8, 9, 10}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(results))
		}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("skips none if predicate immediately false", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 5)

		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := SkipWhile(ctx, in, func(v int) bool { return false })

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should skip nothing, get all
		if len(results) != 5 {
			t.Fatalf("expected 5 values, got %d", len(results))
		}
	})

	t.Run("skips all if predicate always true", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 5)

		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)

		out := SkipWhile(ctx, in, func(v int) bool { return true })

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should skip all
		if len(results) != 0 {
			t.Errorf("expected 0 values, got %d", len(results))
		}
	})

	t.Run("with even number predicate", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)

		// Start with even numbers: 2, 4, 6, then odd: 7, then even: 8, 10
		values := []int{2, 4, 6, 7, 8, 10}
		for _, v := range values {
			in <- v
		}
		close(in)

		out := SkipWhile(ctx, in, func(v int) bool { return v%2 == 0 })

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should skip 2, 4, 6, then get 7, 8, 10 (including even numbers after first odd)
		expected := []int{7, 8, 10}
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
		in := make(chan int)

		out := SkipWhile(ctx, in, func(v int) bool { return v < 5 })

		go func() {
			for i := 1; i <= 20; i++ {
				select {
				case in <- i:
					if i == 10 {
						cancel()
					}
					time.Sleep(10 * time.Millisecond)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should skip 1-4, get some values 5+ before cancellation
		if len(results) == 0 {
			t.Error("expected some values after skipping")
		}
		if results[0] < 5 {
			t.Errorf("expected first value >= 5, got %d", results[0])
		}
	})

	t.Run("with buffered output channel", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)

		for i := 1; i <= 10; i++ {
			in <- i
		}
		close(in)

		out := SkipWhile(ctx, in, func(v int) bool { return v <= 3 }, WithBuffer[int](5))

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should skip 1-3, get 4-10
		if len(results) != 7 {
			t.Fatalf("expected 7 values, got %d", len(results))
		}
	})

	t.Run("empty input channel", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)
		close(in)

		out := SkipWhile(ctx, in, func(v int) bool { return true })

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should receive nothing
		if len(results) != 0 {
			t.Errorf("expected 0 values, got %d", len(results))
		}
	})

	t.Run("slow producer", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int)

		out := SkipWhile(ctx, in, func(v int) bool { return v < 4 })

		go func() {
			for i := 1; i <= 6; i++ {
				in <- i
				time.Sleep(20 * time.Millisecond)
			}
			close(in)
		}()

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should skip 1-3, get 4-6
		expected := []int{4, 5, 6}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(results))
		}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("continues after predicate becomes false", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan int, 10)

		// Negative, positive, negative pattern
		values := []int{-1, -2, -3, 5, -10, -20, 7}
		for _, v := range values {
			in <- v
		}
		close(in)

		// Skip while negative
		out := SkipWhile(ctx, in, func(v int) bool { return v < 0 })

		var results []int
		for val := range out {
			results = append(results, val)
		}

		// Should skip -1, -2, -3, then get all remaining including negatives
		expected := []int{5, -10, -20, 7}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(results))
		}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("with string type", func(t *testing.T) {
		ctx := context.Background()
		in := make(chan string, 10)

		words := []string{"a", "an", "the", "hello", "world"}
		for _, w := range words {
			in <- w
		}
		close(in)

		// Skip while length <= 3
		out := SkipWhile(ctx, in, func(s string) bool { return len(s) <= 3 })

		var results []string
		for val := range out {
			results = append(results, val)
		}

		// Should skip "a", "an", "the", get "hello", "world"
		expected := []string{"hello", "world"}
		if len(results) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(results))
		}
		for i, v := range results {
			if v != expected[i] {
				t.Errorf("at index %d: expected %s, got %s", i, expected[i], v)
			}
		}
	})
}
