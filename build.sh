#!/bin/bash

# pkgz cross-compilation build script
set -e

VERSION="0.1.9"
OUTPUT_DIR="build"
BINARY_NAME="pkgz"

# Create build directory
mkdir -p "$OUTPUT_DIR"

echo "🔨 Building pkgz v$VERSION for multiple architectures..."

# Build for different architectures
platforms=(
    "linux/amd64"
    "linux/386" 
    "linux/arm64"
    "linux/arm"
    "darwin/amd64"
    "darwin/arm64"
    "freebsd/amd64"
    "openbsd/amd64"
)

for platform in "${platforms[@]}"; do
    GOOS=$(echo $platform | cut -d'/' -f1)
    GOARCH=$(echo $platform | cut -d'/' -f2)
    
    output_name="$BINARY_NAME-$GOOS-$GOARCH"
    if [ $GOOS = "windows" ]; then
        output_name="$output_name.exe"
    fi
    
    echo "Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "$OUTPUT_DIR/$output_name" .
    
    # Compress the binary
    if command -v gzip >/dev/null 2>&1; then
        gzip -c "$OUTPUT_DIR/$output_name" > "$OUTPUT_DIR/$output_name.gz"
        rm "$OUTPUT_DIR/$output_name"
    fi
done

echo "✅ Build complete!"
echo "Binaries are available in the $OUTPUT_DIR directory:"
ls -la "$OUTPUT_DIR"
