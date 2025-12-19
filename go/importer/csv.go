package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// CSVImporter handles CSV and TSV file imports
type CSVImporter struct {
	filePath  string
	delimiter rune
	file      *os.File
	reader    *csv.Reader
	columns   []string
}

// NewCSVImporter creates a new CSV importer
func NewCSVImporter(filePath string, delimiter rune) *CSVImporter {
	return &CSVImporter{
		filePath:  filePath,
		delimiter: delimiter,
	}
}

// Open opens the CSV file and reads the header
func (c *CSVImporter) Open() ([]string, error) {
	file, err := os.Open(c.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	c.file = file

	// Create CSV reader
	c.reader = csv.NewReader(file)
	c.reader.Comma = c.delimiter
	c.reader.LazyQuotes = true
	c.reader.TrimLeadingSpace = true

	// Read header row
	header, err := c.reader.Read()
	if err != nil {
		c.file.Close()
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	c.columns = header
	return header, nil
}

// NextRow reads the next row from the CSV file
func (c *CSVImporter) NextRow() ([]interface{}, error) {
	record, err := c.reader.Read()
	if err == io.EOF {
		return nil, fmt.Errorf("EOF")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read row: %w", err)
	}

	// Convert string slice to interface slice
	row := make([]interface{}, len(record))
	for i, val := range record {
		row[i] = val
	}

	return row, nil
}

// Close closes the CSV file
func (c *CSVImporter) Close() error {
	if c.file != nil {
		return c.file.Close()
	}
	return nil
}
