# GoSpider 🕷️

[![Go Version](https://img.shields.io/badge/Go-1.24.4-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A high-performance, concurrent web crawler written in Go that extracts URLs, downloads content, and converts web pages to markdown format.

![Gospider (1)](https://github.com/user-attachments/assets/cf1e9a52-f372-4b34-a1a3-ff82b18844e1)



## 🎯 Overview

GoSpider is a multi-threaded web crawler that leverages Go's concurrency primitives to crawl websites at scale. It implements a producer-consumer architecture with configurable worker pools, allowing it to process thousands of URLs simultaneously while maintaining memory efficiency and system stability.

The crawler is designed for various use cases including web archiving, content analysis, search engine development, and data mining. It handles cross-domain crawling, maintains URL deduplication, and provides real-time progress monitoring.

> [!NOTE]
> **Screenshot Below:** Processing 100k urls, parsing as markdown, downloading .md file and other assets, all in under half an hour

![Code 2025-06-26 06 49 07](https://github.com/user-attachments/assets/605063ec-e1d3-4f33-91ab-7d91d5305c56)

### Key Features

- **Concurrent Architecture**: Producer-consumer pattern with configurable worker pools (1-1000+ workers)
- **Performance Optimizations**: 
  - Connection pooling with up to 500 connections per host
  - 10,000 URL buffer queue for smooth operation
  - Parallel file writing with 16 dedicated writers and 1MB buffers
  - Directory caching to minimize filesystem operations
- **Smart Crawling**:
  - Domain-based limiting to prevent overwhelming single hosts
  - URL deduplication using in-memory hash maps
  - Graceful queue management with consecutive empty checks
  - Relative to absolute URL conversion
- **Proxy Management**:
  - Random proxy rotation from configurable proxy list
  - Automatic proxy testing and validation
  - Fallback to direct connection on proxy failures
  - Support for HTTP/HTTPS proxies
- **Content Processing**:
  - HTML to Markdown conversion preserving structure and links
  - URL extraction from both raw HTML and converted markdown
  - Content-type detection and appropriate handling
  - Optional image downloading with async processing
- **Monitoring & Statistics**:
  - Real-time progress updates (URLs/second, completion rate)
  - Domain coverage tracking
  - Queue depth monitoring
  - Success/failure rate calculation
- **Resilience Features**:
  - 30-second timeout for slow servers
  - Proper resource cleanup with defer statements
  - Non-blocking operations for auxiliary tasks
  - Error handling with graceful degradation

## 🏗️ Architecture

GoSpider uses a sophisticated producer-consumer architecture optimized for high throughput:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Main Thread   │───▶│  Buffered Queue  │───▶│  Worker Pool    │
│   (Producer)    │    │  (10K capacity)  │    │  (N Consumers)  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                                              │
         │                                              ▼
         ▼                                     ┌─────────────────┐
┌─────────────────┐                           │  HTTP Client    │
│ URL Discovery   │                           │ (Connection Pool)│
│ & Deduplication │                           └─────────────────┘
└─────────────────┘                                    │
                                                       ▼
                              ┌─────────────────────────────────┐
                              │     Content Processing          │
                              ├─────────────────┬───────────────┤
                              │ HTML→Markdown   │ URL Extraction│
                              └─────────────────┴───────────────┘
                                                       │
                                    ┌──────────────────┴───────────────┐
                                    ▼                                  ▼
                           ┌─────────────────┐               ┌─────────────────┐
                           │  File Writers   │               │ Image Downloader│
                           │ (16 parallel)   │               │   (Async)       │
                           └─────────────────┘               └─────────────────┘
```

### Core Components

#### Queue System
- **Buffered Channel**: 10,000 capacity URL queue prevents blocking
- **Thread-Safe Maps**: Concurrent-safe visited URL tracking and domain counting
- **Smart Distribution**: Main thread monitors queue state and distributes work efficiently

#### HTTP Client
- **Singleton Pattern**: Single optimized client instance using `sync.Once`
- **Connection Pooling**: 
  - MaxIdleConns: 2000
  - MaxIdleConnsPerHost: 500
  - MaxConnsPerHost: 500
- **Proxy Support**: Round-robin proxy selection with automatic fallback

#### Content Processing
- **Parallel Operations**: URL extraction runs concurrently with markdown conversion
- **Link Resolution**: Converts relative URLs to absolute using domain context
- **Dual Extraction**: Extracts URLs from both HTML source and markdown output

#### File System Operations
- **Writer Pool**: 16 dedicated goroutines for file writing
- **Buffered Writing**: 1MB buffers reduce system calls
- **Directory Cache**: Avoids repeated directory existence checks
- **Fallback Mode**: Synchronous writing when async queue is full

#### Monitoring System
- **Real-time Updates**: Per-second statistics refresh
- **Metrics Tracked**:
  - Processing rate (URLs/second)
  - Queue depth
  - Domain coverage
  - Success/failure rates
  - Time elapsed

## 🔧 Technical Implementation Details

### Concurrency Model

GoSpider's concurrency model is built around Go's CSP (Communicating Sequential Processes) paradigm:

1. **Main Goroutine (Producer)**:
   - Initializes the URL queue with the seed URL
   - Monitors queue state and worker completion
   - Implements graceful shutdown with consecutive empty checks

2. **Worker Goroutines (Consumers)**:
   - Each worker runs in its own goroutine
   - Pulls URLs from the shared channel
   - Processes content independently
   - Sends discovered URLs back to the queue

3. **File Writer Goroutines**:
   - Separate pool of 16 writers
   - Receives file write requests via dedicated channel
   - Buffers writes to reduce I/O overhead

### Memory Management

- **URL Deduplication**: Uses Go's native map with string keys for O(1) lookup
- **Domain Tracking**: Separate map tracks unique domains encountered
- **Buffer Reuse**: HTTP response bodies are properly closed to allow buffer reuse
- **Goroutine Lifecycle**: Workers use WaitGroup for proper cleanup

### Performance Characteristics

| Metric | Value | Description |
|--------|-------|-------------|
| URL Processing Rate | 100-1000/sec | Depends on network latency and content size |
| Memory Usage | ~500MB-2GB | For 50,000 URLs with typical web content |
| Concurrent Connections | Up to 500/host | Configurable via HTTP client settings |
| File Write Throughput | 50-100 MB/s | With SSD and parallel writers |
| Startup Time | <1 second | Including proxy validation |

### Error Handling Strategy

1. **Network Errors**: Logged but don't stop crawling
2. **Proxy Failures**: Automatic fallback to next proxy or direct connection
3. **File System Errors**: Fallback from async to sync writing
4. **Parse Errors**: Skipped with logging, crawl continues
5. **Timeout Handling**: 30-second timeout prevents hanging on slow servers

## 📦 Installation

### Prerequisites

- Go 1.24.4 or higher
- Git (for installation from source)

### Option 1: Install from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/gospider.git
cd gospider

# Install dependencies
go mod download

# Build the binary
go build -o gospider cmd/main.go

# Run directly
./gospider -url="https://example.com"
```

### Option 2: Using Go Install

```bash
go install github.com/yourusername/gospider/cmd@latest
```

### Option 3: Download Binary

Download the latest binary from the [releases page](https://github.com/yourusername/gospider/releases).

## 🔍 How It Works

### Crawling Process

1. **Initialization**:
   - Loads configuration and validates proxies (if enabled)
   - Creates HTTP client with optimized settings
   - Initializes worker pool and file writers
   - Seeds queue with starting URL

2. **URL Processing Loop**:
   ```
   For each URL in queue:
   ├─ Check if already visited
   ├─ Check domain limits
   ├─ Fetch content (with proxy if configured)
   ├─ Extract URLs from HTML
   ├─ Convert to Markdown
   ├─ Extract URLs from Markdown
   ├─ Queue new URLs for processing
   └─ Save files (if enabled)
   ```

3. **Graceful Shutdown**:
   - Monitors queue emptiness
   - Waits for workers to finish
   - Closes file writers
   - Displays final statistics

### URL Extraction Algorithm

GoSpider uses a dual-extraction approach:

1. **HTML Extraction**:
   - Parses `<a href>` tags
   - Extracts from `<link>` elements
   - Finds URLs in `<script>` tags
   - Discovers in inline styles

2. **Markdown Extraction**:
   - Parses `[text](url)` patterns
   - Extracts reference-style links
   - Finds bare URLs in text

3. **URL Normalization**:
   - Converts relative to absolute URLs
   - Handles protocol-relative URLs
   - Cleans query parameters
   - Removes fragments

## 🚀 Usage

### Basic Usage

```bash
# Crawl a single website
./gospider -url="https://example.com"

# Crawl with custom limits
./gospider -url="https://example.com" -domains=50 -urls=500 -workers=10

# Enable verbose output
./gospider -url="https://example.com" -verbose

# Download images and save files
./gospider -url="https://example.com" -images -save
```

### Advanced Usage

```bash
# Use proxy rotation
./gospider -url="https://example.com" -proxies -workers=20

# Large-scale crawling
./gospider -url="https://example.com" -domains=1000 -urls=50000 -workers=50

# Save everything with detailed logging
./gospider -url="https://example.com" -save -images -verbose -domains=100
```

### Command Line Options

| Flag       | Type   | Default      | Description                                       |
| ---------- | ------ | ------------ | ------------------------------------------------- |
| `-url`     | string | **required** | Starting URL to crawl                             |
| `-domains` | int    | 100          | Maximum number of domains to crawl                |
| `-urls`    | int    | 1000         | Maximum number of URLs to process (0 = unlimited) |
| `-workers` | int    | 5            | Number of concurrent workers                      |
| `-proxies` | bool   | false        | Use proxies from proxies.txt file                 |
| `-images`  | bool   | false        | Download images found during crawling             |
| `-save`    | bool   | false        | Save markdown files to disk                       |
| `-verbose` | bool   | false        | Enable verbose output                             |

## ⚙️ Configuration

### Proxy Configuration

Create a `proxies.txt` file in the project root with one proxy per line:

```
http://proxy1.example.com:8080
http://proxy2.example.com:8080
```

### Output Structure

When using the `-save` flag, GoSpider creates an organized directory structure:

```
output/
├── example.com/
│   ├── index.md
│   ├── about.md
│   └── images/
│       ├── logo.png
│       └── banner.jpg
├── blog.example.com/
│   ├── post-1.md
│   └── post-2.md
└── docs.example.com/
    └── api-reference.md
```

### Proxy Configuration

Create a `proxies.txt` file with one proxy per line:

```
http://proxy1.example.com:8080
http://user:pass@proxy2.example.com:3128
socks5://proxy3.example.com:1080
```

Proxy features:
- Automatic validation on startup
- Random selection per request
- Failure tracking and blacklisting
- Transparent fallback to direct connection

### Performance Tuning

#### For Maximum Speed
```bash
./gospider -url="https://example.com" -workers=50 -urls=0
```

#### For Polite Crawling
```bash
./gospider -url="https://example.com" -workers=2 -domains=10
```

#### For Large Archives
```bash
./gospider -url="https://example.com" -workers=20 -save -images -domains=500
```

## 🤝 Contributing

We welcome contributions! Please follow these guidelines:

### Getting Started

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Add tests for new functionality
5. Run tests: `go test ./...`
6. Commit changes: `git commit -m 'Add amazing feature'`
7. Push to branch: `git push origin feature/amazing-feature`
8. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👤 Author

**Aryan Randeriya**

- GitHub: [@yourusername](https://github.com/yourusername)
- Email: your.email@example.com

## 🙏 Acknowledgments

- [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) for HTML conversion
- Go community for excellent concurrency primitives
- Contributors and testers

⭐ **Star this repository if you find it useful!**
