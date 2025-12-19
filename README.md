# Data Engine

Production-ready Node.js package for importing and exporting **massive datasets** (millions to billions of rows) with a high-performance Go engine.

## Features

✅ **Streaming Architecture** - Handle datasets far larger than system memory  
✅ **Multi-Core Performance** - Fully utilize all CPU cores with worker pools  
✅ **Stable Memory Usage** - Constant memory footprint regardless of dataset size  
✅ **Multiple Formats** - CSV, TSV, JSONL, XLSX (import) + Parquet (export)  
✅ **Database Support** - PostgreSQL and MySQL with optimized batch operations  
✅ **Production-Grade** - Graceful shutdown, error handling, progress reporting  
✅ **Zero Native Compilation** - Prebuilt binaries downloaded automatically

## Installation

```bash
npm install data-engine
```

The package automatically downloads the appropriate prebuilt binary for your platform during installation.

**Supported Platforms:**

- Linux (x64, arm64)
- macOS (x64, arm64)
- Windows (x64)

## Quick Start

### Import CSV to PostgreSQL

```javascript
const { importData } = require("data-engine");

await importData({
  file: "/data/huge-dataset.csv",
  format: "auto", // auto-detect format
  dsn: "postgres://user:pass@localhost/mydb",
  table: "my_table",
  batchSize: 5000,
  workers: 0, // 0 = auto-detect CPU count
});
```

### Export Database to JSONL

```javascript
const { exportData } = require("data-engine");

await exportData({
  output: "/data/export.jsonl",
  format: "jsonl",
  dsn: "postgres://user:pass@localhost/mydb",
  query:
    "SELECT * FROM large_table WHERE created_at > NOW() - INTERVAL '30 days'",
  batchSize: 5000,
  workers: 0,
});
```

## TypeScript Support

The package includes full TypeScript type definitions for enhanced IDE support and type safety.

### TypeScript Usage

```typescript
import {
  importData,
  exportData,
  ImportOptions,
  ExportOptions,
} from "data-engine";

// Full type safety and autocomplete
const options: ImportOptions = {
  file: "./data.csv",
  format: "auto", // Autocomplete: 'auto' | 'csv' | 'tsv' | 'jsonl' | 'xlsx'
  dsn: "postgres://localhost/mydb",
  table: "users",
  batchSize: 10000,
  workers: 0,
};

await importData(options);

// Export with type-safe format
const exportOpts: ExportOptions = {
  output: "./export.parquet",
  format: "parquet", // Autocomplete: 'csv' | 'tsv' | 'jsonl' | 'parquet'
  dsn: "postgres://localhost/mydb",
  query: "SELECT * FROM users",
};

await exportData(exportOpts);
```

See [examples.ts](examples.ts) for more TypeScript examples.

## API Reference

### `importData(options)`

Import data from a file into a database.

**Options:**

- `file` (string, required) - Path to input file
- `format` (string) - File format: `auto`, `csv`, `tsv`, `jsonl`, `xlsx` (default: `auto`)
- `dsn` (string, required) - Database connection string
- `table` (string, required) - Target table name
- `batchSize` (number) - Rows per batch (default: `5000`)
- `workers` (number) - Worker count, `0` = auto-detect (default: `0`)

**Returns:** `Promise<void>`

**Example:**

```javascript
await importData({
  file: "./data.csv",
  format: "csv",
  dsn: "postgres://localhost/db",
  table: "users",
  batchSize: 10000,
});
```

### `exportData(options)`

Export data from a database to a file.

**Options:**

- `output` (string, required) - Path to output file
- `format` (string, required) - Output format: `csv`, `tsv`, `jsonl`, `parquet`
- `dsn` (string, required) - Database connection string
- `query` (string, required) - SQL query to execute
- `batchSize` (number) - Rows per batch (default: `5000`)
- `workers` (number) - Worker count, `0` = auto-detect (default: `0`)

**Returns:** `Promise<void>`

**Example:**

```javascript
await exportData({
  output: "./export.parquet",
  format: "parquet",
  dsn: "mysql://user:pass@localhost/db",
  query: "SELECT * FROM orders WHERE year = 2024",
});
```

## Supported Formats

### Import (File → Database)

| Format | Extension           | Notes                                  |
| ------ | ------------------- | -------------------------------------- |
| CSV    | `.csv`              | Fully supported, streaming             |
| TSV    | `.tsv`              | Tab-separated values                   |
| JSONL  | `.jsonl`, `.ndjson` | Newline-delimited JSON                 |
| XLSX   | `.xlsx`             | **Limited:** 100MB max, streaming only |

**Not Supported:**

- ❌ XLS (legacy Excel) - Convert to XLSX or CSV
- ❌ JSON arrays - Use JSONL instead

### Export (Database → File)

| Format  | Extension  | Notes                                    |
| ------- | ---------- | ---------------------------------------- |
| CSV     | `.csv`     | Comma-separated values                   |
| TSV     | `.tsv`     | Tab-separated values                     |
| JSONL   | `.jsonl`   | Newline-delimited JSON                   |
| Parquet | `.parquet` | Columnar format, optimized for analytics |

## Database Connection Strings

### PostgreSQL

```
postgres://user:password@host:port/database
postgresql://user:password@host:port/database?sslmode=require
```

### MySQL

```
user:password@tcp(host:port)/database
mysql://user:password@host:port/database
```

## Performance Tuning

### Batch Size

- **Smaller batches (1000-2000):** Lower memory, more frequent DB commits
- **Larger batches (10000-20000):** Higher throughput, more memory

### Workers

- **Default (0):** Auto-detects CPU count - recommended for most cases
- **Manual:** Set to CPU count for CPU-bound operations, or higher for I/O-bound

### Example: Tuning for 100M row import

```javascript
await importData({
  file: "/data/100m-rows.csv",
  dsn: "postgres://localhost/db",
  table: "events",
  batchSize: 10000, // Larger batches for throughput
  workers: 8, // Match CPU cores
});
```

**Expected Performance:**

- CSV import: ~100,000 - 500,000 rows/sec (depends on hardware and network)
- Memory usage: Constant (~50-200MB regardless of file size)

## Error Handling

The package provides detailed error messages and proper exit codes:

```javascript
try {
  await importData({
    file: "./data.csv",
    dsn: "postgres://localhost/db",
    table: "users",
  });
  console.log("Import successful!");
} catch (err) {
  console.error("Import failed:", err.message);
  // err.message contains detailed error information
}
```

**Common Errors:**

- Invalid DSN format
- File not found
- Unsupported format (XLS, JSON arrays)
- Database connection failure
- XLSX file exceeds 100MB limit

## Graceful Shutdown

The engine supports graceful shutdown on `SIGINT` (Ctrl+C) and `SIGTERM`:

```javascript
const operation = importData({
  file: "./huge.csv",
  dsn: "postgres://localhost/db",
  table: "data",
});

// User presses Ctrl+C
// Engine will:
// 1. Stop reading new data
// 2. Finish processing current batches
// 3. Exit cleanly without data corruption

await operation; // Will reject with "Operation cancelled by user"
```

## Production Deployment

### Docker Example

```dockerfile
FROM node:18-alpine

WORKDIR /app
COPY package*.json ./
RUN npm ci --production

COPY . .

CMD ["node", "your-script.js"]
```

### Environment Variables

```bash
# Database connection
export DATABASE_DSN="postgres://user:pass@db-host/mydb"

# Tuning
export BATCH_SIZE=10000
export WORKERS=8
```

### Monitoring

The engine outputs progress to stderr:

```
[INFO] Mode: import
[INFO] Workers: 8
[INFO] Batch Size: 5000
[INFO] Detected 15 columns: [id, name, email, ...]
[PROGRESS] Processed 500000 rows (125000 rows/sec)
[PROGRESS] Processed 1000000 rows (130000 rows/sec)
[SUCCESS] Operation completed successfully
[INFO] Import completed: 2500000 rows in 20.5 seconds (121951 rows/sec)
```

## Troubleshooting

### Binary not found

If the postinstall script fails to download the binary:

1. Manually download from [GitHub Releases](https://github.com/yourusername/data-engine/releases)
2. Place in `node_modules/data-engine/bin/`
3. Make executable: `chmod +x node_modules/data-engine/bin/data-engine`

### XLSX file too large

```
Error: XLSX file too large: 150000000 bytes (max 100000000 bytes)
```

**Solution:** Convert to CSV for large files:

```bash
# Using LibreOffice
libreoffice --headless --convert-to csv large-file.xlsx

# Or use online converters
```

### Out of memory

If you encounter OOM errors, reduce `batchSize`:

```javascript
await importData({
  // ... other options
  batchSize: 1000, // Reduce from default 5000
});
```

## License

MIT

## Contributing

Contributions welcome! Please open an issue or PR on [GitHub](https://github.com/yourusername/data-engine).
