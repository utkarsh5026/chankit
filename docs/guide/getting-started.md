# Getting Started

Get up and running with chankit in minutes. This guide will walk you through installation, basic concepts, and your first pipeline.

## Installation

Install chankit using Go modules:

```bash
go get github.com/utkarsh5026/chankit
```

Import it in your Go code:

```go
import "github.com/utkarsh5026/chankit/chankit"
```

::: tip Requirements
- Go 1.24 or later (for generics support)
- Go modules enabled
:::

## Your First Pipeline

Let's create a simple pipeline that processes numbers:

```go
package main

import (
    "context"
    "fmt"
    "github.com/utkarsh5026/chankit/chankit"
)

func main() {
    ctx := context.Background()

    // Create a pipeline: numbers 1-10
    result := chankit.RangePipeline(ctx, 1, 11, 1).
        Map(func(x int) any { return x * 2 }).       // Double each number
        Filter(func(x any) bool { return x.(int) > 5 }). // Keep only > 5
        Take(5).                                     // Take first 5
        ToSlice()                                    // Collect results

    fmt.Println(result) // [6, 8, 10, 12, 14]
}
```

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 my-6 rounded">

### What's Happening Here?

1. **RangePipeline** creates a channel with numbers 1-10
2. **Map** transforms each value (multiplies by 2)
3. **Filter** keeps only values greater than 5
4. **Take** limits the results to 5 items
5. **ToSlice** collects all values into a slice

</div>

## Basic Concepts

### Pipelines

A pipeline is a series of channel operations chained together. Each operation processes values and passes them to the next:

```go
chankit.RangePipeline(ctx, 1, 101, 1).
    Map(squareFunc).        // Transform values
    Filter(isEvenFunc).     // Select values
    Take(10).               // Limit results
    ToSlice()               // Collect output
```

### Context Support

All pipelines are context-aware, allowing for cancellation and timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result := chankit.RangePipeline(ctx, 1, 1000000, 1).
    Map(expensiveOperation).
    ToSlice() // Will stop if timeout is reached
```

### Type Safety

chankit uses Go generics for type safety:

```go
// Type-safe transformations
numbers := chankit.RangePipeline(ctx, 1, 11, 1)

// Map with type conversion
doubled := numbers.Map(func(x int) any {
    return x * 2
})
```

## Common Patterns

### Transform and Filter

Process and filter data in one pipeline:

```go
result := chankit.RangePipeline(ctx, 1, 101, 1).
    Map(func(x int) any { return x * x }).           // Square
    Filter(func(x any) bool { return x.(int)%2 == 0 }). // Even only
    ToSlice()
```

### Generate Custom Data

Create channels from custom generator functions:

```go
pipeline := chankit.GeneratePipeline(ctx, func(i int) any {
    return fmt.Sprintf("Item-%d", i)
}).Take(10)
```

### Combine Multiple Sources

Merge data from multiple channels:

```go
ch1 := chankit.Range(ctx, 1, 6, 1)
ch2 := chankit.Range(ctx, 10, 16, 1)
merged := chankit.Merge(ctx, ch1, ch2)
```

## Next Steps

Now that you understand the basics, explore:

- [Why chankit?](/guide/why-chankit) - Learn about the benefits and philosophy
- [Core Concepts](/guide/core-concepts) - Deep dive into pipelines and operators
- [Examples](/examples/data-pipeline) - See real-world use cases
- [API Reference](/api/pipeline) - Explore all available operators

<div class="grid md:grid-cols-2 gap-4 my-8">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6 card-hover">

### 📚 Learn More

Understand the concepts behind chankit and why it matters.

[Why chankit? →](/guide/why-chankit)

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6 card-hover">

### 🚀 See Examples

Jump into real-world examples and use cases.

[View Examples →](/examples/data-pipeline)

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
