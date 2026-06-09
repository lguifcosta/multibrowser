# Internal/Profile Module

## Responsabilidades
O módulo `profile` gerencia os metadados e a estrutura de diretórios dos perfis do navegador.

- **ProfileManager**: Gerencia o CRUD de perfis.
- **Diretório de Dados**: Por padrão, os perfis ficam em `~/.multibrowser/profiles/`.
- **Metadata**: Cada perfil possui um arquivo `metadata.json` em seu próprio diretório.

## Estrutura de um Perfil
Cada subdiretório em `profiles/` representa um perfil único:
- `metadata.json`: Contém informações como nome, IDs, e timestamps.
- `.lock`: (Gerenciado pelo LockManager) indica se o perfil está em uso.
- `...`: Demais diretórios de dados do Chromium.

## Convenções de Código
- Sempre use o ID (UUID) para referenciar o diretório no filesystem.
- Atualize `last_used` sempre que o perfil for iniciado.
- Certifique-se de que a remoção de um perfil também delete o diretório físico.

## Testes
- Validar criação de diretórios.
- Validar persistência e leitura do `metadata.json`.
- Testar comportamento em caso de diretórios já existentes.
