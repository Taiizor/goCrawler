package crawler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Result represents a single URL that has been crawled
type Result struct {
	URL           string    `json:"url"`
	Title         string    `json:"title"`
	StatusCode    int       `json:"status_code"`
	ContentLength int64     `json:"content_length"`
	Links         []string  `json:"links"`
	Depth         int       `json:"depth"`
	Timestamp     time.Time `json:"timestamp"`
}

// Config holds all configuration parameters for the crawler
type Config struct {
	StartURL   string
	MaxDepth   int
	NumWorkers int
	Timeout    time.Duration
	RateLimit  time.Duration
	Logger     *log.Logger
	Storage    interface {
		Save(results interface{}) error
	}
}

// Crawler represents the web crawler
type Crawler struct {
	config      Config
	client      *http.Client
	wg          sync.WaitGroup
	seen        map[string]bool
	results     []Result
	jobs        chan job
	stopChan    chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
	rateLimiter <-chan time.Time
	mu          sync.Mutex
}

// job represents a URL to be crawled
type job struct {
	url   string
	depth int
}

// New creates a new configured crawler
func New(config Config) *Crawler {
	// Set default values
	if config.NumWorkers <= 0 {
		config.NumWorkers = 5
	}
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}
	if config.MaxDepth <= 0 {
		config.MaxDepth = 2
	}
	if config.RateLimit <= 0 {
		config.RateLimit = 100 * time.Millisecond
	}
	if config.Logger == nil {
		config.Logger = log.New(log.Writer(), "[CRAWLER] ", log.LstdFlags)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Crawler{
		config:      config,
		client:      &http.Client{Timeout: config.Timeout},
		seen:        make(map[string]bool),
		results:     make([]Result, 0),
		jobs:        make(chan job),
		stopChan:    make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
		rateLimiter: time.NewTicker(config.RateLimit).C,
	}
}

// Start begins the crawling process
func (c *Crawler) Start() ([]Result, error) {
	// Parse and normalize the starting URL
	startURL, err := NormalizeURL(c.config.StartURL)
	if err != nil {
		return nil, err
	}

	// Extract the base domain for filtering
	baseURL, err := url.Parse(startURL)
	if err != nil {
		return nil, err
	}
	baseDomain := baseURL.Host

	// Start the worker pool
	for i := 0; i < c.config.NumWorkers; i++ {
		c.wg.Add(1)
		go c.worker(i+1, baseDomain)
	}

	// Enqueue the starting URL
	c.jobs <- job{url: startURL, depth: 0}
	c.markURLSeen(startURL)

	// Wait for completion or cancellation
	go func() {
		c.wg.Wait()
		close(c.stopChan)
	}()

	select {
	case <-c.stopChan:
		c.config.Logger.Println("Crawling completed successfully")
	case <-c.ctx.Done():
		c.config.Logger.Println("Crawling was cancelled")
	}

	// Save results
	if c.config.Storage != nil {
		c.mu.Lock()
		results := make([]Result, len(c.results))
		copy(results, c.results)
		c.mu.Unlock()

		if err := c.config.Storage.Save(results); err != nil {
			return results, errors.New("failed to save results: " + err.Error())
		}
	}

	return c.results, nil
}

// Stop gracefully shuts down the crawler
func (c *Crawler) Stop() {
	c.cancel()
	<-c.stopChan
}

// worker processes jobs from the queue
func (c *Crawler) worker(id int, baseDomain string) {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			c.config.Logger.Printf("Worker %d shutting down", id)
			return
		case currentJob, ok := <-c.jobs:
			if !ok {
				return
			}

			// Rate limiting
			<-c.rateLimiter

			// Process the URL
			c.config.Logger.Printf("Worker %d crawling %s (depth: %d)", id, currentJob.url, currentJob.depth)
			result, err := c.crawlURL(currentJob.url, currentJob.depth)
			if err != nil {
				c.config.Logger.Printf("Error crawling %s: %v", currentJob.url, err)
				continue
			}

			// Store the result
			c.mu.Lock()
			c.results = append(c.results, result)
			c.mu.Unlock()

			// If we haven't reached max depth, add all links to the queue
			if currentJob.depth < c.config.MaxDepth {
				for _, link := range result.Links {
					// Only process URLs we haven't seen yet
					if !c.hasURLBeenSeen(link) {
						// Only follow links on the same domain
						linkURL, err := url.Parse(link)
						if err == nil && linkURL.Host == baseDomain {
							c.markURLSeen(link)
							select {
							case c.jobs <- job{url: link, depth: currentJob.depth + 1}:
							case <-c.ctx.Done():
								return
							}
						}
					}
				}
			}
		}
	}
}

// hasURLBeenSeen checks if a URL has already been seen
func (c *Crawler) hasURLBeenSeen(url string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.seen[url]
}

// markURLSeen marks a URL as seen
func (c *Crawler) markURLSeen(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seen[url] = true
}
