# Parallel Web Crawler

A high-performance web crawler with concurrent processing capabilities written in Go.

## Features

- Parallel crawling using a worker pool architecture
- Domain-specific crawling (stays within the same domain)
- Configurable crawl depth and concurrency
- URL filtering and normalization
- Rate limiting and timeout support 
- Link extraction from HTML pages
- Results export to JSON or CSV formats
- Logging and graceful error handling

## Installation

### Prerequisites

- Go 1.20 or higher

### Steps

1. Clone this repository:
   ```bash
   git clone https://github.com/Taiizor/goCrawler.git
   cd goCrawler
   ```

2. Build the application:
   ```bash
   go build -o goCrawler
   ```

## Usage

Run the crawler with the following command:

```bash
./goCrawler -url "https://www.vegalya.com" -depth 3 -workers 10 -output results.json
```

### Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-depth` | Maximum crawling depth | 2 |
| `-timeout` | HTTP request timeout | 10s |
| `-rate` | Rate limit between requests | 100ms |
| `-workers` | Number of concurrent workers | 5 |
| `-url` | Starting URL for crawling | (required) |
| `-output` | Output file name (CSV or JSON) | results.json |

## Examples

Crawl a website with 10 workers to a depth of 3, saving output as JSON:
```bash
./goCrawler -url "https://www.vegalya.com" -depth 3 -workers 10 -output results.json
```

Crawl a website and save results as CSV:
```bash
./goCrawler -url "https://www.vegalya.com" -output results.csv
```

Crawl with custom timeout and rate limiting:
```bash
./goCrawler -url "https://www.vegalya.com" -timeout 5s -rate 200ms
```

## Output Format

### JSON Output

The JSON output contains:
- `results`: Array of crawled pages
- `count`: Number of pages crawled
- `timestamp`: When the crawl completed

Each page result includes:
- `title`: Page title
- `url`: The page URL
- `status_code`: HTTP status code
- `depth`: Crawl depth of this page
- `timestamp`: When this page was crawled
- `content_length`: Content length in bytes
- `links`: Array of links found on the page

### CSV Output

The CSV output contains one row per page with columns:
- URL
- Title
- Depth
- Timestamp
- StatusCode
- ContentLength
- LinksCount (number of links found)

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 