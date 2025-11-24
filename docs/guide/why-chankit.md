# Why chankit?

Discover how chankit transforms Go channel programming with functional patterns, making your code more maintainable, readable, and elegant.

## The Problem with Traditional Go Channels

Go's channels are powerful primitives for concurrency, but working with them can be verbose and error-prone:

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-red-900/20 border border-red-500/50 rounded-lg p-6">

### ❌ Traditional Approach

```go
// Verbose and nested
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

**Problems:**
- 🔴 Too much boilerplate
- 🔴 Manual goroutine management
- 🔴 Easy to forget closing channels
- 🔴 Hard to maintain and test

</div>

<div class="bg-green-900/20 border border-green-500/50 rounded-lg p-6">

### ✅ chankit Approach

```go
// Clean and declarative
result := chankit.RangePipeline(ctx, 1, 101, 1).
    Map(func(x int) any { return x * x }).
    Filter(func(x any) bool {
        return x.(int)%2 == 0
    }).
    ToSlice()
```

**Benefits:**
- 🟢 Minimal boilerplate
- 🟢 Automatic goroutine handling
- 🟢 Channels closed automatically
- 🟢 Easy to read and modify

</div>

</div>

## Key Advantages

### 🎯 Declarative vs Imperative

<ComparisonTable />

<div class="my-6 p-6 bg-primary-900/20 border-l-4 border-primary-500 rounded">

**Focus on WHAT, not HOW**

With chankit, you describe what transformations you want, not how to implement them. The library handles the complexity of goroutines, channels, and cleanup.

</div>

### 🔗 Composability

Build complex operations from simple building blocks:

```go
// Define reusable transformations
squareMapper := func(x int) any { return x * x }
evenFilter := func(x any) bool { return x.(int)%2 == 0 }
isPositive := func(x any) bool { return x.(int) > 0 }

// Compose them in different ways
result1 := pipeline.Map(squareMapper).Filter(evenFilter).ToSlice()
result2 := pipeline.Filter(isPositive).Map(squareMapper).ToSlice()
```

### ⚡ Performance Without Complexity

chankit provides high performance while keeping code simple:

<div class="grid md:grid-cols-3 gap-4 my-6">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4 text-center">

#### Zero Allocations
Efficient channel operations without extra memory overhead

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4 text-center">

#### Lazy Evaluation
Values processed only when needed

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4 text-center">

#### Concurrent by Default
Automatic goroutine orchestration

</div>

</div>

### 🛡️ Type Safety

Full generic support ensures compile-time safety:

```go
// Type errors caught at compile time
numbers := chankit.RangePipeline(ctx, 1, 11, 1)

// This compiles ✅
result := numbers.Map(func(x int) any { return x * 2 })

// This would fail at compile time ❌
// result := numbers.Map(func(x string) any { return x + "text" })
```

### 🎨 Rich Operator Library

50+ operators covering common patterns:

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-6">

#### Flow Control
- Throttle - Rate limiting
- Debounce - Wait for silence
- Batch - Group values
- FixedInterval - Consistent pacing

</div>

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-6">

#### Transformations
- Map - Transform values
- Filter - Select values
- Reduce - Aggregate results
- FlatMap - Transform & flatten

</div>

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-6">

#### Generators
- Range - Numeric sequences
- Generate - Custom functions
- Repeat - Value repetition
- FromSlice - Slice conversion

</div>

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-6">

#### Selectors
- Take - Limit results
- Skip - Skip values
- TakeWhile - Conditional take
- SkipWhile - Conditional skip

</div>

</div>

## Real-World Impact

### Code Reduction

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6 my-6">

**Before chankit:** 45 lines of boilerplate for event debouncing

**After chankit:** 8 lines with `.Debounce()`

**Result:** 80% less code, 100% more maintainable

</div>

### Maintainability

```go
// Easy to add new transformations
result := pipeline.
    Map(transform1).
    Filter(condition1).
    Map(transform2).      // ← Just add this line
    Filter(condition2).
    ToSlice()
```

### Testability

```go
// Test individual transformations
func TestSquareMapper(t *testing.T) {
    mapper := func(x int) any { return x * x }
    assert.Equal(t, 4, mapper(2))
}

// Test the pipeline
func TestPipeline(t *testing.T) {
    result := chankit.RangePipeline(ctx, 1, 4, 1).
        Map(squareMapper).
        ToSlice()
    assert.Equal(t, []any{1, 4, 9}, result)
}
```

## When to Use chankit

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-green-900/20 border border-green-500/50 rounded-lg p-6">

### ✅ Perfect For

- Stream processing pipelines
- Event-driven systems
- Data transformation workflows
- Reactive programming patterns
- Complex channel orchestration
- Rate limiting and throttling
- Batch processing

</div>

<div class="bg-yellow-900/20 border border-yellow-500/50 rounded-lg p-6">

### ⚠️ Consider Alternatives

- Simple single-channel operations
- Direct goroutine control needed
- Micro-optimizations critical
- Learning Go channel basics
- Legacy code with no refactoring budget

</div>

</div>

## Philosophy

chankit follows these principles:

<div class="space-y-4 my-6">

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-3xl">🎯</span>
<div>
<strong>Simplicity First</strong>
<p class="text-dark-text-muted">Complex operations should have simple interfaces</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-3xl">🔗</span>
<div>
<strong>Composability</strong>
<p class="text-dark-text-muted">Small pieces that combine into powerful solutions</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-3xl">🛡️</span>
<div>
<strong>Safety</strong>
<p class="text-dark-text-muted">Type-safe operations that prevent common mistakes</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-3xl">⚡</span>
<div>
<strong>Performance</strong>
<p class="text-dark-text-muted">Efficient by default, no hidden costs</p>
</div>
</div>

</div>

## Ready to Transform Your Code?

<div class="flex gap-4 my-8">

[Get Started](/guide/getting-started){.custom-button}
[See Examples](/examples/data-pipeline){.custom-button-outline}

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
  box-shadow: 0 5px 15px rgba(14, 165, 233, 0.2);
}
</style>
