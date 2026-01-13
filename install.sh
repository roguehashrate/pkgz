#!/bin/bash

# pkgz installation script
set -e

INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="pkgz"

echo "🚀 Installing pkgz..."

# Create install directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Download and install pkgz
if command -v curl >/dev/null 2>&1; then
    curl -L -o "$INSTALL_DIR/$BINARY_NAME" "https://github.com/roguehashrate/pkgz/releases/latest/download/pkgz"
elif command -v wget >/dev/null 2>&1; then
    wget -O "$INSTALL_DIR/$BINARY_NAME" "https://github.com/roguehashrate/pkgz/releases/latest/download/pkgz"
else
    echo "❌ Neither curl nor wget is available. Please install one of them."
    exit 1
fi

# Make binary executable
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Check if install directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "⚠️  $INSTALL_DIR is not in your PATH"
    echo "Please add the following line to your shell profile:"
    echo "export PATH=\"\$PATH:$INSTALL_DIR\""
    echo ""
    echo "Then run: source ~/.bashrc  # or ~/.zshrc, etc."
fi

echo "✅ pkgz installed successfully!"
echo "Run 'pkgz --version' to verify installation."
