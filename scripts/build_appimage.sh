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

# Detect WebKitGTK directory across distros
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
    mkdir -p AppDir/usr/lib/webkit2gtk-4.1
    # Multiarch path support
    mkdir -p AppDir/usr/lib/x86_64-linux-gnu/webkit2gtk-4.1

    for proc in WebKitNetworkProcess WebKitWebProcess WebKitGPUProcess MiniBrowser; do
        if [ -f "$WEBKIT_DIR/$proc" ]; then
            cp "$WEBKIT_DIR/$proc" AppDir/usr/lib/webkit2gtk-4.1/
            cp "$WEBKIT_DIR/$proc" AppDir/usr/lib/x86_64-linux-gnu/webkit2gtk-4.1/
            echo "  + $proc"
        fi
    done

    if [ -d "$WEBKIT_DIR/injected-bundle" ]; then
        cp -r "$WEBKIT_DIR/injected-bundle" AppDir/usr/lib/webkit2gtk-4.1/
        cp -r "$WEBKIT_DIR/injected-bundle" AppDir/usr/lib/x86_64-linux-gnu/webkit2gtk-4.1/
        echo "  + injected-bundle/"
    fi
fi

if [ ! -f "linuxdeploy-x86_64.AppImage" ]; then
    echo "Downloading linuxdeploy..."
    wget -q https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
    chmod +x linuxdeploy-x86_64.AppImage
fi

# Step 1: Run linuxdeploy
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
        if [[ "$proc" == *"injected-bundle"* ]]; then continue; fi
        ./linuxdeploy-x86_64.AppImage --appdir AppDir --executable "$proc" 2>/dev/null || true
    fi
done
for proc in AppDir/usr/lib/x86_64-linux-gnu/webkit2gtk-4.1/*; do
    if [ -f "$proc" ] && [ -x "$proc" ]; then
        if [[ "$proc" == *"injected-bundle"* ]]; then continue; fi
        ./linuxdeploy-x86_64.AppImage --appdir AppDir --executable "$proc" 2>/dev/null || true
    fi
done

# Step 3: Ensure Icon is everywhere it needs to be
cp multibrowser.png AppDir/.DirIcon
cp multibrowser.png AppDir/multibrowser.png
mkdir -p AppDir/usr/share/icons/hicolor/512x512/apps/
cp multibrowser.png AppDir/usr/share/icons/hicolor/512x512/apps/multibrowser.png

# Step 4: Finalize AppRun with aggressive env vars
rm -f AppDir/AppRun
cat > AppDir/AppRun <<'EOF'
#!/bin/bash
HERE="$(dirname "$(readlink -f "${0}")")"

export LD_LIBRARY_PATH="${HERE}/usr/lib:${HERE}/usr/lib/x86_64-linux-gnu:${HERE}/lib:${HERE}/lib/x86_64-linux-gnu:${LD_LIBRARY_PATH}"

# Essential WebKit paths
export WEBKIT_EXEC_PATH="${HERE}/usr/lib/webkit2gtk-4.1"
export WEBKIT_PROCESS_PATH="${HERE}/usr/lib/webkit2gtk-4.1"
export WEBKIT_INJECTED_BUNDLE_PATH="${HERE}/usr/lib/webkit2gtk-4.1/injected-bundle"

# GLib/GTK fixes
export GIO_MODULE_DIR="${HERE}/usr/lib/gio/modules"
export GSETTINGS_SCHEMA_DIR="${HERE}/usr/share/glib-2.0/schemas"
export XDG_DATA_DIRS="${HERE}/usr/share:${XDG_DATA_DIRS:-/usr/local/share:/usr/share}"

# Graphics stability
export WEBKIT_DISABLE_DMABUF_RENDERER=1

exec "${HERE}/usr/bin/multibrowser" "$@"
EOF
chmod +x AppDir/AppRun

# Step 5: Generate the AppImage using appimagetool
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
