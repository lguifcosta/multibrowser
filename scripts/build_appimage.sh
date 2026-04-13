#!/bin/bash
set -e

BINARY_SOURCE="build/bin/multibrowser"

if [ ! -f "$BINARY_SOURCE" ]; then
    echo "ERROR: Binary not found at $BINARY_SOURCE."
    echo "Please run 'wails build -tags webkit2_41' first."
    exit 1
fi

echo "Using binary: $BINARY_SOURCE"

rm -rf AppDir
mkdir -p AppDir/usr/bin
cp "$BINARY_SOURCE" AppDir/usr/bin/multibrowser

cat > multibrowser.desktop <<EOL
[Desktop Entry]
Name=MultiBrowser
Exec=multibrowser
Icon=multibrowser
Type=Application
Categories=Utility;
Terminal=false
EOL

# Icon must match the Icon= field in .desktop (multibrowser.png)
cp build/appicon.png multibrowser.png

if [ ! -f "linuxdeploy-x86_64.AppImage" ]; then
    echo "Downloading linuxdeploy..."
    wget -q https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
    chmod +x linuxdeploy-x86_64.AppImage
fi

export APPIMAGE_EXTRACT_AND_RUN=1
export NO_STRIP=1
./linuxdeploy-x86_64.AppImage \
    --appdir AppDir \
    --executable "$BINARY_SOURCE" \
    --desktop-file multibrowser.desktop \
    --icon-file multibrowser.png \
    --output appimage

rm -f multibrowser.desktop multibrowser.png
mkdir -p dist
find . -maxdepth 1 -name '*.AppImage' ! -name 'linuxdeploy*' -exec mv {} dist/ \;

echo "AppImage generated successfully in dist/"
