# Selectors

Selector operators choose specific values from a stream based on position or condition.

## Take

Take first N values from the stream.

```go
func (p *Pipeline[T]) Take(n int) *Pipeline[T]
```

**Example:**
```go
// Take first 10
pipeline.Take(10)

// Pagination: first page
chankit.RangePipeline(ctx, 1, 1001, 1).
    Take(20) // First 20 items
```

---

## Skip

Skip first N values, emit the rest.

```go
func (p *Pipeline[T]) Skip(n int) *Pipeline[T]
```

**Example:**
```go
// Skip first 10
pipeline.Skip(10)

// Pagination: second page
chankit.RangePipeline(ctx, 1, 1001, 1).
    Skip(20).    // Skip first page
    Take(20)     // Get second page
```

---

## TakeWhile

Take values while a condition is true, stop when false.

```go
func (p *Pipeline[T]) TakeWhile(fn func(T) bool) *Pipeline[T]
```

**Example:**
```go
// Take while less than 100
pipeline.TakeWhile(func(x int) bool {
    return x < 100
})

// Take until first error
logs.TakeWhile(func(log LogEntry) bool {
    return log.Level != "ERROR"
})
```

---

## SkipWhile

Skip values while a condition is true, then emit rest.

```go
func (p *Pipeline[T]) SkipWhile(fn func(T) bool) *Pipeline[T]
```

**Example:**
```go
// Skip negatives, keep rest
pipeline.SkipWhile(func(x int) bool {
    return x < 0
})

// Skip until first valid
data.SkipWhile(func(d Data) bool {
    return !d.IsValid
})
```

---

## First

Get the first value (terminal operation).

```go
func (p *Pipeline[T]) First() (T, bool)
```

**Returns:** First value and `true`, or zero value and `false` if empty

**Example:**
```go
value, ok := pipeline.First()
if ok {
    fmt.Printf("First value: %v\n", value)
}
```

---

## Last

Get the last value (terminal operation).

```go
func (p *Pipeline[T]) Last() (T, bool)
```

**Returns:** Last value and `true`, or zero value and `false` if empty

**Example:**
```go
value, ok := pipeline.Last()
if ok {
    fmt.Printf("Last value: %v\n", value)
}
```

---

## Any

Check if any value matches a predicate (terminal operation).

```go
func (p *Pipeline[T]) Any(fn func(T) bool) bool
```

**Example:**
```go
hasNegative := pipeline.Any(func(x int) bool {
    return x < 0
})

hasError := logs.Any(func(log LogEntry) bool {
    return log.Level == "ERROR"
})
```

---

## All

Check if all values match a predicate (terminal operation).

```go
func (p *Pipeline[T]) All(fn func(T) bool) bool
```

**Example:**
```go
allPositive := pipeline.All(func(x int) bool {
    return x > 0
})

allValid := data.All(func(d Data) bool {
    return d.IsValid
})
```

---

## Count

Count total values (terminal operation).

```go
func (p *Pipeline[T]) Count() int
```

**Example:**
```go
total := pipeline.Count()
fmt.Printf("Total items: %d\n", total)

errorCount := logs.
    Filter(func(log LogEntry) bool {
        return log.Level == "ERROR"
    }).
    Count()
```

---

## Real-World Examples

### Pagination

```go
const pageSize = 20

func getPage(data []int, page int) []int {
    return chankit.FromSlice(ctx, data).
        Skip((page - 1) * pageSize).  // Skip previous pages
        Take(pageSize).                // Get current page
        ToSlice()
}

// Get page 3
page3 := getPage(allData, 3)
```

### Finding First Match

```go
func findFirstError(logs []LogEntry) (LogEntry, bool) {
    return chankit.FromSlice(ctx, logs).
        Filter(func(log LogEntry) bool {
            return log.Level == "ERROR"
        }).
        First()
}

if errorLog, found := findFirstError(logs); found {
    fmt.Printf("First error: %s\n", errorLog.Message)
}
```

### Sample First N

```go
// Get sample of large dataset
sample := chankit.RangePipeline(ctx, 1, 1000001, 1).
    Take(1000).  // First 1000 items
    ToSlice()

// Quick preview
fmt.Printf("Sample size: %d\n", len(sample))
```

### Skip Header Row

```go
// Skip CSV header
dataRows := chankit.FromSlice(ctx, csvLines).
    Skip(1).  // Skip header
    Map(parseCsvLine).
    ToSlice()
```

### Take Until Condition

```go
// Take all valid entries
validEntries := chankit.FromSlice(ctx, entries).
    TakeWhile(func(e Entry) bool {
        return e.IsValid && e.Timestamp.After(cutoff)
    }).
    ToSlice()
```

### Skip Invalid Prefix

```go
// Skip leading invalid data
cleanData := chankit.FromSlice(ctx, rawData).
    SkipWhile(func(d Data) bool {
        return !d.IsValid
    }).
    ToSlice()
```

## Pagination Pattern

```go
type Page struct {
    Items []any
    Page  int
    Size  int
    Total int
}

func paginate(data []any, pageNum, pageSize int) Page {
    total := len(data)

    items := chankit.FromSlice(ctx, data).
        Skip((pageNum - 1) * pageSize).
        Take(pageSize).
        ToSlice()

    return Page{
        Items: items,
        Page:  pageNum,
        Size:  pageSize,
        Total: total,
    }
}

// Usage
page1 := paginate(allData, 1, 20)  // First page
page2 := paginate(allData, 2, 20)  // Second page
```

## Validation Pattern

```go
type ValidationResult struct {
    AllValid  bool
    AnyErrors bool
    ErrorCount int
}

func validateData(data []Data) ValidationResult {
    pipeline := chankit.FromSlice(ctx, data)

    return ValidationResult{
        AllValid: pipeline.All(func(d Data) bool {
            return d.IsValid
        }),
        AnyErrors: pipeline.Any(func(d Data) bool {
            return d.HasError
        }),
        ErrorCount: pipeline.
            Filter(func(d Data) bool { return d.HasError }).
            Count(),
    }
}
```

## Comparison: Take vs TakeWhile

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-6">

### Take (Fixed Count)

```go
// Always takes exactly N
pipeline.Take(10)

Input:  [1,2,3,4,5,6,7,8,9,10,11,12]
Output: [1,2,3,4,5,6,7,8,9,10]
```

**Use when:**
- You know exact count
- Pagination
- Sampling

</div>

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-6">

### TakeWhile (Conditional)

```go
// Takes until condition false
pipeline.TakeWhile(func(x int) bool {
    return x < 10
})

Input:  [1,2,3,4,5,15,6,7,8,9]
Output: [1,2,3,4,5]
        (stops at 15)
```

**Use when:**
- Condition-based stopping
- Unknown count
- Until sentinel value

</div>

</div>

## Performance Tips

<div class="space-y-4 my-6">

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">💡</span>
<div>
<strong>Use Take to limit processing</strong>
<p class="text-dark-text-muted">Take stops upstream processing early, saving computation</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">💡</span>
<div>
<strong>Skip is cheaper than Filter</strong>
<p class="text-dark-text-muted">If skipping fixed count, use Skip instead of Filter with counter</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">💡</span>
<div>
<strong>First/Last read entire stream</strong>
<p class="text-dark-text-muted">Last must read all values. Use Take(1) if you just need first</p>
</div>
</div>

</div>

## Best Practices

<div class="space-y-4 my-6">

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Take early in pipeline</strong>
<p class="text-dark-text-muted">Apply Take as early as possible to avoid unnecessary work</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Check First/Last return value</strong>
<p class="text-dark-text-muted">Always check the bool return to handle empty pipelines</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Combine Skip + Take for pagination</strong>
<p class="text-dark-text-muted">Use Skip to offset and Take to limit for efficient pagination</p>
</div>
</div>

</div>

## Related APIs

- [Pipeline](/api/pipeline) - Pipeline creation
- [Transformers](/api/transformers) - Filter, Map operations
- [Flow Control](/api/flow-control) - Throttle, Batch

<div class="flex gap-4 my-8">

[Back to API Overview](/api/pipeline){.custom-button}
[Previous: Combiners ←](/api/combiners){.custom-button-outline}

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
