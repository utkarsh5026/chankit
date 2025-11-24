# Event Stream Processing

Build a real-time event processing system that debounces user input, filters events, and batches them for processing.

## Scenario

You're building a real-time event processing system that needs to:
1. Receive high-frequency events (user clicks, mouse moves, etc.)
2. Debounce to reduce noise (wait for activity to settle)
3. Filter events by type
4. Batch events for efficient processing

This pattern is common in UI event handling, logging systems, and real-time analytics.

## Implementation

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/utkarsh5026/chankit/chankit"
)

type Event struct {
    Type      string
    Timestamp time.Time
    Data      string
}

func main() {
    ctx := context.Background()

    // Simulate event stream
    events := make(chan Event, 100)
    go func() {
        defer close(events)
        for i := 0; i < 50; i++ {
            events <- Event{
                Type:      "click",
                Timestamp: time.Now(),
                Data:      fmt.Sprintf("event-%d", i),
            }
            time.Sleep(10 * time.Millisecond)
        }
    }()

    // Process events: debounce, filter, and batch
    batches := chankit.From(ctx, events).
        Debounce(50 * time.Millisecond).                  // Wait for 50ms silence
        Filter(func(e Event) bool {                       // Filter specific events
            return e.Type == "click"
        }).
        Batch(5, 200*time.Millisecond)                    // Batch 5 or every 200ms

    // Process batches
    for batch := range batches {
        fmt.Printf("Processing batch of %d events\n", len(batch))
    }
}
```

## How It Works

<div class="space-y-4 my-6">

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 1: Event Generator
A goroutine simulates high-frequency events, sending them at 10ms intervals.

```go
events <- Event{
    Type:      "click",
    Timestamp: time.Now(),
    Data:      fmt.Sprintf("event-%d", i),
}
```

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 2: Debounce
**Debounce(50 * time.Millisecond)** waits for 50ms of silence before emitting a value.
- Reduces noise from rapid consecutive events
- Only processes events after activity settles
- Perfect for search-as-you-type, scroll handlers, etc.

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 3: Filter
**Filter(func(e Event) bool { return e.Type == "click" })** keeps only "click" events.
- Removes unwanted event types
- Implements business logic for event selection

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 4: Batch
**Batch(5, 200*time.Millisecond)** groups events for efficient processing.
- Batches up to 5 events OR
- Emits batch after 200ms (whichever comes first)
- Reduces overhead of processing individual events

</div>

</div>

## Use Cases

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 🖱️ UI Event Handling
- Mouse movement tracking
- Scroll event processing
- Window resize handlers
- Click debouncing

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 📝 Logging Systems
- Log aggregation
- Batch log writes
- Error reporting
- Audit trail processing

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 📊 Real-time Analytics
- User activity tracking
- Session analytics
- Metrics aggregation
- Event correlation

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 🔔 Notification Systems
- Alert batching
- Notification throttling
- Event deduplication
- Alert aggregation

</div>

</div>

## Variations

### Different Debounce Strategy (Throttle Instead)

```go
// Throttle: Emit at most one value per time period
batches := chankit.From(ctx, events).
    Throttle(100 * time.Millisecond).  // At most 10 events/second
    Filter(isImportant).
    Batch(10, 500*time.Millisecond)
```

### Multiple Event Types

```go
type EventType string

const (
    Click   EventType = "click"
    Hover   EventType = "hover"
    Scroll  EventType = "scroll"
)

// Process different event types separately
clickEvents := chankit.From(ctx, events).
    Filter(func(e Event) bool { return e.Type == string(Click) })

scrollEvents := chankit.From(ctx, events).
    Filter(func(e Event) bool { return e.Type == string(Scroll) })
```

### With Context Timeout

```go
// Stop processing after 30 seconds
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

batches := chankit.From(ctx, events).
    Debounce(50 * time.Millisecond).
    Batch(5, 200*time.Millisecond)

// Automatically stops after timeout
for batch := range batches {
    processBatch(batch)
}
```

### Enriched Event Processing

```go
// Add metadata, filter, and batch
enriched := chankit.From(ctx, events).
    Map(func(e Event) any {
        // Enrich with additional data
        return EnrichedEvent{
            Event:    e,
            UserID:   getUserID(e),
            Session:  getSession(e),
        }
    }).
    Filter(func(e any) bool {
        ee := e.(EnrichedEvent)
        return ee.UserID != "" // Only authenticated users
    }).
    Batch(10, time.Second)
```

## Visual Flow

```
Events Stream (10ms intervals)
    ↓
Debounce (50ms silence)
    ↓
Filter (click events only)
    ↓
Batch (5 events or 200ms)
    ↓
Process batches
```

## Performance Characteristics

<div class="bg-primary-900/20 border border-primary-500 rounded-lg p-6 my-6">

**Latency:** 50ms minimum (debounce) + up to 200ms (batch timeout)

**Throughput:** Handles high-frequency events efficiently

**Memory:** O(batch_size) per batch, automatic cleanup

**Concurrency:** Safe for concurrent event producers

</div>

## Best Practices

<div class="space-y-4 my-6">

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Choose appropriate debounce time</strong>
<p class="text-dark-text-muted">Too short: doesn't reduce noise. Too long: feels laggy</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Size batches appropriately</strong>
<p class="text-dark-text-muted">Balance between processing overhead and memory usage</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Always use context</strong>
<p class="text-dark-text-muted">Allows graceful shutdown of event processing</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Handle backpressure</strong>
<p class="text-dark-text-muted">Use buffered channels or drop strategies for high-volume streams</p>
</div>
</div>

</div>

## Related Examples

- [Search Results](/examples/search-results) - Debounced search
- [Real-time Analytics](/examples/analytics) - Sensor data processing
- [Rate Limiting](/examples/rate-limiting) - API throttling

<div class="flex gap-4 my-8">

[Next: Search Results →](/examples/search-results){.custom-button}
[Previous: Data Pipeline ←](/examples/data-pipeline){.custom-button-outline}

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
