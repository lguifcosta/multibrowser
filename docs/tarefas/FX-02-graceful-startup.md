# FX-02: Tratamento de Erros no Startup

## Descrição
Substituir os `panic` no método `startup` do `app.go` por um tratamento de erro gracioso.

## Requisitos
- [x] Remover todos os `panic` de `app.go`.
- [x] Se falhar ao inicializar diretórios ou managers, logar o erro.
- [x] Implementar uma forma de notificar o usuário sobre falha crítica (ex: Wails Dialog).

## Critérios de Aceite
- [x] App não fecha sem aviso em caso de erro de permissão de pasta.
- [x] Logs refletem falhas de inicialização.
