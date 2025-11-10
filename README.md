# chankit

A comprehensive Go toolkit for elegant and powerful channel operations. Built with generics for type safety and composability.

## Features

- **Flow Control**: Throttle, debounce, batch, and fixed-interval processing
- **Transformations**: Map, filter, and reduce operations on channels
- **Generators**: Create channels from functions, ranges, and repeated values
- **Conversions**: Seamless slice-to-channel and channel-to-slice conversions
- **Context-Aware**: All operations respect context cancellation
- **Type-Safe**: Full generic support for compile-time type safety
- **Production-Ready**: Comprehensive test coverage

## Installation

```bash
go get github.com/utkarsh5026/chankit
```

## Quick Start

```go
import (
    "context"
    "fmt"
    "time"
    "github.com/utkarsh5026/chankit/chankit"
)

func main() {
    ctx := context.Background()

    // Create a range of numbers
    numbers := chankit.Range(ctx, 1, 10, 1)

    // Transform with Map
    doubled := chankit.Map(ctx, numbers, func(x int) int {
        return x * 2
    })

    // Filter even numbers
    evens := chankit.Filter(ctx, doubled, func(x int) bool {
        return x%4 == 0
    })

    // Collect to slice
    result := chankit.ChanToSlice(ctx, evens)
    fmt.Println(result) // [4, 8, 12, 16]
}
```

## API Reference

### Flow Control

#### Throttle

Limits the rate of values by emitting only the most recent value within each time interval. Intermediate values are dropped.

```go
// Rate limit to 1 value per 100ms (drops intermediate values)
throttled := chankit.Throttle(ctx, input, 100*time.Millisecond)

// With buffered output
throttled := chankit.Throttle(ctx, input, 100*time.Millisecond,
    chankit.WithBuffer[int](10))
```

**Use Cases**: UI updates, high-frequency event handling, reducing API call frequency

#### Debounce

Emits values only after a period of silence. Timer resets on each new value. Perfect for handling bursts where only the final value matters.

```go
// Wait 200ms of silence before emitting
debounced := chankit.Debounce(ctx, input, 200*time.Millisecond)

// Search box example
searchResults := chankit.Debounce(ctx, userInput, 300*time.Millisecond)
```

**Use Cases**: Search boxes, form validation, resize events, scroll handling

**Key Difference from Throttle**:
- **Throttle**: Emits at fixed intervals (e.g., every 100ms)
- **Debounce**: Waits for silence before emitting (e.g., 100ms after last input)

#### FixedInterval

Processes every value at a fixed rate. Unlike Throttle, no values are dropped - they're queued and emitted with consistent spacing.

```go
// Emit one value every 100ms (all values preserved)
paced := chankit.FixedInterval(ctx, input, 100*time.Millisecond)
```

**Use Cases**: Rate-limited API calls, paced downloads, consistent processing rate

#### Batch

Groups values into batches based on size or timeout, whichever comes first.

```go
// Batch by size (3 items) or timeout (100ms)
batches := chankit.Batch(ctx, input, 3, 100*time.Millisecond)

// Process batches
for batch := range batches {
    fmt.Printf("Processing %d items\n", len(batch))
    // Process batch...
}
```

**Use Cases**: Bulk database inserts, batch API requests, log aggregation

### Transformations

#### Map

Transforms each value using a function. Can change the type.

```go
// Double values
doubled := chankit.Map(ctx, numbers, func(x int) int {
    return x * 2
})

// Convert types
strings := chankit.Map(ctx, numbers, func(x int) string {
    return fmt.Sprintf("num_%d", x)
})
```

#### Filter

Emits only values that satisfy a predicate.

```go
// Even numbers only
evens := chankit.Filter(ctx, numbers, func(x int) bool {
    return x%2 == 0
})

// Positive numbers
positive := chankit.Filter(ctx, numbers, func(x int) bool {
    return x > 0
})
```

#### Reduce

Aggregates all values into a single result. This is a blocking operation.

```go
// Sum all values
sum := chankit.Reduce(ctx, numbers, func(acc, x int) int {
    return acc + x
}, 0)

// Find maximum
max := chankit.Reduce(ctx, numbers, func(max, x int) int {
    if x > max {
        return x
    }
    return max
}, 0)

// Concatenate strings
result := chankit.Reduce(ctx, words, func(acc, word string) string {
    return acc + " " + word
}, "")
```

### Generators

#### Generate

Creates a channel from a generator function.

```go
// Fibonacci sequence
fib := chankit.Generate(ctx, func() (int, bool) {
    // Return (value, true) to emit, or (zero, false) to stop
    val := computeNext()
    return val, val < 1000
})

// With buffering
buffered := chankit.Generate(ctx, genFunc, chankit.WithBuffer[int](10))
```

#### Repeat

Infinitely repeats a value until context is cancelled.

```go
// Repeat "ping" forever
pings := chankit.Repeat(ctx, "ping")

// Use with context timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
heartbeats := chankit.Repeat(ctx, "heartbeat")
```

#### Range

Generates a sequence of numbers.

```go
// 0 to 9
range1 := chankit.Range(ctx, 0, 10, 1)

// Countdown: 10 to 1
countdown := chankit.Range(ctx, 10, 0, -1)

// Even numbers: 0, 2, 4, 6, 8
evens := chankit.Range(ctx, 0, 10, 2)

// Works with float64
decimals := chankit.Range(ctx, 0.0, 1.0, 0.1)
```

### Conversions

#### SliceToChan

Converts a slice to a channel.

```go
slice := []int{1, 2, 3, 4, 5}

// Unbuffered
ch := chankit.SliceToChan(ctx, slice)

// Buffered with custom size
ch := chankit.SliceToChan(ctx, slice, chankit.WithBuffer[int](10))

// Auto-sized buffer (matches slice length)
ch := chankit.SliceToChan(ctx, slice, chankit.WithBufferAuto[int]())
```

#### ChanToSlice

Collects all values from a channel into a slice. Blocks until channel closes.

```go
// Default capacity
slice := chankit.ChanToSlice(ctx, ch)

// Pre-allocate for better performance
slice := chankit.ChanToSlice(ctx, ch, chankit.WithCapacity[int](100))
```

## Configuration Options

### Channel Buffering

Control output channel buffer sizes for better performance:

```go
// Unbuffered (default)
ch := chankit.Map(ctx, input, fn)

// Custom buffer size
ch := chankit.Map(ctx, input, fn, chankit.WithBuffer[int](50))

// Auto-sized (for SliceToChan only)
ch := chankit.SliceToChan(ctx, slice, chankit.WithBufferAuto[int]())
```

### Slice Pre-allocation

Optimize memory allocation when collecting channels:

```go
// Default (zero capacity)
slice := chankit.ChanToSlice(ctx, ch)

// Pre-allocate if you know approximate size
slice := chankit.ChanToSlice(ctx, ch, chankit.WithCapacity[int](1000))
```

## Patterns and Examples

### Pipeline Pattern

Chain operations for elegant data processing:

```go
ctx := context.Background()

// Generate -> Transform -> Filter -> Collect
result := chankit.ChanToSlice(ctx,
    chankit.Filter(ctx,
        chankit.Map(ctx,
            chankit.Range(ctx, 1, 100, 1),
            func(x int) int { return x * x },
        ),
        func(x int) bool { return x%2 == 0 },
    ),
)
```

### Debounced Search

Implement efficient search-as-you-type:

```go
func handleSearch(ctx context.Context, userInput <-chan string) <-chan []Result {
    // Debounce user input (wait 300ms after typing stops)
    debounced := chankit.Debounce(ctx, userInput, 300*time.Millisecond)

    // Map to search results
    return chankit.Map(ctx, debounced, func(query string) []Result {
        return performSearch(query)
    })
}
```

### Rate-Limited API Calls

Process items with rate limiting:

```go
func processItems(ctx context.Context, items []Item) {
    // Convert to channel
    itemChan := chankit.SliceToChan(ctx, items)

    // Process at most 10 items per second
    paced := chankit.FixedInterval(ctx, itemChan, 100*time.Millisecond)

    // Make API calls
    for item := range paced {
        apiClient.Process(item)
    }
}
```

### Batch Processing

Efficiently batch database operations:

```go
func saveRecords(ctx context.Context, records <-chan Record) {
    // Batch 100 records or every 5 seconds
    batches := chankit.Batch(ctx, records, 100, 5*time.Second)

    for batch := range batches {
        db.BulkInsert(batch)
    }
}
```

### Throttled UI Updates

Prevent excessive UI redraws:

```go
func updateUI(ctx context.Context, events <-chan Event) {
    // Update at most every 16ms (~60 FPS)
    throttled := chankit.Throttle(ctx, events, 16*time.Millisecond)

    for event := range throttled {
        renderUI(event)
    }
}
```

## Context Cancellation

All operations respect context cancellation:

```go
// Timeout after 5 seconds
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Will stop after timeout
result := chankit.ChanToSlice(ctx, longRunningChannel)

// Manual cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(1 * time.Second)
    cancel() // Stop all operations
}()
```

## Performance Tips

1. **Use buffered channels** for high-throughput scenarios:
   ```go
   chankit.Map(ctx, input, fn, chankit.WithBuffer[int](100))
   ```

2. **Pre-allocate slices** when collecting known sizes:
   ```go
   chankit.ChanToSlice(ctx, ch, chankit.WithCapacity[int](expectedSize))
   ```

3. **Choose the right flow control**:
   - `Throttle`: Drop intermediate values for rate limiting
   - `Debounce`: Wait for activity to stop
   - `FixedInterval`: Preserve all values with consistent spacing
   - `Batch`: Group multiple values for bulk processing

## Testing

The package includes comprehensive test coverage:

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestDebounce ./chankit
```

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Comparison with Standard Library

While Go's standard library provides excellent channel primitives, `chankit` adds:

- **Higher-level abstractions**: Map, Filter, Reduce for functional-style programming
- **Flow control utilities**: Throttle, Debounce, Batch for common patterns
- **Type-safe generics**: Compile-time type checking across all operations
- **Composability**: Easy chaining of operations for complex pipelines
- **Production-tested**: Comprehensive test coverage for reliability
