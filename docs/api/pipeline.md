# Pipeline API

The Pipeline is the core abstraction in chankit. It represents a stream of values flowing through a series of operations.

## Creating Pipelines

### RangePipeline

Create a pipeline from a numeric range.

```go
func RangePipeline(ctx context.Context, start, end, step int) *Pipeline[int]
```

**Parameters:**
- `ctx` - Context for cancellation
- `start` - Starting value (inclusive)
- `end` - Ending value (exclusive)
- `step` - Step increment

**Example:**
```go
// Create pipeline with numbers 1-100
pipeline := chankit.RangePipeline(ctx, 1, 101, 1)

// Create even numbers 0-100
evens := chankit.RangePipeline(ctx, 0, 101, 2)

// Countdown from 10 to 1
countdown := chankit.RangePipeline(ctx, 10, 0, -1)
```

---

### FromSlice

Create a pipeline from a slice.

```go
func FromSlice[T any](ctx context.Context, slice []T) *Pipeline[T]
```

**Parameters:**
- `ctx` - Context for cancellation
- `slice` - Source slice

**Example:**
```go
numbers := []int{1, 2, 3, 4, 5}
pipeline := chankit.FromSlice(ctx, numbers)

users := []User{{ID: 1}, {ID: 2}}
userPipeline := chankit.FromSlice(ctx, users)
```

---

### From

Create a pipeline from an existing channel.

```go
func From[T any](ctx context.Context, ch <-chan T) *Pipeline[T]
```

**Parameters:**
- `ctx` - Context for cancellation
- `ch` - Existing channel

**Example:**
```go
ch := make(chan int)
go func() {
    defer close(ch)
    for i := 0; i < 10; i++ {
        ch <- i
    }
}()

pipeline := chankit.From(ctx, ch)
```

---

### NewPipeline

Create an empty pipeline (advanced usage).

```go
func NewPipeline[T any](ch <-chan T) *Pipeline[T]
```

**Parameters:**
- `ch` - Channel to wrap

**Example:**
```go
ch := make(chan int, 10)
pipeline := chankit.NewPipeline(ch)
```

---

## Pipeline Methods

### Chan

Get the underlying channel from a pipeline.

```go
func (p *Pipeline[T]) Chan() <-chan T
```

**Returns:** Read-only channel

**Example:**
```go
pipeline := chankit.RangePipeline(ctx, 1, 11, 1)
ch := pipeline.Chan()

for value := range ch {
    fmt.Println(value)
}
```

---

## Chaining Operations

Pipelines support method chaining for composing operations:

```go
result := chankit.RangePipeline(ctx, 1, 101, 1).
    Filter(func(x int) bool { return x%2 == 0 }).
    Map(func(x int) any { return x * x }).
    Take(10).
    ToSlice()
```

## Pipeline Lifecycle

<div class="space-y-4 my-6">

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### 1. Creation
Pipeline is created from a source (range, slice, channel, generator).

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### 2. Transformation
Intermediate operations (Map, Filter, etc.) build the pipeline structure.

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### 3. Terminal Operation
A terminal operation (ToSlice, ForEach, etc.) triggers execution.

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### 4. Cleanup
Channels are automatically closed and goroutines cleaned up.

</div>

</div>

## Context Support

All pipeline creation functions accept a context:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
pipeline := chankit.RangePipeline(ctx, 1, 1000000, 1)

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
pipeline := chankit.FromSlice(ctx, data)

// Cancel to stop pipeline
go func() {
    time.Sleep(2 * time.Second)
    cancel() // All operations stop
}()
```

## Best Practices

<div class="space-y-4 my-6">

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Always pass context</strong>
<p class="text-dark-text-muted">Never use context.Background() in production - use request context or create appropriate timeouts</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Prefer FromSlice for small datasets</strong>
<p class="text-dark-text-muted">Use FromSlice for in-memory data, From for existing channels</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Chain operations naturally</strong>
<p class="text-dark-text-muted">Build pipelines in a single chain for readability</p>
</div>
</div>

</div>

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/utkarsh5026/chankit/chankit"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Create pipeline from range
    result := chankit.RangePipeline(ctx, 1, 101, 1).
        Filter(func(x int) bool {
            return x%2 == 0
        }).
        Map(func(x int) any {
            return x * x
        }).
        Take(10).
        ToSlice()

    fmt.Printf("Result: %v\n", result)
    // Output: [4 16 36 64 100 144 196 256 324 400]
}
```

## Related APIs

- [Transformers](/api/transformers) - Map, Filter, FlatMap, Reduce
- [Flow Control](/api/flow-control) - Throttle, Debounce, Batch
- [Selectors](/api/selectors) - Take, Skip, First, Last

<div class="flex gap-4 my-8">

[Next: Flow Control →](/api/flow-control){.custom-button}
[Back to Guide](/guide/core-concepts){.custom-button-outline}

</div>

<style scoped>
.custom-button {
  display: inline-flex;
  align-items: center;
  padding: 0.75rem 1.5rem;
  border-radius: 0.5rem;
  font-weight: 600;
  background: linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%);
  color: white;
  text-decoration: none;
  transition: all 0.2s;
}

.custom-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 25px rgba(14, 165, 233, 0.3);
}

.custom-button-outline {
  display: inline-flex;
  align-items: center;
  padding: 0.75rem 1.5rem;
  border-radius: 0.5rem;
  font-weight: 600;
  border: 2px solid #0ea5e9;
  color: #38bdf8;
  text-decoration: none;
  transition: all 0.2s;
}

.custom-button-outline:hover {
  background: rgba(14, 165, 233, 0.1);
}
</style>
