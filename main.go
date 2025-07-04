package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Taiizor/goCrawler/crawler"
	"github.com/Taiizor/goCrawler/storage"
)

func main() {
	// Command line flags
	startURL := flag.String("url", "", "Starting URL for crawling")
	maxDepth := flag.Int("depth", 2, "Maximum crawling depth")
	numWorkers := flag.Int("workers", 5, "Number of concurrent workers")
	outputFile := flag.String("output", "results.json", "Output file name (CSV or JSON)")
	timeout := flag.Duration("timeout", 10*time.Second, "HTTP request timeout")
	rateLimit := flag.Duration("rate", 100*time.Millisecond, "Rate limit between requests")
	crawlTimeout := flag.Duration("crawl-timeout", 5*time.Minute, "Maximum time for crawling to run")
	flag.Parse()

	if *startURL == "" {
		fmt.Println("Please provide a starting URL with -url flag")
		flag.Usage()
		os.Exit(1)
	}

	// Setup logger
	logFile, err := os.OpenFile("crawler.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	// Setup storage based on file extension
	var store storage.Storage
	if storage.IsJSONFile(*outputFile) {
		store = storage.NewJSONStorage(*outputFile)
	} else {
		store = storage.NewCSVStorage(*outputFile)
	}

	// Create and configure crawler
	c := crawler.New(crawler.Config{
		StartURL:   *startURL,
		MaxDepth:   *maxDepth,
		NumWorkers: *numWorkers,
		Timeout:    *timeout,
		RateLimit:  *rateLimit,
		Logger:     logger,
		Storage:    store,
	})

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Add a timer at this point to stop the crawler after a certain time
	var timeoutChan <-chan time.Time
	if *crawlTimeout > 0 {
		timeoutChan = time.After(*crawlTimeout)
		fmt.Printf("Crawler will automatically stop after %s if not completed\n", *crawlTimeout)
	}

	// Create a channel for graceful shutdown
	done := make(chan struct{})

	// Goroutine for graceful shutdown
	go func() {
		select {
		case <-sigChan:
			fmt.Println("\nReceived interrupt signal. Shutting down gracefully...")
		case <-timeoutChan:
			if timeoutChan != nil {
				fmt.Printf("\nCrawler timeout reached (%s). Shutting down...\n", *crawlTimeout)
			}
		}
		c.Stop()
		close(done) // Close the done channel to notify other goroutines that the process is finished
	}()

	// Start crawling
	fmt.Printf("Starting crawler at %s with %d workers and max depth %d\n",
		*startURL, *numWorkers, *maxDepth)

	// Start a goroutine to show progress
	if *maxDepth > 0 {
		fmt.Println("Crawling in progress... Press Ctrl+C to stop")
		ticker := time.NewTicker(5 * time.Second)
		go func() {
			dots := 0
			for {
				select {
				case <-ticker.C:
					// Add a progress dot every 5 seconds
					fmt.Print(".")
					dots++
					if dots%10 == 0 {
						fmt.Println() // New line every 10 dots
					}
				case <-done:
					ticker.Stop()
					return
				}
			}
		}()
	}

	start := time.Now()
	results, err := c.Start()
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("\nCrawler error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nCrawling completed in %s\n", elapsed)
	fmt.Printf("Found %d unique URLs\n", len(results))
	fmt.Printf("Results saved to %s\n", *outputFile)
}
