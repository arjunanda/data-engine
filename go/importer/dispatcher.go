package importer

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/datamill/data-engine/go/db"
	"github.com/datamill/data-engine/go/worker"
)

// ImportData orchestrates the import process
func ImportData(ctx context.Context, config *Config) error {
	// Open database connection
	connector, err := db.NewConnector(config.DSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer connector.Close()

	// Detect and route to appropriate importer
	var importer Importer
	switch config.InputFormat {
	case "csv":
		importer = NewCSVImporter(config.InputFile, ',')
	case "tsv":
		importer = NewCSVImporter(config.InputFile, '\t')
	case "jsonl":
		importer = NewJSONLImporter(config.InputFile)
	case "xlsx":
		importer = NewXLSXImporter(config.InputFile)
	default:
		return fmt.Errorf("unsupported input format: %s", config.InputFormat)
	}

	// Open the importer
	columns, err := importer.Open()
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer importer.Close()

	fmt.Fprintf(os.Stderr, "[INFO] Detected %d columns: %v\n", len(columns), columns)

	// Create worker pool
	pool := worker.NewPool(ctx, config.Workers, config.BatchSize)

	// Progress tracking
	var rowCount int64
	startTime := time.Now()

	// Start progress reporter
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				count := atomic.LoadInt64(&rowCount)
				elapsed := time.Since(startTime).Seconds()
				rate := float64(count) / elapsed
				fmt.Fprintf(os.Stderr, "[PROGRESS] Processed %d rows (%.0f rows/sec)\n", count, rate)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Batch processor
	processBatch := func(ctx context.Context, rows [][]interface{}) error {
		if err := connector.BatchInsert(ctx, config.Table, columns, rows); err != nil {
			return err
		}
		atomic.AddInt64(&rowCount, int64(len(rows)))
		return nil
	}

	// Start workers
	pool.Start(processBatch)

	// Read and submit rows
	for {
		row, err := importer.NextRow()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			pool.Cancel()
			return fmt.Errorf("failed to read row: %w", err)
		}

		if err := pool.Submit(row); err != nil {
			return fmt.Errorf("failed to submit row: %w", err)
		}
	}

	// Wait for all workers to finish
	if err := pool.Close(); err != nil {
		return fmt.Errorf("worker pool error: %w", err)
	}

	finalCount := atomic.LoadInt64(&rowCount)
	elapsed := time.Since(startTime).Seconds()
	fmt.Fprintf(os.Stderr, "[INFO] Import completed: %d rows in %.2f seconds (%.0f rows/sec)\n", 
		finalCount, elapsed, float64(finalCount)/elapsed)

	return nil
}

// Importer is the interface for file importers
type Importer interface {
	Open() (columns []string, err error)
	NextRow() ([]interface{}, error)
	Close() error
}

// Config is imported from parent package
type Config struct {
	DSN           string
	InputFile     string
	InputFormat   string
	Table         string
	BatchSize     int
	Workers       int
	ProgressEvery int
}
