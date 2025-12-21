#!/bin/bash
set -e

# Define colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Helper functions
info() { echo -e "${NC}$1"; }
success() { echo -e "${GREEN}$1${NC}"; }
error() { echo -e "${RED}Error: $1${NC}"; exit 1; }
verbose() { [ "$VERBOSE" = "true" ] && echo -e "${NC}VERBOSE: $1"; }

# Setup variables
OWNER="MagdielCAS"
REPO="magi-cli"
BINARY_NAME="magi"
VERBOSE="false"

# Check for verbose flag
for arg in "$@"; do
    if [ "$arg" == "--verbose" ] || [ "$arg" == "-v" ]; then
        VERBOSE="true"
    fi
done

info "Starting installation for $OWNER/$REPO..."

# Detect OS and Architecture
OS=$(uname -s)
ARCH=$(uname -m)
OS_LOWER=$(echo "$OS" | tr '[:upper:]' '[:lower:]')
ARCH_LOWER=$(echo "$ARCH" | tr '[:upper:]' '[:lower:]')

verbose "Detected OS: $OS ($OS_LOWER)"
verbose "Detected Arch: $ARCH ($ARCH_LOWER)"

# Set aliases
declare -a OS_ALIASES
declare -a ARCH_ALIASES

# OS Aliases
if [ "$OS_LOWER" == "linux" ]; then
    OS_ALIASES=("linux")
elif [ "$OS_LOWER" == "darwin" ]; then
    OS_ALIASES=("darwin" "macos" "osx")
else
    error "Unsupported OS: $OS"
fi

# Arch Aliases
if [ "$ARCH" == "x86_64" ]; then
    ARCH_ALIASES=("amd64" "x86_64" "x86-64" "x64")
elif [ "$ARCH" == "i386" ] || [ "$ARCH" == "i686" ]; then
    ARCH_ALIASES=("386" "i386" "i686" "x86")
elif [[ "$ARCH" =~ ^arm ]]; then
    if [ "$ARCH" == "arm64" ] || [ "$ARCH" == "aarch64" ]; then
        ARCH_ALIASES=("arm64" "aarch64")
    else
        ARCH_ALIASES=("arm" "armv7" "armv6" "armv8" "armv8l" "armv7l" "armv6l")
    fi
else
    error "Unsupported Architecture: $ARCH"
fi

verbose "OS Aliases: ${OS_ALIASES[*]}"
verbose "Arch Aliases: ${ARCH_ALIASES[*]}"

# Create temp directory
TMP_DIR=$(mktemp -d)
verbose "Created temp dir: $TMP_DIR"
trap 'rm -rf "$TMP_DIR"' EXIT

# Fetch latest release
RELEASE_URL="https://api.github.com/repos/$OWNER/$REPO/releases/latest"
info "Fetching latest release from GitHub..."
RELEASE_JSON=$(curl -sSL "$RELEASE_URL")

if [ -z "$RELEASE_JSON" ]; then
    error "Failed to fetch release information."
fi

TAG_NAME=$(echo "$RELEASE_JSON" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
info "Found latest version: $TAG_NAME"

# Parse assets
ASSETS=$(echo "$RELEASE_JSON" | grep "browser_download_url" | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$ASSETS" ]; then
    error "No assets found for this release."
fi

verbose "Found assets:"
verbose "$ASSETS"

# Asset Selection Scoring
BEST_ASSET_URL=""
MAX_SCORE=0
BEST_ASSET_NAME=""

for ASSET_URL in $ASSETS; do
    SCORE=0
    FILENAME=$(basename "$ASSET_URL")
    FILENAME_LOWER=$(echo "$FILENAME" | tr '[:upper:]' '[:lower:]')
    
    # +10 for matching OS
    for ALIAS in "${OS_ALIASES[@]}"; do
        if [[ "$FILENAME_LOWER" == *"$ALIAS"* ]]; then
            SCORE=$((SCORE + 10))
            break
        fi
    done

    # +2 for matching Arch
    for ALIAS in "${ARCH_ALIASES[@]}"; do
        if [[ "$FILENAME_LOWER" == *"$ALIAS"* ]]; then
            verbose "  +2 Arch match ($ALIAS): $FILENAME"
            SCORE=$((SCORE + 2))
            break
        fi
    done

    # +1 for archive extension
    if [[ "$FILENAME_LOWER" == *".tar.gz" ]] || [[ "$FILENAME_LOWER" == *".zip" ]] || [[ "$FILENAME_LOWER" == *".tar.bz2" ]]; then
        SCORE=$((SCORE + 1))
    fi
    
    # +2 if it contains repo name, +1 if exact match (unlikely for archive, but good logic)
    if [[ "$FILENAME_LOWER" == *"$REPO"* ]]; then
        SCORE=$((SCORE + 2))
    fi

    verbose "Scored $FILENAME: $SCORE"

    if [ $SCORE -gt $MAX_SCORE ]; then
        MAX_SCORE=$SCORE
        BEST_ASSET_URL=$ASSET_URL
        BEST_ASSET_NAME=$FILENAME
    fi
done

if [ -z "$BEST_ASSET_URL" ] || [ $MAX_SCORE -lt 10 ]; then
    error "Could not find a suitable asset for your system."
fi

info "Selected asset: $BEST_ASSET_NAME"

# Download
info "Downloading..."
curl -sL "$BEST_ASSET_URL" -o "$TMP_DIR/$BEST_ASSET_NAME"

# Extract
info "Extracting..."
if [[ "$BEST_ASSET_NAME" == *".tar.gz" ]]; then
    tar -xzf "$TMP_DIR/$BEST_ASSET_NAME" -C "$TMP_DIR"
elif [[ "$BEST_ASSET_NAME" == *".tar.bz2" ]]; then
    tar -xjf "$TMP_DIR/$BEST_ASSET_NAME" -C "$TMP_DIR"
elif [[ "$BEST_ASSET_NAME" == *".zip" ]]; then
    unzip "$TMP_DIR/$BEST_ASSET_NAME" -d "$TMP_DIR" >/dev/null
else
    error "Unknown archive format: $BEST_ASSET_NAME"
fi

# Find Binary
BINARY_PATH=$(find "$TMP_DIR" -type f -name "$BINARY_NAME" | head -n 1)

# If not found exactly, try to find any executable that might be it (heuristic)
if [ -z "$BINARY_PATH" ]; then
     verbose "Binary not found by exact name, searching for executable files..."
     # Look for files that are executable and not the archive itself
     # This is a bit risky but standard for installers if names mismatch
     BINARY_PATH=$(find "$TMP_DIR" -type f -maxdepth 2 -perm +111 -not -name "*.*" | head -n 1)
fi

if [ -z "$BINARY_PATH" ]; then
    error "Could not locate binary '$BINARY_NAME' in the downloaded archive."
fi

verbose "Found binary at: $BINARY_PATH"

# Install Location
INSTALL_DIR="$HOME/.local/bin"
if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR"
fi

# Move Binary
info "Installing to $INSTALL_DIR..."
mv "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Path Configuration
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    info "Warning: $INSTALL_DIR is not in your PATH."
    
    SHELL_CONFIG=""
    if [ -n "$ZSH_VERSION" ]; then
        SHELL_CONFIG="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
        SHELL_CONFIG="$HOME/.bashrc"
    else
        SHELL_CONFIG="$HOME/.profile"
    fi

    if [ -f "$SHELL_CONFIG" ]; then
        info "You can add it by running:"
        echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> $SHELL_CONFIG"
        echo "  source $SHELL_CONFIG"
    else
        echo "Please add '$INSTALL_DIR' to your PATH manually."
    fi
fi

success "Installation complete! Try running '$BINARY_NAME --version'"
