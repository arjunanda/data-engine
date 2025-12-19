package exporter

import (
	"fmt"
	"os"

	"github.com/parquet-go/parquet-go"
)

// ParquetExporter handles Parquet file exports
type ParquetExporter struct {
	filePath string
	columns  []string
	file     *os.File
	writer   *parquet.GenericWriter[map[string]interface{}]
}

// NewParquetExporter creates a new Parquet exporter
func NewParquetExporter(filePath string, columns []string) *ParquetExporter {
	return &ParquetExporter{
		filePath: filePath,
		columns:  columns,
	}
}

// Open opens the Parquet file
func (p *ParquetExporter) Open() error {
	file, err := os.Create(p.filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	p.file = file

	// Create Parquet writer with generic schema
	// Note: This uses a simple map-based approach
	// For better performance, define a struct schema
	p.writer = parquet.NewGenericWriter[map[string]interface{}](file)

	return nil
}

// WriteRow writes a row to the Parquet file
func (p *ParquetExporter) WriteRow(row []interface{}) error {
	// Convert row to map
	record := make(map[string]interface{}, len(p.columns))
	for i, col := range p.columns {
		if i < len(row) {
			record[col] = convertParquetValue(row[i])
		} else {
			record[col] = nil
		}
	}

	// Write row
	if _, err := p.writer.Write([]map[string]interface{}{record}); err != nil {
		return fmt.Errorf("failed to write row: %w", err)
	}

	return nil
}

// Flush flushes the Parquet writer
func (p *ParquetExporter) Flush() error {
	// Parquet writer flushes automatically
	return nil
}

// Close closes the Parquet file
func (p *ParquetExporter) Close() error {
	if p.writer != nil {
		if err := p.writer.Close(); err != nil {
			return err
		}
	}
	if p.file != nil {
		return p.file.Close()
	}
	return nil
}

// convertParquetValue converts database values to Parquet-compatible types
func convertParquetValue(val interface{}) interface{} {
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case []byte:
		// Convert byte arrays to strings for Parquet
		return string(v)
	default:
		return v
	}
}
