#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const os = require('os');

const binDir = path.join(__dirname, 'bin');
const platform = os.platform();
const arch = os.arch();

let srcFile;

if (platform === 'win32') {
  srcFile = 'mockserver.exe';
} else if (platform === 'darwin') {
  srcFile = arch === 'arm64'
    ? 'mockserver-macos-arm64'
    : 'mockserver-macos';
} else if (platform === 'linux') {
  srcFile = arch === 'arm64'
    ? 'mockserver-linux-arm64'
    : 'mockserver-linux';
} else {
  console.error(`Unsupported platform/arch: ${platform}/${arch}`);
  process.exit(1);
}

const srcPath = path.join(binDir, srcFile);
const destPath = path.join(binDir, 'mockserver');

if (!fs.existsSync(srcPath)) {
  console.error(`Binary not found: ${srcPath}`);
  process.exit(1);
}

fs.copyFileSync(srcPath, destPath);

if (platform !== 'win32') {
  fs.chmodSync(destPath, 0o755);
}

console.log(`Setup completed for ${platform}/${arch}: ${destPath}`);
