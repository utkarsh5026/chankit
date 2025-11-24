# Transformers

Transformation operators modify, filter, and aggregate values flowing through a pipeline.

## Map

Transform each value to a new value.

```go
func (p *Pipeline[T]) Map(fn func(T) any) *Pipeline[any]
```

**Example:**
```go
// Square each number
pipeline.Map(func(x int) any { return x * x })

// Extract field
users.Map(func(u User) any { return u.Email })

// Complex transformation
data.Map(func(d Data) any {
    return ProcessedData{
        ID:    d.ID,
        Value: transform(d.Value),
    }
})
```

---

## MapTo (Type-Safe)

Transform with explicit output type.

```go
func MapTo[T, R any](p *Pipeline[T], fn func(T) R) *Pipeline[R]
```

**Example:**
```go
// Type-safe int to string
strings := chankit.MapTo(numbers, func(x int) string {
    return fmt.Sprintf("Number: %d", x)
})
```

---

## Filter

Keep only values that match a predicate.

```go
func (p *Pipeline[T]) Filter(fn func(T) bool) *Pipeline[T]
```

**Example:**
```go
// Keep even numbers
pipeline.Filter(func(x int) bool { return x%2 == 0 })

// Filter active users
users.Filter(func(u User) bool { return u.Active })

// Complex condition
data.Filter(func(d Data) bool {
    return d.Value > 100 && d.Status == "valid"
})
```

---

## FlatMap

Transform each value to multiple values (one-to-many).

```go
func (p *Pipeline[T]) FlatMap(fn func(T) <-chan any) *Pipeline[any]
```

**Example:**
```go
// Split strings into words
lines.FlatMap(func(line string) <-chan any {
    ch := make(chan any)
    go func() {
        defer close(ch)
        for _, word := range strings.Split(line, " ") {
            ch <- word
        }
    }()
    return ch
})

// Expand ranges
ranges.FlatMap(func(r Range) <-chan any {
    return chankit.Range(ctx, r.Start, r.End, 1)
})
```

---

## Reduce

Aggregate all values into a single result.

```go
func (p *Pipeline[T]) Reduce(initial any, fn func(acc, val any) any) any
```

**Example:**
```go
// Sum all numbers
sum := pipeline.Reduce(0, func(acc, val any) any {
    return acc.(int) + val.(int)
})

// Concatenate strings
result := strings.Reduce("", func(acc, val any) any {
    return acc.(string) + val.(string)
})

// Build map
counts := words.Reduce(make(map[string]int), func(acc, val any) any {
    m := acc.(map[string]int)
    word := val.(string)
    m[word]++
    return m
})
```

---

## Tap

Observe values without modifying them (side effects).

```go
func (p *Pipeline[T]) Tap(fn func(T)) *Pipeline[T]
```

**Example:**
```go
// Log values
pipeline.Tap(func(x int) {
    log.Printf("Processing: %d", x)
})

// Metrics
pipeline.Tap(func(x int) {
    metrics.Increment("processed_count")
})

// Debug
pipeline.Tap(func(x int) {
    if x > 1000 {
        fmt.Printf("Large value: %d\n", x)
    }
})
```

---

## Transformation Patterns

### Chaining Transformations

```go
result := pipeline.
    Filter(isValid).           // Remove invalid
    Map(normalize).            // Normalize format
    Filter(isNonEmpty).        // Remove empty
    Map(enrich).               // Add metadata
    ToSlice()
```

### Map + Filter Pattern

```go
// Transform then filter
results := data.
    Map(func(d Data) any {
        return process(d)
    }).
    Filter(func(r any) bool {
        return r.(Result).Success
    })
```

### Conditional Transformation

```go
results := pipeline.Map(func(x int) any {
    switch {
    case x > 100:
        return x * 2
    case x > 50:
        return x * 1.5
    default:
        return x
    }
})
```

## Real-World Examples

### Data Processing Pipeline

```go
type Record struct {
    ID    int
    Name  string
    Email string
    Age   int
}

processed := chankit.FromSlice(ctx, records).
    Filter(func(r Record) bool {
        return r.Age >= 18                  // Adults only
    }).
    Filter(func(r Record) bool {
        return isValidEmail(r.Email)        // Valid emails
    }).
    Map(func(r Record) any {
        return struct {
            ID    int
            Email string
        }{
            ID:    r.ID,
            Email: strings.ToLower(r.Email), // Normalize
        }
    }).
    ToSlice()
```

### Text Processing

```go
// Count word frequency
wordCounts := chankit.FromSlice(ctx, lines).
    FlatMap(func(line string) <-chan any {
        // Split into words
        ch := make(chan any)
        go func() {
            defer close(ch)
            words := strings.Fields(line)
            for _, word := range words {
                ch <- strings.ToLower(word)
            }
        }()
        return ch
    }).
    Reduce(make(map[string]int), func(acc, val any) any {
        counts := acc.(map[string]int)
        word := val.(string)
        counts[word]++
        return counts
    })

fmt.Println(wordCounts) // {"hello": 5, "world": 3, ...}
```

### Log Analysis

```go
type LogEntry struct {
    Timestamp time.Time
    Level     string
    Message   string
}

errorLogs := chankit.FromSlice(ctx, logs).
    Filter(func(log LogEntry) bool {
        return log.Level == "ERROR"          // Errors only
    }).
    Tap(func(log LogEntry) {
        // Send alert for critical errors
        if strings.Contains(log.Message, "CRITICAL") {
            sendAlert(log)
        }
    }).
    Map(func(log LogEntry) any {
        // Summarize
        return fmt.Sprintf("[%s] %s",
            log.Timestamp.Format(time.RFC3339),
            log.Message)
    }).
    ToSlice()
```

## Performance Considerations

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-green-900/20 border border-green-500/50 rounded-lg p-6">

### ✅ Efficient

```go
// Filter early
pipeline.
    Filter(cheapCheck).    // Fast filter first
    Map(expensiveOp)       // Then expensive op
```

```go
// Combine operations
pipeline.Map(func(x int) any {
    // Do all transformations at once
    return transform3(transform2(transform1(x)))
})
```

</div>

<div class="bg-red-900/20 border border-red-500/50 rounded-lg p-6">

### ❌ Inefficient

```go
// Map then filter (wastes work)
pipeline.
    Map(expensiveOp).      // Expensive op on all
    Filter(check)          // Then filter
```

```go
// Multiple Maps (more goroutines)
pipeline.
    Map(transform1).
    Map(transform2).
    Map(transform3)
```

</div>

</div>

## Best Practices

<div class="space-y-4 my-6">

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Filter before Map</strong>
<p class="text-dark-text-muted">Apply cheap filters early to reduce work in expensive transformations</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Keep functions pure</strong>
<p class="text-dark-text-muted">Map and Filter functions should have no side effects. Use Tap for side effects</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Combine simple transformations</strong>
<p class="text-dark-text-muted">Merge multiple Maps into one to reduce goroutine overhead</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Use Tap for debugging</strong>
<p class="text-dark-text-muted">Insert Tap operations during development to inspect values</p>
</div>
</div>

</div>

## Related APIs

- [Pipeline](/api/pipeline) - Creating pipelines
- [Selectors](/api/selectors) - Take, Skip operations
- [Flow Control](/api/flow-control) - Throttle, Debounce

<div class="flex gap-4 my-8">

[Next: Generators →](/api/generators){.custom-button}
[Previous: Flow Control ←](/api/flow-control){.custom-button-outline}

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
