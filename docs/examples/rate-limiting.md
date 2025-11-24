# Rate-Limited API Calls

Make API calls at a controlled rate to avoid hitting rate limits and overwhelming external services.

## Scenario

You need to make multiple API calls but must:
1. Respect rate limits (e.g., 10 requests per second)
2. Avoid overwhelming the API server
3. Process requests in order
4. Track timing and progress

This is essential when working with third-party APIs that enforce rate limits.

## Implementation

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/utkarsh5026/chankit/chankit"
)

type APIRequest struct {
    ID       int
    Endpoint string
}

func callAPI(req APIRequest) string {
    // Simulate API call
    return fmt.Sprintf("Response for request %d", req.ID)
}

func main() {
    ctx := context.Background()

    // Create 20 API requests
    requests := make([]APIRequest, 20)
    for i := 0; i < 20; i++ {
        requests[i] = APIRequest{
            ID:       i + 1,
            Endpoint: "/api/data",
        }
    }

    // Process at most 10 requests per second (100ms interval)
    fmt.Println("Making rate-limited API calls...")
    start := time.Now()

    chankit.FromSlice(ctx, requests).
        FixedInterval(100 * time.Millisecond).            // 10 per second max
        Tap(func(req APIRequest) {                        // Log progress
            fmt.Printf("[%s] Processing request %d\n",
                time.Since(start).Round(time.Millisecond), req.ID)
        }).
        Map(func(req APIRequest) any {                    // Make API call
            return callAPI(req)
        }).
        ForEach(func(resp any) {                          // Handle responses
            // Process response
        })

    fmt.Printf("Completed in %s\n", time.Since(start).Round(time.Millisecond))
}
```

## Output

```
Making rate-limited API calls...
[0s] Processing request 1
[100ms] Processing request 2
[200ms] Processing request 3
[300ms] Processing request 4
[400ms] Processing request 5
[500ms] Processing request 6
[600ms] Processing request 7
[700ms] Processing request 8
[800ms] Processing request 9
[900ms] Processing request 10
[1s] Processing request 11
...
Completed in 2s
```

## How It Works

<div class="space-y-4 my-6">

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 1: Create Request List
Generate a list of API requests to process.

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 2: Fixed Interval (100ms)
**FixedInterval(100 * time.Millisecond)** ensures exactly 100ms between emissions.
- 1000ms / 100ms = **10 requests per second**
- Guarantees consistent pacing regardless of processing time

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 3: Progress Logging
**Tap** allows side effects (logging) without modifying the stream.

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 4: API Call
**Map** performs the actual API call and transforms the request to a response.

</div>

</div>

## Rate Limiting Strategies

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-6">

### FixedInterval (Recommended)
```go
// Exactly N requests per second
.FixedInterval(100 * time.Millisecond) // 10/sec
```

**Pros:**
- Predictable timing
- Respects rate limits precisely
- Smooth, consistent load

**Best for:** API rate limits, consistent load

</div>

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-6">

### Throttle (Drop Excess)
```go
// At most N requests per second (drops excess)
.Throttle(100 * time.Millisecond)
```

**Pros:**
- Never blocks
- Handles bursts
- Drops unnecessary requests

**Best for:** UI events, non-critical requests

</div>

</div>

## Use Cases

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 🌐 External APIs
- Third-party API calls
- Rate-limited endpoints
- Webhook delivery
- Data synchronization

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 📧 Email/Notifications
- Bulk email sending
- SMS notifications
- Push notifications
- Alert delivery

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 🗄️ Database Operations
- Batch inserts/updates
- Migration scripts
- Data export
- Index rebuilding

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 🔄 Data Sync
- File uploads
- Backup operations
- Cache warming
- Data replication

</div>

</div>

## Common Rate Limits

<div class="space-y-4 my-6">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

**1 request/second** → `FixedInterval(1 * time.Second)`
```go
.FixedInterval(time.Second) // 1 req/sec
```

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

**10 requests/second** → `FixedInterval(100 * time.Millisecond)`
```go
.FixedInterval(100 * time.Millisecond) // 10 req/sec
```

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

**100 requests/minute** → `FixedInterval(600 * time.Millisecond)`
```go
.FixedInterval(600 * time.Millisecond) // 100 req/min
```

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

**1000 requests/hour** → `FixedInterval(3.6 * time.Second)`
```go
.FixedInterval(3600 * time.Millisecond) // 1000 req/hour
```

</div>

</div>

## Variations

### With Retry Logic

```go
maxRetries := 3

chankit.FromSlice(ctx, requests).
    FixedInterval(100 * time.Millisecond).
    Map(func(req APIRequest) any {
        var resp string
        var err error

        for i := 0; i < maxRetries; i++ {
            resp, err = callAPIWithError(req)
            if err == nil {
                break
            }
            time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
        }

        return APIResponse{Response: resp, Error: err}
    })
```

### With Burst Allowance

```go
// Allow initial burst, then rate limit
burst := 5

chankit.FromSlice(ctx, requests).
    Take(burst).                                    // First 5 go immediately
    Concat(
        chankit.FromSlice(ctx, requests[burst:]).
            FixedInterval(100 * time.Millisecond),  // Rest are rate limited
    )
```

### Dynamic Rate Adjustment

```go
currentRate := 100 * time.Millisecond

chankit.FromSlice(ctx, requests).
    Map(func(req APIRequest) any {
        // Adjust rate based on response
        resp, err := callAPI(req)

        if err == RateLimitError {
            currentRate *= 2 // Slow down
        } else if currentRate > 50*time.Millisecond {
            currentRate = currentRate * 9 / 10 // Speed up by 10%
        }

        time.Sleep(currentRate)
        return resp
    })
```

### With Priority Queue

```go
type PriorityRequest struct {
    Request  APIRequest
    Priority int
}

// Sort by priority first
sort.Slice(priorityRequests, func(i, j int) bool {
    return priorityRequests[i].Priority > priorityRequests[j].Priority
})

// Process high-priority first, rate-limited
chankit.FromSlice(ctx, priorityRequests).
    FixedInterval(100 * time.Millisecond).
    Map(func(pr PriorityRequest) any {
        return callAPI(pr.Request)
    })
```

## Monitoring and Metrics

```go
type Metrics struct {
    Total      int
    Successful int
    Failed     int
    Latency    []time.Duration
}

metrics := &Metrics{}

chankit.FromSlice(ctx, requests).
    FixedInterval(100 * time.Millisecond).
    Map(func(req APIRequest) any {
        start := time.Now()
        resp, err := callAPIWithError(req)

        metrics.Total++
        if err != nil {
            metrics.Failed++
        } else {
            metrics.Successful++
        }
        metrics.Latency = append(metrics.Latency, time.Since(start))

        return resp
    })

fmt.Printf("Total: %d, Success: %d, Failed: %d\n",
    metrics.Total, metrics.Successful, metrics.Failed)
```

## Performance Impact

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-red-900/20 border border-red-500/50 rounded-lg p-6">

### ❌ Without Rate Limiting
- Risk of 429 errors (Too Many Requests)
- API blocks or throttles your IP
- Wasted failed requests
- Unpredictable latency
- Service disruption

</div>

<div class="bg-green-900/20 border border-green-500/50 rounded-lg p-6">

### ✅ With Rate Limiting
- Consistent throughput
- No 429 errors
- Better API relationships
- Predictable completion time
- Reliable service

</div>

</div>

## Best Practices

<div class="space-y-4 my-6">

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Stay below the limit</strong>
<p class="text-dark-text-muted">Set your rate to 80-90% of the API's limit for safety margin</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Handle errors gracefully</strong>
<p class="text-dark-text-muted">Implement retry logic with exponential backoff for transient failures</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Monitor and adjust</strong>
<p class="text-dark-text-muted">Track success rates and adjust timing if needed</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Use context timeouts</strong>
<p class="text-dark-text-muted">Set reasonable timeouts to avoid hanging indefinitely</p>
</div>
</div>

</div>

## Related Examples

- [Data Pipeline](/examples/data-pipeline) - Batch processing
- [Event Stream](/examples/event-stream) - Event throttling
- [Search Results](/examples/search-results) - Debouncing

<div class="flex gap-4 my-8">

[Next: Data Transformation →](/examples/transformation){.custom-button}
[Previous: Search Results ←](/examples/search-results){.custom-button-outline}

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
