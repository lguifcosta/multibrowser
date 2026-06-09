# FX-01: Melhoria de Cobertura de Testes

## Descrição
Implementar testes unitários para os módulos que atualmente não possuem cobertura, garantindo a regra de 1:1 do AGENT.md.

## Módulos Alvo
- [x] `internal/config`
- [x] `internal/logger`
- [x] `internal/process`

## Critérios de Aceite
- [x] Testes passando para todos os módulos.
- [x] Suite completa (`go test ./...`) sem falhas.
- [x] Nenhum código de negócio alterado sem necessidade.
