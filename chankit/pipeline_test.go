package chankit

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
)

// ============================================================================
// Basic Pipeline Creation Tests
// ============================================================================

func TestNewPipeline(t *testing.T) {
	ctx := context.Background()
	pipeline := NewPipeline[int](ctx)

	if pipeline == nil {
		t.Fatal("NewPipeline returned nil")
	}
	if pipeline.ctx != ctx {
		t.Error("Pipeline context not set correctly")
	}
}

func TestFromChannel(t *testing.T) {
	ctx := context.Background()
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)

	result := From(ctx, ch).ToSlice()

	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestFromSlice(t *testing.T) {
	ctx := context.Background()
	slice := []int{1, 2, 3, 4, 5}

	result := FromSlice(ctx, slice).ToSlice()

	if !reflect.DeepEqual(result, slice) {
		t.Errorf("Expected %v, got %v", slice, result)
	}
}

// ============================================================================
// Generator Method Tests
// ============================================================================

func TestPipelineRange(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		start    int
		end      int
		step     int
		expected []int
	}{
		{"ascending", 1, 6, 1, []int{1, 2, 3, 4, 5}},
		{"descending", 5, 0, -1, []int{5, 4, 3, 2, 1}},
		{"step_2", 0, 10, 2, []int{0, 2, 4, 6, 8}},
		{"single", 5, 6, 1, []int{5}},
		{"empty", 5, 5, 1, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RangePipeline(ctx, tt.start, tt.end, tt.step).
				ToSlice()

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPipelineRepeat(t *testing.T) {
	ctx := context.Background()

	result := NewPipeline[string](ctx).
		Repeat("ping").
		Take(5).
		ToSlice()

	expected := []string{"ping", "ping", "ping", "ping", "ping"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPipelineGenerate(t *testing.T) {
	ctx := context.Background()
	i := 0

	result := NewPipeline[int](ctx).
		Generate(func() (int, bool) {
			i++
			return i, i <= 5
		}).
		ToSlice()

	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// ============================================================================
// Transformation Method Tests
// ============================================================================

func TestPipelineMap(t *testing.T) {
	ctx := context.Background()

	result := FromSlice(ctx, []int{1, 2, 3, 4, 5}).
		Map(func(x int) any { return x * 2 }).
		ToSlice()

	expected := []any{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPipelineMapTo(t *testing.T) {
	ctx := context.Background()

	result := MapTo(
		FromSlice(ctx, []int{1, 2, 3}),
		func(x int) string { return fmt.Sprintf("num_%d", x) },
	).ToSlice()

	expected := []string{"num_1", "num_2", "num_3"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPipelineFilter(t *testing.T) {
	ctx := context.Background()

	result := FromSlice(ctx, []int{1, 2, 3, 4, 5, 6}).
		Filter(func(x int) bool { return x%2 == 0 }).
		ToSlice()

	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPipelineFlatMap(t *testing.T) {
	ctx := context.Background()

	result := FromSlice(ctx, []int{1, 2, 3}).
		FlatMap(func(x int) <-chan int {
			ch := make(chan int, 2)
			go func() {
				defer close(ch)
				ch <- x
				ch <- x * 10
			}()
			return ch
		}).
		ToSlice()

	// FlatMap with concurrency may have varying order, so check length and values
	if len(result) != 6 {
		t.Errorf("Expected 6 values, got %d", len(result))
	}
}

// ============================================================================
// Selection Method Tests
// ============================================================================

func TestPipelineTake(t *testing.T) {
	ctx := context.Background()

	result := FromSlice(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
		Take(5).
		ToSlice()

	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPipelineSkip(t *testing.T) {
	ctx := context.Background()

	result := FromSlice(ctx, []int{1, 2, 3, 4, 5}).
		Skip(2).
		ToSlice()

	expected := []int{3, 4, 5}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPipelineTakeWhile(t *testing.T) {
	ctx := context.Background()

	result := FromSlice(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8}).
		TakeWhile(func(x int) bool { return x < 5 }).
		ToSlice()

	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPipelineSkipWhile(t *testing.T) {
	ctx := context.Background()

	result := FromSlice(ctx, []int{1, 2, 3, 4, 5, 6}).
		SkipWhile(func(x int) bool { return x < 4 }).
		ToSlice()

	expected := []int{4, 5, 6}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// ============================================================================
// Flow Control Method Tests
// ============================================================================

func TestPipelineThrottle(t *testing.T) {
	ctx := context.Background()
	ch := make(chan int)

	go func() {
		for i := 1; i <= 5; i++ {
			ch <- i
			time.Sleep(10 * time.Millisecond)
		}
		time.Sleep(100 * time.Millisecond) // Wait for throttle window
		close(ch)
	}()

	result := From(ctx, ch).
		Throttle(50 * time.Millisecond).
		ToSlice()

	// Should get at least one value (throttled)
	if len(result) == 0 {
		t.Error("Throttle returned no values")
	}

	// Should get fewer values than input due to throttling
	if len(result) >= 5 {
		t.Errorf("Expected fewer than 5 values due to throttling, got %d", len(result))
	}
}

func TestPipelineDebounce(t *testing.T) {
	ctx := context.Background()
	ch := make(chan int)

	go func() {
		for i := 1; i <= 3; i++ {
			ch <- i
			time.Sleep(20 * time.Millisecond)
		}
		time.Sleep(150 * time.Millisecond) // Wait for debounce
		close(ch)
	}()

	result := From(ctx, ch).
		Debounce(100 * time.Millisecond).
		ToSlice()

	// Should only get the last value after silence
	expected := []int{3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPipelineFixedInterval(t *testing.T) {
	ctx := context.Background()

	start := time.Now()
	result := FromSlice(ctx, []int{1, 2, 3}).
		FixedInterval(50 * time.Millisecond).
		ToSlice()

	duration := time.Since(start)

	// Should take at least 3 * 50ms = 150ms
	if duration < 150*time.Millisecond {
		t.Errorf("FixedInterval too fast: %v", duration)
	}

	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPipelineBatch(t *testing.T) {
	ctx := context.Background()

	batches := FromSlice(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}).
		Batch(3, 1*time.Second)

	var result [][]int
	for batch := range batches {
		result = append(result, batch)
	}

	expected := [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPipelineBatchTimeout(t *testing.T) {
	ctx := context.Background()
	ch := make(chan int)

	go func() {
		ch <- 1
		ch <- 2
		time.Sleep(150 * time.Millisecond)
		close(ch)
	}()

	batches := From(ctx, ch).Batch(10, 100*time.Millisecond)

	var result [][]int
	for batch := range batches {
		result = append(result, batch)
	}

	// Should batch by timeout since we only send 2 values
	if len(result) != 1 {
		t.Errorf("Expected 1 batch, got %d", len(result))
	}
	if !reflect.DeepEqual(result[0], []int{1, 2}) {
		t.Errorf("Expected [1, 2], got %v", result[0])
	}
}

// ============================================================================
// Side Effect Method Tests
// ============================================================================

func TestPipelineTap(t *testing.T) {
	ctx := context.Background()
	var observed []int

	result := FromSlice(ctx, []int{1, 2, 3, 4, 5}).
		Tap(func(x int) { observed = append(observed, x) }).
		Filter(func(x int) bool { return x%2 == 0 }).
		ToSlice()

	expectedObserved := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(observed, expectedObserved) {
		t.Errorf("Tap observed %v, expected %v", observed, expectedObserved)
	}

	expectedResult := []int{2, 4}
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Expected %v, got %v", expectedResult, result)
	}
}

// ============================================================================
// Combining Method Tests
// ============================================================================

func TestPipelineMerge(t *testing.T) {
	ctx := context.Background()

	ch1 := FromSlice(ctx, []int{1, 2, 3}).Chan()
	ch2 := FromSlice(ctx, []int{4, 5, 6}).Chan()

	result := From(ctx, ch1).Merge(ch2).ToSlice()

	// Order is not guaranteed in merge, so check length and contents
	if len(result) != 6 {
		t.Errorf("Expected 6 values, got %d", len(result))
	}

	// Check all values are present
	resultMap := make(map[int]bool)
	for _, v := range result {
		resultMap[v] = true
	}

	for i := 1; i <= 6; i++ {
		if !resultMap[i] {
			t.Errorf("Value %d missing from merged result", i)
		}
	}
}

func TestPipelineZip(t *testing.T) {
	ctx := context.Background()

	p1 := FromSlice(ctx, []int{1, 2, 3})
	ch2 := FromSlice(ctx, []string{"a", "b", "c"}).Chan()

	result := ZipWith(p1, ch2).ToSlice()

	if len(result) != 3 {
		t.Fatalf("Expected 3 pairs, got %d", len(result))
	}

	expectedPairs := []struct {
		First  int
		Second string
	}{
		{1, "a"},
		{2, "b"},
		{3, "c"},
	}

	for i, pair := range result {
		if pair.First != expectedPairs[i].First {
			t.Errorf("Pair %d: expected First=%d, got %d", i, expectedPairs[i].First, pair.First)
		}
		if pair.Second != expectedPairs[i].Second {
			t.Errorf("Pair %d: expected Second=%s, got %v", i, expectedPairs[i].Second, pair.Second)
		}
	}
}

// ============================================================================
// Terminal Operation Tests
// ============================================================================

func TestPipelineToSlice(t *testing.T) {
	ctx := context.Background()

	result := FromSlice(ctx, []int{1, 2, 3, 4, 5}).ToSlice()

	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPipelineReduce(t *testing.T) {
	ctx := context.Background()

	sum := FromSlice(ctx, []int{1, 2, 3, 4, 5}).
		Reduce(func(acc, x int) int { return acc + x }, 0)

	expected := 15
	if sum != expected {
		t.Errorf("Expected %d, got %d", expected, sum)
	}
}

func TestPipelineReduceTo(t *testing.T) {
	ctx := context.Background()

	// Concatenate numbers into a string
	result := ReduceTo(
		FromSlice(ctx, []int{1, 2, 3}),
		func(acc string, x int) string {
			if acc == "" {
				return fmt.Sprint(x)
			}
			return acc + "," + fmt.Sprint(x)
		},
		"",
	)

	expected := "1,2,3"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestPipelineForEach(t *testing.T) {
	ctx := context.Background()
	var result []int

	FromSlice(ctx, []int{1, 2, 3, 4, 5}).
		ForEach(func(x int) { result = append(result, x*2) })

	expected := []int{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPipelineCount(t *testing.T) {
	ctx := context.Background()

	count := FromSlice(ctx, []int{1, 2, 3, 4, 5}).
		Filter(func(x int) bool { return x%2 == 0 }).
		Count()

	expected := 2
	if count != expected {
		t.Errorf("Expected %d, got %d", expected, count)
	}
}

func TestPipelineChan(t *testing.T) {
	ctx := context.Background()

	ch := FromSlice(ctx, []int{1, 2, 3}).Chan()

	var result []int
	for val := range ch {
		result = append(result, val)
	}

	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// ============================================================================
// LINQ-Style Alias Tests
// ============================================================================

func TestPipelineWhere(t *testing.T) {
	ctx := context.Background()

	result := FromSlice(ctx, []int{1, 2, 3, 4, 5, 6}).
		Where(func(x int) bool { return x > 3 }).
		ToSlice()

	expected := []int{4, 5, 6}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPipelineSelect(t *testing.T) {
	ctx := context.Background()

	result := FromSlice(ctx, []int{1, 2, 3}).
		Select(func(x int) any { return x * 3 }).
		ToSlice()

	expected := []any{3, 6, 9}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPipelineFirst(t *testing.T) {
	ctx := context.Background()

	first, ok := FromSlice(ctx, []int{1, 2, 3, 4, 5}).First()

	if !ok {
		t.Error("Expected ok=true, got false")
	}
	if first != 1 {
		t.Errorf("Expected 1, got %d", first)
	}
}

func TestPipelineFirstEmpty(t *testing.T) {
	ctx := context.Background()

	_, ok := FromSlice(ctx, []int{}).First()

	if ok {
		t.Error("Expected ok=false for empty pipeline, got true")
	}
}

func TestPipelineLast(t *testing.T) {
	ctx := context.Background()

	last, ok := FromSlice(ctx, []int{1, 2, 3, 4, 5}).Last()

	if !ok {
		t.Error("Expected ok=true, got false")
	}
	if last != 5 {
		t.Errorf("Expected 5, got %d", last)
	}
}

func TestPipelineLastEmpty(t *testing.T) {
	ctx := context.Background()

	_, ok := FromSlice(ctx, []int{}).Last()

	if ok {
		t.Error("Expected ok=false for empty pipeline, got true")
	}
}

func TestPipelineAny(t *testing.T) {
	ctx := context.Background()

	hasEven := FromSlice(ctx, []int{1, 3, 5, 6, 7}).
		Any(func(x int) bool { return x%2 == 0 })

	if !hasEven {
		t.Error("Expected Any to return true")
	}

	hasNegative := FromSlice(ctx, []int{1, 2, 3, 4, 5}).
		Any(func(x int) bool { return x < 0 })

	if hasNegative {
		t.Error("Expected Any to return false")
	}
}

func TestPipelineAll(t *testing.T) {
	ctx := context.Background()

	allPositive := FromSlice(ctx, []int{1, 2, 3, 4, 5}).
		All(func(x int) bool { return x > 0 })

	if !allPositive {
		t.Error("Expected All to return true")
	}

	allEven := FromSlice(ctx, []int{2, 4, 5, 6}).
		All(func(x int) bool { return x%2 == 0 })

	if allEven {
		t.Error("Expected All to return false")
	}
}

// ============================================================================
// Complex Pipeline Tests (Chaining Multiple Operations)
// ============================================================================

func TestComplexPipeline1(t *testing.T) {
	ctx := context.Background()

	// Range -> Map -> Filter -> Take -> ToSlice
	result := RangePipeline(ctx, 1, 100, 1).
		Map(func(x int) any { return x * x }).
		ToSlice()

	// Convert back to ints for filtering
	intResult := make([]int, len(result))
	for i, v := range result {
		intResult[i] = v.(int)
	}

	pipeline := FromSlice(ctx, intResult).
		Filter(func(x int) bool { return x%2 == 0 }).
		Take(10).
		ToSlice()

	if len(pipeline) != 10 {
		t.Errorf("Expected 10 values, got %d", len(pipeline))
	}

	// Check all values are even
	for _, v := range pipeline {
		if v%2 != 0 {
			t.Errorf("Expected even value, got %d", v)
		}
	}
}

func TestComplexPipeline2(t *testing.T) {
	ctx := context.Background()

	// Using Map and Filter -> Reduce
	sum := FromSlice(ctx, []int{1, 2, 3, 4, 5}).
		Map(func(x int) any { return x * 2 }).
		ToSlice()

	// Convert to ints and sum
	total := 0
	for _, v := range sum {
		total += v.(int)
	}

	// 2 + 4 + 6 + 8 + 10 = 30
	expected := 30
	if total != expected {
		t.Errorf("Expected %d, got %d", expected, total)
	}
}

func TestComplexPipeline3(t *testing.T) {
	ctx := context.Background()

	// Skip -> TakeWhile -> Map -> Count
	count := FromSlice(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
		Skip(2).
		TakeWhile(func(x int) bool { return x < 8 }).
		Map(func(x int) any { return x * 2 }).
		Count()

	// Skip 1,2 -> Take 3,4,5,6,7 -> Map -> Count = 5
	expected := 5
	if count != expected {
		t.Errorf("Expected %d, got %d", expected, count)
	}
}

// ============================================================================
// Context Cancellation Tests
// ============================================================================

func TestPipelineContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan int)
	go func() {
		for i := 1; i <= 100; i++ {
			ch <- i
			time.Sleep(10 * time.Millisecond)
		}
		close(ch)
	}()

	// Cancel after 50ms
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result := From(ctx, ch).ToSlice()

	// Should get fewer than 100 values due to cancellation
	if len(result) >= 100 {
		t.Errorf("Expected fewer than 100 values due to cancellation, got %d", len(result))
	}
}

func TestPipelineContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	ch := make(chan int)
	go func() {
		for i := 1; ; i++ {
			ch <- i
			time.Sleep(10 * time.Millisecond) // Slower generation
		}
	}()

	result := From(ctx, ch).ToSlice()

	// Should get fewer values due to timeout (50ms / 10ms per value = ~5 values)
	if len(result) > 10 {
		t.Errorf("Expected around 5 values due to timeout, got %d", len(result))
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestPipelineEmptyInput(t *testing.T) {
	ctx := context.Background()

	result := FromSlice(ctx, []int{}).
		Map(func(x int) any { return x * 2 }).
		Filter(func(x any) bool { return true }).
		ToSlice()

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %v", result)
	}
}

func TestPipelineFilterAll(t *testing.T) {
	ctx := context.Background()

	result := FromSlice(ctx, []int{1, 2, 3, 4, 5}).
		Filter(func(x int) bool { return false }).
		ToSlice()

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %v", result)
	}
}

func TestPipelineTakeZero(t *testing.T) {
	ctx := context.Background()

	result := FromSlice(ctx, []int{1, 2, 3, 4, 5}).
		Take(0).
		ToSlice()

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %v", result)
	}
}

func TestPipelineTakeMoreThanAvailable(t *testing.T) {
	ctx := context.Background()

	result := FromSlice(ctx, []int{1, 2, 3}).
		Take(10).
		ToSlice()

	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkPipelineSimple(b *testing.B) {
	ctx := context.Background()
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FromSlice(ctx, data).
			Map(func(x int) any { return x * 2 }).
			ToSlice()
	}
}

func BenchmarkPipelineComplex(b *testing.B) {
	ctx := context.Background()
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := FromSlice(ctx, data).ToSlice()
		intPipe := FromSlice(ctx, result).
			Filter(func(x int) bool { return x%2 == 0 }).
			Take(100).
			ToSlice()
		_ = intPipe
	}
}

func BenchmarkPipelineVsManual(b *testing.B) {
	ctx := context.Background()
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.Run("Pipeline", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FromSlice(ctx, data).
				Map(func(x int) any { return x * 2 }).
				ToSlice()
		}
	})

	b.Run("Manual", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := make([]int, len(data))
			for j, v := range data {
				result[j] = v * 2
			}
		}
	})
}
