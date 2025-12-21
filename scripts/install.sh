#!/bin/bash
set -e

# Define colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Function to print error and exit
fail() {
    echo -e "${RED}Error: $1${NC}"
    exit 1
}

# Function to print success message
success() {
    echo -e "${GREEN}$1${NC}"
}

# Detect OS and Architecture
OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
    Darwin)
        OS_TYPE="Darwin"
        ;;
    Linux)
        OS_TYPE="Linux"
        ;;
    *)
        fail "Unsupported OS: $OS"
        ;;
esac

case "$ARCH" in
    x86_64)
        ARCH_TYPE="x86_64"
        ;;
    arm64|aarch64)
        ARCH_TYPE="arm64"
        ;;
    *)
        fail "Unsupported Architecture: $ARCH"
        ;;
esac

REPO_OWNER="MagdielCAS"
REPO_NAME="magi-cli"
BINARY_NAME="magi"

echo "Detected OS: $OS_TYPE"
echo "Detected Arch: $ARCH_TYPE"

# Fetch latest release data
echo "Fetching latest release..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest")

if [ -z "$LATEST_RELEASE" ]; then
    fail "Could not fetch release information. Please check your internet connection."
fi

# Extract tag name (version)
VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    fail "Could not determine latest version."
fi

echo "Latest version: $VERSION"

# Construct asset name based on goreleaser config
# Template: {{ .ProjectName }}_{{ title .Os }}_{{ .Arch }}{{ if .Arm }}v{{ .Arm }}{{ end }}
# Example: magi-cli_Darwin_arm64.tar.gz
ASSET_NAME="${REPO_NAME}_${OS_TYPE}_${ARCH_TYPE}.tar.gz"

echo "Looking for asset: $ASSET_NAME"

# Get download URL for the specific asset
DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep "browser_download_url" | grep "$ASSET_NAME" | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
    fail "Could not find download URL for $ASSET_NAME in release $VERSION"
fi

echo "Downloading from: $DOWNLOAD_URL"

# Create temp directory
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Download the asset
curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/$ASSET_NAME"

# Extract
echo "Extracting..."
tar -xzf "$TMP_DIR/$ASSET_NAME" -C "$TMP_DIR"

# Install
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
if [ ! -w "$INSTALL_DIR" ]; then
    echo "Sudo permission required to install to $INSTALL_DIR"
    sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
else
    mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
fi

success "Successfully installed $BINARY_NAME $VERSION to $INSTALL_DIR"
