# MultiBrowser — Workflow & Convenções (Alinhado ao MVP Desktop)

## Workflow de Verificação Obrigatório (Anti-Regressão)

Antes de finalizar qualquer tarefa:

1. **Testes obrigatórios (1-to-1)**

   * Toda feature ou bugfix deve ter teste correspondente (Go).
   * Foco: filesystem, locks, processos, backup/restore, criptografia.

2. **Validação de Backend**

   ```bash
   go test ./...
   ```

   Garantir integridade de:

   * locks (.lock + PID)
   * criação/remoção de perfis
   * backup/restore
   * limpeza de cache

3. **Verificação de Execução Real**
   Validar manualmente ou via teste:

   * não abre perfil com `.lock` ativo
   * processo Chromium inicia corretamente
   * processo é finalizado corretamente
   * lock órfão é limpo

4. **Validação de Build Desktop**

   ```bash
   wails build
   ```

   Garantir:

   * build sem erro
   * bindings Go ↔ UI funcionando

5. **Prova de Integridade (regras críticas)**

   * 1 perfil = 1 instância
   * backup NÃO ocorre com browser aberto
   * cache cleaning NÃO ocorre com browser aberto
   * import sempre gera novo ID

6. **Documentação**

   * Atualizar este arquivo e/ou IDEA.md
   * Manter coerência com regras de negócio

---

## Comandos Úteis

```bash
# Desenvolvimento
wails dev
go mod tidy

# Testes
go test ./...

# Build
wails build
```

---

## Pilares do Projeto

* **Isolamento por Diretório**

  * Cada perfil usa `user-data-dir` exclusivo
  * Sem compartilhamento simultâneo

* **Gestão de Processos**

  * Spawn via Go (`exec.Command`)
  * Controle via PID
  * Lock obrigatório com `.lock`

* **Persistência**

  * Baseada em filesystem
  * Metadata em JSON (MVP)

* **Segurança**

  * Criptografia AES para backups
  * Nunca persistir senha
  * Sempre solicitar senha no uso

* **Performance**

  * Sem Docker
  * Execução direta do Chromium
  * Cache cleaning seletivo

---

## Convenções de Código

### Go

* Separar responsabilidades:

  * `ProfileManager`
  * `ProcessManager`
  * `BackupService`
  * `LockManager`

* Tratar TODOS erros de:

  * filesystem
  * processos
  * compressão/criptografia

* Nunca assumir estado válido:

  * sempre validar existência de diretórios
  * sempre validar PID

---

### Locks

* Arquivo: `.lock`
* Conteúdo: PID
* Regras:

  * criar ao iniciar
  * remover ao encerrar
  * se PID não existir → remover automaticamente

---

### Processos

* Sempre iniciar com:

  * `--user-data-dir`
* Nunca reutilizar diretório ativo
* Sempre registrar PID

---

### Backup

* Formato: `.tar.gz`
* Regras:

  * bloquear se profile estiver rodando
  * opção de criptografia AES
  * excluir cache opcionalmente

---

### Cache Cleaning

Remover apenas:

* `Cache/`
* `Code Cache/`
* `GPUCache/`
* `Crashpad/`

Regra:

* nunca executar com browser aberto

---

## Anti-Patterns (PROIBIDO)

* Abrir mesmo profile duas vezes
* Backup com profile em execução
* Sincronização em tempo real de profiles
* Uso de Docker no desktop
* Compartilhamento simultâneo de profile

---

## Definição de Pronto (Definition of Done)

Uma task só está concluída se:

* testes passam (`go test ./...`)
* build funciona (`wails build`)
* comportamento real validado (processo + lock + filesystem)
* nenhuma regra de negócio foi violada
* documentação atualizada

---

## Prioridade de Desenvolvimento

1. Core (filesystem + metadata)
2. Gestão de processos
3. Lock robusto
4. Backup/restore
5. Criptografia
6. Cache cleaning
7. UI

---

## Objetivo Técnico

Manter o sistema:

* determinístico (sem estados ocultos)
* resiliente a crash
* sem corrupção de dados
* simples de portar (Linux → Windows)

