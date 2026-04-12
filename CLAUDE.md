# MultiBrowser — Guia de Desenvolvimento & Arquitetura

Este documento serve como a fonte da verdade para o desenvolvimento do MultiBrowser. Ele orquestra o comportamento esperado de todas as partes do sistema.

## 🏗️ Estrutura do Projeto

O projeto é dividido em módulos dentro de `internal/`, cada um com seu próprio arquivo `CLAUDE.md` detalhado:

- **[internal/profile/](./internal/profile/CLAUDE.md)**: Gerenciamento de metadados e diretórios de perfis.
- **[internal/lock/](./internal/lock/CLAUDE.md)**: Sistema de lock baseado em PID para evitar múltiplas instâncias.
- **[internal/process/](./internal/process/CLAUDE.md)**: Orquestração de processos Chromium (Spawn/Kill).
- **[internal/backup/](./internal/backup/CLAUDE.md)**: Exportação/Importação de perfis com criptografia AES.
- **[internal/cache/](./internal/cache/CLAUDE.md)**: Limpeza seletiva de arquivos temporários do navegador.
- **[internal/logger/](./internal/logger/)**: Registro de eventos INFO e ERROR em arquivo local.

## 🚀 Comandos Essenciais

### Desenvolvimento & Build
```bash
# Rodar em modo dev (Hot Reload)
wails dev

# Build para Linux (com WebKitGTK 4.1)
wails build -tags "webkit2_41"

# Build para Windows
wails build
```

### Testes & Qualidade
```bash
# Rodar todos os testes de backend
go test ./...

# Sincronizar branch privada com a pública (main)
bash ./scripts/sync_public.sh
```

## 📜 Regras de Ouro (Anti-Regressão)

1. **Um Perfil = Uma Instância**: Nunca permita que dois processos acessem o mesmo `--user-data-dir` simultaneamente. Use o `LockManager`.
2. **Segurança nos Backups**: Backups criptografados nunca devem ter a senha salva em texto claro. Sempre solicite no ato da operação.
3. **Limpeza Segura**: Cache cleaning e Backup devem ser bloqueados se o perfil estiver em execução.
4. **Isolamento de SO**: Toda lógica que dependa do sistema operacional (caminhos de arquivo, atributos de processo) deve ser isolada em arquivos `_unix.go` ou `_windows.go`.
5. **IDs Únicos**: Ao importar um perfil, sempre gere um novo UUID para evitar conflitos de diretórios.

## 🛠️ Convenções de Código

- **Go**: Use `internal/` para lógica de negócio. Use `app.go` apenas para expor métodos ao Wails (bindings).
- **Frontend**: Localizado em `frontend/src/`. Use Vanilla JS e CSS puro para manter o binário leve.
- **Logs**: Localizados em `~/.multibrowser/app.log`. Sempre logue erros críticos e eventos de ciclo de vida de processos.

## ✅ Definição de Pronto (DoD)
Uma tarefa só está completa quando:
- Os testes em `go test ./...` passam.
- O build (`wails build`) não apresenta erros.
- A documentação em `CLAUDE.md` (root ou módulo) foi atualizada se necessário.
- O script de sincronia foi rodado para atualizar a branch pública.
