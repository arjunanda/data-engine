package main

import (
	"context"

	"github.com/datamill/data-engine/go/exporter"
	"github.com/datamill/data-engine/go/importer"
)

// ImportData bridges to the importer package
func ImportData(ctx context.Context, config *Config) error {
	importConfig := &importer.Config{
		DSN:           config.DSN,
		InputFile:     config.InputFile,
		InputFormat:   config.InputFormat,
		Table:         config.Table,
		BatchSize:     config.BatchSize,
		Workers:       config.Workers,
		ProgressEvery: config.ProgressEvery,
	}
	return importer.ImportData(ctx, importConfig)
}

// ExportData bridges to the exporter package
func ExportData(ctx context.Context, config *Config) error {
	exportConfig := &exporter.Config{
		DSN:           config.DSN,
		OutputFile:    config.OutputFile,
		OutputFormat:  config.OutputFormat,
		Query:         config.Query,
		BatchSize:     config.BatchSize,
		Workers:       config.Workers,
		ProgressEvery: config.ProgressEvery,
	}
	return exporter.ExportData(ctx, exportConfig)
}
