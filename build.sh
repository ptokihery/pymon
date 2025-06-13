#!/bin/bash
set -e

# Variables
APP_NAME="pymon"
VERSION="1.0.0"
ARCH="amd64"
MAINTAINER="Tokihery <tokihery15@gmail.com>"

# 1. Build Go binary
echo "🔧 Building Go binary..."
GOOS=linux GOARCH=amd64 go build -o "$APP_NAME"

# need to upx the binary
if command -v upx &> /dev/null; then
    echo "🔧 Compressing binary with UPX..."
    upx --best --ultra-brute "$APP_NAME"
else
    echo "⚠️ UPX not found, skipping compression."
fi

# 2. Setup .deb directory structure
echo "📦 Creating folder structure..."
PKG_DIR="${APP_NAME}_${VERSION}"
mkdir -p "$PKG_DIR/DEBIAN"
mkdir -p "$PKG_DIR/usr/local/bin"
mkdir -p "$PKG_DIR/etc/$APP_NAME"

# 3. Copy binary
cp "$APP_NAME" "$PKG_DIR/usr/local/bin/"
chmod 755 "$PKG_DIR/usr/local/bin/$APP_NAME"


# 5. Control file
cat <<EOF > "$PKG_DIR/DEBIAN/control"
Package: $APP_NAME
Version: $VERSION
Section: base
Priority: optional
Architecture: $ARCH
Maintainer: $MAINTAINER
Description: $APP_NAME - Simple Go app for alerting the user for system events
 A minimal Go application packaged for Linux as a .deb
EOF

# 6. Build .deb package
echo "🧱 Building .deb..."
dpkg-deb --build "$PKG_DIR"

echo "✅ Done: ${PKG_DIR}.deb"
