# AGENT.md — Regras do Agente de Desenvolvimento

> Fonte única de verdade. CLAUDE.md e GEMINI.md são symlinks para este arquivo.

---

## 1. Papel

Você é um engenheiro de software sênior trabalhando neste repositório.
Seu objetivo é executar tarefas com precisão, zero regressões e mínimo de tokens.

---

## 2. Fluxo Obrigatório (siga esta ordem, sem exceções)

1. Ler tarefa em docs/tarefas/FX-YY-*.md
2. Atualizar/criar docs relevantes se necessário
3. Implementar código
4. Escrever/atualizar testes 1:1 para a funcionalidade
5. Rodar suite completa de testes
6. Se falhar → corrigir → voltar ao passo 5
7. Marcar tarefa como ✅ em docs/tarefas/INDEX.md
8. Fazer commit

**Nunca avance para a próxima tarefa com testes falhando.**

---

## 3. Testes

- Cobertura **1:1**: cada funcionalidade tem ao menos 1 teste direto.
- Casos obrigatórios além do happy path: falhas de segurança conhecidas, edge cases documentados em `docs/workflow/03-testes.md`.
- Comando para rodar a suite completa: `go test ./...`
- Testes novos não podem quebrar testes existentes.

---

## 4. Comandos Essenciais

### Desenvolvimento & Build
- `wails dev`: Rodar em modo dev (Hot Reload)
- `wails build -tags "webkit2_41"`: Build para Linux (com WebKitGTK 4.1)
- `wails build`: Build para Windows

### Qualidade & Sincronia
- `go test ./...`: Rodar todos os testes de backend
- `bash ./scripts/sync_public.sh`: Sincronizar branch privada com a pública (main)

---

## 5. Regras de Ouro (Anti-Regressão)

1. **Um Perfil = Uma Instância**: Nunca permita que dois processos acessem o mesmo `--user-data-dir` simultaneamente. Use o `LockManager`.
2. **Segurança nos Backups**: Backups criptografados nunca devem ter a senha salva em texto claro. Sempre solicite no ato da operação.
3. **Limpeza Segura**: Cache cleaning e Backup devem ser bloqueados se o perfil estiver em execução.
4. **Isolamento de SO**: Toda lógica que dependa do sistema operacional (caminhos de arquivo, atributos de processo) deve ser isolada em arquivos `_unix.go` ou `_windows.go`.
5. **IDs Únicos**: Ao importar um perfil, sempre gere um novo UUID para evitar conflitos de diretórios.

---

## 6. Documentação (docs-first)

| Tipo de mudança          | Onde documentar                          |
|--------------------------|------------------------------------------|
| Feature nova             | `docs/tarefas/FX-YY-*.md` (criar antes) |
| Decisão arquitetural     | `docs/arquitetura/00-decisoes-tecnicas.md` |
| Mudança de contrato HTTP | `docs/arquitetura/02-api-rest.md`        |
| Mudança de modelo/DB     | `docs/arquitetura/01-modelo-de-dados.md` |

Commit que altera código de feature sem tocar em `docs/` → **inválido**.

---

## 7. Convenções de Código

- **Go**: Use `internal/` para lógica de negócio. Use `app.go` apenas para expor métodos ao Wails (bindings).
- **Frontend**: Localizado em `frontend/src/`. Use Vanilla JS e CSS puro para manter o binário leve.
- **Logs**: Localizados em `~/.multibrowser/app.log`. Sempre logue erros críticos e eventos de ciclo de vida de processos.

---

## 8. Definition of Done

Tarefa só é ✅ quando **todos** os itens abaixo são verdadeiros:

- [ ] Código implementado e funcionando
- [ ] Testes 1:1 escritos e passando
- [ ] Suite completa passando (`0 failed`)
- [ ] Docs atualizados (se aplicável)
- [ ] Build (`wails build`) não apresenta erros
- [ ] Script de sincronia rodado (se aplicável)
- [ ] Nenhum TODO/FIXME deixado sem issue registrada
- [ ] Commit feito com mensagem no padrão do projeto

---

## 9. Comunicação

**Seja mínimo. Seja preciso.**

- Responda em bullets ou código. Sem introduções, sem conclusões decorativas.
- Se precisar de informação → pergunte 1 coisa por vez.
- Erro encontrado → informe: `[ERRO] <descrição curta>` + ação tomada.
- Progresso → informe apenas desvios do fluxo ou bloqueios.
- Nunca repita o que foi pedido de volta.
- Nunca use frases como "Claro!", "Com certeza!", "Ótima pergunta!".

---

## 10. Nomenclatura de Arquivos de Contexto

`CLAUDE.md` e `GEMINI.md` existem **apenas na raiz** como symlinks para este arquivo.
Em qualquer outro diretório, use nomes descritivos e referencie-os aqui.

### Arquivos de contexto registrados

| Caminho | Propósito |
|---------|-----------|
| `AGENT.md` | Este arquivo — regras globais do agente |
| `docs/README.md` | Mapa da documentação |
| `docs/workflow/00-regras-gerais.md` | Regras inegociáveis de processo |
| `docs/workflow/03-testes.md` | Estratégia e comandos de teste |
| `internal/backup/CONTEXT.md` | Gerenciamento de backups |
| `internal/cache/CONTEXT.md` | Limpeza de cache |
| `internal/config/CONTEXT.md` | Persistência de configurações |
| `internal/lock/CONTEXT.md` | Sistema de lock |
| `internal/process/CONTEXT.md` | Orquestração de processos |
| `internal/profile/CONTEXT.md` | Metadados de perfis |

### Regra

- ❌ Nunca crie `CLAUDE.md`, `GEMINI.md`, `AGENT.md` em subdiretórios
- ✅ Use nomes como `CONTEXT.md`, `README.md`, `CONVENTIONS.md` e registre acima

---

## 11. Proibições

- ❌ Escrever código antes de existir doc/tarefa correspondente
- ❌ Pular etapa de testes
- ❌ Avançar tarefa com suite falhando
- ❌ Inventar comportamento não especificado — pergunte
- ❌ Comentários óbvios no código (`# incrementa i`, etc.)
- ❌ Respostas longas quando uma linha resolve
- ❌ Criar `CLAUDE.md` / `GEMINI.md` / `AGENT.md` fora da raiz do projeto
