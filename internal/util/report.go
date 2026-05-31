package util

import (
    "convert_cbz/internal/types"
    "fmt"
    "math"
    "strings"
    "time"
)

type VisualLine struct {
    raw     strings.Builder
    visible int
}

func (v *VisualLine) Add(s, ansi string) *VisualLine {
    v.raw.WriteString(ansi)
    v.raw.WriteString(s)
    v.visible += len([]rune(s))
    return v
}

func (v *VisualLine) String() string { return v.raw.String() }

var ansiReset = "\033[0m"
var ansiBold = "\033[1m"
var ansiPurple = "\033[35m"
var ansiGreen = "\033[32m"
var ansiYellow = "\033[33m"
var ansiRed = "\033[31m"
var ansiMuted = "\033[90m"

func newLine() *VisualLine { return &VisualLine{} }

func (v *VisualLine) Plain(s string) *VisualLine     { return v.Add(s, "") }
func (v *VisualLine) Styled(s, a string) *VisualLine { return v.Add(s, a+ansiBold) }
func (v *VisualLine) Muted(s string) *VisualLine     { return v.Add(s, ansiMuted) }
func (v *VisualLine) Color(s, a string) *VisualLine  { return v.Add(s, a) }

func box(content *VisualLine, W int) string {
    pad := W - content.visible
    if pad < 0 {
        pad = 0
    }
    return "│ " + content.String() + ansiReset + strings.Repeat(" ", pad) + " │"
}

func PrintFinalStats(stats *types.ConversionStats, buf *types.SafeWriter, elapsed time.Duration) {
    stats.Mutex.Lock()
    defer stats.Mutex.Unlock()

    buf.Mutex.Lock()
    logContent := buf.Buffer.String()
    buf.Mutex.Unlock()

    var failures []struct{ name, reason string }
    for _, line := range strings.Split(logContent, "\n") {
        if !strings.HasPrefix(line, "[ERROR]") {
            continue
        }
        if idx := strings.Index(line, "Conversion failed: "); idx != -1 {
            reason := line[idx+len("Conversion failed: "):]
            name := ""
            parts := strings.SplitN(line, "] ", 3)
            if len(parts) == 3 {
                name = strings.TrimSpace(parts[1])
            }
            failures = append(failures, struct{ name, reason string }{name, reason})
        }
    }

    processed := stats.Success + stats.Errors
    if processed > 0 {
        stats.Success = stats.Success / processed * 100
    } else if stats.Skipped == stats.Total {
        // If all conversion were skipped, we count it as success
        stats.Success = stats.Total
    }

    const W = 60

    hr := func(l, r string) string {
        return l + strings.Repeat("─", W+2) + r
    }
    top := hr("┌", "┐")
    mid := hr("├", "┤")
    bot := hr("└", "┘")

    filledCount := func(n int) int {
        if stats.Total == 0 {
            return 0
        }
        f := int(float64(20) * float64(n) / float64(stats.Total))
        if f > 20 {
            return 20
        }
        return f
    }

    makeBar := func(label string, color string, n int) *VisualLine {
        f := filledCount(n)
        pct := 0
        if stats.Total > 0 {
            pct = int(math.Round(float64(n) / float64(stats.Total) * 100))
        }
        l := newLine()
        l.Plain(fmt.Sprintf("%-8s ", label))       // 9 visible chars
        l.Color(strings.Repeat("█", f), color)     // f visible chars
        l.Muted(strings.Repeat("░", 20-f))         // 20-f visible chars
        l.Color(fmt.Sprintf(" %3d%%", pct), color) // 5 visible chars
        return l                                   // total: 9+20+5 = 34
    }

    elapsedStr := FmtDuration(elapsed)

    // Header
    h := newLine()
    h.Styled("CONVERSION COMPLETE", ansiPurple)
    h.Muted("  done in " + elapsedStr)
    fmt.Println(top)
    fmt.Println(box(h, W))
    fmt.Println(mid)

    // Metric labels
    lb := newLine()
    lb.Muted(fmt.Sprintf("%-13s%-13s%-13s%s", "TOTAL", "OK", "SKIPPED", "ERRORS"))
    fmt.Println(box(lb, W))

    // Metric values
    v := newLine()
    v.Styled(fmt.Sprintf("%-13d", stats.Total), ansiPurple)
    v.Styled(fmt.Sprintf("%-13d", stats.Success), ansiGreen)
    v.Styled(fmt.Sprintf("%-13d", stats.Skipped), ansiYellow)
    v.Styled(fmt.Sprintf("%d", stats.Errors), ansiRed)
    fmt.Println(box(v, W))
    fmt.Println(mid)

    // Bars
    // Always show success rate bar

    fmt.Println(box(makeBar("success", ansiGreen, stats.Success), W))

    if stats.Skipped > 0 {
        fmt.Println(box(makeBar("skipped", ansiYellow, stats.Skipped), W))
    }
    if stats.Errors > 0 {
        fmt.Println(box(makeBar("errors", ansiRed, stats.Errors), W))
    }
    if stats.NonImageFiles > 0 {
        fmt.Println(box(makeBar("excluded", ansiMuted, stats.NonImageFiles), W))
    }

    // Failures
    if len(failures) > 0 {
        fmt.Println(mid)
        fh := newLine()
        fh.Styled("✗ failed conversions", ansiRed)
        fmt.Println(box(fh, W))
        for _, f := range failures {
            name := TruncateString(f.name, 32)
            reason := TruncateString(f.reason, 14)
            fl := newLine()
            fl.Color("✗ ", ansiRed)
            fl.Plain(fmt.Sprintf("%-32s ", name))
            fl.Muted(reason)
            fmt.Println(box(fl, W))
        }
    }

    // Footer
    logStr := "git@git.jelius.dev:jelius-sama/convert_cbz.git"
    pad := W - len([]rune(logStr))
    if pad < 0 {
        pad = 0
    }

    fmt.Println(mid)
    ft := newLine()
    ft.Muted(logStr)
    ft.Plain(strings.Repeat(" ", pad))
    fmt.Println(box(ft, W))
    fmt.Println(bot)
}

