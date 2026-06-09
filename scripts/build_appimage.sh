#!/bin/bash
set -e

BINARY_SOURCE="${BINARY_SOURCE:-build/bin/multibrowser}"
ICON_SOURCE="${ICON_SOURCE:-build/appicon.png}"

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

cp "$ICON_SOURCE" multibrowser.png

# Detect WebKitGTK directory across distros (Debian/Ubuntu, Fedora, Arch, etc.)
WEBKIT_DIRS=(
    "/usr/lib/webkit2gtk-4.1"
    "/usr/lib/x86_64-linux-gnu/webkit2gtk-4.1"
    "/usr/lib64/webkit2gtk-4.1"
    "/usr/libexec/webkit2gtk-4.1"
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

    # Log the WebKit version for traceability (helps debugging distro-specific issues)
    if command -v pkg-config >/dev/null 2>&1 && pkg-config --exists webkit2gtk-4.1; then
        WEBKIT_VERSION="$(pkg-config --modversion webkit2gtk-4.1)"
        echo "WebKitGTK version: $WEBKIT_VERSION"
    elif command -v dpkg >/dev/null 2>&1; then
        dpkg -l libwebkit2gtk-4.1-0 2>/dev/null | awk '/^ii/ {print "WebKitGTK package: " $2 " " $3}'
    elif command -v pacman >/dev/null 2>&1; then
        pacman -Qi webkit2gtk-4.1 2>/dev/null | awk '/^Version/ {print "WebKitGTK version: " $3}'
    elif command -v rpm >/dev/null 2>&1; then
        rpm -q webkit2gtk4.1 2>/dev/null | head -1 | sed 's/^/WebKitGTK package: /'
    fi

    mkdir -p AppDir/usr/lib/webkit2gtk-4.1
    for proc in WebKitNetworkProcess WebKitWebProcess WebKitGPUProcess MiniBrowser; do
        if [ -f "$WEBKIT_DIR/$proc" ]; then
            cp "$WEBKIT_DIR/$proc" AppDir/usr/lib/webkit2gtk-4.1/
            echo "  + $proc"
        fi
    done

    # Also copy injected bundle directory if present
    if [ -d "$WEBKIT_DIR/injected-bundle" ]; then
        cp -r "$WEBKIT_DIR/injected-bundle" AppDir/usr/lib/webkit2gtk-4.1/
        echo "  + injected-bundle/"
    fi
else
    echo "WARNING: Could not find WebKitGTK directory. AppImage may not run on other distros."
fi

if [ ! -f "linuxdeploy-x86_64.AppImage" ]; then
    echo "Downloading linuxdeploy..."
    wget -q https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
    chmod +x linuxdeploy-x86_64.AppImage
fi

# Step 1: Let linuxdeploy bundle shared libraries and create AppDir structure
export APPIMAGE_EXTRACT_AND_RUN=1
export NO_STRIP=1
./linuxdeploy-x86_64.AppImage \
    --appdir AppDir \
    --executable "$BINARY_SOURCE" \
    --desktop-file multibrowser.desktop \
    --icon-file multibrowser.png

# Step 2: Deploy dependencies of WebKit helper binaries too
for proc in AppDir/usr/lib/webkit2gtk-4.1/*; do
    if [ -f "$proc" ] && [ -x "$proc" ]; then
        # Skip injected-bundle dir as linuxdeploy tries to patch rpath on .so files and fails
        if [[ "$proc" == *"injected-bundle"* ]]; then continue; fi
        ./linuxdeploy-x86_64.AppImage --appdir AppDir --executable "$proc" 2>/dev/null || true
    fi
done

# Step 3: Replace linuxdeploy's default AppRun with a custom one that sets
# WEBKIT_EXEC_PATH so WebKitGTK can find the bundled helpers at runtime.
# This is REQUIRED for portability across distros (Ubuntu/ZorinOS/Fedora/etc).
rm -f AppDir/AppRun
cat > AppDir/AppRun <<'EOF'
#!/bin/bash
if [ -n "$APPIMAGE" ] && [ -n "$APPDIR" ]; then
    HERE="$APPDIR"
else
    HERE="$(cd "$(dirname "$0")" && pwd)"
fi
export LD_LIBRARY_PATH="${HERE}/usr/lib:${HERE}/usr/lib/x86_64-linux-gnu:${LD_LIBRARY_PATH}"
export WEBKIT_EXEC_PATH="${HERE}/usr/lib/webkit2gtk-4.1"
export WEBKIT_INJECTED_BUNDLE_PATH="${HERE}/usr/lib/webkit2gtk-4.1/injected-bundle"
export GIO_MODULE_DIR="${HERE}/usr/lib/gio/modules"
# Disable DMA-BUF renderer — avoids black screen / rendering failures on
# Ubuntu 22.04 / ZorinOS and systems with older graphics stacks.
export WEBKIT_DISABLE_DMABUF_RENDERER=1
exec "${HERE}/usr/bin/multibrowser" "$@"
EOF
chmod +x AppDir/AppRun

# Step 4: Generate the AppImage using appimagetool
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
