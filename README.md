# CBZ Converter

A high-performance, concurrent tool for converting folders containing images into CBZ (Comic Book Archive) files. Built in Go for speed and reliability.

![convert_cbz logo](https://jelius.dev/assets/compressed/convert_cbz.webp)

## Features

- **High Performance**: Multi-threaded processing with configurable concurrency
- **Flexible Input**: Support for both recursive directory scanning and direct folder conversion
- **Smart Detection**: MIME-type based image detection (supports JPEG, PNG, GIF, WebP, HEIF, AVIF, and more)
- **Cross-Platform**: Runs on Linux, macOS, Windows, and various Unix systems
- **Professional Logging**: Color-coded output with detailed progress tracking
- **Safe Operations**: Skips existing files, handles errors gracefully
- **Comprehensive Reporting**: Detailed statistics and non-image file detection

## Installation

### Quick Build

For your current system:
```bash
./build.sh
```

For all supported architectures:
```bash
./build.sh all
```

The compiled binaries will be placed in the `./bin/` directory.

### Manual Build

If you prefer to build manually:
```bash
go build -o convert_cbz main.go
```

## Usage

### Basic Syntax
```bash
convert-cbz -input <input_path> [-input <input_path>...] -output <output_folder> [options]
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-input` | Input directory (can be specified multiple times) | *required* |
| `-output` | Output directory for CBZ files | *required* |
| `-recursive` | Process subdirectories recursively | `false` |
| `-threads` | Number of concurrent processing threads | `4` |
| `-dumb` | Archive all files without filtering | `false` (smart mode) |
| `-help` | Show usage information | - |
| `-version` | Show version information | - |

## Processing Modes

### Recursive Mode
Scans input directories and converts each subdirectory into a separate CBZ file. This is the original behavior for batch processing multiple manga/comic folders.

**Example structure:**
```
./mangas/
├── Manga Title 1/
├── Manga Title 2/
└── Manga Title 3/
```

**Usage:**
```bash
convert-cbz -recursive -input ./mangas -output ./cbz
```

**Result:** Creates `Manga Title 1.cbz`, `Manga Title 2.cbz`, `Manga Title 3.cbz`

### Direct Mode (Default)
Converts specified directories directly into CBZ files without recursion. Perfect for converting specific folders or when you want precise control.

**Usage:**
```bash
# Single folder
convert-cbz -input "./mangas/Manga Title 1" -output ./cbz

# Multiple folders
convert-cbz -input "./mangas/Manga Title 1" -input "./mangas/Manga Title 2" -output ./cbz
```

**Result:** Creates CBZ files only for the specified directories

### Multiple Input Directories
Both modes support multiple input paths:

```bash
# Recursive mode with multiple sources
convert-cbz -recursive -input ./mangas -input ./fav-mangas -output ./cbz

# Direct mode with multiple sources
convert-cbz -input ./folder1 -input ./folder2 -input ./folder3 -output ./cbz
```

## Examples

### Recursive Processing (Batch Conversion)
```bash
# Process all subdirectories in manga folder
convert-cbz -recursive -input ./manga -output ./cbz

# Process multiple source directories
convert-cbz -recursive -input ./manga -input ./comics -output ./cbz

# With custom thread count
convert-cbz -recursive -threads 8 -input ./manga -output ./cbz
```

### Direct Processing (Specific Folders)
```bash
# Convert a single specific folder
convert-cbz -input "./manga/One Piece Chapter 1" -output ./cbz

# Convert multiple specific folders
convert-cbz -input "./manga/Chapter 1" -input "./manga/Chapter 2" -output ./cbz

# Using dumb mode for complete archiving
convert-cbz -dumb -input "./raw/chapter 1" -output ./archives
```

### Advanced Usage
```bash
# High-performance recursive processing
convert-cbz -recursive -threads 16 -input ./large_collection -output ./cbz

# Archive everything without filtering
convert-cbz -recursive -dumb -input ./raw_scans -output ./archives

# Process specific chapters with smart filtering
convert-cbz -input "./Ch1" -input "./Ch2" -input "./Ch3" -output ./out
```

## Content Filtering Modes

### Smart Mode (Default)
**Includes:**
- **Images**: JPEG, PNG, GIF, WebP, HEIF, AVIF, BMP, TIFF
- **Text files**: TXT, MD, NFO, INFO (metadata/descriptions)  
- **Video files**: MP4, AVI, MKV, MOV (supplementary content)
- **Any content with image/text/video MIME types**

**Excludes:**
- **System files**: .DS_Store, Thumbs.db, desktop.ini
- **Version control**: .git, .svn, .hg directories and files
- **IDE files**: .vscode, .idea, .sublime-project
- **Temporary files**: .swp, .swo, *~ backup files

### Dumb Mode (`-dumb`)
**Includes:** Everything - all files and folders are archived without any filtering whatsoever

**Use cases:**
- Preserving complete directory structures
- Archiving mixed content where filtering might remove needed files
- When you want maximum control over what gets included

## How It Works

1. **Input Processing**: 
   - **Recursive Mode**: Scans input directories for subdirectories
   - **Direct Mode**: Uses specified directories directly
2. **Content Analysis**: 
   - **Smart Mode**: Uses MIME type analysis and filename patterns to identify useful content
   - **Dumb Mode**: Includes all files without any filtering
3. **Archive Creation**: Creates compressed ZIP archives with `.cbz` extension
4. **Concurrent Processing**: Distributes work across multiple threads for optimal performance
5. **Progress Reporting**: Provides real-time feedback with colored logging

## Supported Content Types

### Smart Mode Detection
- **Images**: Automatic MIME type detection for all formats
- **Text**: Extensions (.txt, .md, .nfo) + text/* MIME types  
- **Video**: Extensions (.mp4, .avi, .mkv) + video/* MIME types
- **Unknown**: Fail-safe inclusion for unidentifiable files

### All Formats Include
- **Image formats**: JPEG, PNG, GIF, BMP, TIFF, WebP, HEIF, AVIF
- **Video formats**: MP4, AVI, MKV, MOV, WMV, FLV, WebM
- **Text formats**: TXT, MD, NFO, INFO, README

## Output Structure

### Recursive Mode
```
input/
├── Manga Title 1/
│   ├── page001.jpg
│   ├── page002.png
│   └── info.txt
└── Manga Title 2/
    ├── 01.jpg
    └── 02.jpg

output/
├── Manga Title 1.cbz
└── Manga Title 2.cbz
```

### Direct Mode
```
input/
└── mangas/
    ├── Chapter 1/
    │   └── pages...
    └── Chapter 2/
        └── pages...

Command: convert-cbz -input "./mangas/Chapter 1" -output ./cbz

output/
└── Chapter 1.cbz
```

## Logging and Feedback

The tool provides professional logging with color-coded output:

- **[INFO]** - General information (blue)
- **[OK]** - Successful operations (green)  
- **[WARN]** - Warnings and skipped items (yellow)
- **[ERROR]** - Error conditions (red)

### Sample Output
```
[INFO] Starting CBZ conversion with 4 threads
[INFO] Output: ./cbz
[INFO] Mode: SMART - filtering files intelligently
[INFO] Mode: RECURSIVE - processing subdirectories
[INFO] Input: ./manga (199 subdirectories)
[INFO] Found 199 folders to process
[WORKER 1] Processing: [Author] Title Chapter 1
[OK] [WORKER 1] Created: Title Chapter 1.cbz
[WARN] [WORKER 2] Found 2 non-image files (excluded from CBZ)
...
[INFO] Conversion completed
[INFO] Total folders:     199
[OK] Successful:        197
[WARN] Skipped:           2
[INFO] Files excluded:    15 (smart filtering)
[INFO] Success rate:      100.0%
```

## Performance Considerations

- **Thread Count**: Default is 4 threads. Increase for faster processing on multi-core systems
- **Memory Usage**: Each worker uses minimal memory; safe to run many threads
- **I/O Optimization**: Uses buffered operations and compression for efficiency
- **Resource Limits**: Automatically caps threads at 2× CPU cores to prevent system overload

## Error Handling

The tool handles various error conditions gracefully:

- **Missing directories**: Clear error messages with warnings for invalid paths
- **Permission issues**: Skips inaccessible files with warnings
- **Corrupted files**: Uses fail-safe approach to include ambiguous files
- **Existing files**: Skips existing CBZ files to prevent overwriting
- **Individual failures**: Continues processing other folders if one fails
- **Duplicate paths**: Detects and skips duplicate input directories

## Technical Details

- **Language**: Go 1.19+
- **Archive Format**: ZIP with DEFLATE compression
- **MIME Detection**: Uses Go's `http.DetectContentType()` for robust file type identification
- **Concurrency**: Worker pool pattern with bounded channels
- **Cross-Platform**: Builds for 20+ OS/architecture combinations

## Supported Platforms

The build script supports:
- **Linux**: AMD64, ARM, ARM64, PowerPC, MIPS, S390X
- **macOS**: Intel (AMD64) and Apple Silicon (ARM64)
- **FreeBSD**: AMD64, 386
- **OpenBSD**: AMD64, 386, ARM64
- **NetBSD**: AMD64, 386, ARM
- **Other**: DragonFlyBSD, Solaris, Plan 9

## Contributing

1. Fork the repository
2. Create your feature branch
3. Add tests for new functionality
4. Ensure all builds pass: `./build.sh all`
5. Submit a pull request

## License

This project is released under the MIT License.

## Troubleshooting

### Common Issues

**Q: "No files found to archive" error**
- Smart mode: Check that folders contain images, text, or video files
- Dumb mode: Verify the folder actually contains files
- Check file permissions are readable

**Q: Too many/few files being included**
- Use `-dumb` for complete archiving without filtering
- Smart mode intentionally excludes system files and VCS data
- Check the excluded files count in the final statistics

**Q: Difference between recursive and direct mode?**
- Recursive: Scans for subdirectories and converts each one
- Direct: Converts only the specified directories
- Use recursive for batch processing, direct for specific folders

**Q: How to process specific chapters?**
- Use direct mode without `-recursive` flag
- Specify each folder with separate `-input` flags
- Example: `convert-cbz -input "./Ch1" -input "./Ch2" -output ./cbz`

**Q: CBZ files not opening in comic readers**
- Ensure input folders contain valid image files
- Some readers may need specific file ordering
- Try both smart and dumb modes to see which works better

**Q: Permission denied errors**
- Check read permissions on input directory  
- Check write permissions on output directory
- Run with appropriate user privileges

**Q: High memory usage**
- Reduce thread count with `-threads` flag
- Process smaller batches of folders

### Performance Tips

- Use SSD storage for both input and output directories
- Set thread count to match your CPU cores (or slightly higher)
- Use recursive mode for batch processing large collections
- Use direct mode when you need precise control over what gets converted
- Close other resource-intensive applications during conversion
