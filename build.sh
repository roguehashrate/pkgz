#!/bin/bash

# pkgz interactive build script
set -e

VERSION="0.1.9"
OUTPUT_DIR="build"
BINARY_NAME="pkgz"

# Available options
OS_OPTIONS=("linux" "darwin" "freebsd" "openbsd")
ARCH_OPTIONS=("amd64" "386" "arm64" "arm")

echo "🔨 pkgz v$VERSION Interactive Builder"
echo "=================================="
echo

# Ask for OS selection
echo "Available operating systems:"
for i in "${!OS_OPTIONS[@]}"; do
    echo "$((i+1)). ${OS_OPTIONS[i]}"
done
echo

while true; do
    read -p "Select operating system [1-${#OS_OPTIONS[@]}]: " os_choice
    if [[ "$os_choice" =~ ^[0-9]+$ ]] && [ "$os_choice" -ge 1 ] && [ "$os_choice" -le ${#OS_OPTIONS[@]} ]; then
        SELECTED_OS=${OS_OPTIONS[$((os_choice-1))]}
        break
    else
        echo "❌ Invalid choice. Please enter a number between 1 and ${#OS_OPTIONS[@]}"
    fi
done

echo
echo "Available architectures for $SELECTED_OS:"
for i in "${!ARCH_OPTIONS[@]}"; do
    echo "$((i+1)). ${ARCH_OPTIONS[i]}"
done
echo

while true; do
    read -p "Select architecture [1-${#ARCH_OPTIONS[@]}]: " arch_choice
    if [[ "$arch_choice" =~ ^[0-9]+$ ]] && [ "$arch_choice" -ge 1 ] && [ "$arch_choice" -le ${#ARCH_OPTIONS[@]} ]; then
        SELECTED_ARCH=${ARCH_OPTIONS[$((arch_choice-1))]}
        break
    else
        echo "❌ Invalid choice. Please enter a number between 1 and ${#ARCH_OPTIONS[@]}"
    fi
done

echo
echo "Building pkgz v$VERSION for $SELECTED_OS/$SELECTED_ARCH..."

# Create output directory
mkdir -p "$OUTPUT_DIR/$SELECTED_OS/$SELECTED_ARCH"

# Create output path
output_path="$OUTPUT_DIR/$SELECTED_OS/$SELECTED_ARCH/$BINARY_NAME"

# Build the binary
GOOS=$SELECTED_OS GOARCH=$SELECTED_ARCH go build -ldflags="-s -w" -o "$output_path" .

# Compress binary
if command -v gzip >/dev/null 2>&1; then
    echo "Compressing binary..."
    gzip "$output_path"
    
    if [ -f "$output_path.gz" ]; then
        echo "✅ Build complete!"
        echo "📦 Created: $output_path.gz"
        echo "💡 To install: cp $output_path.gz ~/.local/bin/ && cd ~/.local/bin && gunzip pkgz.gz && chmod +x pkgz"
    else
        echo "❌ Compression failed, using uncompressed binary"
        echo "✅ Build complete!"
        echo "📦 Created: $output_path"
        echo "💡 To install: cp $output_path ~/.local/bin/ && chmod +x ~/.local/bin/pkgz"
    fi
else
    echo "⚠️  gzip not found, skipping compression"
    echo "✅ Build complete!"
    echo "📦 Created: $output_path"
    echo "💡 To install: cp $output_path ~/.local/bin/ && chmod +x ~/.local/bin/pkgz"
fi

echo
echo "📂 All builds are in the $OUTPUT_DIR directory"
