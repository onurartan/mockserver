#!/bin/bash
set -euo pipefail

# ----------------------------------------
# Config
# ----------------------------------------
PROJECT_NAME="mockserver"
BIN_DIR="./npm/bin"
CMD_DIR="."     # Go main package location
VERSION="1.0.0"

# Color codes
RED="\033[0;31m"
GREEN="\033[0;32m"
YELLOW="\033[1;33m"
BLUE="\033[0;34m"
NC="\033[0m" # No Color

# Create bin directory if it doesn't exist
mkdir -p "$BIN_DIR"

# ----------------------------------------
# Function to build for a specific OS/ARCH
# ----------------------------------------
build() {
  local GOOS=$1
  local GOARCH=$2
  local OUTPUT=$3

  echo -e "${BLUE}Building $PROJECT_NAME for $GOOS/$GOARCH...${NC}"
  env GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT" "$CMD_DIR"
  chmod +x "$OUTPUT"
  echo -e "${GREEN}Built: $OUTPUT${NC}"
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
echo -e "${BLUE}./$BIN_DIR/$PROJECT_NAME start --config example/mockserver.yaml${NC}"
echo -e "${BLUE}Version: $VERSION${NC}"
