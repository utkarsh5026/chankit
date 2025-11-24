# Data Transformation Stream

Transform user data by filtering active users, extracting emails, and processing in batches.

## Scenario

You have a list of users and need to:
1. Filter for active users only
2. Extract specific fields (emails)
3. Transform the data (uppercase)
4. Batch for efficient processing

This pattern is common in ETL pipelines, user data processing, and email campaigns.

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

type User struct {
    ID     int
    Name   string
    Email  string
    Active bool
}

func main() {
    ctx := context.Background()

    // Sample users
    users := []User{
        {1, "Alice", "alice@example.com", true},
        {2, "Bob", "bob@example.com", false},
        {3, "Charlie", "charlie@example.com", true},
        {4, "Diana", "diana@example.com", true},
        {5, "Eve", "eve@example.com", false},
    }

    // Transform: active users → emails → uppercase → batch
    batches := chankit.FromSlice(ctx, users).
        Filter(func(u User) bool { return u.Active }).    // Active users only
        Map(func(u User) any { return u.Email }).         // Extract emails
        Map(func(e any) any {                             // Uppercase emails
            return strings.ToUpper(e.(string))
        }).
        Batch(2, 100*time.Millisecond)                    // Batch for processing

    // Process email batches
    for batch := range batches {
        fmt.Printf("Email batch: %v\n", batch)
    }
    // Output:
    // Email batch: [ALICE@EXAMPLE.COM CHARLIE@EXAMPLE.COM]
    // Email batch: [DIANA@EXAMPLE.COM]
}
```

## Output

```
Email batch: [ALICE@EXAMPLE.COM CHARLIE@EXAMPLE.COM]
Email batch: [DIANA@EXAMPLE.COM]
```

## How It Works

<div class="space-y-4 my-6">

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 1: Start with User Slice
Convert slice of users to a pipeline using **FromSlice**.

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 2: Filter Active Users
**Filter(func(u User) bool { return u.Active })** keeps only active users.
- Alice (active) ✅
- Bob (inactive) ❌
- Charlie (active) ✅
- Diana (active) ✅
- Eve (inactive) ❌

Result: 3 users (Alice, Charlie, Diana)

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 3: Extract Emails
**Map(func(u User) any { return u.Email })** extracts email field.
- Alice → "alice@example.com"
- Charlie → "charlie@example.com"
- Diana → "diana@example.com"

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 4: Transform to Uppercase
**Map(func(e any) any { return strings.ToUpper(e.(string)) })** uppercases emails.
- "alice@example.com" → "ALICE@EXAMPLE.COM"
- "charlie@example.com" → "CHARLIE@EXAMPLE.COM"
- "diana@example.com" → "DIANA@EXAMPLE.COM"

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 5: Batch for Processing
**Batch(2, 100*time.Millisecond)** groups emails into batches of 2.
- Batch 1: [ALICE@EXAMPLE.COM, CHARLIE@EXAMPLE.COM]
- Batch 2: [DIANA@EXAMPLE.COM]

</div>

</div>

## Use Cases

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 📧 Email Campaigns
- Filter subscribers
- Extract contact info
- Normalize data
- Batch for sending

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 🗄️ ETL Pipelines
- Extract from source
- Transform data
- Load to destination
- Handle in batches

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 👥 User Management
- Filter by criteria
- Export user data
- Format for reports
- Process in chunks

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 📊 Data Export
- Select records
- Transform format
- Normalize values
- Batch writes

</div>

</div>

## Transformation Patterns

### Chaining Multiple Maps

```go
// Each Map adds a transformation step
result := chankit.FromSlice(ctx, users).
    Filter(isActive).
    Map(extractEmail).                 // Step 1: Extract
    Map(normalizeEmail).               // Step 2: Normalize
    Map(validateEmail).                // Step 3: Validate
    Map(enrichWithMetadata).           // Step 4: Enrich
    ToSlice()
```

### Complex Filtering

```go
// Multiple filter conditions
result := chankit.FromSlice(ctx, users).
    Filter(func(u User) bool {
        return u.Active &&                      // Active
               len(u.Email) > 0 &&              // Has email
               strings.Contains(u.Email, "@") && // Valid format
               u.ID > 100                       // New users only
    })
```

### Conditional Transformation

```go
result := chankit.FromSlice(ctx, users).
    Map(func(u User) any {
        email := u.Email

        // Apply transformation based on condition
        if strings.HasSuffix(email, "@company.com") {
            return strings.ToUpper(email) // Internal: uppercase
        }
        return strings.ToLower(email)     // External: lowercase
    })
```

## Advanced Patterns

### Parallel Field Extraction

```go
type UserSummary struct {
    ID    int
    Email string
    Name  string
}

summaries := chankit.FromSlice(ctx, users).
    Filter(func(u User) bool { return u.Active }).
    Map(func(u User) any {
        return UserSummary{
            ID:    u.ID,
            Email: strings.ToLower(u.Email),
            Name:  strings.Title(u.Name),
        }
    }).
    ToSlice()
```

### With Validation

```go
import "net/mail"

validEmails := chankit.FromSlice(ctx, users).
    Filter(func(u User) bool { return u.Active }).
    Map(func(u User) any { return u.Email }).
    Filter(func(e any) bool {
        _, err := mail.ParseAddress(e.(string))
        return err == nil // Keep only valid emails
    }).
    ToSlice()
```

### Aggregation with Reduce

```go
// Count active users by domain
domainCounts := make(map[string]int)

chankit.FromSlice(ctx, users).
    Filter(func(u User) bool { return u.Active }).
    ForEach(func(u User) {
        parts := strings.Split(u.Email, "@")
        if len(parts) == 2 {
            domainCounts[parts[1]]++
        }
    })

fmt.Println(domainCounts) // {"example.com": 3}
```

### Deduplication

```go
seen := make(map[string]bool)

uniqueEmails := chankit.FromSlice(ctx, users).
    Map(func(u User) any { return u.Email }).
    Filter(func(e any) bool {
        email := e.(string)
        if seen[email] {
            return false // Already seen
        }
        seen[email] = true
        return true
    }).
    ToSlice()
```

## Performance Optimization

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-red-900/20 border border-red-500/50 rounded-lg p-6">

### ❌ Inefficient

```go
// Multiple passes over data
activeUsers := filterActive(users)
emails := extractEmails(activeUsers)
upper := toUppercase(emails)
batches := batchEmails(upper)
```

**Issues:**
- Multiple iterations
- Intermediate allocations
- Higher memory usage

</div>

<div class="bg-green-900/20 border border-green-500/50 rounded-lg p-6">

### ✅ Efficient

```go
// Single pipeline, streaming
chankit.FromSlice(ctx, users).
    Filter(isActive).
    Map(extractEmail).
    Map(toUpper).
    Batch(10, time.Second)
```

**Benefits:**
- Single pass
- Streaming processing
- Lower memory footprint

</div>

</div>

## Real-World Example: User Export

```go
type ExportRecord struct {
    ID        int       `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

func exportActiveUsers(ctx context.Context, users []User) error {
    file, err := os.Create("active_users.jsonl")
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)

    return chankit.FromSlice(ctx, users).
        Filter(func(u User) bool { return u.Active }).
        Map(func(u User) any {
            return ExportRecord{
                ID:        u.ID,
                Email:     strings.ToLower(u.Email),
                Name:      u.Name,
                CreatedAt: time.Now(),
            }
        }).
        Batch(100, time.Second).              // Write in batches
        ForEachErr(func(batch any) error {
            records := batch.([]ExportRecord)
            for _, record := range records {
                if err := encoder.Encode(record); err != nil {
                    return err
                }
            }
            return nil
        })
}
```

## Best Practices

<div class="space-y-4 my-6">

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Filter early</strong>
<p class="text-dark-text-muted">Apply filters before expensive transformations to reduce work</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Combine simple operations</strong>
<p class="text-dark-text-muted">Merge multiple Maps into one to reduce goroutine overhead</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Use batching for I/O</strong>
<p class="text-dark-text-muted">Batch operations before database writes or API calls</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Stream when possible</strong>
<p class="text-dark-text-muted">Use ForEach instead of ToSlice for large datasets</p>
</div>
</div>

</div>

## Related Examples

- [Data Pipeline](/examples/data-pipeline) - Basic transformations
- [Real-time Analytics](/examples/analytics) - Sensor data processing
- [Rate Limiting](/examples/rate-limiting) - Batch API calls

<div class="flex gap-4 my-8">

[Next: Real-time Analytics →](/examples/analytics){.custom-button}
[Previous: Rate Limiting ←](/examples/rate-limiting){.custom-button-outline}

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
