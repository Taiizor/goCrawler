package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// JSONStorage implements Storage interface for JSON format
type JSONStorage struct {
	filePath string
}

// NewJSONStorage creates a new JSONStorage instance
func NewJSONStorage(filePath string) *JSONStorage {
	return &JSONStorage{
		filePath: filePath,
	}
}

// Save writes the crawl results to a JSON file
func (s *JSONStorage) Save(results interface{}) error {
	// Create the file
	file, err := os.Create(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	// Add metadata to the output
	output := map[string]interface{}{
		"results":   results,
		"count":     getResultCount(results),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Create an encoder with pretty printing
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	// Encode the data
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// getResultCount tries to determine the count of results
func getResultCount(results interface{}) int {
	// Try to get length of slice
	if slice, ok := results.([]interface{}); ok {
		return len(slice)
	}

	// Try with specific type
	if typedResults, ok := results.([]struct{}); ok {
		return len(typedResults)
	}

	// If we can't determine the count, return 0
	return 0
} 