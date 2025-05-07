package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

// CSVStorage implements Storage interface for CSV format
type CSVStorage struct {
	filePath string
}

// NewCSVStorage creates a new CSVStorage instance
func NewCSVStorage(filePath string) *CSVStorage {
	return &CSVStorage{
		filePath: filePath,
	}
}

// Save writes the crawl results to a CSV file
func (s *CSVStorage) Save(results interface{}) error {
	// Create the file
	file, err := os.Create(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Try to convert results to a slice
	val := reflect.ValueOf(results)
	if val.Kind() != reflect.Slice {
		return fmt.Errorf("expected slice but got %T", results)
	}

	if val.Len() == 0 {
		// Write header for empty results
		err = writer.Write([]string{"URL", "Title", "StatusCode", "ContentLength", "Depth", "Timestamp", "LinksCount"})
		if err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}
		return nil
	}

	// Extract first item to determine structure
	firstItem := val.Index(0).Interface()
	firstItemValue := reflect.ValueOf(firstItem)
	firstItemType := firstItemValue.Type()

	// If first item is a struct, write headers based on struct fields
	if firstItemType.Kind() == reflect.Struct {
		var headers []string
		for i := 0; i < firstItemType.NumField(); i++ {
			field := firstItemType.Field(i)
			// Skip unexported fields
			if field.PkgPath != "" {
				continue
			}
			
			// Use JSON tag if available, otherwise use field name
			tag := field.Tag.Get("json")
			if tag == "" {
				headers = append(headers, field.Name)
			} else {
				// Split to handle tag options like omitempty
				headers = append(headers, tag)
			}
		}

		// Add special header for links count
		headers = append(headers, "LinksCount")

		// Write headers
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}
	}

	// Write data rows
	for i := 0; i < val.Len(); i++ {
		item := val.Index(i).Interface()
		itemValue := reflect.ValueOf(item)

		var row []string
		if itemValue.Kind() == reflect.Struct {
			for j := 0; j < itemValue.NumField(); j++ {
				field := itemValue.Field(j)
				
				// Skip unexported fields
				if itemValue.Type().Field(j).PkgPath != "" {
					continue
				}

				// Handle different field types
				var fieldStr string
				switch field.Kind() {
				case reflect.String:
					fieldStr = field.String()
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					fieldStr = strconv.FormatInt(field.Int(), 10)
				case reflect.Float32, reflect.Float64:
					fieldStr = strconv.FormatFloat(field.Float(), 'f', 2, 64)
				case reflect.Bool:
					fieldStr = strconv.FormatBool(field.Bool())
				case reflect.Struct:
					// Handle Time type specifically
					if field.Type() == reflect.TypeOf(time.Time{}) {
						t := field.Interface().(time.Time)
						fieldStr = t.Format(time.RFC3339)
					} else {
						fieldStr = fmt.Sprintf("%v", field.Interface())
					}
				case reflect.Slice:
					// For links field, just store the count
					if itemValue.Type().Field(j).Name == "Links" {
						row = append(row, strconv.Itoa(field.Len()))
						continue
					} else {
						fieldStr = fmt.Sprintf("%v", field.Interface())
					}
				default:
					fieldStr = fmt.Sprintf("%v", field.Interface())
				}
				
				row = append(row, fieldStr)
			}

		} else {
			// Handle non-struct items (unlikely in our case)
			row = append(row, fmt.Sprintf("%v", item))
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
} 