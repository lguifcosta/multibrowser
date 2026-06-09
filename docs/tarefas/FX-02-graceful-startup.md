# FX-02: Tratamento de Erros no Startup

## Descrição
Substituir os `panic` no método `startup` do `app.go` por um tratamento de erro gracioso.

## Requisitos
- [ ] Remover todos os `panic` de `app.go`.
- [ ] Se falhar ao inicializar diretórios ou managers, logar o erro.
- [ ] Implementar uma forma de notificar o usuário sobre falha crítica (ex: Wails Dialog).

## Critérios de Aceite
- [ ] App não fecha sem aviso em caso de erro de permissão de pasta.
- [ ] Logs refletem falhas de inicialização.
