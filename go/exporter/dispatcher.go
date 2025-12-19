package exporter

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/datamill/data-engine/go/db"
)

// ExportData orchestrates the export process
func ExportData(ctx context.Context, config *Config) error {
	// Open database connection
	connector, err := db.NewConnector(config.DSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer connector.Close()

	// Execute query and get streaming cursor
	rows, err := connector.StreamQuery(ctx, config.Query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := connector.GetColumns(rows)
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[INFO] Exporting %d columns: %v\n", len(columns), columns)

	// Select appropriate exporter
	var exporter Exporter
	switch config.OutputFormat {
	case "csv":
		exporter = NewCSVExporter(config.OutputFile, ',', columns)
	case "tsv":
		exporter = NewCSVExporter(config.OutputFile, '\t', columns)
	case "jsonl":
		exporter = NewJSONLExporter(config.OutputFile, columns)
	case "parquet":
		exporter = NewParquetExporter(config.OutputFile, columns)
	default:
		return fmt.Errorf("unsupported output format: %s", config.OutputFormat)
	}

	// Open exporter
	if err := exporter.Open(); err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer exporter.Close()

	// Progress tracking
	var rowCount int64
	startTime := time.Now()

	// Start progress reporter
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				count := atomic.LoadInt64(&rowCount)
				elapsed := time.Since(startTime).Seconds()
				rate := float64(count) / elapsed
				fmt.Fprintf(os.Stderr, "[PROGRESS] Exported %d rows (%.0f rows/sec)\n", count, rate)
			case <-done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
	defer close(done)

	// Prepare value scanners
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Read and write rows
	for rows.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert to clean values
		row := make([]interface{}, len(values))
		copy(row, values)

		if err := exporter.WriteRow(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}

		atomic.AddInt64(&rowCount, 1)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("row iteration error: %w", err)
	}

	// Flush any remaining data
	if err := exporter.Flush(); err != nil {
		return fmt.Errorf("failed to flush output: %w", err)
	}

	finalCount := atomic.LoadInt64(&rowCount)
	elapsed := time.Since(startTime).Seconds()
	fmt.Fprintf(os.Stderr, "[INFO] Export completed: %d rows in %.2f seconds (%.0f rows/sec)\n", 
		finalCount, elapsed, float64(finalCount)/elapsed)

	return nil
}

// Exporter is the interface for file exporters
type Exporter interface {
	Open() error
	WriteRow(row []interface{}) error
	Flush() error
	Close() error
}

// Config is imported from parent package
type Config struct {
	DSN           string
	OutputFile    string
	OutputFormat  string
	Query         string
	BatchSize     int
	Workers       int
	ProgressEvery int
}

// scanRow scans a SQL row into a slice
func scanRow(rows *sql.Rows, columns []string) ([]interface{}, error) {
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	return values, nil
}
