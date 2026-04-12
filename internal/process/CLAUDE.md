# Internal/Process Module

## Responsabilidades
O módulo `process` orquestra a execução de instâncias do Chromium.

- **ProcessManager**: Gerencia o ciclo de vida (spawn e kill) de processos do navegador.
- **Detecção de Chromium**: Localiza o binário do Chromium no sistema operacional.

## Lógica de Funcionamento
1. **Spawn**: Inicia o Chromium com a flag `--user-data-dir` apontando para o diretório do perfil.
2. **Lock**: Chama o `LockManager` para adquirir o lock do diretório antes do spawn.
3. **Wait**: Monitora a saída do processo e libera o lock automaticamente quando o navegador é fechado.
4. **Kill**: Envia `SIGTERM` (e `SIGKILL` como fallback) para fechar o navegador de forma controlada.

## Convenções de Código
- **Independência de OS**: Use os arquivos `sys_unix.go` e `sys_windows.go` para lógica específica de plataforma (como `Setpgid`).
- **Flags**: Sempre use `--no-first-run` e `--no-default-browser-check` para evitar interrupções na interface do Chromium.

## Testes
- Validar se o comando é construído com o `--user-data-dir` correto.
- Testar detecção de binário por plataforma.
- Validar a lógica de `waitForExit`.
