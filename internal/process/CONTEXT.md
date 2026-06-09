# Internal/Process Module

## Responsabilidades
O modulo `process` orquestra a execucao de instancias de navegadores baseados em Chromium.

- **ProcessManager**: Gerencia o ciclo de vida (spawn e kill) de processos do navegador.
- **DetectBrowsers**: Localiza todos os navegadores Chromium-based instalados no sistema (Chrome, Chromium, Brave, Edge, Vivaldi, Opera).
- **Configuracao de navegador**: Permite ao usuario escolher qual navegador usar, com persistencia via `ConfigManager`.

## Navegadores Suportados
| Navegador       | Linux                               | Windows                                        |
|-----------------|--------------------------------------|-------------------------------------------------|
| Google Chrome   | google-chrome, google-chrome-stable  | Program Files\Google\Chrome\Application         |
| Chromium        | chromium, chromium-browser           | Program Files\Chromium\Application              |
| Brave           | brave-browser, brave-browser-stable  | BraveSoftware\Brave-Browser\Application         |
| Microsoft Edge  | microsoft-edge, microsoft-edge-stable| Program Files\Microsoft\Edge\Application        |
| Vivaldi         | vivaldi, vivaldi-stable              | Vivaldi\Application                             |
| Opera           | opera                                | Programs\Opera                                  |
| Opera GX        | -                                    | Programs\Opera GX                               |

## Logica de Funcionamento
1. **Inicializacao**: Verifica config persistida. Se nao houver, detecta navegadores e usa o primeiro encontrado.
2. **Spawn**: Inicia o navegador com a flag `--user-data-dir` apontando para o diretorio do perfil.
3. **Lock**: Chama o `LockManager` para adquirir o lock do diretorio antes do spawn.
4. **Wait**: Monitora a saida do processo e libera o lock automaticamente quando o navegador e fechado.
5. **Kill**: Envia `SIGTERM` (e `SIGKILL` como fallback) para fechar o navegador de forma controlada.
6. **SetBrowser**: Permite trocar o navegador em runtime com persistencia no config.

## Convencoes de Codigo
- **Independencia de OS**: Use os arquivos `sys_unix.go` e `sys_windows.go` para logica especifica de plataforma (como `Setpgid`).
- **Flags**: Sempre use `--no-first-run` e `--no-default-browser-check` para evitar interrupcoes na interface.
- **Caminho customizado**: O usuario pode informar qualquer executavel Chromium-based. Validar existencia antes de salvar.

## Testes
- Validar se o comando e construido com o `--user-data-dir` correto.
- Testar deteccao de binarios por plataforma.
- Validar a logica de `waitForExit`.
- Testar persistencia do navegador selecionado via config.
