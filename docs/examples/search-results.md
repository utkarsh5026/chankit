# Search-As-You-Type with Debouncing

Implement an efficient search system that waits for the user to stop typing before executing the search.

## Scenario

Build a search feature that:
1. Receives rapid user input as they type
2. Waits for the user to stop typing (debounce)
3. Filters out queries that are too short
4. Executes search only when necessary

This prevents excessive API calls and improves user experience.

## Implementation

```go
package main

import (
    "context"
    "fmt"
    "strings"
    "time"
    "github.com/utkarsh5026/chankit/chankit"
)

func performSearch(query string) []string {
    // Simulate search results
    database := []string{"apple", "application", "apply", "banana", "band", "bandana"}
    results := []string{}
    for _, item := range database {
        if strings.Contains(strings.ToLower(item), strings.ToLower(query)) {
            results = append(results, item)
        }
    }
    return results
}

func main() {
    ctx := context.Background()

    // Simulate user typing
    userInput := make(chan string, 10)
    go func() {
        defer close(userInput)
        queries := []string{"a", "ap", "app", "appl", "apple"}
        for _, q := range queries {
            userInput <- q
            time.Sleep(50 * time.Millisecond)
        }
    }()

    // Debounce and search
    results := chankit.From(ctx, userInput).
        Debounce(200 * time.Millisecond).                 // Wait for typing to stop
        Filter(func(q string) bool { return len(q) >= 2 }).// Min 2 chars
        Map(func(q string) any {                          // Perform search
            fmt.Printf("Searching for: %s\n", q)
            return performSearch(q)
        })

    // Display results
    results.ForEach(func(r any) {
        fmt.Printf("Results: %v\n", r)
    })
}
```

## Output

```
Searching for: apple
Results: [apple application apply]
```

## How It Works

<div class="space-y-4 my-6">

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 1: User Input Simulation
Simulates a user typing "apple" character by character with 50ms delays.

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 2: Debounce (200ms)
Waits for 200ms of silence after the last keystroke before processing.
- **"a"** → Wait 50ms → **"ap"** arrives → Reset timer
- **"ap"** → Wait 50ms → **"app"** arrives → Reset timer
- **"app"** → Wait 50ms → **"appl"** arrives → Reset timer
- **"appl"** → Wait 50ms → **"apple"** arrives → Reset timer
- **"apple"** → Wait 200ms → **No more input** → Emit "apple"

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 3: Filter Short Queries
Only processes queries with 2+ characters to avoid unnecessary searches.

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 4: Execute Search
Performs the actual search only once, after the user stops typing.

</div>

</div>

## Benefits

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-green-900/20 border border-green-500/50 rounded-lg p-6">

### ✅ Without Debouncing
```
User types "apple" (5 characters)
→ 5 API calls (one per character)
→ High server load
→ Wasted bandwidth
→ Results arrive out of order
```

</div>

<div class="bg-green-900/20 border border-green-500/50 rounded-lg p-6">

### ✅ With Debouncing
```
User types "apple" (5 characters)
→ 1 API call (after typing stops)
→ Low server load
→ Efficient bandwidth use
→ Only relevant results shown
```

</div>

</div>

## Use Cases

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 🔍 Search Features
- Autocomplete
- Search suggestions
- Instant search
- Typeahead

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 📝 Form Validation
- Email validation
- Username availability
- Real-time validation
- Input sanitization

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 🎨 Live Preview
- Markdown editors
- Code formatters
- Color pickers
- Size calculators

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 💬 Chat Features
- Typing indicators
- Message drafts
- Command detection
- Auto-save

</div>

</div>

## Variations

### With Caching

```go
cache := make(map[string][]string)

results := chankit.From(ctx, userInput).
    Debounce(200 * time.Millisecond).
    Filter(func(q string) bool { return len(q) >= 2 }).
    Map(func(q string) any {
        // Check cache first
        if cached, ok := cache[q]; ok {
            return cached
        }

        // Perform search
        results := performSearch(q)
        cache[q] = results
        return results
    })
```

### With Error Handling

```go
type SearchResult struct {
    Query   string
    Results []string
    Error   error
}

results := chankit.From(ctx, userInput).
    Debounce(200 * time.Millisecond).
    Filter(func(q string) bool { return len(q) >= 2 }).
    Map(func(q string) any {
        results, err := performSearchWithError(q)
        return SearchResult{
            Query:   q,
            Results: results,
            Error:   err,
        }
    })
```

### With Loading States

```go
type SearchState int

const (
    Idle SearchState = iota
    Searching
    Complete
)

// Separate channels for state and results
states := make(chan SearchState, 10)
results := make(chan []string, 10)

chankit.From(ctx, userInput).
    Tap(func(_ string) {
        states <- Searching
    }).
    Debounce(200 * time.Millisecond).
    Filter(func(q string) bool { return len(q) >= 2 }).
    ForEach(func(q string) {
        res := performSearch(q)
        results <- res
        states <- Complete
    })
```

## Performance Impact

<div class="bg-primary-900/20 border border-primary-500 rounded-lg p-6 my-6">

### Without Debouncing
- **5-character query:** 5 API calls
- **10-character query:** 10 API calls
- **100 users typing:** 500-1000 API calls

### With 200ms Debouncing
- **5-character query:** 1 API call
- **10-character query:** 1 API call
- **100 users typing:** 100 API calls

**Result: 80-90% reduction in API calls**

</div>

## Tuning Debounce Time

<div class="space-y-4 my-6">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

**100ms** - Very responsive, minimal debouncing
- Good for: Local searches, instant preview
- Risk: Still many API calls for slow typers

</div>

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-4">

**200ms** - Balanced (recommended)
- Good for: Most search implementations
- Sweet spot between responsiveness and efficiency

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-4">

**300-500ms** - Conservative
- Good for: Expensive searches, API rate limits
- Risk: May feel slightly laggy

</div>

</div>

## Related Examples

- [Event Stream Processing](/examples/event-stream) - Event handling
- [Rate Limiting](/examples/rate-limiting) - API throttling
- [Data Transformation](/examples/transformation) - Data processing

<div class="flex gap-4 my-8">

[Next: Rate Limiting →](/examples/rate-limiting){.custom-button}
[Previous: Event Stream ←](/examples/event-stream){.custom-button-outline}

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
