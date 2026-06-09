# Internal/Backup Module

## Responsabilidades
O módulo `backup` implementa a exportação e importação de perfis com suporte opcional a criptografia.

- **BackupService**: Gerencia a criação de arquivos `.tar.gz`.
- **Criptografia AES-GCM**: Protege os arquivos de backup com senha.

## Lógica de Funcionamento
1. **Exportar**:
   - Bloqueia o perfil se ele estiver rodando.
   - Opcionalmente apaga o cache antes de compactar.
   - Gera um arquivo `.tar.gz`.
   - Aplica criptografia se uma senha for fornecida.
2. **Importar**:
   - Descriptografa (se necessário).
   - Extrai o conteúdo para um novo diretório de perfil.
   - Gera um novo UUID para evitar colisões com perfis existentes.

## Convenções de Código
- Sempre valide senhas em backups criptografados antes da extração.
- Não permita sobrescrever perfis existentes no import; sempre gere um novo ID.
- Use `tar.gz` para máxima portabilidade.

## Testes
- Validar exportação de arquivos válidos.
- Testar descriptografia com senha correta e incorreta.
- Verificar se o import gera um novo ID único.
