---
layout: home

hero:
  name: "chankit"
  text: "Elegant Channel Operations for Go"
  tagline: "Transform the way you work with Go channels using modern, composable pipelines with type-safe functional programming patterns"
  image:
    src: /logo.svg
    alt: chankit
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: View Examples
      link: /examples/data-pipeline
    - theme: alt
      text: API Reference
      link: /api/pipeline

features:
  - icon: 🎛️
    title: Flow Control
    details: Control data flow with powerful timing operators - Throttle, Debounce, Batch, and FixedInterval for precise timing control
    link: /api/flow-control

  - icon: 🔄
    title: Transformations
    details: Transform and process data functionally with Map, Filter, Reduce, and FlatMap - all fully type-safe with generics
    link: /api/transformers

  - icon: 🏭
    title: Generators
    details: Create channels from various sources - Range, Generate, Repeat, and FromSlice for flexible data generation
    link: /api/generators

  - icon: 🔗
    title: Pipeline API
    details: Modern fluent interface for chaining operations - build complex data processing pipelines with ease
    link: /api/pipeline

  - icon: ⚡
    title: Context-Aware
    details: Built-in support for context cancellation and timeouts - respects Go's concurrency best practices
    link: /guide/core-concepts

  - icon: 🛡️
    title: Type-Safe
    details: Full generic support ensures compile-time type safety - catch errors before runtime
    link: /guide/core-concepts
---

<div class="max-w-6xl mx-auto px-4 py-16">

## Why chankit?

<div class="grid md:grid-cols-2 gap-8 my-8">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6 card-hover">

### ❌ Traditional Go
```go
// Verbose, nested, hard to maintain
ch := make(chan int)
go func() {
    for i := 1; i <= 100; i++ {
        ch <- i
    }
    close(ch)
}()

ch2 := make(chan int)
go func() {
    defer close(ch2)
    for v := range ch {
        ch2 <- v * v
    }
}()

result := []int{}
for v := range ch2 {
    if v%2 == 0 {
        result = append(result, v)
    }
}
```
</div>

<div class="bg-dark-bg-soft border border-primary-500 rounded-lg p-6 card-hover">

### ✅ With chankit
```go
// Clean, expressive, maintainable
result := chankit.RangePipeline(ctx, 1, 101, 1).
    Map(func(x int) any { return x * x }).
    Filter(func(x any) bool {
        return x.(int)%2 == 0
    }).
    ToSlice()
```

<div class="mt-4 p-4 bg-primary-900/20 border-l-4 border-primary-500 rounded">
<strong>🎯 70% less code, 100% more readable</strong>
</div>

</div>

</div>

## Quick Example

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/utkarsh5026/chankit"
)

func main() {
    ctx := context.Background()

    // Create a pipeline that processes numbers 1-10
    result := chankit.RangePipeline(ctx, 1, 11, 1).
        Map(func(x int) any { return x * 2 }).       // Double each number
        Filter(func(x any) bool { return x.(int) > 5 }). // Keep only > 5
        Take(5).                                     // Take first 5
        ToSlice()                                    // Collect results

    fmt.Println(result) // [6, 8, 10, 12, 14]
}
```

## Key Features

<div class="grid md:grid-cols-3 gap-6 my-12">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6 card-hover">

### 🚀 Performance
Zero-allocation pipelines with efficient channel operations. Built for high-throughput scenarios.

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6 card-hover">

### 🧪 Well-Tested
Comprehensive test coverage with race detection. Battle-tested for production use.

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6 card-hover">

### 📚 Rich API
50+ operators covering all common patterns - from basic transformations to advanced flow control.

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6 card-hover">

### 🎨 Composable
Chain operations naturally with fluent API. Build complex pipelines from simple building blocks.

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6 card-hover">

### 🔒 Goroutine-Safe
All operations are safe for concurrent use. No shared state, no race conditions.

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6 card-hover">

### 📖 Well-Documented
Extensive documentation with real-world examples. Learn by doing.

</div>

</div>

## Ready to Get Started?

<div class="flex gap-4 my-8">

[Get Started](/guide/getting-started){.custom-button}
[View Examples](/examples/data-pipeline){.custom-button-outline}
[API Reference](/api/pipeline){.custom-button-outline}

</div>

</div>

<style scoped>
.card-hover {
  transition: all 0.3s ease;
}

.card-hover:hover {
  transform: translateY(-4px);
  box-shadow: 0 10px 30px rgba(14, 165, 233, 0.2);
}

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
  box-shadow: 0 5px 15px rgba(14, 165, 233, 0.2);
}
</style>
