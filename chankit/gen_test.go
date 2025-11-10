package chankit

import (
	"context"
	"testing"
	"time"
)

// TestGenerate tests the Generate function
func TestGenerate(t *testing.T) {
	t.Run("basic generation", func(t *testing.T) {
		ctx := context.Background()
		counter := 0
		genFunc := func() (int, bool) {
			if counter < 5 {
				counter++
				return counter, true
			}
			return 0, false
		}

		ch := Generate(ctx, genFunc)
		var result []int
		for val := range ch {
			result = append(result, val)
		}

		expected := []int{1, 2, 3, 4, 5}
		if len(result) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("with buffer", func(t *testing.T) {
		ctx := context.Background()
		counter := 0
		genFunc := func() (int, bool) {
			if counter < 10 {
				counter++
				return counter, true
			}
			return 0, false
		}

		ch := Generate(ctx, genFunc, WithBuffer[int](5))
		var result []int
		for val := range ch {
			result = append(result, val)
		}

		if len(result) != 10 {
			t.Fatalf("expected 10 values, got %d", len(result))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		counter := 0
		genFunc := func() (int, bool) {
			counter++
			return counter, true // infinite generator
		}

		ch := Generate(ctx, genFunc)

		// Read a few values
		<-ch
		<-ch
		<-ch

		// Cancel context
		cancel()

		// Channel should close
		timeout := time.After(100 * time.Millisecond)
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					return // success
				}
			case <-timeout:
				t.Fatal("channel did not close after context cancellation")
			}
		}
	})

	t.Run("context cancelled before generation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		genFunc := func() (int, bool) {
			return 1, true
		}

		ch := Generate(ctx, genFunc)

		// Channel should close immediately
		timeout := time.After(100 * time.Millisecond)
		select {
		case _, ok := <-ch:
			if ok {
				t.Fatal("expected channel to be closed")
			}
		case <-timeout:
			t.Fatal("channel did not close")
		}
	})

	t.Run("empty generation", func(t *testing.T) {
		ctx := context.Background()
		genFunc := func() (int, bool) {
			return 0, false // immediately stop
		}

		ch := Generate(ctx, genFunc)
		var result []int
		for val := range ch {
			result = append(result, val)
		}

		if len(result) != 0 {
			t.Fatalf("expected 0 values, got %d", len(result))
		}
	})

	t.Run("string generation", func(t *testing.T) {
		ctx := context.Background()
		words := []string{"hello", "world", "test"}
		idx := 0
		genFunc := func() (string, bool) {
			if idx < len(words) {
				word := words[idx]
				idx++
				return word, true
			}
			return "", false
		}

		ch := Generate(ctx, genFunc)
		var result []string
		for val := range ch {
			result = append(result, val)
		}

		if len(result) != len(words) {
			t.Fatalf("expected %d values, got %d", len(words), len(result))
		}
		for i, v := range result {
			if v != words[i] {
				t.Errorf("at index %d: expected %s, got %s", i, words[i], v)
			}
		}
	})
}

// TestRepeat tests the Repeat function
func TestRepeat(t *testing.T) {
	t.Run("basic repeat", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ch := Repeat(ctx, 42)

		// Read first 10 values
		for i := 0; i < 10; i++ {
			val := <-ch
			if val != 42 {
				t.Errorf("expected 42, got %d", val)
			}
		}

		cancel()

		// Channel should close
		timeout := time.After(100 * time.Millisecond)
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					return // success
				}
			case <-timeout:
				t.Fatal("channel did not close after context cancellation")
			}
		}
	})

	t.Run("with buffer", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ch := Repeat(ctx, "test", WithBuffer[string](5))

		// Read some values
		for i := 0; i < 20; i++ {
			val := <-ch
			if val != "test" {
				t.Errorf("expected 'test', got %s", val)
			}
		}
	})

	t.Run("context cancelled immediately", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		ch := Repeat(ctx, 99)

		// Channel should close immediately or after at most one value
		timeout := time.After(100 * time.Millisecond)
		valueCount := 0
		done := false
		for !done {
			select {
			case _, ok := <-ch:
				if !ok {
					done = true
					break
				}
				valueCount++
				if valueCount > 1 {
					t.Fatal("received more than one value after immediate cancellation")
				}
			case <-timeout:
				t.Fatal("channel did not close")
			}
		}
	})

	t.Run("repeat struct", func(t *testing.T) {
		type point struct {
			x, y int
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		p := point{x: 1, y: 2}
		ch := Repeat(ctx, p)

		// Read some values
		for i := 0; i < 5; i++ {
			val := <-ch
			if val != p {
				t.Errorf("expected %v, got %v", p, val)
			}
		}
	})
}

// TestRange tests the Range function
func TestRange(t *testing.T) {
	t.Run("ascending range", func(t *testing.T) {
		ctx := context.Background()
		ch := Range(ctx, 0, 10, 1)

		var result []int
		for val := range ch {
			result = append(result, val)
		}

		expected := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		if len(result) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("descending range", func(t *testing.T) {
		ctx := context.Background()
		ch := Range(ctx, 10, 0, -1)

		var result []int
		for val := range ch {
			result = append(result, val)
		}

		expected := []int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
		if len(result) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("range with step 2", func(t *testing.T) {
		ctx := context.Background()
		ch := Range(ctx, 0, 10, 2)

		var result []int
		for val := range ch {
			result = append(result, val)
		}

		expected := []int{0, 2, 4, 6, 8}
		if len(result) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("descending range with step -2", func(t *testing.T) {
		ctx := context.Background()
		ch := Range(ctx, 10, 0, -2)

		var result []int
		for val := range ch {
			result = append(result, val)
		}

		expected := []int{10, 8, 6, 4, 2}
		if len(result) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("empty range - zero step", func(t *testing.T) {
		ctx := context.Background()
		ch := Range(ctx, 0, 10, 0)

		var result []int
		timeout := time.After(50 * time.Millisecond)
		done := false
		for !done {
			select {
			case val, ok := <-ch:
				if !ok {
					done = true
					break
				}
				result = append(result, val)
			case <-timeout:
				done = true
			}
		}

		if len(result) != 0 {
			t.Fatalf("expected 0 values with zero step, got %d", len(result))
		}
	})

	t.Run("empty range - start equals end", func(t *testing.T) {
		ctx := context.Background()
		ch := Range(ctx, 5, 5, 1)

		var result []int
		for val := range ch {
			result = append(result, val)
		}

		if len(result) != 0 {
			t.Fatalf("expected 0 values when start equals end, got %d", len(result))
		}
	})

	t.Run("with buffer", func(t *testing.T) {
		ctx := context.Background()
		ch := Range(ctx, 0, 100, 1, WithBuffer[int](10))

		var result []int
		for val := range ch {
			result = append(result, val)
		}

		if len(result) != 100 {
			t.Fatalf("expected 100 values, got %d", len(result))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ch := Range(ctx, 0, 1000, 1)

		// Read a few values
		<-ch
		<-ch
		<-ch

		// Cancel context
		cancel()

		// Channel should close
		timeout := time.After(100 * time.Millisecond)
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					return // success
				}
			case <-timeout:
				t.Fatal("channel did not close after context cancellation")
			}
		}
	})

	t.Run("float64 range", func(t *testing.T) {
		ctx := context.Background()
		ch := Range(ctx, 0.0, 5.0, 0.5)

		var result []float64
		for val := range ch {
			result = append(result, val)
		}

		expected := []float64{0.0, 0.5, 1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5}
		if len(result) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %f, got %f", i, expected[i], v)
			}
		}
	})

	t.Run("uint8 range", func(t *testing.T) {
		ctx := context.Background()
		ch := Range(ctx, uint8(0), uint8(10), uint8(1))

		var result []uint8
		for val := range ch {
			result = append(result, val)
		}

		if len(result) != 10 {
			t.Fatalf("expected 10 values, got %d", len(result))
		}
		for i, v := range result {
			if v != uint8(i) {
				t.Errorf("at index %d: expected %d, got %d", i, i, v)
			}
		}
	})

	t.Run("negative range", func(t *testing.T) {
		ctx := context.Background()
		ch := Range(ctx, -5, 5, 1)

		var result []int
		for val := range ch {
			result = append(result, val)
		}

		expected := []int{-5, -4, -3, -2, -1, 0, 1, 2, 3, 4}
		if len(result) != len(expected) {
			t.Fatalf("expected %d values, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})
}
