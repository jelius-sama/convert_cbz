package processor

import (
    "archive/zip"
    "compress/flate"
    "convert_cbz/internal/types"
    "io"
    "os"
    "path/filepath"
    "sync"
)

var (
    compression     types.CompressionMode
    compressionOnce sync.Once
)

// Once program starts there's no way to change compression mode so just cache it
func getCompression() types.CompressionMode {
    compressionOnce.Do(func() {
        compression = types.ToCompressionMode(os.Getenv(types.CKey.String()))
    })
    return compression
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
    compression := getCompression()

    switch compression {
    case types.CMDefault:
        header.Method = zip.Deflate
        zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
            return flate.NewWriter(out, flate.DefaultCompression)
        })

    case types.CMFast:
        header.Method = zip.Deflate
        zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
            return flate.NewWriter(out, flate.BestSpeed)
        })

    case types.CMSlow:
        header.Method = zip.Deflate
        zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
            return flate.NewWriter(out, flate.BestCompression)
        })

    default:
        header.Method = zip.Store
    }

    // Create ZIP entry
    writer, err := zipWriter.CreateHeader(header)
    if err != nil {
        return err
    }

    // Copy file content to ZIP entry
    _, err = io.Copy(writer, sourceFile)
    return err
}

