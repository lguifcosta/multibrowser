#!/bin/bash
set -e

APP=multibrowser

# Determine the binary source
# 1. First check GoReleaser path (as per .goreleaser.yml ID 'linux')
# 2. Fallback to local 'wails build' path for easier local testing
BINARY_SOURCE="dist/linux_linux_amd64/multibrowser"

if [ ! -f "$BINARY_SOURCE" ]; then
    echo "GoReleaser binary not found at $BINARY_SOURCE, checking local wails build..."
    BINARY_SOURCE="build/bin/multibrowser"
fi

if [ ! -f "$BINARY_SOURCE" ]; then
    echo "ERROR: Binary not found at $BINARY_SOURCE."
    echo "Please run 'wails build -tags webkit2_41' first."
    exit 1
fi

echo "Using binary: $BINARY_SOURCE"

mkdir -p AppDir/usr/bin
cp "$BINARY_SOURCE" AppDir/usr/bin/multibrowser

mkdir -p AppDir/usr/share/applications
cat > AppDir/usr/share/applications/multibrowser.desktop <<EOL
[Desktop Entry]
Name=MultiBrowser
Exec=multibrowser
Icon=multibrowser
Type=Application
Categories=Utility;
Terminal=false
EOL

mkdir -p AppDir/usr/share/icons/hicolor/256x256/apps
cp build/appicon.png AppDir/usr/share/icons/hicolor/256x256/apps/multibrowser.png

if [ ! -f "linuxdeploy-x86_64.AppImage" ]; then
    echo "Downloading linuxdeploy..."
    wget -q https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
    chmod +x linuxdeploy-x86_64.AppImage
fi

export APPIMAGE_EXTRACT_AND_RUN=1
export NO_STRIP=1
./linuxdeploy-x86_64.AppImage --appdir AppDir --output appimage

# Create dist if it doesn't exist (for local runs)
mkdir -p dist
mv *.AppImage dist/

echo "AppImage generated successfully in dist/"
