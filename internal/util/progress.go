package util

import (
    "convert_cbz/internal/types"
    "fmt"
    "strings"
    "sync/atomic"
    "time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⼼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
    stats   *types.ConversionStats
    total   int
    current atomic.Value // current item name
    done    chan struct{}
}

func NewSpinner(stats *types.ConversionStats, total int) *Spinner {
    s := &Spinner{
        stats: stats,
        total: total,
        done:  make(chan struct{}),
    }
    s.current.Store("")
    return s
}

func (s *Spinner) SetCurrent(name string) {
    s.current.Store(name)
}

func (s *Spinner) Start() {
    go func() {
        start := time.Now()
        frame := 0
        // Hide cursor
        fmt.Print("\033[?25l")
        defer fmt.Print("\033[?25h") // restore on exit

        for {
            select {
            case <-s.done:
                s.render(frame, time.Since(start), true)
                fmt.Println()
                return
            default:
                s.render(frame, time.Since(start), false)
                frame = (frame + 1) % len(spinnerFrames)
                time.Sleep(80 * time.Millisecond)
            }
        }
    }()
}

func (s *Spinner) Stop() {
    close(s.done)
    time.Sleep(120 * time.Millisecond) // let final render flush
}

func (s *Spinner) render(frame int, elapsed time.Duration, final bool) {
    s.stats.Mutex.Lock()
    done := s.stats.Success + s.stats.Errors + s.stats.Skipped
    success := s.stats.Success
    errors := s.stats.Errors
    s.stats.Mutex.Unlock()

    sp := spinnerFrames[frame]
    pct := 0.0
    if s.total > 0 {
        pct = float64(done) / float64(s.total) * 100
    }

    // ETA
    eta := ""
    if done > 0 && done < s.total {
        perItem := elapsed / time.Duration(done)
        remaining := perItem * time.Duration(s.total-done)
        eta = fmt.Sprintf("  eta %s", FmtDuration(remaining))
    }

    // Progress bar (30 chars wide)
    const barWidth = 30
    filled := int(float64(barWidth) * float64(done) / float64(s.total))
    if filled > barWidth {
        filled = barWidth
    }
    bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

    // Status counts
    counts := fmt.Sprintf("\033[32m✓ %d ok\033[0m", success)
    if errors > 0 {
        counts += fmt.Sprintf("  \033[31m✗ %d failed\033[0m", errors)
    }

    // Current item
    current := s.current.Load().(string)
    currentLine := ""
    if !final && current != "" {
        currentLine = fmt.Sprintf("\n  \033[2m%s  %s\033[0m", sp, current)
    }

    prefix := fmt.Sprintf("\033[35m%s\033[0m", sp)
    if final {
        prefix = "\033[32m✓\033[0m"
        eta = fmt.Sprintf("  done in %s", FmtDuration(elapsed))
    }

    // Move cursor up to overwrite previous render (3 lines)
    fmt.Print("\033[3A\033[J")
    fmt.Printf(
        "%s converting \033[35m%d/%d\033[0m folders\n  \033[35m%s\033[0m \033[90m%3.0f%%%s\033[0m\n  %s%s\n",
        prefix, done, s.total,
        bar, pct, eta,
        counts, currentLine,
    )
}

