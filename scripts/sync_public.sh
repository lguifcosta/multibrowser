#!/bin/bash
# Script para sincronizar a branch 'privado' com a 'main' (pública)
# Garante que nenhum histórico ou arquivo privado vaze para o GitHub público.

set -e

# Configurações
PRIVATE_BRANCH="privado"
PUBLIC_BRANCH="main"
PRIVATE_FILES=("CLAUDE.md" "IDEA.md" "NEWSTEP.md")
REMOTE_PUBLIC="origin-publico"
REMOTE_PRIVATE="origin-privado"

echo "🔄 Iniciando sincronização: $PRIVATE_BRANCH -> $PUBLIC_BRANCH"

# 1. Verificar se há alterações não commitadas
if ! git diff-index --quiet HEAD --; then
    echo "❌ Erro: Você tem alterações não salvas. Faça commit ou stash antes de continuar."
    exit 1
fi

# 2. Garantir que estamos na branch privado e atualizados
git checkout $PRIVATE_BRANCH

# 3. Mudar para a branch pública
git checkout $PUBLIC_BRANCH

# 4. Sobrescrever o conteúdo da branch pública com o da privada (sem merge de histórico)
# O comando abaixo faz a branch pública 'parecer' com a privada no sistema de arquivos
git checkout $PRIVATE_BRANCH -- .

# 5. Remover explicitamente os arquivos privados
echo "🛡️  Removendo arquivos privados..."
for file in "${PRIVATE_FILES[@]}"; do
    if [ -f "$file" ]; then
        git rm -f "$file" > /dev/null 2>&1 || true
        echo "   - $file removido."
    fi
done

# 6. Commitar as mudanças se houver algo novo
git add .
if git diff-index --quiet HEAD --; then
    echo "✅ Nada novo para sincronizar."
else
    git commit -m "feat: sync changes from $PRIVATE_BRANCH $(date +'%Y-%m-%d %H:%M')"
    echo "✅ Sincronização concluída localmente na branch $PUBLIC_BRANCH."
fi

# 7. Voltar para a branch privado
git checkout $PRIVATE_BRANCH

echo ""
echo "🚀 Próximos passos recomendados:"
echo "1. Enviar para o Privado: git push $REMOTE_PRIVATE $PRIVATE_BRANCH"
echo "2. Enviar para o Público: git push $REMOTE_PUBLIC $PUBLIC_BRANCH"
