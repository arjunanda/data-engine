package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	// Utilize all CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Set up context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		fmt.Fprintf(os.Stderr, "\n[SIGNAL] Received %v, initiating graceful shutdown...\n", sig)
		cancel()
	}()

	// Read configuration from stdin
	var config Config
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&config); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Normalize configuration (auto-detect workers, format, etc.)
	if err := config.Normalize(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Configuration normalization failed: %v\n", err)
		os.Exit(1)
	}

	// Log configuration (without sensitive data)
	fmt.Fprintf(os.Stderr, "[INFO] Mode: %s\n", config.Mode)
	fmt.Fprintf(os.Stderr, "[INFO] Workers: %d\n", config.Workers)
	fmt.Fprintf(os.Stderr, "[INFO] Batch Size: %d\n", config.BatchSize)

	// Dispatch to appropriate mode
	var err error
	switch config.Mode {
	case "import":
		err = runImport(ctx, &config)
	case "export":
		err = runExport(ctx, &config)
	default:
		fmt.Fprintf(os.Stderr, "[ERROR] Unknown mode: %s\n", config.Mode)
		os.Exit(1)
	}

	// Handle execution errors
	if err != nil {
		if ctx.Err() != nil {
			// Graceful shutdown
			fmt.Fprintf(os.Stderr, "[INFO] Operation cancelled, shutting down gracefully\n")
			os.Exit(130) // Standard exit code for SIGINT
		}
		fmt.Fprintf(os.Stderr, "[ERROR] Operation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "[SUCCESS] Operation completed successfully\n")
	os.Exit(0)
}

func runImport(ctx context.Context, config *Config) error {
	fmt.Fprintf(os.Stderr, "[INFO] Starting import: %s -> %s\n", config.InputFile, config.Table)
	return ImportData(ctx, config)
}

func runExport(ctx context.Context, config *Config) error {
	fmt.Fprintf(os.Stderr, "[INFO] Starting export: %s -> %s\n", config.Query, config.OutputFile)
	return ExportData(ctx, config)
}
