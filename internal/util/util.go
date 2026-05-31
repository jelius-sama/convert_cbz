package util

import (
    "fmt"
    "os"
    "sort"
    "time"
)

func TruncateString(s string, maxLen int) string {
    runes := []rune(s)
    if len(runes) <= maxLen {
        return s
    }
    if maxLen <= 1 {
        return string(runes[:maxLen])
    }
    // Use ellipsis to indicate truncation
    return string(runes[:maxLen-1]) + "…"
}

func GetFolders(dir string) ([]string, error) {
    var folders []string

    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil, err
    }

    for _, entry := range entries {
        if entry.IsDir() {
            folders = append(folders, entry.Name())
        }
    }

    // Sort for consistent processing order
    sort.Strings(folders)
    return folders, nil
}

func FmtDuration(d time.Duration) string {
    if d < time.Second {
        return "<1s"
    }
    s := int(d.Seconds())
    if s < 60 {
        return fmt.Sprintf("%ds", s)
    }
    return fmt.Sprintf("%dm%ds", s/60, s%60)
}

