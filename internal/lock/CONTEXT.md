# Internal/Lock Module

## Responsabilidades
O módulo `lock` implementa um sistema de exclusão mútua simples para garantir que apenas um processo gerencie um perfil por vez.

- **LockManager**: Gerencia a criação e verificação do arquivo `.lock`.
- **Detecção de Lock Órfão**: Verifica se o PID gravado no `.lock` ainda é um processo ativo.

## Lógica de Funcionamento
1. **Acquire**: Escreve o PID atual no arquivo `.lock` se não houver um.
2. **Release**: Deleta o arquivo `.lock`.
3. **IsLocked**: 
   - Se o `.lock` não existe, está liberado.
   - Se o `.lock` existe, lê o PID.
   - Se o PID não está rodando no sistema, limpa o `.lock` automaticamente e libera.

## Convenções de Código
- O arquivo `.lock` deve conter apenas o PID em texto simples.
- Use `os.FindProcess` (Go) para validar se um PID é real.

## Testes
- Validar criação de arquivo `.lock`.
- Testar limpeza automática de lock órfão (PID inexistente).
- Testar falha ao tentar adquirir um lock já ativo.
