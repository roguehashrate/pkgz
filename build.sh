#!/bin/bash

# pkgz build script
#   build.sh          interactive build (gzip-compressed binary)
#   build.sh --dev    quick host build + .tar.gz and .deb packages
set -e

VERSION="1.1.0"
OUTPUT_DIR="build"
BINARY_NAME="pkgz"

# Available options
OS_OPTIONS=("linux")
ARCH_OPTIONS=("amd64" "386" "arm64" "arm")

DEV_MODE=false
if [ "$1" == "--dev" ]; then
    DEV_MODE=true
fi

# Build a tarball distribution (.tar.gz) from the uncompressed binary
package_tarball() {
    local base="$(dirname "$output_path")"
    local tar_dir="$base/pkgz-dist"
    local tar_path="$base/pkgz-v$VERSION-$SELECTED_OS-$SELECTED_ARCH.tar.gz"

    echo "📦 Building tarball..."
    rm -rf "$tar_dir"
    mkdir -p "$tar_dir"
    cp "$output_path" "$tar_dir/pkgz"
    cp LICENSE README.md "$tar_dir/" 2>/dev/null || true
    tar -czf "$tar_path" -C "$base" "pkgz-dist"
    rm -rf "$tar_dir"
    echo "📦 Created: $tar_path"
}

# Build a .deb package (only meaningful for Linux)
package_deb() {
    local deb_arch=""
    case "$SELECTED_ARCH" in
        amd64) deb_arch="amd64" ;;
        386)   deb_arch="i386" ;;
        arm64) deb_arch="arm64" ;;
        arm)   deb_arch="armhf" ;;
    esac

    if [ "$SELECTED_OS" != "linux" ] || [ -z "$deb_arch" ] || ! command -v dpkg-deb >/dev/null 2>&1; then
        echo "⚠️  Skipping .deb (requires linux + dpkg-deb)"
        return
    fi

    local base="$(dirname "$output_path")"
    local stage="$base/deb-stage"
    local deb_path="$base/pkgz_${VERSION}_${deb_arch}.deb"

    echo "📦 Building .deb package..."
    rm -rf "$stage"
    mkdir -p "$stage/DEBIAN" "$stage/usr/local/bin"
    cp "$output_path" "$stage/usr/local/bin/pkgz"
    cat > "$stage/DEBIAN/control" <<EOF
Package: pkgz
Version: $VERSION
Section: utils
Priority: optional
Architecture: $deb_arch
Maintainer: roguehashrate <https://github.com/roguehashrate>
Description: Fast, extensible CLI tool for managing multiple package types on Linux.
EOF
    dpkg-deb --root-owner-group --build "$stage" "$deb_path" >/dev/null
    rm -rf "$stage"
    echo "📦 Created: $deb_path"
}

# --- Dev build: non-interactive host build with packages ---------------------
if [ "$DEV_MODE" == "true" ]; then
    SELECTED_OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$(uname -m)" in
        x86_64|amd64) SELECTED_ARCH="amd64" ;;
        aarch64|arm64) SELECTED_ARCH="arm64" ;;
        i386|i686)     SELECTED_ARCH="386" ;;
        arm*)          SELECTED_ARCH="arm" ;;
        *)             SELECTED_ARCH="amd64" ;;
    esac

    echo "🔨 pkgz v$VERSION Dev Build for $SELECTED_OS/$SELECTED_ARCH"
    echo "========================================================="

    output_path="$OUTPUT_DIR/$SELECTED_OS/$SELECTED_ARCH/$BINARY_NAME"
    mkdir -p "$(dirname "$output_path")"
    rm -f "$output_path" "$output_path.gz"

    GOOS=$SELECTED_OS GOARCH=$SELECTED_ARCH go build -ldflags="-s -w" -o "$output_path" .
    echo "✅ Built: $output_path"

    package_tarball
    package_deb

    echo
    echo "📂 Dev builds are in the $OUTPUT_DIR directory"
    exit 0
fi

# --- Interactive build -------------------------------------------------------
echo "🔨 pkgz v$VERSION Interactive Builder"
echo "=================================="
echo

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

output_path="$OUTPUT_DIR/$SELECTED_OS/$SELECTED_ARCH/$BINARY_NAME"
mkdir -p "$OUTPUT_DIR/$SELECTED_OS/$SELECTED_ARCH"
rm -f "$output_path" "$output_path.gz"

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