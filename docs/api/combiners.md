# Combiners

Combiner operators merge multiple pipelines or channels into a single stream.

## Merge

Combine multiple channels into one stream.

```go
func (p *Pipeline[T]) Merge(channels ...<-chan T) *Pipeline[T]
func Merge[T any](ctx context.Context, channels ...<-chan T) <-chan T
```

**Behavior:** Emits values from all input channels as they arrive (interleaved)

**Example:**
```go
ch1 := chankit.Range(ctx, 1, 6, 1)       // 1,2,3,4,5
ch2 := chankit.Range(ctx, 10, 16, 1)    // 10,11,12,13,14,15

merged := chankit.Merge(ctx, ch1, ch2)
// Output (order may vary): 1,10,2,11,3,12,4,13,5,14,15
```

---

## Zip / ZipWith

Pair values from two channels.

```go
func ZipWith[T, U any](ch1 <-chan T, ch2 <-chan U) <-chan Pair[T, U]
```

**Behavior:** Pairs first value from ch1 with first from ch2, etc. Stops when either ends.

**Example:**
```go
numbers := chankit.Range(ctx, 1, 6, 1)      // 1,2,3,4,5
letters := chankit.FromSlice(ctx, []string{"a", "b", "c"})

pairs := chankit.ZipWith(numbers, letters)
// Output: [(1,a), (2,b), (3,c)]
// Stops when letters ends
```

---

## Concat

Concatenate pipelines (one after another).

```go
func (p *Pipeline[T]) Concat(other <-chan T) *Pipeline[T]
```

**Behavior:** Emits all values from first channel, then all from second

**Example:**
```go
first := chankit.Range(ctx, 1, 4, 1)     // 1,2,3
second := chankit.Range(ctx, 10, 13, 1)  // 10,11,12

result := chankit.From(ctx, first).
    Concat(second).
    ToSlice()
// Output: [1,2,3,10,11,12]
```

---

## Real-World Examples

### Combining API Responses

```go
// Fetch from multiple endpoints in parallel
users1 := fetchUsers("/api/users?page=1")
users2 := fetchUsers("/api/users?page=2")
users3 := fetchUsers("/api/users?page=3")

allUsers := chankit.Merge(ctx, users1, users2, users3).
    ToSlice()
```

### Fan-Out / Fan-In Pattern

```go
func processData(ctx context.Context, data []int) []int {
    // Split into even and odd pipelines
    source := chankit.FromSlice(ctx, data)

    evens := source.Filter(func(x int) bool { return x%2 == 0 }).
        Map(func(x int) any { return x * 2 })

    odds := source.Filter(func(x int) bool { return x%2 != 0 }).
        Map(func(x int) any { return x * 3 })

    // Merge results
    merged := evens.Merge(odds.Chan())
    return merged.ToSlice()
}
```

### Pairing IDs with Names

```go
type User struct {
    ID   int
    Name string
}

ids := chankit.Range(ctx, 1, 101, 1)           // 1-100
names := chankit.FromSlice(ctx, userNames)      // Names slice

users := chankit.ZipWith(ids, names).
    Map(func(pair Pair[int, string]) any {
        return User{
            ID:   pair.First,
            Name: pair.Second,
        }
    }).
    ToSlice()
```

### Merging Real-Time Streams

```go
// Merge logs from multiple services
service1Logs := streamLogs(ctx, "service1")
service2Logs := streamLogs(ctx, "service2")
service3Logs := streamLogs(ctx, "service3")

allLogs := chankit.Merge(ctx,
    service1Logs,
    service2Logs,
    service3Logs,
).
    Tap(func(log LogEntry) {
        fmt.Printf("[%s] %s\n", log.Service, log.Message)
    })
```

### Sequential Processing

```go
// Process batches sequentially
batch1 := processBatch(ctx, data[0:1000])
batch2 := processBatch(ctx, data[1000:2000])
batch3 := processBatch(ctx, data[2000:3000])

results := chankit.From(ctx, batch1).
    Concat(batch2).
    Concat(batch3).
    ToSlice()
```

## Merge vs Concat

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-6">

### Merge (Parallel)

```
Ch1: 1---2---3---4
Ch2: a---b---c---d
Merge: 1-a-2-b-3-c-4-d
       (interleaved)
```

**Use when:**
- Processing in parallel
- Order doesn't matter
- Want concurrent execution

</div>

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-6">

### Concat (Sequential)

```
Ch1: 1---2---3---4
Ch2: a---b---c---d
Concat: 1-2-3-4-a-b-c-d
        (one after another)
```

**Use when:**
- Sequential processing
- Order matters
- Pagination/batches

</div>

</div>

## Advanced Patterns

### Merge with Transformation

```go
// Process multiple sources differently
source1 := chankit.Range(ctx, 1, 6, 1).
    Map(func(x int) any { return x * 10 })

source2 := chankit.Range(ctx, 1, 6, 1).
    Map(func(x int) any { return x * 100 })

merged := source1.Merge(source2.Chan())
// Contains: 10,100,20,200,30,300,40,400,50,500
```

### Zip with Error Handling

```go
results := make(chan Result, 10)
errors := make(chan error, 10)

type ProcessedItem struct {
    Result Result
    Error  error
}

combined := chankit.ZipWith(results, errors).
    Filter(func(pair Pair[Result, error]) bool {
        return pair.Second == nil // Filter out errors
    }).
    Map(func(pair Pair[Result, error]) any {
        return pair.First
    })
```

### Dynamic Merging

```go
// Merge variable number of channels
sources := []<-chan int{
    chankit.Range(ctx, 1, 6, 1),
    chankit.Range(ctx, 10, 16, 1),
    chankit.Range(ctx, 20, 26, 1),
}

merged := chankit.Merge(ctx, sources...)
```

## Performance Considerations

<div class="space-y-4 my-6">

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">💡</span>
<div>
<strong>Merge is concurrent</strong>
<p class="text-dark-text-muted">All input channels are read concurrently. Good for parallel processing</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">💡</span>
<div>
<strong>Concat is sequential</strong>
<p class="text-dark-text-muted">Waits for first channel to complete before starting second. Better for ordered processing</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">💡</span>
<div>
<strong>Zip stops at shortest</strong>
<p class="text-dark-text-muted">ZipWith stops when the shorter channel ends. Plan accordingly</p>
</div>
</div>

</div>

## Best Practices

<div class="space-y-4 my-6">

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Use Merge for parallel work</strong>
<p class="text-dark-text-muted">When fetching from multiple sources, Merge provides natural parallelism</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Use Concat for ordered sequences</strong>
<p class="text-dark-text-muted">When order matters or processing batches, use Concat</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Check Zip length mismatch</strong>
<p class="text-dark-text-muted">Ensure zipped channels have compatible lengths or handle early termination</p>
</div>
</div>

</div>

## Related APIs

- [Pipeline](/api/pipeline) - Pipeline creation
- [Generators](/api/generators) - Creating source pipelines
- [Transformers](/api/transformers) - Transform combined results

<div class="flex gap-4 my-8">

[Next: Selectors →](/api/selectors){.custom-button}
[Previous: Generators ←](/api/generators){.custom-button-outline}

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
