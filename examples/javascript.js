const { importData, exportData } = require('../index');

/**
 * Example: Import CSV to PostgreSQL
 */
async function exampleImport() {
  console.log('=== Import Example ===\n');
  
  try {
    await importData({
      file: './test-data/sample.csv',
      format: 'auto',
      dsn: 'postgres://user:password@localhost:5432/testdb',
      table: 'users',
      batchSize: 5000,
      workers: 0  // Auto-detect CPU count
    });
    
    console.log('✓ Import completed successfully!');
  } catch (err) {
    console.error('✗ Import failed:', err.message);
  }
}

/**
 * Example: Export database to JSONL
 */
async function exampleExport() {
  console.log('\n=== Export Example ===\n');
  
  try {
    await exportData({
      output: './test-data/export.jsonl',
      format: 'jsonl',
      dsn: 'postgres://user:password@localhost:5432/testdb',
      query: 'SELECT * FROM users LIMIT 10000',
      batchSize: 5000,
      workers: 0
    });
    
    console.log('✓ Export completed successfully!');
  } catch (err) {
    console.error('✗ Export failed:', err.message);
  }
}

/**
 * Example: Import JSONL to MySQL
 */
async function exampleJSONLImport() {
  console.log('\n=== JSONL Import Example ===\n');
  
  try {
    await importData({
      file: './test-data/data.jsonl',
      format: 'jsonl',
      dsn: 'user:password@tcp(localhost:3306)/testdb',
      table: 'events',
      batchSize: 10000,
      workers: 4
    });
    
    console.log('✓ JSONL import completed successfully!');
  } catch (err) {
    console.error('✗ JSONL import failed:', err.message);
  }
}

/**
 * Example: Export to Parquet for analytics
 */
async function exampleParquetExport() {
  console.log('\n=== Parquet Export Example ===\n');
  
  try {
    await exportData({
      output: './test-data/analytics.parquet',
      format: 'parquet',
      dsn: 'postgres://user:password@localhost:5432/analytics',
      query: `
        SELECT 
          date_trunc('day', created_at) as day,
          COUNT(*) as event_count,
          AVG(value) as avg_value
        FROM events
        WHERE created_at > NOW() - INTERVAL '30 days'
        GROUP BY day
        ORDER BY day
      `,
      batchSize: 5000
    });
    
    console.log('✓ Parquet export completed successfully!');
  } catch (err) {
    console.error('✗ Parquet export failed:', err.message);
  }
}

// Run examples
async function main() {
  console.log('Data Engine - Usage Examples\n');
  console.log('Note: These examples require a running database.');
  console.log('Update the DSN strings with your actual database credentials.\n');
  
  // Uncomment the examples you want to run:
  
  // await exampleImport();
  // await exampleExport();
  // await exampleJSONLImport();
  // await exampleParquetExport();
  
  console.log('\nTo run these examples:');
  console.log('1. Set up a PostgreSQL or MySQL database');
  console.log('2. Update the DSN connection strings');
  console.log('3. Uncomment the example functions above');
  console.log('4. Run: node examples.js');
}

main().catch(console.error);
