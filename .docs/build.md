This documentation provides a comprehensive guide to the **MockServer** build ecosystem. The project utilizes a sophisticated Go-based builder for cross-platform releases and shell scripts for specialized distribution tasks.

---

# Build Guide: MockServer

## Overview

The MockServer build system is designed to provide high-performance, lightweight binaries across multiple platforms including Linux, macOS, and Windows. The core build logic is handled by a dedicated Go script, supplemented by shell scripts for specific environments like NPM and compressed distributions.

---

## 🛠️ The Primary Builder: `builder.go`

The `scripts/builder.go` script is the central engine of the build process. It manages environment-specific flags and cross-compilation.

### Usage and Flags

You can execute the builder directly via `go run` or through the provided **Makefile**.

| Flag | Description |
| --- | --- |
| `-current` | Builds the binary only for the current operating system and architecture. |
| `-all` | Triggers cross-platform builds for Linux, macOS, and Windows. |
| `-npm` | Targets the build specifically for NPM distribution, outputting to `./npm/bin`. |
| `-out` | Specifies the output directory (default is typically `releases/latest`). |

### Makefile Shortcuts

For convenience, use the following commands:

* `make build`: Fast build for the current system.
* `make build-all`: Full cross-platform release build including NPM assets.
* `make build-npm`: Specialized build for NPM package distribution.

---

## Shell Scripts Automation

While `builder.go` handles the core logic, two shell scripts are provided for specialized workflows.

### 1. `build.sh`

This script is a streamlined wrapper for the NPM distribution pipeline.

* **Purpose**: It ensures the environment is correctly set up for NPM-specific binaries.
* **Mechanism**: It essentially triggers the `builder.go` script with the `-npm` flag to populate the necessary directories for the `mockserverx` package.

### 2. `build_with_upx.sh`

This script is used for creating highly compressed binaries.

* **Purpose**: It uses the **UPX (Ultimate Packer for eXecutables)** tool to significantly reduce the file size of the generated binaries.
* **Trade-off**: While it results in much smaller files, it may impact startup time slightly due to decompression.

> [!WARNING]
> **Security and False Positives**
> Executables packed with **UPX** are sometimes flagged as "Suspicious" or "Malicious" by various antivirus software and file integrity scanners. This is because the packer obfuscates the original binary structure, which is a technique also used by malware. Use this script only for environments where storage size is critical and you can whitelist the binary.

---

## 📋 Build Requirements

* **Go**: Version 1.21 or higher is required.
* **UPX**: Required only if using `build_with_upx.sh`.
* **CGO**: CGO is disabled (`CGO_ENABLED=0`) by default during builds to ensure maximum portability and smaller static binaries.

---

## Execution Pipeline

When you trigger a build, the system follows this internal hierarchy:

1. **Validation**: Checks for required Go modules and environment variables.
2. **Compilation**: The Go compiler generates the Linux/macOS/Windows binaries using specific `GOOS` and `GOARCH` settings.
3. **Artifact Management**: Binaries are moved to the designated `./bin`, `./npm/bin`, or `./releases` folders.
4. **Compression (Optional)**: If using the UPX script, the binaries are packed after the initial compilation.
