package types

import (
    "bytes"
    "strings"
    "sync"
)

// ConversionStats tracks overall conversion statistics
type ConversionStats struct {
    Mutex         sync.Mutex
    Total         int
    Success       int
    Errors        int
    Skipped       int
    NonImageFiles int
}

// WorkItem represents a single conversion job
type WorkItem struct {
    FolderName string
    SourcePath string
    OutputPath string
    DumbMode   bool
}

// StringSliceFlag allows multiple string flags
type StringSliceFlag []string

func (s *StringSliceFlag) String() string {
    return strings.Join(*s, ", ")
}

func (s *StringSliceFlag) Set(value string) error {
    *s = append(*s, value)
    return nil
}

type SafeWriter struct {
    Mutex  sync.Mutex
    Buffer bytes.Buffer
}

func (w *SafeWriter) Write(p []byte) (n int, err error) {
    w.Mutex.Lock()
    defer w.Mutex.Unlock()
    return w.Buffer.Write(p)
}

