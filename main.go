package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"github.com/jelius-sama/logger"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

const VERSION = "v2.0.0"

// ConversionStats tracks overall conversion statistics
type ConversionStats struct {
	mu            sync.Mutex
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

func main() {
	// Command line argument parsing
	var inputPaths StringSliceFlag
	var (
		outputDir   = flag.String("output", "", "Output directory for CBZ files (required)")
		threads     = flag.Int("threads", 4, "Number of concurrent threads")
		dumbMode    = flag.Bool("dumb", false, "Archive all files without filtering (default: smart filtering)")
		recursive   = flag.Bool("recursive", false, "Process subdirectories recursively (default: direct conversion)")
		showHelp    = flag.Bool("help", false, "Show usage information")
		showVersion = flag.Bool("version", false, "Show version information")
	)

	flag.Var(&inputPaths, "input", "Input directory/directories (can be specified multiple times)")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Println("CBZ Converter " + VERSION)
		fmt.Println("Converts folders containing images to CBZ comic book archives")
		return
	}

	// Handle help flag or missing required arguments
	if *showHelp || len(inputPaths) == 0 || *outputDir == "" {
		showUsage()
		return
	}

	// Validate thread count - ensure reasonable bounds
	if *threads < 1 {
		*threads = 1
	} else if *threads > runtime.NumCPU()*2 {
		// Limit to 2x CPU cores to prevent resource exhaustion
		*threads = runtime.NumCPU() * 2
		logger.Info(fmt.Sprintf("Thread count limited to %d (2x CPU cores)", *threads))
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		logger.Fatal(fmt.Sprintf("Failed to create output directory: %v", err))
	}

	logger.Info(fmt.Sprintf("Starting CBZ conversion with %d threads", *threads))
	logger.Info(fmt.Sprintf("Output: %s", *outputDir))

	if *dumbMode {
		logger.Info("Mode: DUMB - archiving all files without filtering")
	} else {
		logger.Info("Mode: SMART - filtering files intelligently")
	}

	if *recursive {
		logger.Info("Mode: RECURSIVE - processing subdirectories")
	} else {
		logger.Info("Mode: DIRECT - converting specified directories only")
	}

	// Collect all work items based on input paths and mode
	var workItems []WorkItem
	var err error

	if *recursive {
		// Recursive mode: scan each input path for subdirectories
		workItems, err = collectRecursiveWorkItems(inputPaths, *outputDir, *dumbMode)
	} else {
		// Direct mode: convert specified directories directly
		workItems, err = collectDirectWorkItems(inputPaths, *outputDir, *dumbMode)
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
	stats := &ConversionStats{Total: len(workItems)}
	processConcurrently(workItems, *threads, stats)

	// Print final statistics
	printFinalStats(stats)
}

func showUsage() {
	fmt.Println("CBZ Converter - Convert image folders to CBZ comic book archives")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Printf("  %s -input <dir> [-input <dir>...] -output <folder> [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("REQUIRED:")
	fmt.Println("  -input   string    Input directory (can be specified multiple times)")
	fmt.Println("  -output  string    Output directory for CBZ files")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -recursive        Process subdirectories recursively (default: false)")
	fmt.Println("  -threads int      Number of concurrent threads (default: 4)")
	fmt.Println("  -dumb            Archive all files without filtering (default: false)")
	fmt.Println("  -help            Show this help message")
	fmt.Println("  -version         Show version information")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  1. Recursive Mode:")
	fmt.Println("     Process every subdirectory inside root folders:")
	fmt.Printf("       %s -recursive -input ./mangas -output ./cbz\n", os.Args[0])
	fmt.Printf("       %s -recursive -input ./mangas -input ./fav-mangas -output ./cbz\n", os.Args[0])
	fmt.Println()
	fmt.Println("  2. Direct Mode (single specific folder):")
	fmt.Println("     Convert only the specified directory:")
	fmt.Printf("       %s -input \"./mangas/some manga\" -output ./cbz\n", os.Args[0])
	fmt.Println()
	fmt.Println("  3. Direct Mode (multiple specific folders):")
	fmt.Println("     Convert each specified directory (no recursion):")
	fmt.Printf("       %s -input \"./mangas/some manga\" -input \"./mangas/some manga part 2\" -output ./cbz\n", os.Args[0])
	fmt.Println()
	fmt.Println("  4. With additional options:")
	fmt.Printf("       %s -recursive -threads 8 -input ./mangas -output ./cbz\n", os.Args[0])
	fmt.Printf("       %s -dumb -input \"./raw/chapter 1\" -output ./archives\n", os.Args[0])
	fmt.Println()
	fmt.Println("MODES:")
	fmt.Println("  RECURSIVE (-recursive):")
	fmt.Println("    Scans input directories and converts each subdirectory into a CBZ")
	fmt.Println("    Example: ./mangas/ contains [manga1/, manga2/, manga3/]")
	fmt.Println("             → Creates manga1.cbz, manga2.cbz, manga3.cbz")
	fmt.Println()
	fmt.Println("  DIRECT (default):")
	fmt.Println("    Converts the specified directories directly into CBZ files")
	fmt.Println("    Example: -input \"./mangas/manga1\"")
	fmt.Println("             → Creates manga1.cbz (only this folder)")
	fmt.Println()
	fmt.Println("  SMART (default):")
	fmt.Println("    Intelligently filters files to include:")
	fmt.Println("      • Image files (JPEG, PNG, GIF, WebP, HEIF, etc.)")
	fmt.Println("      • Text files (TXT, MD, NFO - metadata)")
	fmt.Println("      • Video files (MP4, AVI, MKV - supplementary content)")
	fmt.Println("      • Excludes: system files (.DS_Store, Thumbs.db), VCS (.git, .svn)")
	fmt.Println()
	fmt.Println("  DUMB (-dumb):")
	fmt.Println("    Archives everything without any filtering")
}

// collectRecursiveWorkItems scans input directories for subdirectories (original behavior)
func collectRecursiveWorkItems(inputPaths []string, outputDir string, dumbMode bool) ([]WorkItem, error) {
	var workItems []WorkItem
	seenPaths := make(map[string]bool) // Prevent duplicates

	for _, inputPath := range inputPaths {
		// Validate input directory exists
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			logger.Warning(fmt.Sprintf("Input directory does not exist, skipping: %s", inputPath))
			continue
		}

		// Get subdirectories
		folders, err := getFolders(inputPath)
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

			workItems = append(workItems, WorkItem{
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
func collectDirectWorkItems(inputPaths []string, outputDir string, dumbMode bool) ([]WorkItem, error) {
	var workItems []WorkItem
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

		workItems = append(workItems, WorkItem{
			FolderName: folderName,
			SourcePath: absPath,
			OutputPath: outputPath,
			DumbMode:   dumbMode,
		})
	}

	return workItems, nil
}

func getFolders(dir string) ([]string, error) {
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

func processConcurrently(workItems []WorkItem, numThreads int, stats *ConversionStats) {
	// Create work channel with buffer to prevent blocking
	workChan := make(chan WorkItem, numThreads)

	// Create wait group to track completion
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := range numThreads {
		wg.Add(1)
		go worker(i+1, workChan, &wg, stats)
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
}

func worker(id int, workChan <-chan WorkItem, wg *sync.WaitGroup, stats *ConversionStats) {
	defer wg.Done()

	for item := range workChan {
		// Process single conversion job
		processWorkItem(id, item, stats)

		// Small delay to prevent overwhelming the system
		time.Sleep(5 * time.Millisecond)
	}
}

func processWorkItem(workerID int, item WorkItem, stats *ConversionStats) {
	prefix := fmt.Sprintf("[WORKER %d]", workerID)

	logger.Info(fmt.Sprintf("%s Processing: %s", prefix, truncateString(item.FolderName, 60)))

	// Check if output already exists
	if _, err := os.Stat(item.OutputPath); err == nil {
		logger.Warning(fmt.Sprintf("%s CBZ already exists, skipping: %s", prefix, filepath.Base(item.OutputPath)))
		stats.mu.Lock()
		stats.Skipped++
		stats.mu.Unlock()
		return
	}

	// Convert folder to CBZ
	nonImageCount, err := convertToCBZ(item.SourcePath, item.OutputPath, item.DumbMode)
	if err != nil {
		logger.Error(fmt.Sprintf("%s Conversion failed: %v", prefix, err))
		stats.mu.Lock()
		stats.Errors++
		stats.mu.Unlock()
		return
	}

	// Update statistics
	stats.mu.Lock()
	stats.Success++
	stats.NonImageFiles += nonImageCount
	stats.mu.Unlock()

	logger.Okay(fmt.Sprintf("%s Created: %s", prefix, filepath.Base(item.OutputPath)))

	// Report non-image files if found
	if nonImageCount > 0 {
		logger.Warning(fmt.Sprintf("%s Found %d non-image files (excluded from CBZ)", prefix, nonImageCount))
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

// getAllFiles gets all files in directory for DUMB mode (no filtering)
func getAllFiles(dir string) ([]string, error) {
	var allFiles []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Include all files, skip only directories
		if !d.IsDir() {
			allFiles = append(allFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort files for consistent ordering
	sort.Strings(allFiles)
	return allFiles, nil
}

// getSmartFilteredFiles intelligently filters files for SMART mode
func getSmartFilteredFiles(dir string) ([]string, int, error) {
	var includedFiles []string
	var excludedFiles []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		fileName := d.Name()

		// Check if file should be excluded (system files, VCS, etc.)
		if shouldExcludeFile(fileName) {
			excludedFiles = append(excludedFiles, fileName)
			return nil
		}

		// For remaining files, check if they're useful content
		isUseful, err := isUsefulFile(path)
		if err != nil {
			// If we can't determine, include it (fail-safe approach)
			logger.Warning(fmt.Sprintf("Could not analyze file %s, including anyway", fileName))
			includedFiles = append(includedFiles, path)
		} else if isUseful {
			includedFiles = append(includedFiles, path)
		} else {
			excludedFiles = append(excludedFiles, fileName)
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	// Sort files for consistent ordering
	sort.Strings(includedFiles)
	return includedFiles, len(excludedFiles), nil
}

// shouldExcludeFile checks for obvious system/VCS files to exclude
func shouldExcludeFile(fileName string) bool {
	fileName = strings.ToLower(fileName)

	// System files
	systemFiles := []string{
		".ds_store", "thumbs.db", "desktop.ini", ".directory",
		"folder.jpg", "albumartsmall.jpg", ".picasa.ini",
	}

	if exits := slices.Contains(systemFiles, fileName); exits == true {
		return true
	}

	// VCS directories/files
	vcsPatterns := []string{
		".git", ".svn", ".hg", ".bzr",
		".gitignore", ".gitattributes", ".hgignore",
	}

	for _, pattern := range vcsPatterns {
		if strings.Contains(fileName, pattern) {
			return true
		}
	}

	// IDE/Editor files
	idePatterns := []string{
		".vscode", ".idea", ".sublime-",
		"*.swp", "*.swo", "*~",
	}

	for _, pattern := range idePatterns {
		if strings.Contains(fileName, pattern) {
			return true
		}
	}

	return false
}

// isUsefulFile determines if a file is useful content for comic archives
func isUsefulFile(filePath string) (bool, error) {
	// First check by extension for quick decisions
	ext := strings.ToLower(filepath.Ext(filePath))

	// Text files that might contain metadata
	textExtensions := map[string]bool{
		".txt": true, ".md": true, ".nfo": true, ".info": true,
		".readme": true, ".description": true, ".notes": true,
	}

	if textExtensions[ext] {
		return true, nil
	}

	// Video files that might be supplementary content
	videoExtensions := map[string]bool{
		".mp4": true, ".avi": true, ".mkv": true, ".mov": true,
		".wmv": true, ".flv": true, ".webm": true, ".m4v": true,
	}

	if videoExtensions[ext] {
		return true, nil
	}

	// For files without clear extensions, use MIME detection
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read first 512 bytes for MIME type detection
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}

	mimeType := http.DetectContentType(buffer)

	// Include images, text, and video content
	usefulMimeTypes := []string{"image/", "text/", "video/"}

	for _, prefix := range usefulMimeTypes {
		if strings.HasPrefix(mimeType, prefix) {
			return true, nil
		}
	}

	return false, nil
}

func addFileToZip(zipWriter *zip.Writer, filePath, baseDir string) error {
	// Calculate relative path for the ZIP entry
	// This preserves the directory structure within the archive
	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return err
	}

	// Convert to forward slashes for ZIP standard compliance
	relPath = filepath.ToSlash(relPath)

	// Open source file
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Get file information for archive header
	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	// Create ZIP file header
	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	// Set compression method and file path
	header.Name = relPath
	header.Method = zip.Deflate // Use compression to reduce file size

	// Create ZIP entry
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to ZIP entry
	_, err = io.Copy(writer, sourceFile)
	return err
}

func printFinalStats(stats *ConversionStats) {
	stats.mu.Lock()
	defer stats.mu.Unlock()

	logger.Info("Conversion completed")
	logger.Info(fmt.Sprintf("Total folders:     %d", stats.Total))
	logger.Okay(fmt.Sprintf("Successful:        %d", stats.Success))

	if stats.Skipped > 0 {
		logger.Warning(fmt.Sprintf("Skipped:           %d", stats.Skipped))
	}

	if stats.Errors > 0 {
		logger.Error(fmt.Sprintf("Errors:            %d", stats.Errors))
	}

	if stats.NonImageFiles > 0 {
		logger.Info(fmt.Sprintf("Files excluded:    %d (smart filtering)", stats.NonImageFiles))
	}

	// Calculate success rate
	processed := stats.Success + stats.Errors
	if processed > 0 {
		successRate := float64(stats.Success) / float64(processed) * 100
		logger.Info(fmt.Sprintf("Success rate:      %.1f%%", successRate))
	}
}

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	// Use ellipsis to indicate truncation
	return string(runes[:maxLen-3]) + "..."
}
