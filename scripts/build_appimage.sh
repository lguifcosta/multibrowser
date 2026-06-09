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

# Bundle GTK runtime data (NOT traced by linuxdeploy: loaded at runtime).
# Missing these is the main cause of crashes on systems without GTK installed.
LIB_PREFIXES=(
    "/usr/lib/x86_64-linux-gnu"
    "/usr/lib64"
    "/usr/lib"
)

find_in_prefixes() {
    # $1 = relative path under a lib prefix; prints first match
    for p in "${LIB_PREFIXES[@]}"; do
        if [ -e "$p/$1" ]; then echo "$p/$1"; return 0; fi
    done
    return 1
}

# 1) GSettings schemas (required: GTK aborts on g_settings_new if missing)
if [ -f "/usr/share/glib-2.0/schemas/gschemas.compiled" ]; then
    echo "Bundling GSettings schemas"
    mkdir -p AppDir/usr/share/glib-2.0/schemas
    cp /usr/share/glib-2.0/schemas/*.xml AppDir/usr/share/glib-2.0/schemas/ 2>/dev/null || true
    cp /usr/share/glib-2.0/schemas/*.gschema.override AppDir/usr/share/glib-2.0/schemas/ 2>/dev/null || true
    if command -v glib-compile-schemas >/dev/null 2>&1 && ls AppDir/usr/share/glib-2.0/schemas/*.xml >/dev/null 2>&1; then
        glib-compile-schemas AppDir/usr/share/glib-2.0/schemas/ >/dev/null 2>&1 || \
            cp /usr/share/glib-2.0/schemas/gschemas.compiled AppDir/usr/share/glib-2.0/schemas/
    else
        cp /usr/share/glib-2.0/schemas/gschemas.compiled AppDir/usr/share/glib-2.0/schemas/
    fi
fi

# 2) GIO modules (TLS via glib-networking/gnutls -> HTTPS)
if GIO_SRC=$(find_in_prefixes "gio/modules"); then
    echo "Bundling GIO modules from: $GIO_SRC"
    mkdir -p AppDir/usr/lib/gio/modules
    cp "$GIO_SRC"/*.so AppDir/usr/lib/gio/modules/ 2>/dev/null || true
    if command -v gio-querymodules >/dev/null 2>&1; then
        gio-querymodules AppDir/usr/lib/gio/modules >/dev/null 2>&1 || true
    fi
fi

# 3) gdk-pixbuf loaders (external on Debian/Ubuntu; built-in on Arch gdk-pixbuf>=2.44)
if PIXBUF_BASE=$(find_in_prefixes "gdk-pixbuf-2.0/2.10.0/loaders"); then
    if ls "$PIXBUF_BASE"/*.so >/dev/null 2>&1; then
        echo "Bundling gdk-pixbuf loaders from: $PIXBUF_BASE"
        mkdir -p AppDir/usr/lib/gdk-pixbuf-2.0/2.10.0/loaders
        cp "$PIXBUF_BASE"/*.so AppDir/usr/lib/gdk-pixbuf-2.0/2.10.0/loaders/ 2>/dev/null || true
        # query-loaders echoes the argument paths verbatim, so we feed absolute
        # build paths; AppRun rewrites the known suffix to the relocated dir.
        QUERY=$(command -v gdk-pixbuf-query-loaders-64 gdk-pixbuf-query-loaders 2>/dev/null | head -1)
        if [ -n "$QUERY" ]; then
            "$QUERY" "$PWD"/AppDir/usr/lib/gdk-pixbuf-2.0/2.10.0/loaders/*.so \
                > AppDir/usr/lib/gdk-pixbuf-2.0/2.10.0/loaders.cache 2>/dev/null || true
        fi
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

# gdk-pixbuf loaders: rewrite the build-time absolute paths in the cache to the
# relocated loader dir, into a writable per-run cache file.
PIXBUF_LOADERDIR="${HERE}/usr/lib/gdk-pixbuf-2.0/2.10.0/loaders"
if [ -f "${HERE}/usr/lib/gdk-pixbuf-2.0/2.10.0/loaders.cache" ]; then
    export GDK_PIXBUF_MODULEDIR="${PIXBUF_LOADERDIR}"
    RUNTIME_CACHE="$(mktemp 2>/dev/null || echo /tmp/mb-pixbuf.cache)"
    sed "s|\"[^\"]*/gdk-pixbuf-2.0/2.10.0/loaders/|\"${PIXBUF_LOADERDIR}/|g" \
        "${HERE}/usr/lib/gdk-pixbuf-2.0/2.10.0/loaders.cache" > "${RUNTIME_CACHE}" 2>/dev/null \
        && export GDK_PIXBUF_MODULE_FILE="${RUNTIME_CACHE}"
fi

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
