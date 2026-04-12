import './style.css';
import './app.css';

import {
    CreateProfile,
    ListProfiles,
    DeleteProfile,
    RenameProfile,
    CloneProfile,
    LaunchProfile,
    StopProfile,
    IsProfileRunning,
    ExportBackup,
    ImportBackup,
    CleanCache,
    GetChromiumPath
} from '../wailsjs/go/main/App';

let profiles = [];
let pollInterval = null;

function showToast(message, type = 'success') {
    const existing = document.querySelector('.toast');
    if (existing) existing.remove();

    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 3000);
}

function formatDate(dateStr) {
    if (!dateStr) return '-';
    const d = new Date(dateStr);
    return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit'});
}

function showModal(title, fields, onSubmit) {
    const overlay = document.createElement('div');
    overlay.className = 'modal-overlay';

    let fieldsHtml = fields.map(f => {
        if (f.type === 'checkbox') {
            return `<div class="checkbox-group">
                <input type="checkbox" id="modal-${f.name}" ${f.checked ? 'checked' : ''}>
                <label for="modal-${f.name}">${f.label}</label>
            </div>`;
        }
        return `<div class="form-group">
            <label for="modal-${f.name}">${f.label}</label>
            <input type="${f.type || 'text'}" id="modal-${f.name}" value="${f.value || ''}" placeholder="${f.placeholder || ''}">
        </div>`;
    }).join('');

    overlay.innerHTML = `
        <div class="modal">
            <h2>${title}</h2>
            ${fieldsHtml}
            <div class="modal-actions">
                <button class="btn btn-secondary" id="modal-cancel">Cancelar</button>
                <button class="btn btn-primary" id="modal-confirm">Confirmar</button>
            </div>
        </div>
    `;

    document.body.appendChild(overlay);

    const firstInput = overlay.querySelector('input[type="text"], input[type="password"]');
    if (firstInput) firstInput.focus();

    overlay.querySelector('#modal-cancel').onclick = () => overlay.remove();
    overlay.querySelector('#modal-confirm').onclick = () => {
        const values = {};
        fields.forEach(f => {
            const el = document.getElementById(`modal-${f.name}`);
            values[f.name] = f.type === 'checkbox' ? el.checked : el.value;
        });
        overlay.remove();
        onSubmit(values);
    };

    // Enter to confirm
    overlay.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') overlay.querySelector('#modal-confirm').click();
        if (e.key === 'Escape') overlay.remove();
    });
}

async function refreshProfiles() {
    try {
        profiles = await ListProfiles();
        render();
    } catch (err) {
        showToast('Erro ao carregar perfis: ' + err, 'error');
    }
}

async function handleCreate() {
    showModal('Novo Perfil', [
        {name: 'name', label: 'Nome do perfil', placeholder: 'Meu Perfil'}
    ], async (values) => {
        if (!values.name.trim()) return;
        try {
            await CreateProfile(values.name.trim());
            showToast('Perfil criado');
            refreshProfiles();
        } catch (err) {
            showToast('Erro: ' + err, 'error');
        }
    });
}

async function handleLaunch(id) {
    try {
        await LaunchProfile(id);
        showToast('Browser iniciado');
        refreshProfiles();
    } catch (err) {
        showToast('Erro: ' + err, 'error');
    }
}

async function handleStop(id) {
    try {
        await StopProfile(id);
        showToast('Browser encerrado');
        refreshProfiles();
    } catch (err) {
        showToast('Erro: ' + err, 'error');
    }
}

async function handleDelete(id, name) {
    if (!confirm(`Deletar perfil "${name}"? Esta acao nao pode ser desfeita.`)) return;
    try {
        await DeleteProfile(id);
        showToast('Perfil deletado');
        refreshProfiles();
    } catch (err) {
        showToast('Erro: ' + err, 'error');
    }
}

async function handleRename(id, currentName) {
    showModal('Renomear Perfil', [
        {name: 'name', label: 'Novo nome', value: currentName}
    ], async (values) => {
        if (!values.name.trim()) return;
        try {
            await RenameProfile(id, values.name.trim());
            showToast('Perfil renomeado');
            refreshProfiles();
        } catch (err) {
            showToast('Erro: ' + err, 'error');
        }
    });
}

async function handleClone(id, name) {
    showModal('Clonar Perfil', [
        {name: 'name', label: 'Nome do clone', value: name + ' (copia)'}
    ], async (values) => {
        if (!values.name.trim()) return;
        try {
            await CloneProfile(id, values.name.trim());
            showToast('Perfil clonado');
            refreshProfiles();
        } catch (err) {
            showToast('Erro: ' + err, 'error');
        }
    });
}

async function handleExport(id) {
    showModal('Exportar Backup', [
        {name: 'excludeCache', label: 'Excluir cache', type: 'checkbox', checked: true},
        {name: 'password', label: 'Senha (opcional)', type: 'password', placeholder: 'Deixe vazio para backup sem criptografia'}
    ], async (values) => {
        try {
            const path = await ExportBackup(id, values.excludeCache, values.password);
            showToast('Backup salvo em: ' + path);
        } catch (err) {
            showToast('Erro: ' + err, 'error');
        }
    });
}

async function handleImport() {
    showModal('Importar Backup', [
        {name: 'path', label: 'Caminho do arquivo', placeholder: '/caminho/para/backup.tar.gz'},
        {name: 'name', label: 'Nome do perfil', placeholder: 'Perfil Importado'},
        {name: 'password', label: 'Senha (se criptografado)', type: 'password'}
    ], async (values) => {
        if (!values.path.trim() || !values.name.trim()) return;
        try {
            await ImportBackup(values.path.trim(), values.name.trim(), values.password);
            showToast('Backup importado');
            refreshProfiles();
        } catch (err) {
            showToast('Erro: ' + err, 'error');
        }
    });
}

async function handleCleanCache(id, name) {
    if (!confirm(`Limpar cache do perfil "${name}"?`)) return;
    try {
        const freed = await CleanCache(id);
        const mb = (freed / 1024 / 1024).toFixed(2);
        showToast(`Cache limpo: ${mb} MB liberados`);
    } catch (err) {
        showToast('Erro: ' + err, 'error');
    }
}

function renderProfileCard(p) {
    const isRunning = p.status === 'running';
    const statusClass = isRunning ? 'status-running' : 'status-stopped';
    const statusText = isRunning ? 'Running' : 'Stopped';

    return `
        <div class="profile-card" data-id="${p.id}">
            <div class="profile-info">
                <span class="profile-name">${p.name}</span>
                <span class="profile-meta">
                    <span class="profile-status ${statusClass}">${statusText}</span>
                    &middot; Criado: ${formatDate(p.created_at)}
                    &middot; Usado: ${formatDate(p.last_used)}
                </span>
            </div>
            <div class="profile-actions">
                ${isRunning
                    ? `<button class="btn btn-danger btn-sm" data-action="stop" data-id="${p.id}">Parar</button>`
                    : `<button class="btn btn-success btn-sm" data-action="launch" data-id="${p.id}">Abrir</button>`
                }
                <button class="btn btn-secondary btn-sm" data-action="rename" data-id="${p.id}" data-name="${p.name}" ${isRunning ? 'disabled' : ''}>Renomear</button>
                <button class="btn btn-secondary btn-sm" data-action="clone" data-id="${p.id}" data-name="${p.name}" ${isRunning ? 'disabled' : ''}>Clonar</button>
                <button class="btn btn-secondary btn-sm" data-action="export" data-id="${p.id}" ${isRunning ? 'disabled' : ''}>Backup</button>
                <button class="btn btn-warning btn-sm" data-action="clean" data-id="${p.id}" data-name="${p.name}" ${isRunning ? 'disabled' : ''}>Cache</button>
                <button class="btn btn-danger btn-sm" data-action="delete" data-id="${p.id}" data-name="${p.name}" ${isRunning ? 'disabled' : ''}>Deletar</button>
            </div>
        </div>
    `;
}

function render() {
    const container = document.getElementById('profiles-list');

    if (!profiles || profiles.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <p>Nenhum perfil encontrado</p>
                <button class="btn btn-primary" id="empty-create">Criar Perfil</button>
            </div>
        `;
        document.getElementById('empty-create').onclick = handleCreate;
        return;
    }

    container.innerHTML = profiles.map(renderProfileCard).join('');
}

// Event delegation for profile actions
document.addEventListener('click', (e) => {
    const btn = e.target.closest('[data-action]');
    if (!btn) return;

    const action = btn.dataset.action;
    const id = btn.dataset.id;
    const name = btn.dataset.name;

    switch (action) {
        case 'launch': handleLaunch(id); break;
        case 'stop': handleStop(id); break;
        case 'delete': handleDelete(id, name); break;
        case 'rename': handleRename(id, name); break;
        case 'clone': handleClone(id, name); break;
        case 'export': handleExport(id); break;
        case 'clean': handleCleanCache(id, name); break;
    }
});

// Initialize
async function init() {
    let chromiumPath = '';
    try {
        chromiumPath = await GetChromiumPath();
    } catch (e) {
        chromiumPath = 'Chromium nao detectado';
    }

    document.querySelector('#app').innerHTML = `
        <header>
            <div>
                <h1>MultiBrowser</h1>
                <span class="chromium-path">${chromiumPath}</span>
            </div>
            <div class="toolbar">
                <button class="btn btn-primary" id="btn-create">Novo Perfil</button>
                <button class="btn btn-secondary" id="btn-import">Importar</button>
            </div>
        </header>
        <div class="profiles-container" id="profiles-list"></div>
    `;

    document.getElementById('btn-create').onclick = handleCreate;
    document.getElementById('btn-import').onclick = handleImport;

    await refreshProfiles();

    // Poll for status changes every 3 seconds
    pollInterval = setInterval(refreshProfiles, 3000);
}

init();
