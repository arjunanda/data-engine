package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config represents the complete configuration for import/export operations
type Config struct {
	// Common fields
	Mode           string `json:"mode"`            // "import" or "export"
	DSN            string `json:"dsn"`             // Database connection string
	BatchSize      int    `json:"batch_size"`      // Number of rows per batch
	Workers        int    `json:"workers"`         // Number of worker goroutines (0 = auto)
	ProgressEvery  int    `json:"progress_every"`  // Report progress every N rows

	// Import-specific fields
	InputFile   string `json:"input_file"`   // Path to input file
	InputFormat string `json:"input_format"` // "auto", "csv", "tsv", "jsonl", "xlsx"
	Table       string `json:"table"`        // Target database table

	// Export-specific fields
	OutputFile   string `json:"output_file"`   // Path to output file
	OutputFormat string `json:"output_format"` // "csv", "tsv", "jsonl", "parquet"
	Query        string `json:"query"`         // SQL query for export
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate mode
	if c.Mode != "import" && c.Mode != "export" {
		return fmt.Errorf("mode must be 'import' or 'export', got: %s", c.Mode)
	}

	// Validate DSN
	if c.DSN == "" {
		return fmt.Errorf("dsn is required")
	}

	// Validate batch size
	if c.BatchSize <= 0 {
		c.BatchSize = 5000 // Default
	}
	if c.BatchSize > 50000 {
		return fmt.Errorf("batch_size too large (max 50000): %d", c.BatchSize)
	}

	// Validate workers
	if c.Workers < 0 {
		return fmt.Errorf("workers cannot be negative: %d", c.Workers)
	}

	// Validate progress reporting
	if c.ProgressEvery <= 0 {
		c.ProgressEvery = 100000 // Default
	}

	// Mode-specific validation
	if c.Mode == "import" {
		return c.validateImport()
	}
	return c.validateExport()
}

// validateImport validates import-specific configuration
func (c *Config) validateImport() error {
	// Validate input file
	if c.InputFile == "" {
		return fmt.Errorf("input_file is required for import mode")
	}
	if _, err := os.Stat(c.InputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", c.InputFile)
	}

	// Validate table
	if c.Table == "" {
		return fmt.Errorf("table is required for import mode")
	}

	// Validate format
	if c.InputFormat == "" {
		c.InputFormat = "auto"
	}
	validFormats := []string{"auto", "csv", "tsv", "jsonl", "xlsx"}
	if !contains(validFormats, c.InputFormat) {
		return fmt.Errorf("invalid input_format: %s (must be one of: %s)", c.InputFormat, strings.Join(validFormats, ", "))
	}

	return nil
}

// validateExport validates export-specific configuration
func (c *Config) validateExport() error {
	// Validate output file
	if c.OutputFile == "" {
		return fmt.Errorf("output_file is required for export mode")
	}

	// Validate query
	if c.Query == "" {
		return fmt.Errorf("query is required for export mode")
	}

	// Validate format
	if c.OutputFormat == "" {
		return fmt.Errorf("output_format is required for export mode")
	}
	validFormats := []string{"csv", "tsv", "jsonl", "parquet"}
	if !contains(validFormats, c.OutputFormat) {
		return fmt.Errorf("invalid output_format: %s (must be one of: %s)", c.OutputFormat, strings.Join(validFormats, ", "))
	}

	return nil
}

// Normalize applies defaults and auto-detection
func (c *Config) Normalize() error {
	// Auto-detect worker count
	if c.Workers == 0 {
		c.Workers = runtime.NumCPU()
	}

	// Auto-detect input format for import
	if c.Mode == "import" && c.InputFormat == "auto" {
		detected, err := detectFormat(c.InputFile)
		if err != nil {
			return fmt.Errorf("failed to detect input format: %w", err)
		}
		c.InputFormat = detected
		fmt.Fprintf(os.Stderr, "[INFO] Auto-detected format: %s\n", detected)
	}

	return nil
}

// detectFormat attempts to detect file format from extension and content
func detectFormat(filePath string) (string, error) {
	// Check extension first
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".csv":
		return "csv", nil
	case ".tsv":
		return "tsv", nil
	case ".jsonl", ".ndjson":
		return "jsonl", nil
	case ".xlsx":
		return "xlsx", nil
	case ".xls":
		return "", fmt.Errorf("XLS format is not supported (legacy Excel format). Please convert to XLSX or CSV")
	case ".json":
		// Check if it's JSONL or JSON array
		return "", fmt.Errorf("JSON arrays are not supported. Use JSONL (newline-delimited JSON) instead")
	}

	// Try to detect from content (read first few bytes)
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && n == 0 {
		return "", fmt.Errorf("cannot read file to detect format")
	}

	// Check for XLSX magic bytes (ZIP signature)
	if n >= 4 && header[0] == 0x50 && header[1] == 0x4B && header[2] == 0x03 && header[3] == 0x04 {
		return "xlsx", nil
	}

	// Check for XLS magic bytes (OLE2 signature)
	if n >= 8 && header[0] == 0xD0 && header[1] == 0xCF && header[2] == 0x11 && header[3] == 0xE0 {
		return "", fmt.Errorf("XLS format detected (legacy Excel format). Please convert to XLSX or CSV")
	}

	// Default to CSV for text files
	return "csv", nil
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
