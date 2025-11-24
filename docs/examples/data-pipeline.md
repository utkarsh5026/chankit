# Data Processing Pipeline

Process a stream of numbers - square them, filter for even results, skip the first 10, and take the next 20.

## Scenario

You have a large stream of numbers that need to be:
1. Squared
2. Filtered to keep only even values
3. Skip the first 10 results (pagination/offset)
4. Take the next 20 results (limit)

This is a common pattern in data processing, ETL pipelines, and paginated result sets.

## Implementation

```go
package main

import (
    "context"
    "fmt"
    "github.com/utkarsh5026/chankit/chankit"
)

func main() {
    ctx := context.Background()

    // Process numbers 1-100
    result := chankit.RangePipeline(ctx, 1, 101, 1).
        Map(func(x int) any { return x * x }).           // Square each number
        Filter(func(x any) bool {                         // Keep only even squares
            return x.(int)%2 == 0
        }).
        Skip(10).                                         // Skip first 10 results
        Take(20).                                         // Take next 20
        ToSlice()                                         // Collect to slice

    fmt.Printf("Processed %d values: %v\n", len(result), result)
}
```

## Output

```
Processed 20 values: [484 576 676 784 900 1024 1156 1296 1444 1600 1764 1936 2116 2304 2500 2704 2916 3136 3364 3600]
```

## How It Works

<div class="space-y-4 my-6">

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 1: Generate Range
**RangePipeline(ctx, 1, 101, 1)** creates a channel that emits numbers from 1 to 100.

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 2: Map (Square)
**Map(func(x int) any { return x * x })** transforms each number by squaring it.
- 1 → 1
- 2 → 4
- 3 → 9
- ...
- 100 → 10000

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 3: Filter (Even Only)
**Filter(func(x any) bool { return x.(int)%2 == 0 })** keeps only even squares.
- Keeps: 4, 16, 36, 64, 100, ...
- Drops: 1, 9, 25, 49, 81, ...

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 4: Skip First 10
**Skip(10)** skips the first 10 even squares.
- Skipped: 4, 16, 36, 64, 100, 144, 196, 256, 324, 400

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 5: Take Next 20
**Take(20)** takes the next 20 values.
- Result starts at: 484 (22²)

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 6: Collect Results
**ToSlice()** collects all values into a slice for final output.

</div>

</div>

## Use Cases

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 📊 Data Analytics
- Process large datasets with transformations
- Apply filters to clean data
- Implement pagination for results

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 🔢 Mathematical Processing
- Bulk mathematical transformations
- Statistical filtering
- Sample selection from datasets

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 📄 Paginated Results
- Implement offset/limit patterns
- Process paginated API results
- Handle large result sets efficiently

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 🔄 ETL Pipelines
- Extract-Transform-Load workflows
- Data cleaning and normalization
- Multi-stage data processing

</div>

</div>

## Variations

### Without Skip/Take (Process All)

```go
result := chankit.RangePipeline(ctx, 1, 101, 1).
    Map(func(x int) any { return x * x }).
    Filter(func(x any) bool { return x.(int)%2 == 0 }).
    ToSlice()
// All 50 even squares from 1² to 100²
```

### With Different Transformations

```go
result := chankit.RangePipeline(ctx, 1, 101, 1).
    Map(func(x int) any { return x * 3 }).        // Triple instead
    Filter(func(x any) bool { return x.(int) > 100 }). // Filter by value
    Take(10).
    ToSlice()
```

### Stream Instead of Collect

```go
// More memory efficient for large datasets
chankit.RangePipeline(ctx, 1, 1000001, 1).
    Map(squareFunc).
    Filter(isEvenFunc).
    Skip(10).
    Take(20).
    ForEach(func(val any) {
        processValue(val) // Process immediately, no memory buildup
    })
```

## Performance Characteristics

<div class="bg-primary-900/20 border border-primary-500 rounded-lg p-6 my-6">

**Memory:** O(1) for streaming (ForEach), O(n) for collecting (ToSlice)

**Time Complexity:** O(n) where n is the range size

**Concurrency:** Each operator runs in its own goroutine, allowing parallel processing

</div>

## Related Examples

- [Event Stream Processing](/examples/event-stream) - Real-time event handling
- [Data Transformation](/examples/transformation) - Complex transformations
- [Real-time Analytics](/examples/analytics) - Sensor data processing

<div class="flex gap-4 my-8">

[Next: Event Stream →](/examples/event-stream){.custom-button}
[Back to Guide](/guide/core-concepts){.custom-button-outline}

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
