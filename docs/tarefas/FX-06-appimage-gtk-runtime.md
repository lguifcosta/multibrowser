# FX-06: AppImage quebra em distros sem GTK (v1.2.0)

## Descrição
O AppImage da v1.2.0 inicia na máquina de build mas falha em sistemas que não
têm GTK instalado. O `linuxdeploy` empacota apenas as bibliotecas `.so` ligadas
diretamente ao binário; ele **não** rastreia os dados de runtime do GTK,
carregados via `dlopen`/lookup em tempo de execução.

## Causa raiz
O `AppRun` já exporta `GIO_MODULE_DIR` e `GSETTINGS_SCHEMA_DIR` apontando para
pastas dentro do AppDir, mas o `build_appimage.sh` **nunca copiava esses dados**
para lá. Em máquinas sem GTK do sistema (sem fallback via `XDG_DATA_DIRS`):

- **GSettings schemas ausentes** → GTK aborta em `g_settings_new(...)`
  (ex.: `org.gtk.Settings.FileChooser is not installed`). Causa principal.
- **GIO modules ausentes** → falha de TLS/HTTPS (`glib-networking`/gnutls).
- **gdk-pixbuf loaders ausentes** → ícones/imagens/SVG não carregam
  (apenas em hosts onde os loaders são externos, ex.: Ubuntu; no Arch/Manjaro
  o gdk-pixbuf 2.44 traz os loaders embutidos no `libgdk_pixbuf`).

## Requisitos
- [x] Empacotar `gschemas.compiled` em `usr/share/glib-2.0/schemas` e recompilar.
- [x] Empacotar `gio/modules` (TLS) em `usr/lib/gio/modules` + `giomodule.cache`.
- [x] Empacotar gdk-pixbuf loaders + `loaders.cache` relocável (quando externos).
- [x] `AppRun` exporta `GDK_PIXBUF_MODULE_FILE`/`GDK_PIXBUF_MODULEDIR` e regenera
      o cache do pixbuf em runtime quando há loaders empacotados.
- [x] Detecção multiarch dos diretórios de origem (Arch, Debian/Ubuntu, Fedora).

## Logo/ícone da janela ausente
O Wails só aplica o ícone da janela quando `options.App.Linux.Icon != nil`
(`window.go:137`); ele **não** usa `build/appicon.png` automaticamente. O
`main.go` nunca preenchia esse campo → janela/taskbar sem logo.

- [x] `main.go` embute `build/appicon.png` e o passa em `Linux.Icon`.
- [x] O ícone é decodificado via `GdkPixbufLoader` → depende dos loaders do
      pixbuf empacotados acima (caso contrário falha silenciosamente em distros
      com loaders externos, ex.: Ubuntu).

## WebKit helper path hardcoded (defeito real do build de release)
O build de release roda no Ubuntu. O `libwebkit2gtk-4.1` do Ubuntu **embute o
caminho absoluto** dos processos auxiliares (`/usr/lib/x86_64-linux-gnu/webkit2gtk-4.1/`)
em tempo de compilação e **ignora `WEBKIT_EXEC_PATH`** em builds de produção
(string ausente na lib; sem resolução relativa via dladdr). Em qualquer distro
não-Debian esse caminho não existe → `Unable to spawn WebKitNetworkProcess` e
abort. Os helpers empacotados nunca eram usados fora do Ubuntu.

- [x] `scripts/webkit_exec_fix.c`: shim `LD_PRELOAD` que intercepta
      `posix_spawn`/`execve`/`execv` e reescreve o prefixo hardcoded
      (`MB_WEBKIT_SYS`) para o dir empacotado (`MB_WEBKIT_DIR`).
- [x] `build_appimage.sh` compila o shim e o AppRun o injeta via `LD_PRELOAD`,
      baqueando `MB_WEBKIT_SYS` = caminho do webkit no host de build.
- [x] Validado: AppImage do release (Ubuntu) que crashava na Manjaro passa a
      iniciar com o shim ativo.

## Sintomas locais que NÃO são defeito do artefato
SIGSEGV em `gtk_init` ao rodar e "sumiço" do arquivo após executar são causados
pelo **AppImageLauncher** (binfmt) + FUSE nesta máquina (kernel 6.18). Provado:
`APPIMAGE_EXTRACT_AND_RUN=1` e o `AppRun` extraído rodam sem erros.

## Critérios de Aceite
- [x] AppImage extrai e contém `gschemas.compiled` e `gio/modules` populados.
- [x] AppImage inicia em ambiente sem GTK do sistema (libs forçadas via bundle).
- [x] Ícone da janela definido em runtime via `Linux.Icon`.
- [x] Sem regressão: continua funcionando na máquina de build.
