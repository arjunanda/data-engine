const https = require('https');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const GITHUB_REPO = 'yourusername/data-engine';
const VERSION = require('./package.json').version;

/**
 * Download and install the prebuilt binary for the current platform
 */
async function install() {
  const platform = process.platform;
  const arch = process.arch;
  
  console.log(`Installing data-engine for ${platform}/${arch}...`);

  // Map Node.js platform/arch to binary naming
  const platformMap = {
    'darwin': 'darwin',
    'linux': 'linux',
    'win32': 'windows'
  };
  
  const archMap = {
    'x64': 'amd64',
    'arm64': 'arm64'
  };
  
  const mappedPlatform = platformMap[platform];
  const mappedArch = archMap[arch];
  
  if (!mappedPlatform || !mappedArch) {
    console.error(`Unsupported platform/architecture: ${platform}/${arch}`);
    console.error('Supported platforms: darwin/linux/windows with x64/arm64');
    process.exit(1);
  }
  
  // Construct binary name
  let binaryName = 'data-engine';
  const remoteBinaryName = `data-engine-${mappedPlatform}-${mappedArch}`;
  
  if (platform === 'win32') {
    binaryName += '.exe';
  }
  
  // Create bin directory if it doesn't exist
  const binDir = path.join(__dirname, 'bin');
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }
  
  const binaryPath = path.join(binDir, binaryName);
  
  // Check if binary already exists (for local development)
  if (fs.existsSync(binaryPath)) {
    console.log('Binary already exists, skipping download');
    makeExecutable(binaryPath);
    console.log('Installation complete!');
    return;
  }
  
  // Download from GitHub Releases
  const downloadUrl = `https://github.com/${GITHUB_REPO}/releases/download/v${VERSION}/${remoteBinaryName}`;
  
  console.log(`Downloading from: ${downloadUrl}`);
  
  try {
    await downloadFile(downloadUrl, binaryPath);
    makeExecutable(binaryPath);
    console.log('Installation complete!');
  } catch (err) {
    console.error('Failed to download binary:', err.message);
    console.error('\nYou can manually download the binary from:');
    console.error(downloadUrl);
    console.error(`And place it at: ${binaryPath}`);
    
    // Don't fail installation - allow manual setup
    console.warn('\nWarning: Binary not installed. Manual installation required.');
  }
}

/**
 * Download a file from a URL
 */
function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    
    https.get(url, (response) => {
      // Handle redirects
      if (response.statusCode === 301 || response.statusCode === 302) {
        file.close();
        fs.unlinkSync(dest);
        return downloadFile(response.headers.location, dest)
          .then(resolve)
          .catch(reject);
      }
      
      if (response.statusCode !== 200) {
        file.close();
        fs.unlinkSync(dest);
        return reject(new Error(`HTTP ${response.statusCode}: ${response.statusMessage}`));
      }
      
      response.pipe(file);
      
      file.on('finish', () => {
        file.close();
        resolve();
      });
    }).on('error', (err) => {
      file.close();
      fs.unlinkSync(dest);
      reject(err);
    });
    
    file.on('error', (err) => {
      file.close();
      fs.unlinkSync(dest);
      reject(err);
    });
  });
}

/**
 * Make the binary executable on Unix systems
 */
function makeExecutable(filePath) {
  if (process.platform !== 'win32') {
    try {
      fs.chmodSync(filePath, 0o755);
    } catch (err) {
      console.warn('Warning: Could not make binary executable:', err.message);
    }
  }
}

// Run installation
install().catch((err) => {
  console.error('Installation failed:', err);
  process.exit(1);
});
