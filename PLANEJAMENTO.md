# Planejamento: Compatibilidade e Distribuição Ampliada

## Contexto

O MultiBrowser atualmente distribui apenas AppImage (Linux x86_64) e `.exe` solto (Windows). O AppImage não funciona corretamente no ZorinOS, indicando problemas de portabilidade. O objetivo deste planejamento é resolver essa falha e expandir os formatos de distribuição de forma incremental.

---

## Fase 1 — Corrigir AppImage no ZorinOS (prioridade máxima)

### Diagnóstico do problema

O ZorinOS é baseado em Ubuntu LTS. Os prováveis culpados são:

**1. AppRun ausente no script local**
O `scripts/build_appimage.sh` **não gera** um AppRun customizado — ele depende do padrão do linuxdeploy. Já o workflow de CI (`release.yml`) cria um AppRun com `WEBKIT_EXEC_PATH` e `LD_LIBRARY_PATH`. Isso significa que AppImages gerados localmente e via CI têm comportamento diferente.

**2. Bundling incompleto de helpers do WebKit**
O script atual copia apenas `WebKitNetworkProcess` e `WebKitWebProcess`. Versões mais recentes do WebKitGTK também requerem:
- `WebKitGPUProcess`
- Typelibs GObject Introspection (`*.typelib` em `/usr/lib/girepository-1.0/`)
- Backend `libWPEBackend-fdo-1.0.so`
- `libwpe-1.0.so`

**3. Versão do WebKitGTK no ZorinOS**
O ZorinOS pode ter uma versão de webkit2gtk-4.1 diferente (mais antiga ou mais nova) do que a usada no ambiente de build. Bibliotecas `SONAME` incompatíveis causam falha silenciosa.

**4. FUSE não disponível**
AppImages requerem FUSE 2 ou FUSE 3. ZorinOS pode precisar de `libfuse2` instalado. O flag `APPIMAGE_EXTRACT_AND_RUN=1` contorna isso para as ferramentas de build mas não para o AppImage final.

### Tarefas

- [ ] Unificar o AppRun: mover o script AppRun customizado do CI para dentro do `build_appimage.sh`, garantindo que ambos sejam idênticos
- [ ] Ampliar bundling WebKit: incluir `WebKitGPUProcess`, typelibs GI e backends WPE no `AppDir`
- [ ] Adicionar detecção de versão do WebKit: logar qual versão está sendo empacotada para rastreabilidade
- [ ] Testar o AppImage resultante em ZorinOS (VM ou container com imagem Ubuntu 22.04/24.04 como proxy)
- [ ] Documentar requisito de FUSE no README com solução (`libfuse2` ou `--appimage-extract-and-run`)

---

## Fase 2 — Flatpak

### Motivação

O Flatpak resolve estruturalmente o problema de portabilidade do WebKit: o runtime `org.gnome.Platform` já inclui WebKitGTK, eliminando o bundling manual. Funciona em qualquer distro que tenha Flatpak (Ubuntu, Fedora, Arch, ZorinOS, etc.) e pode ser publicado no Flathub.

### Desafio principal: sandboxing vs. lançamento de browsers

O MultiBrowser lança processos externos (Chrome, Brave, etc.), o que conflita com o sandbox do Flatpak. **Estratégia em duas fases:**

1. **Fase 2a — permissivo**: usar `--socket=session-bus` + `--device=all` para validar o Flatpak funcionando sem sandbox restrito
2. **Fase 2b — correto**: migrar `internal/process/` para usar portais XDG (`xdg-desktop-portal`) para lançar aplicações externas de forma compatível com sandbox

### Tarefas

- [ ] Criar `packaging/flatpak/com.multibrowser.MultiBrowser.yml`
  - Runtime: `org.gnome.Platform` (versão compatível com webkit2gtk-4.1)
  - SDK: `org.gnome.Sdk`
  - Permissões iniciais: `--socket=x11`, `--socket=wayland`, `--socket=session-bus`, `--device=all`, `--filesystem=home`
- [ ] Testar build local com `flatpak-builder`
- [ ] Adicionar job `build-flatpak` no CI (release.yml) que gera `.flatpak` como artefato de release
- [ ] Avaliar migração para portais XDG no `internal/process/` (Fase 2b, tarefa separada)
- [ ] Publicar no Flathub (requer review externo, etapa final)

---

## Fase 3 — AUR (Arch Linux / Manjaro)

### Motivação

Arch e Manjaro têm a maior concentração de usuários power-user. Um PKGBUILD no AUR permite instalação via `yay` ou `paru` sem baixar o AppImage manualmente.

### Estratégia

O PKGBUILD baixa o AppImage do release do GitHub e o instala — sem recompilar. Isso mantém o pacote simples e vinculado aos releases existentes.

### Tarefas

- [ ] Criar `packaging/aur/PKGBUILD` que:
  - Baixa o AppImage do release do GitHub
  - Instala em `/opt/multibrowser/`
  - Cria wrapper em `/usr/bin/multibrowser`
  - Instala `.desktop` e ícone via `install -Dm644`
- [ ] Criar conta no AUR (se necessário) e publicar com `makepkg --printsrcinfo > .SRCINFO`
- [ ] Criar script de atualização do PKGBUILD a cada release (`scripts/update_aur.sh`)

---

## Fase 4 — Instalador Windows (NSIS via Wails)

### Motivação

Entregar apenas o `.exe` solto obriga o usuário a colocá-lo manualmente no PATH, sem desinstalador, sem atalho no menu iniciar. O Wails tem suporte nativo a NSIS.

### Tarefas

- [ ] Configurar `wails.json` com metadados NSIS (nome, publisher, ícone, versão)
- [ ] Usar `wails build -nsis` no workflow Windows do CI
- [ ] Incluir o `.exe` instalador (`*-installer.exe`) nos artefatos de release
- [ ] Manter o `.exe` portátil como artefato adicional (usuários que preferem sem instalador)

---

## Sequência de Execução

```
Fase 1: AppImage ZorinOS    ← começar agora, problema em produção
Fase 2: Flatpak             ← maior impacto de distribuição
Fase 3: AUR                 ← menor esforço, alto retorno para o público-alvo
Fase 4: Instalador Windows  ← melhoria UX, sem dependências nas fases anteriores
```

---

## Critério de Conclusão por Fase

| Fase | Pronto quando... |
|---|---|
| 1 — AppImage | AppImage roda no ZorinOS sem intervenção manual |
| 2 — Flatpak | `flatpak install multibrowser.flatpak` funciona e browsers são lançados |
| 3 — AUR | `yay -S multibrowser` instala e executa corretamente |
| 4 — Windows | Instalador NSIS aparece nos assets do release e instala/desinstala sem erros |
