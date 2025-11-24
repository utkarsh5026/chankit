# Core Concepts

Understanding the fundamental concepts behind chankit will help you build powerful channel-based applications.

## Pipelines

A **Pipeline** is the core abstraction in chankit. It represents a series of operations applied to a stream of values flowing through channels.

### Creating Pipelines

There are several ways to create a pipeline:

#### From a Range

```go
// Create a pipeline with numbers 1-10
pipeline := chankit.RangePipeline(ctx, 1, 11, 1)
```

#### From a Generator Function

```go
// Generate custom values
pipeline := chankit.GeneratePipeline(ctx, func(i int) any {
    return fmt.Sprintf("Item-%d", i)
})
```

#### From a Slice

```go
// Convert a slice to a pipeline
data := []int{1, 2, 3, 4, 5}
pipeline := chankit.FromSlicePipeline(ctx, data)
```

#### From an Existing Channel

```go
// Wrap an existing channel
ch := make(chan any)
pipeline := chankit.NewPipeline(ch)
```

### Pipeline Execution

Pipelines are **lazy** - operations don't execute until you call a terminal operation:

```go
// Nothing happens yet (lazy)
pipeline := chankit.RangePipeline(ctx, 1, 1000000, 1).
    Map(expensiveOperation).
    Filter(complexPredicate)

// Execution starts here (eager terminal operation)
result := pipeline.ToSlice()
```

<div class="bg-primary-900/20 border-l-4 border-primary-500 p-4 my-6 rounded">

**Key Insight:** Intermediate operations (Map, Filter, etc.) set up the pipeline structure. Terminal operations (ToSlice, ToChannel, ForEach) trigger execution.

</div>

## Operators

Operators are functions that transform or control the flow of data through a pipeline.

### Transformation Operators

Transform values as they flow through:

<div class="grid md:grid-cols-2 gap-4 my-6">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

#### Map
Transform each value

```go
.Map(func(x int) any {
    return x * 2
})
```

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

#### Filter
Select values by condition

```go
.Filter(func(x any) bool {
    return x.(int) > 10
})
```

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

#### FlatMap
Transform and flatten

```go
.FlatMap(func(x int) []any {
    return []any{x, x * 2}
})
```

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

#### Reduce
Aggregate to single value

```go
.Reduce(0, func(acc, val any) any {
    return acc.(int) + val.(int)
})
```

</div>

</div>

### Flow Control Operators

Control timing and rate of data flow:

<div class="grid md:grid-cols-2 gap-4 my-6">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

#### Throttle
Rate limit (drop excess)

```go
.Throttle(100 * time.Millisecond)
```

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

#### Debounce
Wait for silence

```go
.Debounce(500 * time.Millisecond)
```

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

#### Batch
Group by size or time

```go
.Batch(10, time.Second)
```

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

#### FixedInterval
Consistent pacing

```go
.FixedInterval(time.Second)
```

</div>

</div>

### Selection Operators

Select specific values from the stream:

<div class="grid md:grid-cols-2 gap-4 my-6">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

#### Take
Take first N values

```go
.Take(10)
```

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

#### Skip
Skip first N values

```go
.Skip(5)
```

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

#### TakeWhile
Take while condition true

```go
.TakeWhile(func(x any) bool {
    return x.(int) < 100
})
```

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

#### SkipWhile
Skip while condition true

```go
.SkipWhile(func(x any) bool {
    return x.(int) < 10
})
```

</div>

</div>

## Context Awareness

All chankit operations respect Go's context for cancellation and timeouts.

### Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())

// Start processing
go func() {
    result := chankit.RangePipeline(ctx, 1, 1000000, 1).
        Map(slowOperation).
        ToSlice()
    fmt.Println(result)
}()

// Cancel after 2 seconds
time.Sleep(2 * time.Second)
cancel() // Pipeline stops immediately
```

### Timeouts

```go
// Automatically stop after 5 seconds
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result := chankit.RangePipeline(ctx, 1, 1000000, 1).
    Map(expensiveOperation).
    ToSlice() // Stops at timeout
```

### Deadlines

```go
// Stop at specific time
deadline := time.Now().Add(1 * time.Hour)
ctx, cancel := context.WithDeadline(context.Background(), deadline)
defer cancel()

pipeline := chankit.RangePipeline(ctx, 1, 1000000, 1)
```

## Concurrency Model

### Goroutine Management

chankit handles goroutines automatically:

```go
// Each operator runs in its own goroutine
pipeline := chankit.RangePipeline(ctx, 1, 1001, 1).
    Map(transform1).    // Goroutine 1
    Filter(condition).  // Goroutine 2
    Map(transform2)     // Goroutine 3

// No manual goroutine management needed!
```

### Channel Cleanup

Channels are automatically closed when no longer needed:

```go
// No need to close channels manually
result := pipeline.ToSlice()
// All internal channels cleaned up automatically
```

### Backpressure

chankit handles backpressure naturally through Go's channel semantics:

```go
// Slow consumer naturally slows down the pipeline
pipeline.ForEach(func(value any) {
    time.Sleep(100 * time.Millisecond) // Slow processing
    fmt.Println(value)
})
```

## Type Safety with Generics

chankit uses Go generics for type-safe operations:

### Type Parameters

```go
// Pipeline[T] is generic over the value type
type Pipeline[T any] struct {
    ch <-chan T
}

// Strongly typed operations
func (p *Pipeline[int]) Map(f func(int) int) *Pipeline[int] {
    // ...
}
```

### Type Conversions

When working with `any`, explicit conversions are needed:

```go
result := chankit.RangePipeline(ctx, 1, 11, 1).
    Map(func(x int) any { return x * 2 }).
    Filter(func(x any) bool {
        // Type assertion needed
        val, ok := x.(int)
        return ok && val > 5
    }).
    ToSlice()
```

## Error Handling

### Propagating Errors

Use special error values in the stream:

```go
pipeline := chankit.GeneratePipeline(ctx, func(i int) any {
    result, err := riskyOperation(i)
    if err != nil {
        return err // Propagate error as value
    }
    return result
}).Filter(func(x any) bool {
    // Filter out errors or handle them
    _, isError := x.(error)
    return !isError
})
```

### Panic Recovery

Panics in operators are caught and don't crash the pipeline:

```go
pipeline := chankit.RangePipeline(ctx, 1, 11, 1).
    Map(func(x int) any {
        if x == 5 {
            panic("something went wrong")
        }
        return x * 2
    }) // Panic is caught, pipeline continues
```

## Performance Considerations

### Memory Usage

<div class="grid md:grid-cols-2 gap-4 my-6">

<div class="bg-green-900/20 border border-green-500/50 rounded-lg p-4">

#### ✅ Efficient

```go
// Streaming (low memory)
pipeline.ForEach(process)
```

```go
// Process in chunks
pipeline.Batch(100, time.Second)
```

</div>

<div class="bg-red-900/20 border border-red-500/50 rounded-lg p-4">

#### ❌ Inefficient

```go
// Loads everything (high memory)
result := pipeline.ToSlice()
```

```go
// No batching (overhead)
pipeline.Map(individualProcess)
```

</div>

</div>

### Goroutine Overhead

Each operator spawns a goroutine. For simple operations, consider:

```go
// Multiple operations, multiple goroutines
.Map(op1).Map(op2).Map(op3)

// Better: Combine operations
.Map(func(x int) any {
    return op3(op2(op1(x)))
})
```

## Best Practices

<div class="space-y-4 my-6">

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Always use context</strong>
<p class="text-dark-text-muted">Pass context to all pipelines for proper cancellation</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Prefer streaming over collecting</strong>
<p class="text-dark-text-muted">Use ForEach instead of ToSlice when possible</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Combine related operations</strong>
<p class="text-dark-text-muted">Reduce goroutine overhead by merging simple operations</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Use typed functions</strong>
<p class="text-dark-text-muted">Define transformation functions separately for reusability and testing</p>
</div>
</div>

</div>

## Next Steps

<div class="grid md:grid-cols-2 gap-4 my-8">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6 card-hover">

### 🚀 See It In Action

Explore real-world examples of chankit in use.

[View Examples →](/examples/data-pipeline)

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6 card-hover">

### 📖 Full API Reference

Explore all available operators and their options.

[API Reference →](/api/pipeline)

</div>

</div>

<style scoped>
.card-hover {
  transition: all 0.3s ease;
  cursor: pointer;
}

.card-hover:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(14, 165, 233, 0.15);
  border-color: #0ea5e9;
}
</style>
