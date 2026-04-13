# Internal/Config Module

## Responsabilidades
O modulo `config` persiste configuracoes do usuario em `~/.multibrowser/config.json`.

- **ConfigManager**: Leitura e escrita de configuracoes persistentes.

## Estrutura do config.json
```json
{
  "browser_path": "/usr/bin/brave-browser",
  "browser_name": "Brave"
}
```

## Logica de Funcionamento
1. **Inicializacao**: Le o arquivo `config.json` se existir. Se nao existir, inicia com config vazia.
2. **SetBrowser**: Atualiza o navegador selecionado e persiste no disco.
3. **Get**: Retorna a configuracao atual em memoria.

## Convencoes de Codigo
- Thread-safe com `sync.Mutex`.
- Nunca armazene senhas ou dados sensiveis no config.
- O arquivo e criado sob demanda (primeira escrita).

## Testes
- Validar leitura de config existente.
- Validar criacao de config quando inexistente.
- Validar persistencia apos SetBrowser.
