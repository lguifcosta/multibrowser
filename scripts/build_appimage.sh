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

cp build/appicon.png multibrowser.png

# Bundle WebKitGTK helper processes (required for portability)
WEBKIT_DIRS=(
    "/usr/lib/webkit2gtk-4.1"
    "/usr/lib/x86_64-linux-gnu/webkit2gtk-4.1"
    "/usr/lib64/webkit2gtk-4.1"
)
WEBKIT_DIR=""
for dir in "${WEBKIT_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        WEBKIT_DIR="$dir"
        break
    fi
done

if [ -n "$WEBKIT_DIR" ]; then
    echo "Bundling WebKit helpers from: $WEBKIT_DIR"
    mkdir -p AppDir/usr/lib/webkit2gtk-4.1
    for proc in WebKitNetworkProcess WebKitWebProcess; do
        if [ -f "$WEBKIT_DIR/$proc" ]; then
            cp "$WEBKIT_DIR/$proc" AppDir/usr/lib/webkit2gtk-4.1/
        fi
    done
else
    echo "WARNING: Could not find WebKitGTK directory."
fi

if [ ! -f "linuxdeploy-x86_64.AppImage" ]; then
    echo "Downloading linuxdeploy..."
    wget -q https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
    chmod +x linuxdeploy-x86_64.AppImage
fi

# Step 1: Let linuxdeploy deploy libraries and create structure
export APPIMAGE_EXTRACT_AND_RUN=1
export NO_STRIP=1
./linuxdeploy-x86_64.AppImage \
    --appdir AppDir \
    --executable "$BINARY_SOURCE" \
    --desktop-file multibrowser.desktop \
    --icon-file multibrowser.png

# Step 2: Replace AppRun symlink with custom script that sets WEBKIT_EXEC_PATH
# linuxdeploy creates AppRun as a symlink, must remove before writing
rm -f AppDir/AppRun
tee AppDir/AppRun > /dev/null <<'APPRUN'
#!/bin/bash
if [ -n "$APPIMAGE" ] && [ -n "$APPDIR" ]; then
    HERE="$APPDIR"
else
    HERE="$(cd "$(dirname "$0")" && pwd)"
fi
export LD_LIBRARY_PATH="${HERE}/usr/lib:${LD_LIBRARY_PATH}"
export WEBKIT_EXEC_PATH="${HERE}/usr/lib/webkit2gtk-4.1"
exec "${HERE}/usr/bin/multibrowser" "$@"
APPRUN
chmod +x AppDir/AppRun
# Verify it's a regular file, not symlink
test -f AppDir/AppRun && ! test -L AppDir/AppRun && echo "AppRun is a regular file (OK)"

# Step 3: Deploy WebKit helper libraries too
for proc in AppDir/usr/lib/webkit2gtk-4.1/*; do
    if [ -f "$proc" ]; then
        ./linuxdeploy-x86_64.AppImage --appdir AppDir --executable "$proc" 2>/dev/null || true
    fi
done

# Step 4: Generate AppImage using appimagetool
if [ ! -f "appimagetool-x86_64.AppImage" ]; then
    echo "Downloading appimagetool..."
    wget -q https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-x86_64.AppImage
    chmod +x appimagetool-x86_64.AppImage
fi

ARCH=x86_64 ./appimagetool-x86_64.AppImage AppDir

rm -f multibrowser.desktop multibrowser.png
mkdir -p dist
find . -maxdepth 1 -name 'MultiBrowser*.AppImage' -exec mv {} dist/ \;

echo "AppImage generated successfully in dist/"
