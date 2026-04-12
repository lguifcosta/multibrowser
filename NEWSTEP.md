# MultiBrowser — Build & Release (GoReleaser + GitHub Actions)

## Objetivo

Automatizar completamente:

* build Linux (.AppImage + binário)
* build Windows (.exe)
* versionamento via tag
* publicação automática no GitHub Releases

---

## Stack de Release

* GoReleaser
* GitHub Actions
* linuxdeploy (AppImage)

---

## Estrutura Esperada

```id="z4p2kd"
.
├── build/
├── frontend/
├── AppDir/           # gerado no pipeline
├── .goreleaser.yml
└── .github/workflows/release.yml
```

---

## 1. Configuração do GoReleaser

Cria `.goreleaser.yml`:

```yaml id="goreleaser-config"
project_name: multibrowser

builds:
  - id: app
    main: ./main.go
    binary: multibrowser
    env:
      - CGO_ENABLED=1
    goos:
      - linux
      - windows
    goarch:
      - amd64

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

release:
  github:
    owner: SEU_USER
    name: SEU_REPO

checksum:
  name_template: "checksums.txt"
```

---

## 2. Script de AppImage

Cria `scripts/build_appimage.sh`:

```bash id="appimage-script"
#!/bin/bash
set -e

APP=multibrowser

mkdir -p AppDir/usr/bin
cp dist/${APP}_linux_amd64/${APP} AppDir/usr/bin/${APP}

mkdir -p AppDir/usr/share/applications
cat > AppDir/usr/share/applications/${APP}.desktop <<EOF
[Desktop Entry]
Name=MultiBrowser
Exec=${APP}
Icon=${APP}
Type=Application
Categories=Utility;
EOF

mkdir -p AppDir/usr/share/icons/hicolor/256x256/apps
cp assets/icon.png AppDir/usr/share/icons/hicolor/256x256/apps/${APP}.png

wget -q https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
chmod +x linuxdeploy-x86_64.AppImage

./linuxdeploy-x86_64.AppImage \
  --appdir AppDir \
  --output appimage

mv *.AppImage dist/
```

---

## 3. GitHub Actions (Release)

Cria `.github/workflows/release.yml`:

```yaml id="release-workflow"
name: Release

on:
  push:
    tags:
      - "v*"

jobs:
  release:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install frontend deps
        run: |
          cd frontend
          npm install

      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest

      - name: Build Wails (Linux)
        run: wails build -tags "webkit2_41"

      - name: Install GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest

      - name: Run GoReleaser (Linux + Windows)
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: goreleaser release --clean

      - name: Build AppImage
        run: |
          chmod +x scripts/build_appimage.sh
          ./scripts/build_appimage.sh

      - name: Upload AppImage
        uses: softprops/action-gh-release@v2
        with:
          files: dist/*.AppImage
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## 4. Ajuste importante (Wails)

GoReleaser não entende Wails diretamente.

Estratégia usada:

* Wails build manual (Linux GUI)
* GoReleaser → binários Go (CLI/engine)
* AppImage → empacota binário gerado

Se quiser 100% correto depois:
→ integrar Wails build dentro do pipeline custom

---

## 5. Como usar

### Criar release

```bash id="tag-release"
git tag v0.1.0
git push origin v0.1.0
```

---

### Resultado automático

GitHub Release com:

* multibrowser_linux_amd64.tar.gz
* multibrowser_windows_amd64.zip
* AppImage
* checksums.txt

---

## 6. Problemas esperados

* CGO + Windows cross-build pode falhar
  → solução: job separado Windows

* WebKitGTK não embutido no AppImage
  → pode quebrar em distros antigas

---

## 7. Evolução futura

* pipeline separado por OS
* assinatura de binário
* auto-update
* build reproducível

---

## Resumo seco

* GoReleaser cuida de build + release
* linuxdeploy gera AppImage
* GitHub Actions orquestra tudo
* tag = release automática

