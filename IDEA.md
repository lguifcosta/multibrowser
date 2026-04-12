# Browser Profile Manager (Desktop) — MVP Spec

## Objetivo

Aplicação desktop (Linux-first, portável para Windows) para gerenciar múltiplos perfis isolados de navegador (Chromium), com suporte a execução simultânea, backup/restauração e criptografia opcional.

---

## Stack

* Core: Go
* GUI: Wails
* Browser: Chromium (externo, detectado no sistema)

---

## Conceito Base

* Cada perfil = diretório isolado (`user-data-dir`)
* Execução = processo local (sem Docker)
* Persistência = filesystem
* Isolamento = nativo do Chromium

---

## Estrutura de Diretórios

```
/app-data/
  profiles/
    {profile_id}/
      (dados do chromium)
      .lock
  backups/
  metadata.json
  logs/
```

---

## Metadata (MVP: JSON)

```json
{
  "profiles": [
    {
      "id": "uuid",
      "name": "string",
      "created_at": "timestamp",
      "last_used": "timestamp",
      "status": "running | stopped"
    }
  ]
}
```

---

## Funcionalidades

### Perfis

* Criar perfil
* Listar perfis
* Deletar perfil
* Clonar perfil
* Renomear perfil

### Execução

* Abrir perfil (spawn Chromium)
* Fechar perfil (kill processo)
* Controle de PID
* Bloqueio de execução duplicada (lock)

### Lock (obrigatório)

* Arquivo: `/profiles/{id}/.lock`
* Conteúdo: PID do processo
* Regras:

  * Criar ao abrir
  * Remover ao fechar
  * Se PID não existir → limpar lock

---

## Execução do Browser

```
chromium \
  --user-data-dir=/app-data/profiles/{id} \
  --no-first-run \
  --no-default-browser-check
```

### Requisitos

* Detectar caminho do Chromium automaticamente
* Permitir override manual

---

## Backup

### Exportar

* Formato: `.tar.gz`
* Regras:

  * Não permitir backup com profile aberto
  * Opcional: excluir cache

```
tar -czf backups/{id}.tar.gz profiles/{id}/
```

---

### Importar

* Extrair para `/profiles/{novo_id}/`
* Atualizar metadata

---

## Criptografia (opcional)

### Método

* AES (Go `crypto/aes`)

### Regras

* Senha fornecida pelo usuário
* Não armazenar senha
* Solicitar senha ao:

  * Exportar (criptografado)
  * Importar (descriptografar)

---

## Limpeza de Cache (MVP incluído)

### Objetivo

Reduzir tamanho do profile

### Remover:

* `Cache/`
* `Code Cache/`
* `GPUCache/`
* `Crashpad/`

### Execução

* Manual (botão na UI)
* Nunca executar com browser aberto

---

## UI (Wails)

### Tela principal

* Lista de perfis
* Status (running/stopped)

### Ações

* Criar
* Abrir
* Fechar
* Clonar
* Deletar
* Backup
* Importar
* Limpar cache

---

## Logs

* Arquivo em `/logs/app.log`
* Registrar:

  * spawn de processos
  * erros
  * backup/import

---

## Regras de Negócio

* 1 perfil = 1 instância ativa
* Não permitir abrir perfil com lock ativo
* Backup só com perfil parado
* Cache cleaning só com perfil parado
* Import sempre gera novo ID

---

## Tratamento de Erros

* Lock órfão → validar PID
* Chromium não encontrado → erro + opção de configurar path
* Falha no spawn → log + feedback UI
* Backup corrompido → abortar import

---

## Sequência de Implementação

1. Estrutura de diretórios + metadata
2. CRUD de perfis
3. Execução (spawn/kill)
4. Lock funcional
5. Backup/export
6. Import
7. Criptografia
8. Cache cleaning
9. UI (Wails)

---

## Fora do Escopo (MVP)

* Sync remoto (Google Drive, etc)
* Multiusuário simultâneo
* Fingerprint spoofing
* Proxy avançado
* Automação (Playwright)

---

## Riscos

* Corrupção de perfil (backup com browser aberto)
* Lock inconsistente
* Caminho do Chromium variar por distro
* Crescimento excessivo de disco

---

## Resultado Esperado

App desktop leve que:

* gerencia perfis isolados
* executa múltiplos navegadores
* permite backup seguro
* mantém simplicidade e portabilidade

