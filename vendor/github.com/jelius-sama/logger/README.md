# Logger

A high-performance, colorized logging library for Go applications with customizable output styles and comprehensive logging levels.

[![Go Reference](https://pkg.go.dev/badge/github.com/jelius-sama/logger.svg)](https://pkg.go.dev/github.com/jelius-sama/logger)
[![Go Report Card](https://goreportcard.com/badge/github.com/jelius-sama/logger)](https://goreportcard.com/report/github.com/jelius-sama/logger)

## Features

- 🎨 **Colorized Output** - Visual distinction between log levels with ANSI color codes
- ⚡ **High Performance** - Optimized for minimal system calls and memory efficiency
- 🔧 **Customizable Styles** - Switch between bracket `[INFO]` and colon `INFO:` formats
- ⏱️ **Timestamped Logging** - Optional timestamp prefix for all log levels
- 🛡️ **Graceful Panic Handling** - Proper defer execution before program termination
- 📊 **Multiple Log Levels** - Debug, Info, Warning, Error, Fatal, and Panic levels

## Installation

```bash
go get github.com/jelius-sama/logger
```

## Quick Start

```go
package main

import "github.com/jelius-sama/logger"

func main() {
    logger.Info("Application starting")
    logger.Warning("This is a warning message")
    logger.Error("Something went wrong")
    logger.Okay("Operation completed successfully")
}
```

## Log Levels

The logger provides six distinct log levels, each with its own color coding and behavior:

| Level | Color | Output | Behavior |
|-------|-------|--------|----------|
| `Debug` | Blue | stdout | Development information |
| `Info` | Cyan | stdout | General information |
| `Okay` | Green | stdout | Success messages |
| `Warning` | Yellow | stdout | Warning messages |
| `Error` | Red | stderr | Error messages |
| `Fatal` | Red | stderr | Exits program immediately with `os.Exit(-1)` |
| `Panic` | Red | stderr | Executes deferred functions, then panics |

### Basic Usage

```go
logger.Debug("Debugging application flow")
logger.Info("User logged in successfully")
logger.Okay("Database connection established")
logger.Warning("Deprecated function called")
logger.Error("Failed to connect to database")
logger.Fatal("Critical system failure")   // Exits immediately
logger.Panic("Unrecoverable error")        // Executes defers, then panics
```

## Timestamped Logging

All log levels have timestamped variants that automatically prepend the current timestamp:

```go
logger.TimedInfo("User action recorded")    // 2006/01/02 15:04:05 User action recorded
logger.TimedError("Database timeout")       // 2006/01/02 15:04:05 Database timeout
logger.TimedWarning("Cache miss")          // 2006/01/02 15:04:05 Cache miss
```

Available timed functions:
- `TimedDebug()`
- `TimedInfo()`
- `TimedOkay()`
- `TimedWarning()`
- `TimedError()`
- `TimedFatal()`
- `TimedPanic()`

## Output Styles

The logger supports two output styles that can be changed at runtime:

### Bracket Style (Default)
```go
logger.SetStyle("brackets")
logger.Info("Hello World")
// Output: [INFO] Hello World
```

### Colon Style
```go
logger.SetStyle("colon")
logger.Info("Hello World")
// Output: INFO: Hello World
```

## Fatal vs Panic

Understanding the difference between `Fatal` and `Panic` is crucial:

### Fatal
- Immediately terminates the program with `os.Exit(-1)`
- **Does NOT execute deferred functions**
- Use for unrecoverable errors where cleanup isn't possible

```go
func main() {
    defer fmt.Println("This will NOT be printed")
    logger.Fatal("Critical error")
}
```

### Panic
- Executes all deferred functions in LIFO order
- Unwinds the stack gracefully
- Can be recovered using `recover()`
- Use when you need cleanup before termination

```go
func main() {
    defer fmt.Println("This WILL be printed")
    logger.Panic("Recoverable error")
}
```

## Performance Characteristics

This logger is optimized for performance through several design decisions:

### Minimal System Calls
Instead of multiple `fmt.Print*` calls, the logger uses a single `append` operation followed by one system call. This approach trades a small amount of memory for significantly better performance:

```go
// Inefficient: 3 system calls
fmt.Print("[INFO] ")
fmt.Print("Your message")
fmt.Print("\n")

// Efficient: 1 system call
fmt.Println(append(append([]any{"[INFO]"}, a...), "\n")...)
```

### Memory Efficiency
The `append` strategy creates a single slice that gets passed to the underlying write system call, reducing:
- System call overhead
- Kernel-userspace transitions
- I/O buffer fragmentation

### Benchmarking
You can benchmark the logger performance:

```bash
go test -bench=.
```

## Advanced Usage

### Custom Error Handling
```go
func processData() {
    defer func() {
        if r := recover(); r != nil {
            logger.Error("Recovered from panic:", r)
        }
    }()
    
    // Your code here
    if criticalError {
        logger.Panic("Data corruption detected")
    }
}
```

### Conditional Logging
```go
const DEBUG = true

func debugLog(msg string) {
    if DEBUG {
        logger.Debug(msg)
    }
}
```

### Style Switching
```go
// Switch styles based on environment
if os.Getenv("LOG_STYLE") == "colon" {
    logger.SetStyle("colon")
} else {
    logger.SetStyle("brackets")
}
```

## Color Output

The logger uses ANSI escape codes for colorization:

- **Red (`\033[31m`)**: Error, Fatal, Panic
- **Blue (`\033[34m`)**: Debug
- **Cyan (`\033[0;36m`)**: Info
- **Green (`\033[32m`)**: Okay
- **Yellow (`\033[33m`)**: Warning
- **Reset (`\033[0m`)**: Returns to default color

Colors automatically work in most modern terminals. If your terminal doesn't support colors, the escape codes will be visible as text.

## Thread Safety

**Important**: This logger is **not thread-safe** by design to maintain maximum performance. If you need concurrent logging, consider:

1. Using a mutex around logger calls
2. Implementing a channel-based logging system
3. Using separate logger instances per goroutine

## Examples

### Basic Web Server Logging
```go
package main

import (
    "net/http"
    "github.com/jelius-sama/logger"
)

func handler(w http.ResponseWriter, r *http.Request) {
    logger.TimedInfo("Request:", r.Method, r.URL.Path)
    
    // Process request...
    
    logger.TimedOkay("Response sent for:", r.URL.Path)
}

func main() {
    logger.SetStyle("colon")
    logger.Info("Starting web server on :8080")
    
    http.HandleFunc("/", handler)
    if err := http.ListenAndServe(":8080", nil); err != nil {
        logger.Fatal("Server failed to start:", err)
    }
}
```

### Database Connection with Retry
```go
func connectDB() {
    maxRetries := 3
    
    for i := 0; i < maxRetries; i++ {
        logger.TimedInfo("Attempting database connection, try", i+1)
        
        if err := db.Connect(); err != nil {
            logger.TimedWarning("Connection failed:", err)
            if i == maxRetries-1 {
                logger.Fatal("Max retries exceeded, giving up")
            }
            continue
        }
        
        logger.TimedOkay("Database connected successfully")
        return
    }
}
```

## Testing

Run the test suite:

```bash
# Run all tests
go test -v

# Run with coverage
go test -cover

# Run benchmarks
go test -bench=.
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test -v`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Performance Notes

- Single system call per log message
- Minimal memory allocations
- ANSI color codes add ~10 bytes per message
- Timestamp formatting uses Go's optimized time package
- No external dependencies beyond Go standard library

## Changelog
### v1.0.2
- `TimedX` now uses UTC time universally.

### v1.0.1
- Added a `Configure` function to configure current environment (`dev` VS `prod`).
- Debug logs only gets logged during `dev` mode.

### v1.0.0
- Initial release
- Basic logging levels with colorization
- Bracket and colon output styles
- Timestamped logging variants
- Comprehensive test suite
- Performance optimizations
