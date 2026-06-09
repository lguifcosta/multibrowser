# FX-05: Configurações Globais e Padrões

## Descrição
Implementar um sistema de configurações globais que servem como padrão para novos perfis e perfis que não foram modificados explicitamente.

## Requisitos
- [ ] Expandir o `internal/config` para incluir flags padrão (ex: `RestoreLastSession`).
- [ ] Alterar lógica de `Launch` para priorizar flag do perfil, mas cair para a global se for nula/padrão.
- [ ] UI para editar configurações globais.

## Critérios de Aceite
- [ ] Novo perfil criado respeita a configuração global de "Restaurar última sessão".
- [ ] Alterar a configuração global afeta perfis existentes que não possuem essa configuração marcada manualmente.
