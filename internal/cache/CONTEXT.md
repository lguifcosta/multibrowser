# Internal/Cache Module

## Responsabilidades
O módulo `cache` é responsável pela limpeza de arquivos temporários do Chromium para economizar espaço em disco.

- **CacheCleaner**: Gerencia a remoção seletiva de diretórios de cache.

## Lógica de Funcionamento
1. **Verificação de Lock**: Nunca inicia a limpeza se o perfil estiver ativo no momento.
2. **Remoção Seletiva**: Deleta recursivamente os diretórios:
   - `Cache/`
   - `Code Cache/`
   - `GPUCache/`
   - `Crashpad/`
3. **Contagem de Espaço**: Retorna o total de bytes liberados para a interface do usuário.

## Convenções de Código
- Não apague dados críticos como `Local Storage` ou `Cookies`.
- Sempre verifique a existência do diretório antes de tentar a remoção.

## Testes
- Validar se a limpeza é bloqueada para perfis ativos.
- Testar se os diretórios especificados são deletados.
- Validar o cálculo de bytes liberados.
