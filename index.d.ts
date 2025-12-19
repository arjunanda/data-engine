/**
 * Data Engine - TypeScript Type Definitions
 * Production-ready package for importing and exporting massive datasets
 */

/**
 * Supported input file formats for import operations
 */
export type InputFormat = 'auto' | 'csv' | 'tsv' | 'jsonl' | 'xlsx';

/**
 * Supported output file formats for export operations
 */
export type OutputFormat = 'csv' | 'tsv' | 'jsonl' | 'parquet';

/**
 * Options for importing data from a file into a database
 */
export interface ImportOptions {
  /**
   * Path to the input file
   */
  file: string;

  /**
   * File format. Use 'auto' to auto-detect from file extension and content.
   * @default 'auto'
   */
  format?: InputFormat;

  /**
   * Database connection string (DSN)
   * 
   * PostgreSQL: postgres://user:password@host:port/database
   * MySQL: user:password@tcp(host:port)/database
   */
  dsn: string;

  /**
   * Target database table name
   */
  table: string;

  /**
   * Number of rows to batch together before inserting into database
   * @default 5000
   */
  batchSize?: number;

  /**
   * Number of worker goroutines for parallel processing
   * Set to 0 to auto-detect based on CPU count
   * @default 0
   */
  workers?: number;
}

/**
 * Options for exporting data from a database to a file
 */
export interface ExportOptions {
  /**
   * Path to the output file
   */
  output: string;

  /**
   * Output file format
   */
  format: OutputFormat;

  /**
   * Database connection string (DSN)
   * 
   * PostgreSQL: postgres://user:password@host:port/database
   * MySQL: user:password@tcp(host:port)/database
   */
  dsn: string;

  /**
   * SQL query to execute for data export
   */
  query: string;

  /**
   * Number of rows to batch together during export
   * @default 5000
   */
  batchSize?: number;

  /**
   * Number of worker goroutines for parallel processing
   * Set to 0 to auto-detect based on CPU count
   * @default 0
   */
  workers?: number;
}

/**
 * Import data from a file into a database
 * 
 * Supports massive datasets (millions to billions of rows) with:
 * - Streaming architecture (constant memory usage)
 * - Multi-core parallel processing
 * - Automatic format detection
 * - Progress reporting
 * - Graceful shutdown on SIGINT/SIGTERM
 * 
 * @param options - Import configuration options
 * @returns Promise that resolves when import completes successfully
 * @throws Error if import fails (invalid config, file not found, database error, etc.)
 * 
 * @example
 * ```typescript
 * import { importData } from 'data-engine';
 * 
 * await importData({
 *   file: './large-dataset.csv',
 *   format: 'auto',
 *   dsn: 'postgres://user:pass@localhost/mydb',
 *   table: 'events',
 *   batchSize: 10000,
 *   workers: 0  // auto-detect CPU count
 * });
 * ```
 */
export function importData(options: ImportOptions): Promise<void>;

/**
 * Export data from a database to a file
 * 
 * Supports massive datasets (millions to billions of rows) with:
 * - Streaming architecture (constant memory usage)
 * - Multi-core parallel processing
 * - Multiple output formats (CSV, JSONL, Parquet)
 * - Progress reporting
 * - Graceful shutdown on SIGINT/SIGTERM
 * 
 * @param options - Export configuration options
 * @returns Promise that resolves when export completes successfully
 * @throws Error if export fails (invalid config, database error, write error, etc.)
 * 
 * @example
 * ```typescript
 * import { exportData } from 'data-engine';
 * 
 * await exportData({
 *   output: './export.jsonl',
 *   format: 'jsonl',
 *   dsn: 'postgres://user:pass@localhost/mydb',
 *   query: 'SELECT * FROM events WHERE created_at > NOW() - INTERVAL \'30 days\'',
 *   batchSize: 5000
 * });
 * ```
 */
export function exportData(options: ExportOptions): Promise<void>;
