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
	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		c.Stop()
	}()

	// Start crawling
	fmt.Printf("Starting crawler at %s with %d workers and max depth %d\n",
		*startURL, *numWorkers, *maxDepth)

	start := time.Now()
	results, err := c.Start()
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("Crawler error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Crawling completed in %s\n", elapsed)
	fmt.Printf("Found %d unique URLs\n", len(results))
	fmt.Printf("Results saved to %s\n", *outputFile)
}
