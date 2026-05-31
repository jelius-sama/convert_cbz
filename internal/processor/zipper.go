package processor

import (
    "archive/zip"
    "io"
    "os"
    "path/filepath"
)

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

