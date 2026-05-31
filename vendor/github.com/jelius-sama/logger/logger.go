// Package logger provides high-performance, colorized logging functionality
// with customizable output styles and comprehensive logging levels.
//
// The logger supports two output styles: "brackets" ([INFO]) and "colon" (INFO:),
// and provides both regular and timestamped variants of all logging functions.
// It's optimized for performance by using single system calls per log message
// rather than multiple print operations.
//
// Thread Safety: This logger is NOT thread-safe by design for maximum performance.
// Use appropriate synchronization mechanisms if logging from multiple goroutines.
package logger

import (
    "fmt"
    "os"
    "strings"
    "time"
)

var (
    // LoggerStyle defines the current output format style.
    // Valid values are "brackets" for [LEVEL] format and "colon" for LEVEL: format.
    // Defaults to "brackets" style.
    LoggerStyle string = "brackets"

    isDebugMode *bool
)

// Configure sets up debug mode detection (call once at startup)
func Configure(envVar, devValue string) {
    enabled := os.Getenv(envVar) == devValue
    isDebugMode = &enabled

    if enabled {
        Info("DEBUG MODE ENABLED")
        Error("If you see this in production, STOP immediately!")
    }
}

// SetStyle changes the logger output format style.
// Accepts "brackets" for [LEVEL] format or "colon" for LEVEL: format.
// Any invalid style will default to "brackets" with a warning message.
// The style change is applied globally to all subsequent log messages.
//
// Example:
//
//	SetStyle("colon")    // Changes to "INFO: message" format
//	SetStyle("brackets") // Changes to "[INFO] message" format
func SetStyle(s string) {
    switch s {
    case "brackets":
        LoggerStyle = "brackets"
        Okay("Logger style set to `" + LoggerStyle + "`.")
        return

    case "colon":
        LoggerStyle = "colon"
        Okay("Logger style set to `" + LoggerStyle + "`.")
        return

    default:
        LoggerStyle = "brackets"
        Warning("Logger style " + s + " does not exists, setting to default instead!")
        return
    }
}

// applyStyle formats a label according to the current LoggerStyle setting.
// This is an internal helper function that wraps labels with brackets or colons.
// It takes a format string and a label, returning the formatted result.
// Falls back to brackets format if an invalid LoggerStyle is encountered.
func applyStyle(format string, label string) string {
    switch LoggerStyle {
    case "brackets":
        return fmt.Sprintf(format, "["+label+"]")

    case "colon":
        return fmt.Sprintf(format, label+":")

    default:
        Error("Unreachable code reached!")
        return fmt.Sprintf(format, "["+label+"]")
    }
}

// Error logs an error message to stderr with red coloring.
// Messages are prefixed with [ERROR] or ERROR: depending on the current style.
// Uses a single system call for optimal performance by combining all output
// elements into one slice before writing.
//
// Example:
//
//	Error("Database connection failed")
//	Error("Invalid input:", userInput, "expected number")
func Error(a ...any) {
    fmt.Fprintln(os.Stderr, append(append([]any{applyStyle("\033[31m%s", "ERROR")}, a...), []any{"\033[0m"}...)...)
}

// Debug logs a debug message to stdout with blue coloring.
// Messages are prefixed with [DEBUG] or DEBUG: depending on the current style.
// Intended for development and troubleshooting information.
//
// Example:
//
//	Debug("Processing user request")
//	Debug("Variable value:", someVar)
func Debug(a ...any) {
    // INFO: Only if we are in dev mode we print the debug logs.
    if isDebugMode == nil {
        Error("Debug not configured! Call Configure() before using Debug()")
        return
    }

    if *isDebugMode {
        fmt.Println(append(append([]any{applyStyle("\033[34m%s", "DEBUG")}, a...), []any{"\033[0m"}...)...)
    }
}

// Fatal logs a fatal error message to stderr with red coloring and immediately
// terminates the program with exit code -1.
// WARNING: Does NOT execute deferred functions - use Panic() if cleanup is needed.
// Messages are prefixed with [FATAL] or FATAL: depending on the current style.
//
// Example:
//
//	Fatal("Critical system failure - cannot continue")
func Fatal(a ...any) {
    fmt.Fprintln(os.Stderr, append(append([]any{applyStyle("\033[31m%s", "FATAL")}, a...), []any{"\033[0m"}...)...)
    os.Exit(-1)
}

// Panic logs a panic message to stderr with red coloring and triggers a panic.
// Unlike Fatal(), this DOES execute deferred functions before termination.
// The panic can be recovered using recover() if needed.
// Messages are prefixed with [PANIC] or PANIC: depending on the current style.
//
// Example:
//
//	Panic("Unrecoverable error occurred")
//
//	// With defer (will execute):
//	defer cleanup()
//	Panic("Something went wrong")  // cleanup() will run
func Panic(a ...any) {
    // Print the formatted panic message to stderr first
    fmt.Fprintln(os.Stderr, append(append([]any{applyStyle("\033[31m%s", "PANIC")}, a...), []any{"\033[0m"}...)...)

    // Create panic message and trigger panic
    panic(strings.TrimSuffix(fmt.Sprintln(a...), "\n"))
}

// Info logs an informational message to stdout with cyan coloring.
// Messages are prefixed with [INFO] or INFO: depending on the current style.
// Used for general application information and status updates.
//
// Example:
//
//	Info("Application started successfully")
//	Info("Processing", itemCount, "items")
func Info(a ...any) {
    fmt.Println(append(append([]any{applyStyle("\033[0;36m%s", "INFO")}, a...), []any{"\033[0m"}...)...)
}

// Okay logs a success message to stdout with green coloring.
// Messages are prefixed with [OK] or OK: depending on the current style.
// Used to indicate successful operations and positive status updates.
//
// Example:
//
//	Okay("Database connection established")
//	Okay("File saved successfully")
func Okay(a ...any) {
    fmt.Println(append(append([]any{applyStyle("\033[32m%s", "OK")}, a...), []any{"\033[0m"}...)...)
}

// Warning logs a warning message to stdout with yellow coloring.
// Messages are prefixed with [WARN] or WARN: depending on the current style.
// Used for non-critical issues that should be brought to attention.
//
// Example:
//
//	Warning("Configuration file not found, using defaults")
//	Warning("API rate limit approaching")
func Warning(a ...any) {
    fmt.Println(append(append([]any{applyStyle("\033[33m%s", "WARN")}, a...), []any{"\033[0m"}...)...)
}

// TimedError logs an error message with a timestamp prefix.
// Combines the current timestamp (YYYY/MM/DD HH:MM:SS format) with the error message.
// Outputs to stderr with red coloring like Error().
//
// Example:
//
//	TimedError("Connection timeout")
//	// Output: [ERROR] 2006/01/02 15:04:05 Connection timeout
func TimedError(a ...any) {
    Error(append([]any{time.Now().UTC().Format("2006/01/02 15:04:05")}, a...)...)
}

// TimedDebug logs a debug message with a timestamp prefix.
// Combines the current timestamp (YYYY/MM/DD HH:MM:SS format) with the debug message.
// Outputs to stdout with blue coloring like Debug().
//
// Example:
//
//	TimedDebug("Cache miss for key:", key)
//	// Output: [DEBUG] 2006/01/02 15:04:05 Cache miss for key: user123
func TimedDebug(a ...any) {
    Debug(append([]any{time.Now().UTC().Format("2006/01/02 15:04:05")}, a...)...)
}

// TimedFatal logs a fatal error message with a timestamp prefix and exits.
// Combines the current timestamp (YYYY/MM/DD HH:MM:SS format) with the fatal message.
// Immediately terminates the program without executing deferred functions.
//
// Example:
//
//	TimedFatal("System corruption detected")
//	// Output: [FATAL] 2006/01/02 15:04:05 System corruption detected
func TimedFatal(a ...any) {
    Fatal(append([]any{time.Now().UTC().Format("2006/01/02 15:04:05")}, a...)...)
}

// TimedPanic logs a panic message with a timestamp prefix and triggers panic.
// Combines the current timestamp (YYYY/MM/DD HH:MM:SS format) with the panic message.
// Executes deferred functions before program termination, unlike TimedFatal().
//
// Example:
//
//	TimedPanic("Critical state reached")
//	// Output: [PANIC] 2006/01/02 15:04:05 Critical state reached
func TimedPanic(a ...any) {
    Panic(append([]any{time.Now().UTC().Format("2006/01/02 15:04:05")}, a...)...)
}

// TimedInfo logs an informational message with a timestamp prefix.
// Combines the current timestamp (YYYY/MM/DD HH:MM:SS format) with the info message.
// Outputs to stdout with cyan coloring like Info().
//
// Example:
//
//	TimedInfo("User login successful")
//	// Output: [INFO] 2006/01/02 15:04:05 User login successful
func TimedInfo(a ...any) {
    Info(append([]any{time.Now().UTC().Format("2006/01/02 15:04:05")}, a...)...)
}

// TimedOkay logs a success message with a timestamp prefix.
// Combines the current timestamp (YYYY/MM/DD HH:MM:SS format) with the success message.
// Outputs to stdout with green coloring like Okay().
//
// Example:
//
//	TimedOkay("Backup completed")
//	// Output: [OK] 2006/01/02 15:04:05 Backup completed
func TimedOkay(a ...any) {
    Okay(append([]any{time.Now().UTC().Format("2006/01/02 15:04:05")}, a...)...)
}

// TimedWarning logs a warning message with a timestamp prefix.
// Combines the current timestamp (YYYY/MM/DD HH:MM:SS format) with the warning message.
// Outputs to stdout with yellow coloring like Warning().
//
// Example:
//
//	TimedWarning("Disk space low")
//	// Output: [WARN] 2006/01/02 15:04:05 Disk space low
func TimedWarning(a ...any) {
    Warning(append([]any{time.Now().UTC().Format("2006/01/02 15:04:05")}, a...)...)
}

