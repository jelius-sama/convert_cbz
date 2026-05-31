package main

import (
    "convert_cbz/internal/processor"
    "convert_cbz/internal/types"
    "convert_cbz/internal/util"
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "time"

    "github.com/jelius-sama/logger"
)

const VERSION = "v2.1.1"

func main() {
    start := time.Now()
    // Command line argument parsing
    var (
        outputDir   string
        threads     int
        dumbMode    bool
        recursive   bool
        showHelp    bool
        showVersion bool
        inputPaths  types.StringSliceFlag
    )

    flag.StringVar(&outputDir, "output", "", "Output directory")
    flag.StringVar(&outputDir, "o", "", "Output directory")

    flag.IntVar(&threads, "threads", 4, "Number of concurrent threads")
    flag.IntVar(&threads, "t", 4, "Number of concurrent threads")
    flag.IntVar(&threads, "j", 4, "Number of concurrent threads")

    flag.BoolVar(&dumbMode, "dumb", false, "Archive all files without filtering")
    flag.BoolVar(&dumbMode, "d", false, "Archive all files without filtering")

    flag.BoolVar(&recursive, "recursive", false, "Process subdirectories recursively")
    flag.BoolVar(&recursive, "r", false, "Process subdirectories recursively")

    flag.BoolVar(&showHelp, "help", false, "Show usage information")
    flag.BoolVar(&showHelp, "h", false, "Show usage information")

    flag.BoolVar(&showVersion, "version", false, "Show version information")
    flag.BoolVar(&showVersion, "v", false, "Show version information")

    flag.Var(&inputPaths, "input", "Input directory/directories (can be specified multiple times)")
    flag.Var(&inputPaths, "i", "Input directory/directories (can be specified multiple times)")

    flag.Usage = showUsage
    flag.Parse()

    // Handle version flag
    if showVersion {
        fmt.Println("CBZ Converter " + VERSION)
        fmt.Println("Converts folders containing images to CBZ comic book archives")
        return
    }

    // Handle help flag or missing required arguments
    if showHelp || len(inputPaths) == 0 || outputDir == "" {
        showUsage()
        return
    }

    // Validate thread count - ensure reasonable bounds
    if threads < 1 {
        threads = 1
    } else if threads > runtime.NumCPU()*2 {
        // Limit to 2x CPU cores to prevent resource exhaustion
        // Too much CPU usage might end up triggering aggresive context switching,
        // which can hurt performance instead of increasing it
        threads = runtime.NumCPU() * 2
        logger.Info(fmt.Sprintf("Thread count limited to %d (2x CPU cores)", threads))
    }

    // Create output directory if it doesn't exist
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        logger.Fatal(fmt.Sprintf("Failed to create output directory: %v", err))
    }

    logger.Info(fmt.Sprintf("Starting CBZ conversion with %d threads", threads))
    logger.Info(fmt.Sprintf("Output: %s", outputDir))

    if dumbMode {
        logger.Info("Mode: DUMB - archiving all files without filtering")
    } else {
        logger.Info("Mode: SMART - filtering files intelligently")
    }

    if recursive {
        logger.Info("Mode: RECURSIVE - processing subdirectories")
    } else {
        logger.Info("Mode: DIRECT - converting specified directories only")
    }

    // Collect all work items based on input paths and mode
    var workItems []types.WorkItem
    var err error

    if recursive {
        // Recursive mode: scan each input path for subdirectories
        workItems, err = collectRecursiveWorkItems(inputPaths, outputDir, dumbMode)
    } else {
        // Direct mode: convert specified directories directly
        workItems, err = collectDirectWorkItems(inputPaths, outputDir, dumbMode)
    }

    if err != nil {
        logger.Fatal(fmt.Sprintf("Failed to collect work items: %v", err))
    }

    if len(workItems) == 0 {
        logger.Warning("No folders found to process")
        return
    }

    logger.Info(fmt.Sprintf("Found %d folders to process", len(workItems)))

    // Process folders concurrently
    stats := &types.ConversionStats{Total: len(workItems)}
    util.PrintFinalStats(stats, processor.ProcessConcurrently(workItems, threads, stats), time.Since(start))
}

// collectRecursiveWorkItems scans input directories for subdirectories (original behavior)
func collectRecursiveWorkItems(inputPaths []string, outputDir string, dumbMode bool) ([]types.WorkItem, error) {
    var workItems []types.WorkItem
    seenPaths := make(map[string]bool) // Prevent duplicates

    for _, inputPath := range inputPaths {
        // Validate input directory exists
        if _, err := os.Stat(inputPath); os.IsNotExist(err) {
            logger.Warning(fmt.Sprintf("Input directory does not exist, skipping: %s", inputPath))
            continue
        }

        // Get subdirectories
        folders, err := util.GetFolders(inputPath)
        if err != nil {
            logger.Warning(fmt.Sprintf("Failed to read directory %s: %v", inputPath, err))
            continue
        }

        logger.Info(fmt.Sprintf("Input: %s (%d subdirectories)", inputPath, len(folders)))

        // Create work items for each subdirectory
        for _, folder := range folders {
            sourcePath := filepath.Join(inputPath, folder)

            // Get absolute path to avoid duplicates
            absPath, err := filepath.Abs(sourcePath)
            if err != nil {
                logger.Warning(fmt.Sprintf("Failed to resolve path %s: %v", sourcePath, err))
                continue
            }

            // Skip if we've already seen this path
            if seenPaths[absPath] {
                continue
            }
            seenPaths[absPath] = true

            outputPath := filepath.Join(outputDir, folder+".cbz")

            workItems = append(workItems, types.WorkItem{
                FolderName: folder,
                SourcePath: absPath,
                OutputPath: outputPath,
                DumbMode:   dumbMode,
            })
        }
    }

    return workItems, nil
}

// collectDirectWorkItems converts specified directories directly
func collectDirectWorkItems(inputPaths []string, outputDir string, dumbMode bool) ([]types.WorkItem, error) {
    var workItems []types.WorkItem
    seenPaths := make(map[string]bool) // Prevent duplicates

    for _, inputPath := range inputPaths {
        // Validate input directory exists
        inputInfo, err := os.Stat(inputPath)
        if os.IsNotExist(err) {
            logger.Warning(fmt.Sprintf("Input path does not exist, skipping: %s", inputPath))
            continue
        }

        // Ensure it's a directory
        if !inputInfo.IsDir() {
            logger.Warning(fmt.Sprintf("Input path is not a directory, skipping: %s", inputPath))
            continue
        }

        // Get absolute path to avoid duplicates
        absPath, err := filepath.Abs(inputPath)
        if err != nil {
            logger.Warning(fmt.Sprintf("Failed to resolve path %s: %v", inputPath, err))
            continue
        }

        // Skip if we've already seen this path
        if seenPaths[absPath] {
            logger.Warning(fmt.Sprintf("Duplicate path, skipping: %s", inputPath))
            continue
        }
        seenPaths[absPath] = true

        // Generate output filename from directory name
        folderName := filepath.Base(absPath)
        outputPath := filepath.Join(outputDir, folderName+".cbz")

        logger.Info(fmt.Sprintf("Input: %s", inputPath))

        workItems = append(workItems, types.WorkItem{
            FolderName: folderName,
            SourcePath: absPath,
            OutputPath: outputPath,
            DumbMode:   dumbMode,
        })
    }

    return workItems, nil
}

