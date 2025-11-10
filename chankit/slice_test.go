package chankit

import (
	"context"
	"testing"
	"time"
)

func TestSliceToChan_Unbuffered(t *testing.T) {
	ctx := context.Background()
	input := []int{1, 2, 3, 4, 5}

	ch := SliceToChan(ctx, input)

	var result []int
	for v := range ch {
		result = append(result, v)
	}

	if len(result) != len(input) {
		t.Errorf("expected %d items, got %d", len(input), len(result))
	}

	for i, v := range result {
		if v != input[i] {
			t.Errorf("at index %d: expected %d, got %d", i, input[i], v)
		}
	}
}

func TestSliceToChan_WithBuffer(t *testing.T) {
	ctx := context.Background()
	input := []int{1, 2, 3, 4, 5}

	ch := SliceToChan(ctx, input, WithBuffer[int](3))

	var result []int
	for v := range ch {
		result = append(result, v)
	}

	if len(result) != len(input) {
		t.Errorf("expected %d items, got %d", len(input), len(result))
	}

	for i, v := range result {
		if v != input[i] {
			t.Errorf("at index %d: expected %d, got %d", i, input[i], v)
		}
	}
}

func TestSliceToChan_WithBufferAuto(t *testing.T) {
	ctx := context.Background()
	input := []int{1, 2, 3, 4, 5}

	ch := SliceToChan(ctx, input, WithBufferAuto[int]())

	var result []int
	for v := range ch {
		result = append(result, v)
	}

	if len(result) != len(input) {
		t.Errorf("expected %d items, got %d", len(input), len(result))
	}

	for i, v := range result {
		if v != input[i] {
			t.Errorf("at index %d: expected %d, got %d", i, input[i], v)
		}
	}
}

func TestSliceToChan_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	ch := SliceToChan(ctx, input)

	// Read a few items then cancel
	count := 0
	for v := range ch {
		count++
		if v == 3 {
			cancel()
		}
		// Give goroutine time to see cancellation
		time.Sleep(10 * time.Millisecond)
	}

	// Should have stopped early due to cancellation
	if count >= len(input) {
		t.Errorf("expected early termination, but got all %d items", count)
	}
}

func TestSliceToChan_EmptySlice(t *testing.T) {
	ctx := context.Background()
	var input []int

	ch := SliceToChan(ctx, input)

	count := 0
	for range ch {
		count++
	}

	if count != 0 {
		t.Errorf("expected 0 items from empty slice, got %d", count)
	}
}

func TestSliceToChan_StringType(t *testing.T) {
	ctx := context.Background()
	input := []string{"hello", "world", "go", "generics"}

	ch := SliceToChan(ctx, input, WithBufferAuto[string]())

	var result []string
	for v := range ch {
		result = append(result, v)
	}

	if len(result) != len(input) {
		t.Errorf("expected %d items, got %d", len(input), len(result))
	}

	for i, v := range result {
		if v != input[i] {
			t.Errorf("at index %d: expected %s, got %s", i, input[i], v)
		}
	}
}

func TestChanToSlice_Basic(t *testing.T) {
	ctx := context.Background()
	ch := make(chan int)

	go func() {
		defer close(ch)
		for i := 1; i <= 5; i++ {
			ch <- i
		}
	}()

	result := ChanToSlice(ctx, ch)

	expected := []int{1, 2, 3, 4, 5}
	if len(result) != len(expected) {
		t.Errorf("expected %d items, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestChanToSlice_WithCapacity(t *testing.T) {
	ctx := context.Background()
	ch := make(chan int)

	go func() {
		defer close(ch)
		for i := 1; i <= 5; i++ {
			ch <- i
		}
	}()

	result := ChanToSlice(ctx, ch, WithCapacity[int](10))

	expected := []int{1, 2, 3, 4, 5}
	if len(result) != len(expected) {
		t.Errorf("expected %d items, got %d", len(expected), len(result))
	}

	if cap(result) < len(expected) {
		t.Errorf("expected capacity >= %d, got %d", len(expected), cap(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestChanToSlice_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	ch := make(chan int)

	go func() {
		defer close(ch)
		for i := 1; i <= 100; i++ {
			ch <- i
			time.Sleep(10 * time.Millisecond) // Slow producer
		}
	}()

	result := ChanToSlice(ctx, ch)

	// Should have been interrupted by context timeout
	if len(result) >= 100 {
		t.Errorf("expected early termination due to context, but got %d items", len(result))
	}
}

func TestChanToSlice_EmptyChannel(t *testing.T) {
	ctx := context.Background()
	ch := make(chan int)
	close(ch)

	result := ChanToSlice(ctx, ch)

	if len(result) != 0 {
		t.Errorf("expected 0 items from closed channel, got %d", len(result))
	}
}

func TestRoundTrip_SliceToChanToSlice(t *testing.T) {
	ctx := context.Background()
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Convert slice to channel
	ch := SliceToChan(ctx, input, WithBufferAuto[int]())

	// Convert back to slice
	result := ChanToSlice(ctx, ch, WithCapacity[int](len(input)))

	if len(result) != len(input) {
		t.Errorf("expected %d items, got %d", len(input), len(result))
	}

	for i, v := range result {
		if v != input[i] {
			t.Errorf("at index %d: expected %d, got %d", i, input[i], v)
		}
	}
}

func TestSliceToChan_StructType(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	ctx := context.Background()
	input := []Person{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
		{Name: "Charlie", Age: 35},
	}

	ch := SliceToChan(ctx, input, WithBuffer[Person](2))

	var result []Person
	for v := range ch {
		result = append(result, v)
	}

	if len(result) != len(input) {
		t.Errorf("expected %d items, got %d", len(input), len(result))
	}

	for i, v := range result {
		if v != input[i] {
			t.Errorf("at index %d: expected %+v, got %+v", i, input[i], v)
		}
	}
}

// Benchmark tests
func BenchmarkSliceToChan_Unbuffered(b *testing.B) {
	ctx := context.Background()
	input := make([]int, 1000)
	for i := range input {
		input[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := SliceToChan(ctx, input)
		for range ch {
		}
	}
}

func BenchmarkSliceToChan_BufferedAuto(b *testing.B) {
	ctx := context.Background()
	input := make([]int, 1000)
	for i := range input {
		input[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := SliceToChan(ctx, input, WithBufferAuto[int]())
		for range ch {
		}
	}
}

func BenchmarkSliceToChan_BufferedCustom(b *testing.B) {
	ctx := context.Background()
	input := make([]int, 1000)
	for i := range input {
		input[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := SliceToChan(ctx, input, WithBuffer[int](100))
		for range ch {
		}
	}
}

func BenchmarkChanToSlice_NoCapacity(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := make(chan int, 1000)
		go func() {
			defer close(ch)
			for j := 0; j < 1000; j++ {
				ch <- j
			}
		}()
		_ = ChanToSlice(ctx, ch)
	}
}

func BenchmarkChanToSlice_WithCapacity(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := make(chan int, 1000)
		go func() {
			defer close(ch)
			for j := 0; j < 1000; j++ {
				ch <- j
			}
		}()
		_ = ChanToSlice(ctx, ch, WithCapacity[int](1000))
	}
}
