package storage

import (
	"path/filepath"
	"strings"
)

// Storage is the interface for saving crawler results
type Storage interface {
	Save(results interface{}) error
}

// IsJSONFile checks if a file path has a .json extension
func IsJSONFile(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".json"
}

// IsCSVFile checks if a file path has a .csv extension
func IsCSVFile(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".csv"
}
