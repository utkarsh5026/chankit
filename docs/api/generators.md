# Generators

Generator functions create pipelines from various sources.

## Range

Generate a sequence of integers.

```go
func Range(ctx context.Context, start, end, step int) <-chan int
func RangePipeline(ctx context.Context, start, end, step int) *Pipeline[int]
```

**Parameters:**
- `start` - Starting value (inclusive)
- `end` - Ending value (exclusive)
- `step` - Increment/decrement

**Examples:**
```go
// Numbers 1-10
chankit.RangePipeline(ctx, 1, 11, 1)

// Even numbers 0-100
chankit.RangePipeline(ctx, 0, 101, 2)

// Countdown 10 to 1
chankit.RangePipeline(ctx, 10, 0, -1)

// Just the channel
ch := chankit.Range(ctx, 1, 11, 1)
```

---

## Generate

Generate values using a custom function.

```go
func Generate(ctx context.Context, fn func(int) any) <-chan any
func GeneratePipeline(ctx context.Context, fn func(int) any) *Pipeline[any]
```

**Parameter:**
- `fn` - Generator function receiving index (0, 1, 2, ...)

**Examples:**
```go
// Generate formatted strings
chankit.GeneratePipeline(ctx, func(i int) any {
    return fmt.Sprintf("Item-%d", i)
})

// Generate timestamps
chankit.GeneratePipeline(ctx, func(i int) any {
    return time.Now().Add(time.Duration(i) * time.Hour)
}).Take(24) // Next 24 hours

// Generate random numbers
chankit.GeneratePipeline(ctx, func(i int) any {
    return rand.Intn(100)
}).Take(10)

// Fibonacci sequence
chankit.GeneratePipeline(ctx, func(i int) any {
    return fibonacci(i)
}).Take(10)
```

---

## Repeat

Repeat a value infinitely or N times.

```go
func Repeat(ctx context.Context, value any) <-chan any
func RepeatN(ctx context.Context, value any, n int) <-chan any
```

**Examples:**
```go
// Infinite stream of zeros
zeros := chankit.Repeat(ctx, 0)

// Repeat 10 times
tens := chankit.RepeatN(ctx, 10, 5) // [10, 10, 10, 10, 10]

// Heartbeat
heartbeat := chankit.Repeat(ctx, "ping").
    FixedInterval(time.Second).
    Take(60) // 60 pings

// Default values
defaults := chankit.RepeatN(ctx, DefaultConfig{}, 100)
```

---

## FromSlice

Convert a slice to a pipeline.

```go
func FromSlice[T any](ctx context.Context, slice []T) *Pipeline[T]
```

**Examples:**
```go
numbers := []int{1, 2, 3, 4, 5}
chankit.FromSlice(ctx, numbers)

users := []User{{ID: 1}, {ID: 2}}
chankit.FromSlice(ctx, users)

// Can chain immediately
result := chankit.FromSlice(ctx, data).
    Filter(isValid).
    Map(transform).
    ToSlice()
```

---

## From

Wrap an existing channel.

```go
func From[T any](ctx context.Context, ch <-chan T) *Pipeline[T]
```

**Examples:**
```go
// Existing channel
ch := make(chan int)
go produceValues(ch)

pipeline := chankit.From(ctx, ch)

// Combine with operations
result := chankit.From(ctx, eventChannel).
    Debounce(100 * time.Millisecond).
    Filter(isImportant).
    ToSlice()
```

---

## Real-World Examples

### Pagination Generator

```go
type Page struct {
    Number int
    Size   int
}

pages := chankit.GeneratePipeline(ctx, func(i int) any {
    return Page{
        Number: i + 1,
        Size:   100,
    }
}).Take(10) // Generate 10 pages

pages.ForEach(func(p Page) {
    data := fetchPage(p.Number, p.Size)
    processData(data)
})
```

### Test Data Generation

```go
type TestUser struct {
    ID    int
    Name  string
    Email string
}

testUsers := chankit.GeneratePipeline(ctx, func(i int) any {
    return TestUser{
        ID:    i + 1,
        Name:  fmt.Sprintf("User%d", i+1),
        Email: fmt.Sprintf("user%d@test.com", i+1),
    }
}).Take(1000)

// Insert into test database
testUsers.Batch(100, time.Second).
    ForEach(func(batch []TestUser) {
        db.BulkInsert(batch)
    })
```

### Time Series Data

```go
// Generate hourly timestamps for last 24 hours
now := time.Now()
timestamps := chankit.GeneratePipeline(ctx, func(i int) any {
    return now.Add(-time.Duration(24-i) * time.Hour)
}).Take(24)

// Fetch metrics for each timestamp
metrics := timestamps.Map(func(t any) any {
    timestamp := t.(time.Time)
    return fetchMetrics(timestamp)
})
```

### Retry with Backoff

```go
maxRetries := 5

retries := chankit.GeneratePipeline(ctx, func(i int) any {
    // Exponential backoff
    backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
    return backoff
}).Take(maxRetries)

retries.ForEach(func(delay any) {
    time.Sleep(delay.(time.Duration))
    if err := attemptOperation(); err == nil {
        return // Success
    }
})
```

## Generator Patterns

### Infinite Sequences

```go
// Infinite counter
counter := chankit.GeneratePipeline(ctx, func(i int) any {
    return i
})

// Use with Take to limit
first100 := counter.Take(100)
```

### Custom Iterators

```go
// Iterator over file lines
func FileLines(ctx context.Context, path string) *Pipeline[string] {
    ch := make(chan string)
    go func() {
        defer close(ch)
        file, _ := os.Open(path)
        defer file.Close()

        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            select {
            case <-ctx.Done():
                return
            case ch <- scanner.Text():
            }
        }
    }()
    return chankit.From(ctx, ch)
}

// Use like any pipeline
lines := FileLines(ctx, "data.txt").
    Filter(func(line string) bool {
        return len(line) > 0
    }).
    Take(100)
```

### Combining Generators

```go
// Generate ranges and concat
result := chankit.RangePipeline(ctx, 1, 6, 1).     // 1-5
    Concat(chankit.Range(ctx, 10, 16, 1)).        // + 10-15
    Concat(chankit.Range(ctx, 20, 26, 1)).        // + 20-25
    ToSlice()
// [1, 2, 3, 4, 5, 10, 11, 12, 13, 14, 15, 20, 21, 22, 23, 24, 25]
```

## Best Practices

<div class="space-y-4 my-6">

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Use Take with infinite generators</strong>
<p class="text-dark-text-muted">Always limit infinite sequences with Take to avoid infinite loops</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Respect context in generators</strong>
<p class="text-dark-text-muted">Custom generators should check ctx.Done() to support cancellation</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Close channels in generators</strong>
<p class="text-dark-text-muted">Always defer close(ch) in custom generator goroutines</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Use FromSlice for in-memory data</strong>
<p class="text-dark-text-muted">Most efficient for existing slices, use Generate for computed sequences</p>
</div>
</div>

</div>

## Related APIs

- [Pipeline](/api/pipeline) - Pipeline creation
- [Transformers](/api/transformers) - Transform generated values
- [Selectors](/api/selectors) - Limit generated sequences

<div class="flex gap-4 my-8">

[Next: Combiners →](/api/combiners){.custom-button}
[Previous: Transformers ←](/api/transformers){.custom-button-outline}

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
