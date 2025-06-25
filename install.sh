#!/usr/bin/env bash

set -e

REPO="roguehashrate/pkgz"
BIN_NAME="pkgz"
INSTALL_DIR="$HOME/.local/bin"
ARCH=$(uname -m)

# Only support x86_64
if [[ "$ARCH" != "x86_64" ]]; then
  echo "❌ Unsupported architecture: $ARCH"
  echo "This installer only supports x86_64 systems."
  exit 1
fi

# Get latest release tag
LATEST=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep "tag_name" | cut -d '"' -f 4)

URL="https://github.com/$REPO/releases/download/$LATEST/$BIN_NAME"

mkdir -p "$INSTALL_DIR"
curl -L "$URL" -o "$INSTALL_DIR/$BIN_NAME"
chmod +x "$INSTALL_DIR/$BIN_NAME"

echo "✅ Installed $BIN_NAME to $INSTALL_DIR"
echo "👉 Make sure $INSTALL_DIR is in your PATH"
