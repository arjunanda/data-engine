/**
 * TypeScript Usage Examples for Data Engine
 * 
 * This file demonstrates how to use the data-engine package with TypeScript
 */

import { importData, exportData, ImportOptions, ExportOptions } from '../index';

/**
 * Example 1: Import CSV to PostgreSQL with full type safety
 */
async function exampleImportCSV(): Promise<void> {
  const options: ImportOptions = {
    file: './data/users.csv',
    format: 'auto',  // TypeScript will autocomplete: 'auto' | 'csv' | 'tsv' | 'jsonl' | 'xlsx'
    dsn: 'postgres://user:password@localhost:5432/mydb',
    table: 'users',
    batchSize: 10000,
    workers: 0  // 0 = auto-detect CPU count
  };

  try {
    await importData(options);
    console.log('✓ Import completed successfully!');
  } catch (error) {
    console.error('✗ Import failed:', error);
    throw error;
  }
}

/**
 * Example 2: Export to JSONL with type-safe options
 */
async function exampleExportJSONL(): Promise<void> {
  const options: ExportOptions = {
    output: './exports/users.jsonl',
    format: 'jsonl',  // TypeScript will autocomplete: 'csv' | 'tsv' | 'jsonl' | 'parquet'
    dsn: 'postgres://user:password@localhost:5432/mydb',
    query: 'SELECT * FROM users WHERE created_at > NOW() - INTERVAL \'30 days\'',
    batchSize: 5000
  };

  await exportData(options);
}

/**
 * Example 3: Import JSONL to MySQL
 */
async function exampleImportJSONLToMySQL(): Promise<void> {
  await importData({
    file: './data/events.jsonl',
    format: 'jsonl',
    dsn: 'user:password@tcp(localhost:3306)/analytics',
    table: 'events',
    batchSize: 15000,
    workers: 8
  });
}

/**
 * Example 4: Export to Parquet for analytics
 */
async function exampleExportParquet(): Promise<void> {
  await exportData({
    output: './analytics/monthly-report.parquet',
    format: 'parquet',
    dsn: 'postgres://analytics:pass@db.example.com:5432/warehouse',
    query: `
      SELECT 
        date_trunc('day', created_at) as day,
        COUNT(*) as event_count,
        AVG(value) as avg_value,
        SUM(revenue) as total_revenue
      FROM events
      WHERE created_at BETWEEN '2024-01-01' AND '2024-12-31'
      GROUP BY day
      ORDER BY day
    `
  });
}

/**
 * Example 5: Batch processing with error handling
 */
async function exampleBatchProcessing(): Promise<void> {
  const files = [
    './data/users-2024-01.csv',
    './data/users-2024-02.csv',
    './data/users-2024-03.csv'
  ];

  for (const file of files) {
    try {
      console.log(`Processing ${file}...`);
      
      await importData({
        file,
        format: 'csv',
        dsn: 'postgres://localhost/mydb',
        table: 'users',
        batchSize: 10000
      });
      
      console.log(`✓ ${file} completed`);
    } catch (error) {
      console.error(`✗ ${file} failed:`, error);
      // Continue with next file or throw based on requirements
    }
  }
}

/**
 * Example 6: Using environment variables for configuration
 */
async function exampleWithEnvVars(): Promise<void> {
  const DATABASE_DSN = process.env.DATABASE_DSN || 'postgres://localhost/mydb';
  const BATCH_SIZE = parseInt(process.env.BATCH_SIZE || '5000', 10);
  const WORKERS = parseInt(process.env.WORKERS || '0', 10);

  await importData({
    file: './data/large-dataset.csv',
    format: 'auto',
    dsn: DATABASE_DSN,
    table: 'events',
    batchSize: BATCH_SIZE,
    workers: WORKERS
  });
}

/**
 * Example 7: Type-safe configuration builder
 */
class DataEngineConfig {
  private options: Partial<ImportOptions | ExportOptions> = {};

  setDSN(dsn: string): this {
    this.options.dsn = dsn;
    return this;
  }

  setBatchSize(size: number): this {
    this.options.batchSize = size;
    return this;
  }

  setWorkers(count: number): this {
    this.options.workers = count;
    return this;
  }

  buildImport(file: string, table: string): ImportOptions {
    return {
      file,
      table,
      format: 'auto',
      dsn: this.options.dsn!,
      batchSize: this.options.batchSize || 5000,
      workers: this.options.workers || 0
    };
  }

  buildExport(output: string, query: string, format: ExportOptions['format']): ExportOptions {
    return {
      output,
      query,
      format,
      dsn: this.options.dsn!,
      batchSize: this.options.batchSize || 5000,
      workers: this.options.workers || 0
    };
  }
}

// Usage of config builder
async function exampleConfigBuilder(): Promise<void> {
  const config = new DataEngineConfig()
    .setDSN('postgres://localhost/mydb')
    .setBatchSize(10000)
    .setWorkers(8);

  // Import
  await importData(config.buildImport('./data.csv', 'users'));

  // Export
  await exportData(config.buildExport(
    './export.jsonl',
    'SELECT * FROM users',
    'jsonl'
  ));
}

/**
 * Example 8: Async/await with proper error handling
 */
async function exampleProperErrorHandling(): Promise<void> {
  try {
    await importData({
      file: './data/users.csv',
      dsn: 'postgres://localhost/mydb',
      table: 'users'
    });
  } catch (error) {
    if (error instanceof Error) {
      // TypeScript knows error is Error type
      console.error('Import failed:', error.message);
      
      // Check for specific error types
      if (error.message.includes('connection refused')) {
        console.error('Database is not running');
      } else if (error.message.includes('file not found')) {
        console.error('Input file does not exist');
      }
    }
    throw error;
  }
}

// Export all examples
export {
  exampleImportCSV,
  exampleExportJSONL,
  exampleImportJSONLToMySQL,
  exampleExportParquet,
  exampleBatchProcessing,
  exampleWithEnvVars,
  exampleConfigBuilder,
  exampleProperErrorHandling
};
