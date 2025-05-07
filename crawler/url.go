package crawler

import (
	"net/url"
	"strings"
)

// NormalizeURL standardizes a URL for consistency
func NormalizeURL(rawURL string) (string, error) {
	// Add scheme if missing
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Ensure the path ends with a slash if it's a directory (no filename)
	if !strings.Contains(parsedURL.Path, ".") && !strings.HasSuffix(parsedURL.Path, "/") {
		parsedURL.Path = parsedURL.Path + "/"
	}

	// Remove common tracking parameters
	query := parsedURL.Query()
	removeParams := []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content"}
	for _, param := range removeParams {
		query.Del(param)
	}
	parsedURL.RawQuery = query.Encode()

	// Remove fragments (anchors)
	parsedURL.Fragment = ""

	// Remove default ports
	if (parsedURL.Scheme == "http" && parsedURL.Port() == "80") ||
		(parsedURL.Scheme == "https" && parsedURL.Port() == "443") {
		// Remove the port
		host := parsedURL.Host
		if colonIndex := strings.IndexByte(host, ':'); colonIndex != -1 {
			parsedURL.Host = host[:colonIndex]
		}
	}

	return parsedURL.String(), nil
}

// IsURLValid checks if a URL is valid and acceptable for crawling
func IsURLValid(rawURL string) bool {
	// Check for empty URL
	if rawURL == "" {
		return false
	}

	// Check for non-HTTP(S) schemes
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return false
	}

	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Reject certain file types that are not webpages
	excludedExtensions := []string{
		".pdf", ".jpg", ".jpeg", ".png", ".gif", ".svg", ".css", ".js",
		".zip", ".tar", ".gz", ".rar", ".exe", ".mp3", ".mp4", ".avi",
		".mov", ".mkv", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
	}

	for _, ext := range excludedExtensions {
		if strings.HasSuffix(strings.ToLower(parsedURL.Path), ext) {
			return false
		}
	}

	return true
}

// ResolveURL resolves a relative URL against a base URL
func ResolveURL(baseURL, relativeURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// If it's already absolute, just return it
	if strings.HasPrefix(relativeURL, "http://") || strings.HasPrefix(relativeURL, "https://") {
		return relativeURL, nil
	}

	rel, err := url.Parse(relativeURL)
	if err != nil {
		return "", err
	}

	resolvedURL := base.ResolveReference(rel)
	return resolvedURL.String(), nil
}
