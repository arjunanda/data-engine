package exporter

import (
	"encoding/csv"
	"fmt"
	"os"
)

// CSVExporter handles CSV and TSV file exports
type CSVExporter struct {
	filePath  string
	delimiter rune
	columns   []string
	file      *os.File
	writer    *csv.Writer
}

// NewCSVExporter creates a new CSV exporter
func NewCSVExporter(filePath string, delimiter rune, columns []string) *CSVExporter {
	return &CSVExporter{
		filePath:  filePath,
		delimiter: delimiter,
		columns:   columns,
	}
}

// Open opens the CSV file and writes the header
func (c *CSVExporter) Open() error {
	file, err := os.Create(c.filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	c.file = file

	// Create CSV writer
	c.writer = csv.NewWriter(file)
	c.writer.Comma = c.delimiter

	// Write header
	if err := c.writer.Write(c.columns); err != nil {
		c.file.Close()
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}

// WriteRow writes a row to the CSV file
func (c *CSVExporter) WriteRow(row []interface{}) error {
	// Convert interface slice to string slice
	record := make([]string, len(row))
	for i, val := range row {
		record[i] = formatValue(val)
	}

	if err := c.writer.Write(record); err != nil {
		return fmt.Errorf("failed to write row: %w", err)
	}

	return nil
}

// Flush flushes the CSV writer
func (c *CSVExporter) Flush() error {
	c.writer.Flush()
	return c.writer.Error()
}

// Close closes the CSV file
func (c *CSVExporter) Close() error {
	if c.writer != nil {
		c.writer.Flush()
	}
	if c.file != nil {
		return c.file.Close()
	}
	return nil
}

// formatValue converts an interface{} to a string for CSV output
func formatValue(val interface{}) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}
