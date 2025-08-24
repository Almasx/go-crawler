# Go Web Crawler

A concurrent web crawler built in Go that analyzes internal link structure of websites.

## Features

- **Concurrent crawling** with configurable concurrency limits
- **Internal link analysis** - only crawls pages within the same domain
- **Link frequency reporting** - shows how many times each page is referenced
- **Configurable limits** for maximum pages and concurrent connections
- **Normalized URL handling** to avoid duplicate crawling

## Usage

```bash
go run . <website-url> [max-concurrency] [max-pages]
```

### Parameters

- `website-url` - The starting URL to crawl (required)
- `max-concurrency` - Maximum number of concurrent crawlers (default: 10)
- `max-pages` - Maximum number of pages to crawl (default: 100)

### Examples

```bash
# Basic crawling with defaults
go run . https://example.com

# Custom concurrency and page limits
go run . https://example.com 5 50
```

## How it works

1. Starts from the provided base URL
2. Fetches HTML content and extracts all links
3. Only follows links within the same domain (internal links)
4. Tracks how many times each page is referenced by other pages
5. Provides a sorted report showing link frequency

## Output

The crawler outputs:
- Progress messages showing which pages are being crawled
- A final report showing all discovered pages sorted by reference count
- Pages with more internal links pointing to them appear first

## Requirements

- Go 1.23.4+
- `golang.org/x/net` dependency (automatically managed by Go modules)

## Installation

```bash
git clone <repository-url>
cd go-crawler
go mod tidy
go run . <website-url>
```