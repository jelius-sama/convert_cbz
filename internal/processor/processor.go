package processor

import (
    "archive/zip"
    "convert_cbz/internal/types"
    "convert_cbz/internal/util"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"

    "github.com/jelius-sama/logger"
)

func ProcessConcurrently(workItems []types.WorkItem, numThreads int, stats *types.ConversionStats) *types.SafeWriter {
    // Create work channel with buffer to prevent blocking
    workChan := make(chan types.WorkItem, numThreads)
    buf := &types.SafeWriter{}

    spinner := util.NewSpinner(stats, len(workItems))
    // Print 4 blank lines so first render has space to overwrite and to make it less cluttered
    fmt.Print("\n\n\n\n")
    spinner.Start()

    // Create wait group to track completion
    var wg sync.WaitGroup

    // Start worker goroutines
    for i := range numThreads {
        wg.Add(1)
        go worker(i+1, workChan, &wg, stats, buf)
    }

    // Send work items to channel
    go func() {
        defer close(workChan)
        for _, item := range workItems {
            workChan <- item
        }
    }()

    // Wait for all workers to complete
    wg.Wait()
    spinner.Stop()

    // flush buffer to disk in one shot
    // This might create problems in devices with low memory
    // TODO: Introduce a flag to disable logging.
    // We might also want to consider a flag that writes directly to a file on disk,
    // although that could introduce slowdowns as disk is slower than memory.
    if err := os.MkdirAll("/tmp/convert-cbz", 0755); err != nil {
        logger.Error(fmt.Sprintf("Failed to write log file: %v", err))
    } else {
        logFilePath := fmt.Sprintf("/tmp/convert-cbz/%s.log", time.Now().Format("2006-01-02-1504"))
        if err := os.WriteFile(logFilePath, buf.Buffer.Bytes(), 0644); err != nil {
            logger.Error(fmt.Sprintf("Failed to write log file: %v", err))
        } else {
            fmt.Println("\033[90m  log written → " + logFilePath + "\033[0m")
        }
    }
    return buf
}

func worker(id int, workChan <-chan types.WorkItem, wg *sync.WaitGroup, stats *types.ConversionStats, buf *types.SafeWriter) {
    defer wg.Done()

    for item := range workChan {
        // Process single conversion job
        processWorkItem(id, item, stats, buf)

        // Small delay to prevent overwhelming the system
        time.Sleep(5 * time.Millisecond)
    }
}

func processWorkItem(workerID int, item types.WorkItem, stats *types.ConversionStats, buf *types.SafeWriter) {
    prefix := fmt.Sprintf("[WORKER %d]", workerID)
    fmt.Fprintf(buf, "[INFO] %s Processing: %s\n", prefix, item.FolderName)

    // Check if output already exists
    if _, err := os.Stat(item.OutputPath); err == nil {
        fmt.Fprintf(buf, "[WARN] %s CBZ already exists, skipping: %s\n", prefix, filepath.Base(item.OutputPath))
        stats.Mutex.Lock()
        stats.Skipped++
        stats.Mutex.Unlock()
        return
    }

    // Convert folder to CBZ
    nonImageCount, err := convertToCBZ(item.SourcePath, item.OutputPath, item.DumbMode)
    if err != nil {
        fmt.Fprintf(buf, "[ERROR] %s Conversion failed: %v\n", prefix, err)
        stats.Mutex.Lock()
        stats.Errors++
        stats.Mutex.Unlock()
        return
    }

    // Update statistics
    stats.Mutex.Lock()
    stats.Success++
    stats.NonImageFiles += nonImageCount
    stats.Mutex.Unlock()

    fmt.Fprintf(buf, "[OK] %s Created: %s\n", prefix, filepath.Base(item.OutputPath))

    // Report non-image files if found
    if nonImageCount > 0 {
        fmt.Fprintf(buf, "[WARN] %s Found %d non-image files (excluded from CBZ)\n", prefix, nonImageCount)
    }
}

func convertToCBZ(sourceDir, cbzPath string, dumbMode bool) (int, error) {
    var includeFiles []string
    var excludedCount int

    if dumbMode {
        // DUMB MODE: Include all files without any filtering
        files, err := getAllFiles(sourceDir)
        if err != nil {
            return 0, fmt.Errorf("failed to scan directory: %w", err)
        }
        includeFiles = files
        excludedCount = 0
    } else {
        // SMART MODE: Intelligently filter files
        var err error
        includeFiles, excludedCount, err = getSmartFilteredFiles(sourceDir)
        if err != nil {
            return 0, fmt.Errorf("failed to analyze directory: %w", err)
        }
    }

    if len(includeFiles) == 0 {
        return 0, fmt.Errorf("no files found to archive")
    }

    // Create CBZ file (which is just a ZIP with .cbz extension)
    cbzFile, err := os.Create(cbzPath)
    if err != nil {
        return 0, fmt.Errorf("failed to create CBZ file: %w", err)
    }
    defer cbzFile.Close()

    // Create ZIP writer with compression
    zipWriter := zip.NewWriter(cbzFile)
    defer zipWriter.Close()

    // Add all selected files to the ZIP archive
    for _, filePath := range includeFiles {
        if err := addFileToZip(zipWriter, filePath, sourceDir); err != nil {
            return 0, fmt.Errorf("failed to add file to archive: %w", err)
        }
    }

    return excludedCount, nil
}

