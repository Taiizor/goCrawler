package crawler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// crawlURL fetches and processes a single URL
func (c *Crawler) crawlURL(url string, depth int) (Result, error) {
	result := Result{
		URL:       url,
		Depth:     depth,
		Timestamp: time.Now(),
		Links:     []string{},
	}

	// Skip invalid URLs
	if !IsURLValid(url) {
		return result, fmt.Errorf("invalid URL: %s", url)
	}

	// Make the HTTP request
	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, url, nil)
	if err != nil {
		return result, err
	}

	// Set a user agent to avoid being blocked by some sites
	req.Header.Set("User-Agent", "goCrawler/1.0 (+https://github.com/Taiizor/goCrawler)")

	// Make the request
	resp, err := c.client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	// Record status code and content length
	result.StatusCode = resp.StatusCode
	result.ContentLength = resp.ContentLength

	// Only process successful responses
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Only process HTML content
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "text/html") {
		return result, fmt.Errorf("non-HTML content type: %s", contentType)
	}

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return result, err
	}

	// Extract the title
	result.Title = strings.TrimSpace(doc.Find("title").Text())

	// Extract all links
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		// Get the href attribute
		href, exists := s.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "#") {
			return
		}

		// Resolve relative URLs
		absoluteURL, err := ResolveURL(url, href)
		if err != nil {
			c.config.Logger.Printf("Error resolving URL %s against %s: %v", href, url, err)
			return
		}

		// Normalize the URL
		normalizedURL, err := NormalizeURL(absoluteURL)
		if err != nil {
			c.config.Logger.Printf("Error normalizing URL %s: %v", absoluteURL, err)
			return
		}

		// Skip invalid URLs
		if !IsURLValid(normalizedURL) {
			return
		}

		// Add the link to the results
		result.Links = append(result.Links, normalizedURL)
	})

	return result, nil
}

// FetchURLContent is a simplified HTTP client for fetching content directly
func FetchURLContent(url string, timeout time.Duration) (string, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}
