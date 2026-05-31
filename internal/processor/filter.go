package processor

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "slices"
    "sort"
    "strings"

    "github.com/jelius-sama/logger"
)

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

