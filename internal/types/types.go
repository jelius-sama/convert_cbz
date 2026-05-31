package types

import (
    "bytes"
    "strings"
    "sync"

    "github.com/jelius-sama/logger"
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

type CompressionMode uint8

const (
    CMNone CompressionMode = iota
    CMDefault
    CMFast
    CMSlow
    CKey
)

func (cm CompressionMode) Set(value string) error {
    cm = ToCompressionMode(value)
    return nil
}

func ToCompressionMode(cm string) CompressionMode {
    switch cm {
    case CMDefault.String():
        return CMDefault
    case CMDefault.String():
        return CMNone
    case CMFast.String():
        return CMFast
    case CMSlow.String():
        return CMSlow
    case CKey.String():
        return CKey
    default:
        logger.Warning("Undefined compression mode used, defaulting to \"none\".")
        return CMNone
    }
}

func (cm CompressionMode) String() string {
    switch cm {
    case CMDefault:
        return "default"
    case CMNone:
        return "none"
    case CMFast:
        return "fast"
    case CMSlow:
        return "slow"
    case CKey:
        return "compression"
    default:
        logger.Warning("Undefined compression mode used, defaulting to \"none\".")
        return "none"
    }
}

