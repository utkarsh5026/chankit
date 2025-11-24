# Real-Time Sensor Data Analytics

Process high-frequency sensor readings, debounce them to reduce noise, transform values, and calculate running averages.

## Scenario

You're building an IoT system that:
1. Receives high-frequency sensor readings (10ms intervals)
2. Debounces to reduce noise and fluctuations
3. Transforms raw sensor values
4. Calculates statistics (averages, trends)

This pattern is common in IoT, monitoring systems, and real-time analytics dashboards.

## Implementation

```go
package main

import (
    "context"
    "fmt"
    "math/rand"
    "time"
    "github.com/utkarsh5026/chankit/chankit"
)

type SensorReading struct {
    SensorID  string
    Value     float64
    Timestamp time.Time
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    // Simulate sensor data stream
    readings := make(chan SensorReading, 100)
    go func() {
        defer close(readings)
        for {
            select {
            case <-ctx.Done():
                return
            case readings <- SensorReading{
                SensorID:  "temp-01",
                Value:     20 + rand.Float64()*10,
                Timestamp: time.Now(),
            }:
                time.Sleep(10 * time.Millisecond)
            }
        }
    }()

    // Process: debounce → transform to Celsius → calculate average
    processed := chankit.From(ctx, readings).
        Debounce(100 * time.Millisecond).                 // Reduce noise
        Map(func(r SensorReading) any {                   // Transform value
            return r.Value * 1.5 // Simulate transformation
        }).
        Take(10)                                          // Take first 10

    // Calculate average
    sum := 0.0
    count := 0
    processed.ForEach(func(v any) {
        sum += v.(float64)
        count++
        fmt.Printf("Reading #%d: %.2f\n", count, v.(float64))
    })

    if count > 0 {
        fmt.Printf("\nAverage: %.2f\n", sum/float64(count))
    }
}
```

## Output

```
Reading #1: 37.45
Reading #2: 39.82
Reading #3: 41.23
Reading #4: 38.91
Reading #5: 40.56
Reading #6: 42.18
Reading #7: 37.89
Reading #8: 41.45
Reading #9: 39.23
Reading #10: 40.78

Average: 39.95
```

## How It Works

<div class="space-y-4 my-6">

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 1: Sensor Data Generation
Simulates a temperature sensor sending readings every 10ms (100 readings/second).

```go
readings <- SensorReading{
    SensorID:  "temp-01",
    Value:     20 + rand.Float64()*10, // 20-30°C
    Timestamp: time.Now(),
}
```

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 2: Debounce (100ms)
**Debounce(100 * time.Millisecond)** reduces noise by waiting for stable readings.
- Filters out rapid fluctuations
- Emits value only after 100ms of stability
- Reduces data points from 100/sec to ~10/sec

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 3: Transform Values
**Map(func(r SensorReading) any { return r.Value * 1.5 })** applies calibration/transformation.
- Convert units (e.g., raw → calibrated)
- Apply scaling factors
- Normalize sensor readings

</div>

<div class="bg-dark-bg-soft border-l-4 border-primary-500 p-4 rounded">

### Step 4: Limit & Calculate
**Take(10)** limits to 10 readings, then calculates the average.

</div>

</div>

## Use Cases

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 🌡️ Temperature Monitoring
- HVAC systems
- Server room monitoring
- Industrial equipment
- Weather stations

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 💓 Health Monitoring
- Heart rate sensors
- Blood pressure monitors
- Fitness trackers
- Medical devices

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 🏭 Industrial IoT
- Pressure sensors
- Flow meters
- Vibration analysis
- Quality control

</div>

<div class="bg-dark-bg-soft border border-dark-border rounded-lg p-6">

### 🚗 Vehicle Telematics
- Speed monitoring
- Fuel consumption
- Engine diagnostics
- GPS tracking

</div>

</div>

## Analytics Patterns

### Rolling Average

```go
windowSize := 5
window := make([]float64, 0, windowSize)

chankit.From(ctx, readings).
    Map(func(r SensorReading) any {
        window = append(window, r.Value)
        if len(window) > windowSize {
            window = window[1:] // Slide window
        }

        // Calculate average
        sum := 0.0
        for _, v := range window {
            sum += v
        }
        return sum / float64(len(window))
    })
```

### Anomaly Detection

```go
threshold := 100.0
baseline := 50.0

chankit.From(ctx, readings).
    Map(func(r SensorReading) any {
        deviation := math.Abs(r.Value - baseline)
        if deviation > threshold {
            return Alert{
                SensorID: r.SensorID,
                Value:    r.Value,
                Message:  "Anomaly detected",
            }
        }
        return nil
    }).
    Filter(func(v any) bool { return v != nil })
```

### Trend Detection

```go
type Trend string

const (
    Rising  Trend = "rising"
    Falling Trend = "falling"
    Stable  Trend = "stable"
)

lastValue := 0.0

chankit.From(ctx, readings).
    Map(func(r SensorReading) any {
        var trend Trend
        if r.Value > lastValue+1.0 {
            trend = Rising
        } else if r.Value < lastValue-1.0 {
            trend = Falling
        } else {
            trend = Stable
        }
        lastValue = r.Value
        return TrendData{Value: r.Value, Trend: trend}
    })
```

## Advanced Patterns

### Multi-Sensor Aggregation

```go
type SensorID string

const (
    Temp1 SensorID = "temp-01"
    Temp2 SensorID = "temp-02"
    Temp3 SensorID = "temp-03"
)

// Collect readings from multiple sensors
sensorData := make(map[SensorID][]float64)

chankit.From(ctx, readings).
    Debounce(100 * time.Millisecond).
    ForEach(func(r SensorReading) {
        sensorData[SensorID(r.SensorID)] = append(
            sensorData[SensorID(r.SensorID)],
            r.Value,
        )
    })

// Calculate per-sensor averages
for id, values := range sensorData {
    avg := calculateAverage(values)
    fmt.Printf("Sensor %s average: %.2f\n", id, avg)
}
```

### Time-Windowed Statistics

```go
type Stats struct {
    Min    float64
    Max    float64
    Avg    float64
    Count  int
    Window time.Time
}

windowDuration := time.Minute

chankit.From(ctx, readings).
    Batch(100, windowDuration).
    Map(func(batch any) any {
        readings := batch.([]SensorReading)

        stats := Stats{
            Min:    math.MaxFloat64,
            Max:    -math.MaxFloat64,
            Window: time.Now(),
        }

        sum := 0.0
        for _, r := range readings {
            stats.Min = math.Min(stats.Min, r.Value)
            stats.Max = math.Max(stats.Max, r.Value)
            sum += r.Value
            stats.Count++
        }

        if stats.Count > 0 {
            stats.Avg = sum / float64(stats.Count)
        }

        return stats
    })
```

### Alert System

```go
type AlertLevel string

const (
    Normal   AlertLevel = "normal"
    Warning  AlertLevel = "warning"
    Critical AlertLevel = "critical"
)

func getAlertLevel(value float64) AlertLevel {
    switch {
    case value < 15:
        return Critical // Too cold
    case value < 20:
        return Warning  // Cold
    case value > 35:
        return Critical // Too hot
    case value > 30:
        return Warning  // Hot
    default:
        return Normal
    }
}

chankit.From(ctx, readings).
    Debounce(100 * time.Millisecond).
    Map(func(r SensorReading) any {
        return struct {
            Reading SensorReading
            Alert   AlertLevel
        }{
            Reading: r,
            Alert:   getAlertLevel(r.Value),
        }
    }).
    Filter(func(v any) bool {
        // Only emit warnings and critical alerts
        data := v.(struct {
            Reading SensorReading
            Alert   AlertLevel
        })
        return data.Alert != Normal
    }).
    ForEach(func(v any) {
        // Handle alert
        sendAlert(v)
    })
```

## Performance Optimization

<div class="grid md:grid-cols-2 gap-6 my-8">

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-6">

### Memory-Efficient Streaming

```go
// Good: Process without accumulation
chankit.From(ctx, readings).
    Debounce(100 * time.Millisecond).
    ForEach(processReading) // O(1) memory
```

</div>

<div class="bg-dark-bg-soft border border-primary-500/30 rounded-lg p-6">

### Batched Processing

```go
// Good: Process in chunks
chankit.From(ctx, readings).
    Batch(100, time.Second).
    ForEach(func(batch any) {
        // Process 100 readings at once
        saveToDB(batch)
    })
```

</div>

</div>

## Real-World Example: Temperature Dashboard

```go
type DashboardUpdate struct {
    Current     float64
    Average     float64
    Min         float64
    Max         float64
    TrendStatus Trend
    Timestamp   time.Time
}

func monitorTemperature(ctx context.Context, readings <-chan SensorReading) <-chan DashboardUpdate {
    window := make([]float64, 0, 10)
    lastValue := 0.0

    return chankit.From(ctx, readings).
        Debounce(500 * time.Millisecond).
        Map(func(r SensorReading) any {
            // Update window
            window = append(window, r.Value)
            if len(window) > 10 {
                window = window[1:]
            }

            // Calculate statistics
            sum, min, max := 0.0, r.Value, r.Value
            for _, v := range window {
                sum += v
                min = math.Min(min, v)
                max = math.Max(max, v)
            }

            // Determine trend
            var trend Trend
            if r.Value > lastValue+1.0 {
                trend = Rising
            } else if r.Value < lastValue-1.0 {
                trend = Falling
            } else {
                trend = Stable
            }
            lastValue = r.Value

            return DashboardUpdate{
                Current:     r.Value,
                Average:     sum / float64(len(window)),
                Min:         min,
                Max:         max,
                TrendStatus: trend,
                Timestamp:   r.Timestamp,
            }
        }).
        ToChannel() // Return as channel for dashboard updates
}
```

## Best Practices

<div class="space-y-4 my-6">

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Debounce high-frequency sensors</strong>
<p class="text-dark-text-muted">Reduce noise and processing load for sensors that update rapidly</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Use rolling windows for trends</strong>
<p class="text-dark-text-muted">Maintain a sliding window of recent values for accurate trend analysis</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Implement circuit breakers</strong>
<p class="text-dark-text-muted">Protect against sensor failures and invalid readings</p>
</div>
</div>

<div class="flex items-start gap-4 bg-dark-bg-soft border border-dark-border rounded-lg p-4">
<span class="text-2xl">✅</span>
<div>
<strong>Batch database writes</strong>
<p class="text-dark-text-muted">Use Batch() to write sensor data efficiently instead of per-reading</p>
</div>
</div>

</div>

## Related Examples

- [Event Stream](/examples/event-stream) - Event processing patterns
- [Data Pipeline](/examples/data-pipeline) - Batch processing
- [Data Transformation](/examples/transformation) - Data cleaning

<div class="flex gap-4 my-8">

[Back to Examples](/examples/data-pipeline){.custom-button}
[Previous: Transformation ←](/examples/transformation){.custom-button-outline}

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
