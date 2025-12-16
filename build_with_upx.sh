#!/bin/bash

# Using UPX for building can sometimes cause problems and may be flagged as dangerous by some file virus scanning software. Please keep this in mind.

set -euo pipefail

# ----------------------------------------
# Config
# ----------------------------------------
PROJECT_NAME="mockserver"
BIN_DIR="./npm/bin"
CMD_DIR="./"         # Go main package location
VERSION="0.0.11"

# Color codes
RED="\033[0;31m"
GREEN="\033[0;32m"
YELLOW="\033[1;33m"
BLUE="\033[0;34m"
NC="\033[0m" # No Color

# Create bin directory if it doesn't exist
mkdir -p "$BIN_DIR"

# ----------------------------------------
# Compress binary with UPX if available
# ----------------------------------------
compress_with_upx() {
  local FILE=$1
  if command -v upx >/dev/null 2>&1; then
    if [[ "$FILE" == *macos* ]]; then
      echo -e "${YELLOW}Skipping UPX for macOS binary: $FILE${NC}"
    else
      echo -e "${BLUE}Compressing $FILE with UPX...${NC}"
      upx --best --lzma --force "$FILE"
      echo -e "${GREEN}Compressed: $FILE${NC}"
    fi
  else
    echo -e "${YELLOW}UPX not found. Skipping compression for $FILE${NC}"
  fi
}

# ----------------------------------------
# Function to build for a specific OS/ARCH
# ----------------------------------------
build() {
  local GOOS=$1
  local GOARCH=$2
  local OUTPUT=$3

  echo -e "${BLUE}Building $PROJECT_NAME for $GOOS/$GOARCH...${NC}"
  # -ldflags "-X main.version=$VERSION"
  env GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT" "$CMD_DIR"
  chmod +x "$OUTPUT"
  echo -e "${GREEN}Built: $OUTPUT${NC}"

  compress_with_upx "$OUTPUT"
}

# ----------------------------------------
# Build for multiple platforms
# ----------------------------------------
build linux amd64 "$BIN_DIR/$PROJECT_NAME-linux"
build darwin amd64 "$BIN_DIR/$PROJECT_NAME-macos"
build darwin arm64 "$BIN_DIR/$PROJECT_NAME-macos-arm64"
build windows amd64 "$BIN_DIR/$PROJECT_NAME.exe"

echo -e "${GREEN}All builds completed!${NC}"

# ----------------------------------------
# Link current OS binary for local testing
# ----------------------------------------
OS=$(uname | tr '[:upper:]' '[:lower:]')

case "$OS" in
  linux*)
    ln -sf "$BIN_DIR/$PROJECT_NAME-linux" "$BIN_DIR/$PROJECT_NAME"
    ;;
  darwin*)
    ln -sf "$BIN_DIR/$PROJECT_NAME-macos" "$BIN_DIR/$PROJECT_NAME"
    ;;
  mingw*|cygwin*|msys*)
    echo -e "${YELLOW}Use $BIN_DIR/$PROJECT_NAME.exe on Windows${NC}"
    ;;
  *)
    echo -e "${RED}Unsupported OS: $OS${NC}"
    exit 1
    ;;
esac

echo -e "${GREEN}Done! You can now run:${NC}"
echo -e "${BLUE}./$BIN_DIR/$PROJECT_NAME start --config mockserver.json${NC}"
echo -e "${BLUE}Version: $VERSION${NC}"
