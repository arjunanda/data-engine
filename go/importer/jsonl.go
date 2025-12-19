package importer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// JSONLImporter handles JSONL (newline-delimited JSON) file imports
type JSONLImporter struct {
	filePath string
	file     *os.File
	scanner  *bufio.Scanner
	columns  []string
	firstRow map[string]interface{}
}

// NewJSONLImporter creates a new JSONL importer
func NewJSONLImporter(filePath string) *JSONLImporter {
	return &JSONLImporter{
		filePath: filePath,
	}
}

// Open opens the JSONL file and reads the first line to detect columns
func (j *JSONLImporter) Open() ([]string, error) {
	file, err := os.Open(j.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	j.file = file

	// Create scanner for line-by-line reading
	j.scanner = bufio.NewScanner(file)
	
	// Increase buffer size for large lines
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	j.scanner.Buffer(buf, maxCapacity)

	// Read first line to detect columns
	if !j.scanner.Scan() {
		j.file.Close()
		return nil, fmt.Errorf("empty file or read error")
	}

	var firstObj map[string]interface{}
	if err := json.Unmarshal(j.scanner.Bytes(), &firstObj); err != nil {
		j.file.Close()
		return nil, fmt.Errorf("invalid JSON on first line: %w", err)
	}

	// Extract column names from first object
	j.columns = make([]string, 0, len(firstObj))
	for key := range firstObj {
		j.columns = append(j.columns, key)
	}

	// Store first row to be returned on first NextRow call
	j.firstRow = firstObj

	return j.columns, nil
}

// NextRow reads the next row from the JSONL file
func (j *JSONLImporter) NextRow() ([]interface{}, error) {
	var obj map[string]interface{}

	// Return first row if available
	if j.firstRow != nil {
		obj = j.firstRow
		j.firstRow = nil
	} else {
		// Read next line
		if !j.scanner.Scan() {
			if err := j.scanner.Err(); err != nil {
				return nil, fmt.Errorf("scanner error: %w", err)
			}
			return nil, fmt.Errorf("EOF")
		}

		// Parse JSON
		if err := json.Unmarshal(j.scanner.Bytes(), &obj); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
	}

	// Convert to row in column order
	row := make([]interface{}, len(j.columns))
	for i, col := range j.columns {
		if val, ok := obj[col]; ok {
			row[i] = val
		} else {
			row[i] = nil
		}
	}

	return row, nil
}

// Close closes the JSONL file
func (j *JSONLImporter) Close() error {
	if j.file != nil {
		return j.file.Close()
	}
	return nil
}
