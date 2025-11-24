# Flow Control

Flow control operators manage the timing and rate of values flowing through a pipeline.

## Throttle

Emit values at a maximum rate, dropping excess values.

```go
func (p *Pipeline[T]) Throttle(duration time.Duration) *Pipeline[T]
```

**Use Case:** Limit rate of high-frequency events (UI updates, logs)

**Behavior:** Emits first value immediately, then enforces minimum interval

**Example:**
```go
// At most 10 values per second (drop extras)
pipeline.Throttle(100 * time.Millisecond)
```

<div class="bg-primary-900/20 border-l-4 border-primary-500 p-4 my-4 rounded">

**Perfect for:** UI events, mouse movements, scroll handlers, metrics sampling

</div>

---

## Debounce

Wait for silence before emitting a value.

```go
func (p *Pipeline[T]) Debounce(duration time.Duration) *Pipeline[T]
```

**Use Case:** Wait for user to stop typing before triggering action

**Behavior:** Resets timer on each new value, emits only after silence

**Example:**
```go
// Wait 300ms after last keystroke
searchInput.Debounce(300 * time.Millisecond)
```

<div class="bg-primary-900/20 border-l-4 border-primary-500 p-4 my-4 rounded">

**Perfect for:** Search boxes, form validation, auto-save, window resize

</div>

---

## FixedInterval

Pace values at exact intervals without dropping any.

```go
func (p *Pipeline[T]) FixedInterval(duration time.Duration) *Pipeline[T]
```

**Use Case:** Rate-limit API calls while processing all requests

**Behavior:** Emits values at consistent intervals, never drops values

**Example:**
```go
// Exactly 10 requests per second
requests.FixedInterval(100 * time.Millisecond)
```

<div class="bg-primary-900/20 border-l-4 border-primary-500 p-4 my-4 rounded">

**Perfect for:** API rate limiting, controlled throughput, consistent pacing

</div>

---

## Batch

Group values by size or timeout.

```go
func (p *Pipeline[T]) Batch(size int, timeout time.Duration) *Pipeline[[]T]
```

**Parameters:**
- `size` - Maximum batch size
- `timeout` - Maximum wait time

**Behavior:** Emits batch when size is reached OR timeout expires (whichever first)

**Example:**
```go
// Batch 100 items or every 5 seconds
pipeline.Batch(100, 5*time.Second)
```

<div class="bg-primary-900/20 border-l-4 border-primary-500 p-4 my-4 rounded">

**Perfect for:** Bulk database inserts, batch API calls, log aggregation

</div>

---

## Comparison

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-6">

### Throttle vs Debounce

**Throttle:** Emits regularly during activity
```
Events:  |x|x|x|x|x|x|x|x|
Throttle:|x|  |x|  |x|  |x|
(100ms)
```

**Debounce:** Emits after activity stops
```
Events:  |x|x|x|x|    (silence)
Debounce:|            x|
(100ms)
```

</div>

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-6">

### Throttle vs FixedInterval

**Throttle:** May drop values
```
Fast input: 1,2,3,4,5,6,7,8
Throttle:   1,  3,  5,  7
```

**FixedInterval:** Never drops
```
Fast input: 1,2,3,4,5,6,7,8
Fixed:      1,2,3,4,5,6,7,8
            (paced at 100ms each)
```

</div>

</div>

## Real-World Examples

### Search Debouncing

```go
searchInput := make(chan string, 10)

results := chankit.From(ctx, searchInput).
    Debounce(300 * time.Millisecond).        // Wait for typing to stop
    Filter(func(q string) bool {
        return len(q) >= 2                   // Minimum 2 characters
    }).
    Map(func(q string) any {
        return performSearch(q)              // Execute search
    })
```

### Rate-Limited API Calls

```go
requests := []APIRequest{...}

chankit.FromSlice(ctx, requests).
    FixedInterval(100 * time.Millisecond).   // 10 req/sec
    Map(func(req APIRequest) any {
        return callAPI(req)
    }).
    ForEach(handleResponse)
```

### Batch Database Inserts

```go
records := make(chan Record, 1000)

chankit.From(ctx, records).
    Batch(100, 5*time.Second).               // 100 records or 5sec
    ForEach(func(batch []Record) {
        db.BulkInsert(batch)
    })
```

### UI Event Throttling

```go
mouseEvents := make(chan MouseEvent, 100)

chankit.From(ctx, mouseEvents).
    Throttle(16 * time.Millisecond).         // 60 FPS
    ForEach(updateUI)
```

## Performance Tips

<div class="space-y-4 my-6">

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">💡</span>
<div>
<strong>Choose the right operator</strong>
<p class="text-dark-text-muted">Throttle for sampling, Debounce for settling, FixedInterval for pacing, Batch for grouping</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">💡</span>
<div>
<strong>Tune timings carefully</strong>
<p class="text-dark-text-muted">Too short: minimal benefit. Too long: feels laggy. Test with real usage patterns</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">💡</span>
<div>
<strong>Consider backpressure</strong>
<p class="text-dark-text-muted">Use buffered channels or Throttle/Debounce to handle burst traffic</p>
</div>
</div>

</div>

## Related APIs

- [Pipeline](/api/pipeline) - Creating pipelines
- [Transformers](/api/transformers) - Map, Filter operations
- [Examples](/examples/event-stream) - Event processing patterns

<div class="flex gap-4 my-8">

[Next: Transformers →](/api/transformers){.custom-button}
[Previous: Pipeline ←](/api/pipeline){.custom-button-outline}

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
