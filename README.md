# MultiBrowser

Gerenciador de perfis isolados de Chromium para desktop (Linux/Windows).

## Funcionalidades

- **Isolamento Total**: Cada perfil possui seu próprio diretório de dados (`--user-data-dir`).
- **Execução Simultânea**: Abra múltiplos perfis ao mesmo tempo com controle de instâncias.
- **Sistema de Lock**: Proteção contra abertura dupla do mesmo perfil usando detecção de PID.
- **Backup e Importação**: Exporte perfis em `.tar.gz` com opção de criptografia AES-GCM.
- **Limpeza de Cache**: Libere espaço em disco removendo caches desnecessários do Chromium.
- **Interface Moderna**: UI escura e intuitiva construída com Wails e Vite.

## Requisitos

- **Go** 1.22+
- **Wails v2**
- **NPM** (para o frontend)
- **Chromium** ou **Google Chrome** instalado no sistema.
- **Linux**: `webkit2gtk-4.0` ou `webkit2gtk-4.1`.

## Desenvolvimento

Para rodar em modo de desenvolvimento:

```bash
wails dev
```

## Build

Para gerar o binário de produção:

### Linux

Se o seu sistema usa `webkit2gtk-4.1` (como Ubuntu 24.04+):

```bash
wails build -tags "webkit2_41"
```

Caso contrário:

```bash
wails build
```

### Windows

```bash
wails build
```

## Testes

```bash
go test ./...
```

## Executando o AppImage

O AppImage requer **FUSE 2** para montar o pacote em tempo de execução. A maioria das distros já vem com ele, mas no Ubuntu 22.04+, Debian 12+ e derivados (incluindo ZorinOS) pode ser necessário instalar:

```bash
sudo apt install libfuse2
```

Caso prefira não instalar o FUSE, é possível rodar o AppImage extraindo-o em memória:

```bash
./MultiBrowser-linux-x86_64.AppImage --appimage-extract-and-run
```

Ou extrair permanentemente e rodar o `AppRun` diretamente:

```bash
./MultiBrowser-linux-x86_64.AppImage --appimage-extract
./squashfs-root/AppRun
```

## Estrutura do Projeto

- `app.go`: Ponto de entrada e integração dos serviços.
- `internal/profile`: Gerenciamento de metadados e diretórios de perfis.
- `internal/process`: Orquestração de processos do Chromium.
- `internal/lock`: Mecanismo de lock baseado em PID.
- `internal/backup`: Serviço de exportação/importação com criptografia.
- `internal/cache`: Limpeza seletiva de arquivos temporários.
- `frontend/`: Interface do usuário (Vite + Vanilla JS).
