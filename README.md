# capitan

[![CI Status](https://github.com/zoobzio/capitan/workflows/CI/badge.svg)](https://github.com/zoobzio/capitan/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/zoobzio/capitan/graph/badge.svg?branch=main)](https://codecov.io/gh/zoobzio/capitan)
[![Go Report Card](https://goreportcard.com/badge/github.com/zoobzio/capitan)](https://goreportcard.com/report/github.com/zoobzio/capitan)
[![CodeQL](https://github.com/zoobzio/capitan/workflows/CodeQL/badge.svg)](https://github.com/zoobzio/capitan/security/code-scanning)
[![Go Reference](https://pkg.go.dev/badge/github.com/zoobzio/capitan.svg)](https://pkg.go.dev/github.com/zoobzio/capitan)
[![License](https://img.shields.io/github/license/zoobzio/capitan)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/zoobzio/capitan)](go.mod)
[![Release](https://img.shields.io/github/v/release/zoobzio/capitan)](https://github.com/zoobzio/capitan/releases)

Type-safe event coordination for Go with zero dependencies.

Emit events with typed fields, hook listeners, and let capitan handle the rest with async processing and backpressure.

## The Power of Simplicity

At its core, capitan provides just three operations:

```go
// Emit an event
capitan.Emit(signal, fields...)

// Hook a listener
listener := capitan.Hook(signal, func(e *capitan.Event) {
    // Handle event
})

// Observe all signals
observer := capitan.Observe(func(e *capitan.Event) {
    // Handle any event
})
```

**That's it.** No schemas to declare, no complex configuration, just events and listeners.

## Quick Start

```go
package main

import (
    "github.com/zoobzio/capitan"
)

func main() {
    // Define signal and keys
    orderCreated := capitan.Signal("order.created")
    orderID := capitan.NewStringKey("order_id")
    total := capitan.NewFloat64Key("total")

    // Hook a listener
    capitan.Hook(orderCreated, func(e *capitan.Event) {
        id, _ := orderID.From(e)
        amount, _ := total.From(e)
        // Process order...
    })

    // Emit an event (async with backpressure)
    capitan.Emit(orderCreated,
        orderID.Field("ORDER-123"),
        total.Field(99.99),
    )

    // Gracefully drain pending events
    capitan.Shutdown()
}
```

## Why capitan?

- **Type-safe**: Typed fields with compile-time safety
- **Zero dependencies**: Just standard library
- **Async by default**: Non-blocking emission with per-signal worker goroutines and backpressure
- **Lazy**: Workers created only when needed
- **Isolated**: Slow listeners don't affect other signals
- **Panic-safe**: Listener panics recovered, system stays running
- **Clean**: No schemas, no boilerplate, no registration ceremony
- **Testable**: Every component independently testable

## Installation

```bash
go get github.com/zoobzio/capitan
```

Requirements: Go 1.24+

## Core Concepts

**Signals** identify event types:
```go
userLogin := capitan.Signal("user.login")
orderShipped := capitan.Signal("order.shipped")
```

**Keys** define typed field names:
```go
userID := capitan.NewStringKey("user_id")
count := capitan.NewIntKey("count")
ratio := capitan.NewFloat64Key("ratio")
active := capitan.NewBoolKey("active")
```

**Fields** carry typed values:
```go
capitan.Emit(userLogin,
    userID.Field("user_123"),
    count.Field(5),
)
```

**Listeners** handle events:
```go
listener := capitan.Hook(userLogin, func(e *capitan.Event) {
    uid := userID.From(e)
    // Handle login...
})
```

**Observers** watch all signals (dynamic):
```go
observer := capitan.Observe(func(e *capitan.Event) {
    // Log all events
})
```

Observers receive events from both existing signals and any signals created after the observer is registered. This is compatible with lazy signal initialization - observers automatically attach to workers as they're created.

### Best Practice: Define Signals and Keys as Constants

**Always define signals and keys as package-level constants:**

```go
// Define signals at package level
const (
    UserLogin    = capitan.Signal("user.login")
    UserLogout   = capitan.Signal("user.logout")
    OrderCreated = capitan.Signal("order.created")
)

// Define keys at package level
var (
    userID   = capitan.NewStringKey("user_id")
    orderID  = capitan.NewStringKey("order_id")
    total    = capitan.NewFloat64Key("total")
)
```

**Why?**

1. **Bounded memory** - Capitan maintains internal registries keyed by signal. A fixed set of signals means predictable memory usage. Dynamic signal creation (e.g., per-user, per-request) will cause unbounded registry growth.

2. **Worker lifecycle** - Each signal gets its own worker goroutine. Dynamic signals create goroutines that persist until explicit cleanup, leading to goroutine leaks in long-running systems.

3. **Predictable behavior** - Const signals make event flow analyzable at compile time. Observers and hooks can be reasoned about statically.

4. **Performance** - Internal lookups are optimized for a stable signal set. Dynamic signals bypass lazy initialization benefits.

**Avoid dynamic signals:**
```go
// BAD: Creates new signal per user
signal := capitan.Signal(fmt.Sprintf("user.%s.login", userID))

// GOOD: Use fields to carry dynamic data
capitan.Emit(UserLogin, userID.Field(id))
```

Use fields to carry variable data, not signal names. Think of signals as event _types_, not event _instances_.

## Real-World Example

Here's a realistic example showing how capitan handles application events:

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/zoobzio/capitan"
)

// Define signals as constants
const (
    orderCreated = capitan.Signal("order.created")
    orderShipped = capitan.Signal("order.shipped")
)

// Define keys
var (
    orderID   = capitan.NewStringKey("order_id")
    userID    = capitan.NewStringKey("user_id")
    total     = capitan.NewFloat64Key("total")
    timestamp = capitan.NewStringKey("timestamp")
)

func main() {
    // Setup logging observer for all events
    capitan.Observe(func(e *capitan.Event) {
        log.Printf("[EVENT] %s at %s",
            e.Signal(),
            e.Timestamp().Format(time.RFC3339))
    })

    // Hook order created handler
    capitan.Hook(orderCreated, func(e *capitan.Event) {
        id, _ := orderID.From(e)
        amount, _ := total.From(e)

        // Send confirmation email
        fmt.Printf("📧 Sending confirmation for order %s (%.2f)\n", id, amount)

        // Update analytics
        fmt.Printf("📊 Recording order metrics\n")
    })

    // Hook order shipped handler
    capitan.Hook(orderShipped, func(e *capitan.Event) {
        id, _ := orderID.From(e)
        user, _ := userID.From(e)

        // Send shipping notification
        fmt.Printf("📦 Notifying user %s: order %s shipped\n", user, id)
    })

    // Emit events (async with backpressure)
    capitan.Emit(orderCreated,
        orderID.Field("ORDER-123"),
        userID.Field("user_456"),
        total.Field(99.99),
    )

    capitan.Emit(orderShipped,
        orderID.Field("ORDER-123"),
        userID.Field("user_456"),
    )

    // Gracefully drain queue before exit
    capitan.Shutdown()
}
```

## Architecture

**Async Execution**: Each signal gets its own worker goroutine, created lazily on first emission. Events are queued (16 buffer default) and processed asynchronously.

**Isolation**: Slow or misbehaving listeners on one signal don't affect other signals. Each signal's queue operates independently.

**Panic Recovery**: Listener panics are caught and recovered silently. One bad listener won't crash your system or prevent other listeners from running.

**Event Pooling**: Events are pooled internally to reduce allocations. Events are returned to the pool after all listeners finish.

**Shutdown**: `Shutdown()` closes all worker goroutines gracefully, processing remaining queued events before exit.

## Concurrency & Ordering

**Per-Signal Ordering**: Events emitted to the same signal are processed in emission order. Each signal's worker processes its queue sequentially.

**Cross-Signal Independence**: No ordering guarantees between different signals. Workers operate concurrently and independently.

```go
Emit("order.created", orderID.Field("123"))
Emit("email.sent", orderID.Field("123"))
Shutdown()

// order.created's listeners might complete before OR after email.sent's listeners
// Each signal processes independently
```

**Shutdown Behavior**: `Shutdown()` waits for all workers to drain their queues, but workers complete independently. Events queued at shutdown time will be processed before exit.

**Backpressure**: Each signal has a buffered queue (16 events by default). If the queue fills, `Emit()` blocks until space is available. This provides natural backpressure - slow listeners will slow down emitters for that signal only, preventing unbounded memory growth. Other signals are unaffected.

## Multiple Instances

While the module-level API uses a default singleton, you can create isolated instances:

```go
c := capitan.New()

c.Hook(signal, handler)
c.Emit(signal, fields...)
c.Shutdown()
```

## Listener Management

**Close individual listeners**:
```go
listener := capitan.Hook(signal, handler)
// ...later
listener.Close() // Stop receiving events
```

**Close observers**:
```go
observer := capitan.Observe(handler)
// ...later
observer.Close() // Stop all observer listeners
```

## Field Types

Capitan supports four primitive field types:

- `StringKey` / `StringField` - string values
- `IntKey` / `IntField` - int values
- `Float64Key` / `Float64Field` - float64 values
- `BoolKey` / `BoolField` - bool values

Access typed values using the From() method:
```go
value, ok := key.From(e)  // Returns (value, ok) tuple
if !ok {
    // Field not present or wrong type
}
```

Or via type assertion on fields:
```go
field := e.Get(key)
if sf, ok := field.(capitan.StringField); ok {
    value := sf.String()
}
```

## Event Access

```go
// Get specific field
field := e.Get(key)

// Get all fields
fields := e.Fields() // Returns []Field

// Access metadata
signal := e.Signal()       // Signal identifier
timestamp := e.Timestamp() // When event was created
```

## Performance

Capitan is designed for performance:

- Lazy worker creation (no upfront cost)
- Event pooling reduces allocations
- Per-signal goroutines prevent contention
- Minimal locking on hot paths
- Zero reflection or runtime type assertions

Run benchmarks:
```bash
go test -bench=.
```

## Testing

Run tests:
```bash
go test -v ./...
```

Run with coverage:
```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Contributing

Contributions welcome! Please ensure:
- Tests pass: `go test ./...`
- Code is formatted: `go fmt ./...`
- No lint errors: `golangci-lint run`

## License

MIT License - see [LICENSE](LICENSE) file for details.
