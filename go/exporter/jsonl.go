package exporter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// JSONLExporter handles JSONL (newline-delimited JSON) file exports
type JSONLExporter struct {
	filePath string
	columns  []string
	file     *os.File
	writer   *bufio.Writer
}

// NewJSONLExporter creates a new JSONL exporter
func NewJSONLExporter(filePath string, columns []string) *JSONLExporter {
	return &JSONLExporter{
		filePath: filePath,
		columns:  columns,
	}
}

// Open opens the JSONL file
func (j *JSONLExporter) Open() error {
	file, err := os.Create(j.filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	j.file = file

	// Create buffered writer for better performance
	j.writer = bufio.NewWriterSize(file, 1024*1024) // 1MB buffer

	return nil
}

// WriteRow writes a row to the JSONL file
func (j *JSONLExporter) WriteRow(row []interface{}) error {
	// Convert row to map
	obj := make(map[string]interface{}, len(j.columns))
	for i, col := range j.columns {
		if i < len(row) {
			obj[col] = row[i]
		} else {
			obj[col] = nil
		}
	}

	// Marshal to JSON
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write JSON line
	if _, err := j.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	// Write newline
	if err := j.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

// Flush flushes the buffered writer
func (j *JSONLExporter) Flush() error {
	return j.writer.Flush()
}

// Close closes the JSONL file
func (j *JSONLExporter) Close() error {
	if j.writer != nil {
		j.writer.Flush()
	}
	if j.file != nil {
		return j.file.Close()
	}
	return nil
}
