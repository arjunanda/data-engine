const { spawn } = require("child_process");
const path = require("path");
const fs = require("fs");

/**
 * Get the path to the data-engine binary
 */
function getBinaryPath() {
  const platform = process.platform;
  const arch = process.arch;

  let binaryName = "data-engine";

  // Map Node.js platform/arch to binary naming
  const platformMap = {
    darwin: "darwin",
    linux: "linux",
    win32: "windows",
  };

  const archMap = {
    x64: "amd64",
    arm64: "arm64",
  };

  const mappedPlatform = platformMap[platform];
  const mappedArch = archMap[arch];

  if (!mappedPlatform || !mappedArch) {
    throw new Error(`Unsupported platform/architecture: ${platform}/${arch}`);
  }

  if (platform === "win32") {
    binaryName += ".exe";
  }

  const binaryPath = path.join(__dirname, "bin", binaryName);

  if (!fs.existsSync(binaryPath)) {
    throw new Error(
      `Binary not found at ${binaryPath}. ` +
        `Please ensure the package was installed correctly. ` +
        `Platform: ${platform}, Arch: ${arch}`
    );
  }

  return binaryPath;
}

/**
 * Import data from a file into a database
 * @param {Object} options - Import options
 * @param {string} options.file - Path to input file
 * @param {string} options.format - File format (auto, csv, tsv, jsonl, xlsx)
 * @param {string} options.dsn - Database connection string
 * @param {string} options.table - Target table name
 * @param {number} [options.batchSize=5000] - Batch size for inserts
 * @param {number} [options.workers=0] - Number of workers (0 = auto)
 * @returns {Promise<void>}
 */
async function importData(options) {
  const {
    file,
    format = "auto",
    dsn,
    table,
    batchSize = 5000,
    workers = 0,
  } = options;

  // Validate required options
  if (!file) throw new Error("file is required");
  if (!dsn) throw new Error("dsn is required");
  if (!table) throw new Error("table is required");

  const config = {
    mode: "import",
    input_file: file,
    input_format: format,
    dsn,
    table,
    batch_size: batchSize,
    workers,
    progress_every: 100000,
  };

  return runEngine(config);
}

/**
 * Export data from a database to a file
 * @param {Object} options - Export options
 * @param {string} options.output - Path to output file
 * @param {string} options.format - Output format (csv, tsv, jsonl, parquet)
 * @param {string} options.dsn - Database connection string
 * @param {string} options.query - SQL query to execute
 * @param {number} [options.batchSize=5000] - Batch size for reads
 * @param {number} [options.workers=0] - Number of workers (0 = auto)
 * @returns {Promise<void>}
 */
async function exportData(options) {
  const { output, format, dsn, query, batchSize = 5000, workers = 0 } = options;

  // Validate required options
  if (!output) throw new Error("output is required");
  if (!format) throw new Error("format is required");
  if (!dsn) throw new Error("dsn is required");
  if (!query) throw new Error("query is required");

  const config = {
    mode: "export",
    output_file: output,
    output_format: format,
    dsn,
    query,
    batch_size: batchSize,
    workers,
    progress_every: 100000,
  };

  return runEngine(config);
}

/**
 * Run the Go engine with the given configuration
 * @param {Object} config - Configuration object
 * @returns {Promise<void>}
 */
function runEngine(config) {
  return new Promise((resolve, reject) => {
    const binaryPath = getBinaryPath();

    // Spawn the Go binary
    const child = spawn(binaryPath, [], {
      stdio: ["pipe", "inherit", "pipe"],
    });

    let stderrData = "";

    // Capture stderr for logging
    child.stderr.on("data", (data) => {
      const text = data.toString();
      stderrData += text;
      // Stream logs to stderr
      process.stderr.write(text);
    });

    // Handle process exit
    child.on("close", (code) => {
      if (code === 0) {
        resolve();
      } else if (code === 130) {
        // Graceful shutdown (SIGINT)
        reject(new Error("Operation cancelled by user"));
      } else {
        reject(
          new Error(`Engine failed with exit code ${code}\n${stderrData}`)
        );
      }
    });

    // Handle process errors
    child.on("error", (err) => {
      reject(new Error(`Failed to start engine: ${err.message}`));
    });

    // Send configuration via stdin
    try {
      child.stdin.write(JSON.stringify(config));
      child.stdin.end();
    } catch (err) {
      child.kill();
      reject(new Error(`Failed to send configuration: ${err.message}`));
    }
  });
}

module.exports = {
  importData,
  exportData,
};
