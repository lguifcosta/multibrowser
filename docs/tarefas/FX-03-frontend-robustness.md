# FX-03: Robustez no Frontend

## Descrição
Implementar captura e exibição de erros nas chamadas de API do frontend.

## Requisitos
- [x] Adicionar blocos `try/catch` ou wrapper global para chamadas ao `wailsjs/go`.
- [x] Exibir erros via `showToast` com tipo 'error'.
- [x] Garantir que o estado da UI (ex: loading spinners) seja resetado após falha.

## Critérios de Aceite
- [x] Falhas no backend (ex: erro ao deletar perfil) são visíveis para o usuário.
- [x] O console do navegador não possui erros não capturados de Promises.
