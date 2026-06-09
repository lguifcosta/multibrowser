import './style.css';
import './app.css';

import {
    CreateProfile, ListProfiles, DeleteProfile, RenameProfile, CloneProfile,
    LaunchProfile, StopProfile,
    ExportBackup, ImportBackup, CleanCache,
    GetBrowserPath, GetBrowserName, HasBrowser, DetectBrowsers, SetBrowser, SetCustomBrowserPath,
    GetProfileFlags, UpdateProfileFlags,
    GetProfileTelemetry, GetProfileInfo,
    CreateGroup, ListGroups, DeleteGroup, RenameGroup, AssignProfileToGroup, ReorderGroups,
    GetConfig, UpdateConfig,
} from '../wailsjs/go/main/App';

let profiles = [];
let groups = [];
let appConfig = {};
let activeTab = null;
let pollInterval = null;
let draggedTabId = null;
let isDragging = false;

// ─── Utilities ───────────────────────────────────────────────────────────────

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
    return d.toLocaleDateString('pt-BR') + ' ' + d.toLocaleTimeString('pt-BR', {hour: '2-digit', minute: '2-digit'});
}

function formatLifetime(sec) {
    if (!sec || sec < 0) return '-';
    const h = Math.floor(sec / 3600);
    const m = Math.floor((sec % 3600) / 60);
    const s = sec % 60;
    if (h > 0) return `${h}h ${m}m`;
    if (m > 0) return `${m}m ${s}s`;
    return `${s}s`;
}

// ─── Generic confirm modal ────────────────────────────────────────────────────

function showModal(title, fields, onSubmit) {
    const overlay = document.createElement('div');
    overlay.className = 'modal-overlay';

    const fieldsHtml = fields.map(f => {
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

    overlay.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') overlay.querySelector('#modal-confirm').click();
        if (e.key === 'Escape') overlay.remove();
    });
}

function showAboutModal() {
    const overlay = document.createElement('div');
    overlay.className = 'modal-overlay';
    overlay.innerHTML = `
        <div class="modal modal-about">
            <h2>MultiBrowser</h2>
            <p class="about-desc">
                Gerenciador de perfis isolados para navegadores baseados em Chromium.
                Execute multiplas instancias simultaneas com configuracoes, proxies e identidades independentes.
            </p>
            <div class="about-stack">
                <span class="about-tag">Go + Wails v2</span>
                <span class="about-tag">Vanilla JS</span>
                <span class="about-tag">Linux / Windows</span>
            </div>
            <div class="modal-actions">
                <button class="btn btn-secondary" id="about-close">Fechar</button>
            </div>
        </div>
    `;
    document.body.appendChild(overlay);
    overlay.querySelector('#about-close').onclick = () => overlay.remove();
    overlay.addEventListener('keydown', e => { if (e.key === 'Escape') overlay.remove(); });
}

// ─── Profile list ─────────────────────────────────────────────────────────────

async function refreshProfiles() {
    if (isDragging) return;
    try {
        const [p, g, c] = await Promise.all([ListProfiles(), ListGroups(), GetConfig()]);
        profiles = p || [];
        groups = g || [];
        appConfig = c || {};

        if (activeTab === null) {
            activeTab = appConfig.default_tab_id || 'all';
        }

        render();
    } catch (err) {
        showToast('Erro ao carregar dados: ' + err, 'error');
    }
}

// ─── Group actions ────────────────────────────────────────────────────────────

async function handleSetDefaultTab(id) {
    try {
        const newConfig = { ...appConfig, default_tab_id: id };
        await UpdateConfig(newConfig);
        appConfig = newConfig;
        showToast('Aba padrao configurada');
        render();
    } catch (err) {
        showToast('Erro: ' + err, 'error');
    }
}

async function handleReorderGroups(ids) {
    try {
        await ReorderGroups(ids);
        groups = groups.sort((a, b) => ids.indexOf(a.id) - ids.indexOf(b.id));
        render();
    } catch (err) {
        showToast('Erro ao reordenar: ' + err, 'error');
    }
}

// ─── Group actions ────────────────────────────────────────────────────────────

async function handleCreateGroup() {
    showModal('Nova Aba (Grupo)', [
        {name: 'name', label: 'Nome da aba', placeholder: 'Trabalho, Pessoal, etc.'}
    ], async (values) => {
        if (!values.name.trim()) return;
        try {
            await CreateGroup(values.name.trim());
            showToast('Grupo criado');
            refreshProfiles();
        } catch (err) {
            showToast('Erro: ' + err, 'error');
        }
    });
}

async function handleRenameGroup(id, currentName) {
    showModal('Renomear Aba', [
        {name: 'name', label: 'Novo nome', value: currentName}
    ], async (values) => {
        if (!values.name.trim()) return;
        try {
            await RenameGroup(id, values.name.trim());
            showToast('Grupo renomeado');
            refreshProfiles();
        } catch (err) {
            showToast('Erro: ' + err, 'error');
        }
    });
}

async function handleDeleteGroup(id, name) {
    if (!confirm(`Deletar aba "${name}"? Os perfis voltarao para "Sem Grupo".`)) return;
    try {
        await DeleteGroup(id);
        if (activeTab === id) activeTab = 'all';
        showToast('Grupo deletado');
        refreshProfiles();
    } catch (err) {
        showToast('Erro: ' + err, 'error');
    }
}

async function handleAssignGroup(profileId, groupId) {
    try {
        await AssignProfileToGroup(profileId, groupId);
        showToast('Perfil movido');
        refreshProfiles();
    } catch (err) {
        showToast('Erro: ' + err, 'error');
    }
}

async function handleBatchAssign(groupId, profileIds, removeIds) {
    try {
        const promises = [
            ...profileIds.map(id => AssignProfileToGroup(id, groupId)),
            ...removeIds.map(id => AssignProfileToGroup(id, ""))
        ];
        await Promise.all(promises);
        showToast(`${profileIds.length + removeIds.length} perfis atualizados`);
        refreshProfiles();
    } catch (err) {
        showToast('Erro no processamento em lote: ' + err, 'error');
    }
}

// ─── Profile actions ──────────────────────────────────────────────────────────

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

// ─── Per-profile settings modal ───────────────────────────────────────────────

async function handleProfileSettings(profileId) {
    const p = profiles.find(x => x.id === profileId);
    if (!p) return;
    let flags = {};
    try { flags = await GetProfileFlags(profileId); } catch (e) {}
    openProfileSettingsModal(p, flags || {});
}

function openProfileSettingsModal(p, flags) {
    const isRunning = p.status === 'running';
    const overlay   = document.createElement('div');
    overlay.className = 'modal-overlay';

    const esc = v => String(v ?? '').replace(/"/g, '&quot;');
    const chk = v => v ? 'checked' : '';
    const dis = isRunning ? 'disabled' : '';

    overlay.innerHTML = `
        <div class="modal modal-profile-settings">
            <div class="ps-header">
                <div class="ps-header-info">
                    <h2>${esc(p.name)}</h2>
                    <span class="ps-subtitle">Configuracoes da sessao</span>
                </div>
                <button class="ps-close" id="ps-close">&#x2715;</button>
            </div>

            <div class="ps-tabs">
                <button class="ps-tab active" data-tab="flags">Configuracoes</button>
                <button class="ps-tab" data-tab="actions">Acoes</button>
                <button class="ps-tab" data-tab="info">Informacoes</button>
            </div>

            <div class="ps-body">

                <!-- Configuracoes -->
                <div class="ps-panel" id="ps-panel-flags">
                    <div class="ps-section">
                        <div class="ps-section-title">Sessao</div>
                        <label class="ps-toggle">
                            <input type="checkbox" id="f-restore" ${chk(flags.restore_last_session)}>
                            <span class="ps-toggle-label">Restaurar ultima sessao</span>
                            <code class="ps-flag">--restore-last-session</code>
                        </label>
                    </div>

                    <div class="ps-section">
                        <div class="ps-section-title">Fingerprint / Anti-detect</div>
                        <div class="ps-field">
                            <label>User Agent</label>
                            <input type="text" id="f-ua" value="${esc(flags.user_agent)}" placeholder="Mozilla/5.0 (X11; Linux x86_64) ...">
                            <code class="ps-flag">--user-agent</code>
                        </div>
                        <div class="ps-field">
                            <label>Idioma</label>
                            <input type="text" id="f-lang" value="${esc(flags.lang)}" placeholder="pt-BR">
                            <code class="ps-flag">--lang</code>
                        </div>
                        <div class="ps-field">
                            <label>Tamanho da janela</label>
                            <input type="text" id="f-winsize" value="${esc(flags.window_size)}" placeholder="1280,800">
                            <code class="ps-flag">--window-size</code>
                        </div>
                    </div>

                    <div class="ps-section">
                        <div class="ps-section-title">Proxy por instancia</div>
                        <div class="ps-field">
                            <label>Servidor proxy</label>
                            <input type="text" id="f-proxy" value="${esc(flags.proxy_server)}" placeholder="http://ip:porta">
                            <code class="ps-flag">--proxy-server</code>
                        </div>
                        <div class="ps-field">
                            <label>Bypass do proxy</label>
                            <input type="text" id="f-bypass" value="${esc(flags.proxy_bypass_list)}" placeholder="&lt;-loopback&gt;">
                            <code class="ps-flag">--proxy-bypass-list</code>
                        </div>
                    </div>

                    <div class="ps-section">
                        <div class="ps-section-title">Organização</div>
                        <div class="ps-field">
                            <label>Aba / Grupo</label>
                            <div class="ps-select-wrapper">
                                <select id="f-group" class="ps-select">
                                    <option value="">Sem Grupo</option>
                                    ${groups.map(g => `<option value="${g.id}" ${p.group_id === g.id ? 'selected' : ''}>${g.name}</option>`).join('')}
                                </select>
                            </div>
                        </div>
                    </div>

                    <div class="ps-section">
                        <div class="ps-section-title">Performance</div>
                        <label class="ps-toggle">
                            <input type="checkbox" id="f-bgnet" ${chk(flags.disable_background_networking)}>
                            <span class="ps-toggle-label">Desativar rede em background</span>
                            <code class="ps-flag">--disable-background-networking</code>
                        </label>
                        <label class="ps-toggle">
                            <input type="checkbox" id="f-bgtimer" ${chk(flags.disable_background_timer_throttling)}>
                            <span class="ps-toggle-label">Desativar throttling de timers</span>
                            <code class="ps-flag">--disable-background-timer-throttling</code>
                        </label>
                        <label class="ps-toggle">
                            <input type="checkbox" id="f-renderbg" ${chk(flags.disable_renderer_backgrounding)}>
                            <span class="ps-toggle-label">Desativar backgrounding do renderer</span>
                            <code class="ps-flag">--disable-renderer-backgrounding</code>
                        </label>
                        <label class="ps-toggle">
                            <input type="checkbox" id="f-translate" ${chk(flags.disable_features_translate_ui)}>
                            <span class="ps-toggle-label">Desativar UI de traducao</span>
                            <code class="ps-flag">--disable-features=TranslateUI</code>
                        </label>
                        <label class="ps-toggle">
                            <input type="checkbox" id="f-ext" ${chk(flags.disable_extensions)}>
                            <span class="ps-toggle-label">Desativar extensoes</span>
                            <code class="ps-flag">--disable-extensions</code>
                        </label>
                        <label class="ps-toggle">
                            <input type="checkbox" id="f-sync" ${chk(flags.disable_sync)}>
                            <span class="ps-toggle-label">Desativar sincronizacao</span>
                            <code class="ps-flag">--disable-sync</code>
                        </label>
                    </div>

                    <div class="ps-actions">
                        <button class="btn btn-secondary" id="ps-cancel">Cancelar</button>
                        <button class="btn btn-primary" id="ps-save">Salvar</button>
                    </div>
                </div>

                <!-- Acoes -->
                <div class="ps-panel hidden" id="ps-panel-actions">
                    ${isRunning ? `<div class="ps-running-notice">Encerre o perfil para realizar acoes de gerenciamento.</div>` : ''}

                    <div class="ps-section">
                        <div class="ps-section-title">Gerenciar</div>
                        <div class="ps-action-list">
                            <button class="ps-action-btn" id="pa-rename" ${dis}>
                                <span class="ps-action-name">Renomear</span>
                                <span class="ps-action-desc">Alterar o nome de exibicao do perfil</span>
                            </button>
                            <button class="ps-action-btn" id="pa-clone" ${dis}>
                                <span class="ps-action-name">Clonar perfil</span>
                                <span class="ps-action-desc">Criar uma copia independente com todos os dados</span>
                            </button>
                        </div>
                    </div>

                    <div class="ps-section">
                        <div class="ps-section-title">Dados</div>
                        <div class="ps-action-list">
                            <button class="ps-action-btn" id="pa-backup" ${dis}>
                                <span class="ps-action-name">Exportar backup</span>
                                <span class="ps-action-desc">Salvar perfil como arquivo comprimido (com criptografia opcional)</span>
                            </button>
                            <button class="ps-action-btn ps-action-warning" id="pa-cache" ${dis}>
                                <span class="ps-action-name">Limpar cache</span>
                                <span class="ps-action-desc">Liberar espaco em disco sem perder dados de sessao</span>
                            </button>
                        </div>
                    </div>

                    <div class="ps-section ps-danger-zone">
                        <div class="ps-section-title">Zona de perigo</div>
                        <button class="ps-action-btn ps-action-danger" id="pa-delete" ${dis}>
                            <span class="ps-action-name">Deletar perfil</span>
                            <span class="ps-action-desc">Remove permanentemente todos os dados deste perfil</span>
                        </button>
                    </div>
                </div>

                <!-- Informacoes -->
                <div class="ps-panel hidden" id="ps-panel-info">
                    <div id="ps-info-content" class="ps-info-loading">
                        <span class="ps-spinner"></span> Carregando...
                    </div>
                </div>

            </div>
        </div>
    `;

    document.body.appendChild(overlay);

    let telemInterval = null;

    function close() {
        clearInterval(telemInterval);
        overlay.remove();
    }

    function switchTab(name) {
        overlay.querySelectorAll('.ps-tab').forEach(t => t.classList.remove('active'));
        overlay.querySelectorAll('.ps-panel').forEach(t => t.classList.add('hidden'));
        overlay.querySelector(`.ps-tab[data-tab="${name}"]`).classList.add('active');
        overlay.querySelector(`#ps-panel-${name}`).classList.remove('hidden');

        clearInterval(telemInterval);
        if (name === 'info') {
            loadInfoTab();
            telemInterval = setInterval(loadInfoTab, 3000);
        }
    }

    async function loadInfoTab() {
        const container = overlay.querySelector('#ps-info-content');
        try {
            const [telem, info] = await Promise.all([
                GetProfileTelemetry(p.id),
                GetProfileInfo(p.id),
            ]);
            const running = telem && telem.is_running;
            const cpu   = running ? telem.cpu_percent.toFixed(1) + '%' : '-';
            const ram   = running ? telem.ram_mb.toFixed(0) + ' MB'    : '-';
            const life  = running ? formatLifetime(telem.lifetime_sec)  : '-';
            const pid   = running ? telem.pid                           : '-';
            const disk  = info    ? info.disk_usage_mb.toFixed(1) + ' MB' : '-';
            const dir   = info    ? info.profile_dir : '-';
            const crashes = info  ? info.crash_count : 0;

            container.className = '';
            container.innerHTML = `
                <div class="ps-section">
                    <div class="ps-section-title">Perfil</div>
                    <div class="ps-info-row"><span class="ps-info-key">Diretorio</span><span class="ps-info-val ps-mono">${dir}</span></div>
                    <div class="ps-info-row"><span class="ps-info-key">Criado em</span><span class="ps-info-val">${formatDate(p.created_at)}</span></div>
                    <div class="ps-info-row"><span class="ps-info-key">Ultimo uso</span><span class="ps-info-val">${formatDate(p.last_used)}</span></div>
                    <div class="ps-info-row"><span class="ps-info-key">Uso em disco</span><span class="ps-info-val">${disk}</span></div>
                    <div class="ps-info-row"><span class="ps-info-key">Crash reports</span><span class="ps-info-val ${crashes > 0 ? 'ps-warn' : ''}">${crashes}</span></div>
                </div>
                <div class="ps-section">
                    <div class="ps-section-title">
                        Telemetria
                        ${running ? '<span class="ps-badge-running">ao vivo</span>' : ''}
                    </div>
                    <div class="ps-info-row"><span class="ps-info-key">PID</span><span class="ps-info-val ps-mono">${pid}</span></div>
                    <div class="ps-info-row"><span class="ps-info-key">CPU</span><span class="ps-info-val">${cpu}</span></div>
                    <div class="ps-info-row"><span class="ps-info-key">RAM</span><span class="ps-info-val">${ram}</span></div>
                    <div class="ps-info-row"><span class="ps-info-key">Tempo de vida</span><span class="ps-info-val">${life}</span></div>
                </div>
            `;
        } catch (err) {
            container.className = 'ps-info-error';
            container.textContent = 'Erro ao carregar: ' + err;
        }
    }

    // Tab navigation
    overlay.querySelectorAll('.ps-tab').forEach(btn => {
        btn.onclick = () => switchTab(btn.dataset.tab);
    });

    // Save flags
    overlay.querySelector('#ps-save').onclick = async () => {
        const newFlags = {
            restore_last_session:                overlay.querySelector('#f-restore').checked,
            user_agent:                          overlay.querySelector('#f-ua').value.trim(),
            lang:                                overlay.querySelector('#f-lang').value.trim(),
            window_size:                         overlay.querySelector('#f-winsize').value.trim(),
            proxy_server:                        overlay.querySelector('#f-proxy').value.trim(),
            proxy_bypass_list:                   overlay.querySelector('#f-bypass').value.trim(),
            disable_background_networking:       overlay.querySelector('#f-bgnet').checked,
            disable_background_timer_throttling: overlay.querySelector('#f-bgtimer').checked,
            disable_renderer_backgrounding:      overlay.querySelector('#f-renderbg').checked,
            disable_features_translate_ui:       overlay.querySelector('#f-translate').checked,
            disable_extensions:                  overlay.querySelector('#f-ext').checked,
            disable_sync:                        overlay.querySelector('#f-sync').checked,
        };
        const newGroupId = overlay.querySelector('#f-group').value;
        try {
            await Promise.all([
                UpdateProfileFlags(p.id, newFlags),
                AssignProfileToGroup(p.id, newGroupId)
            ]);
            showToast('Configuracoes salvas');
            close();
            refreshProfiles();
        } catch (err) {
            showToast('Erro: ' + err, 'error');
        }
    };

    // Actions tab handlers — close modal first, then open the action
    overlay.querySelector('#pa-rename').onclick = () => { close(); handleRename(p.id, p.name); };
    overlay.querySelector('#pa-clone').onclick  = () => { close(); handleClone(p.id, p.name); };
    overlay.querySelector('#pa-backup').onclick = () => { close(); handleExport(p.id); };
    overlay.querySelector('#pa-cache').onclick  = () => { close(); handleCleanCache(p.id, p.name); };
    overlay.querySelector('#pa-delete').onclick = () => { close(); handleDelete(p.id, p.name); };

    overlay.querySelector('#ps-cancel').onclick = close;
    overlay.querySelector('#ps-close').onclick  = close;
    overlay.addEventListener('keydown', e => { if (e.key === 'Escape') close(); });
}

// ─── Browser settings modal ───────────────────────────────────────────────────

async function handleSettings() {
    let browsers = [];
    try { browsers = await DetectBrowsers(); } catch (e) {}
    const currentPath = await GetBrowserPath();
    const config = await GetConfig();

    const overlay = document.createElement('div');
    overlay.className = 'modal-overlay';

    const browserListHtml = (browsers && browsers.length > 0)
        ? `<div class="form-group">
               <label>Navegadores detectados</label>
               <div class="browser-list">
                   ${browsers.map(b => `
                       <button class="browser-option ${b.path === currentPath ? 'browser-selected' : ''}"
                               data-name="${b.name}" data-path="${b.path}">
                           <span class="browser-option-name">${b.name}</span>
                           <span class="browser-option-path">${b.path}</span>
                       </button>
                   `).join('')}
               </div>
           </div>`
        : `<div class="settings-notice">Nenhum navegador Chromium detectado automaticamente. Informe o caminho abaixo.</div>`;

    overlay.innerHTML = `
        <div class="modal modal-settings">
            <h2>Configuracoes Globais</h2>
            
            <div class="ps-section">
                <div class="ps-section-title">Interface</div>
                <label class="ps-toggle">
                    <input type="checkbox" id="set-show-unassigned" ${config.show_unassigned_tab ? 'checked' : ''}>
                    <span class="ps-toggle-label">Mostrar aba "Sem Grupo"</span>
                </label>
            </div>

            <div class="settings-divider"></div>

            <div class="ps-section">
                <div class="ps-section-title">Navegador Padrao</div>
                ${browserListHtml}
                <div class="form-group" style="margin-top: 12px;">
                    <label for="custom-path">Caminho customizado</label>
                    <input type="text" id="custom-path" value="" placeholder="/usr/bin/meu-navegador">
                </div>
                <button class="btn btn-secondary btn-block" id="btn-apply-custom">Usar caminho customizado</button>
            </div>

            <div class="modal-actions">
                <button class="btn btn-secondary" id="modal-cancel">Fechar</button>
                <button class="btn btn-primary" id="modal-save-config">Salvar Configuracoes</button>
            </div>
        </div>
    `;

    document.body.appendChild(overlay);

    overlay.querySelectorAll('.browser-option').forEach(btn => {
        btn.onclick = async () => {
            try {
                await SetBrowser(btn.dataset.name, btn.dataset.path);
                showToast(`Navegador alterado para ${btn.dataset.name}`);
                overlay.querySelectorAll('.browser-option').forEach(b => b.classList.remove('browser-selected'));
                btn.classList.add('browser-selected');
                updateHeader();
            } catch (err) {
                showToast('Erro: ' + err, 'error');
            }
        };
    });

    overlay.querySelector('#btn-apply-custom').onclick = async () => {
        const path = document.getElementById('custom-path').value.trim();
        if (!path) return;
        try {
            await SetCustomBrowserPath(path);
            showToast('Navegador customizado configurado');
            updateHeader();
        } catch (err) {
            showToast('Erro: ' + err, 'error');
        }
    };

    overlay.querySelector('#modal-save-config').onclick = async () => {
        const newConfig = {
            ...config,
            show_unassigned_tab: document.getElementById('set-show-unassigned').checked,
        };
        try {
            await UpdateConfig(newConfig);
            showToast('Configuracoes salvas');
            overlay.remove();
            refreshProfiles();
        } catch (err) {
            showToast('Erro: ' + err, 'error');
        }
    };

    overlay.querySelector('#modal-cancel').onclick = () => overlay.remove();
    overlay.addEventListener('keydown', e => { if (e.key === 'Escape') overlay.remove(); });
}

// ─── Header ───────────────────────────────────────────────────────────────────

async function updateHeader() {
    let browserName = '', browserPath = '', hasBrowser = false;
    try {
        browserName = await GetBrowserName();
        browserPath = await GetBrowserPath();
        hasBrowser  = await HasBrowser();
    } catch (e) {}

    const el = document.querySelector('.browser-info');
    if (el) {
        el.innerHTML = hasBrowser
            ? `<span class="browser-name">${browserName}</span><span class="chromium-path">${browserPath}</span>`
            : `<span class="browser-warning">Nenhum navegador configurado</span>`;
    }
}

// ─── Hamburger menu ───────────────────────────────────────────────────────────

function setupHamburger() {
    const btn  = document.getElementById('hamburger-btn');
    const menu = document.getElementById('hamburger-menu');

    btn.onclick = e => { e.stopPropagation(); menu.classList.toggle('open'); };
    document.addEventListener('click', () => menu.classList.remove('open'));
    menu.addEventListener('click', e => e.stopPropagation());

    document.getElementById('hmenu-browser').onclick = () => { menu.classList.remove('open'); handleSettings(); };
    document.getElementById('hmenu-import').onclick  = () => { menu.classList.remove('open'); handleImport(); };
    document.getElementById('hmenu-about').onclick   = () => { menu.classList.remove('open'); showAboutModal(); };
}

// ─── Render ───────────────────────────────────────────────────────────────────

function renderProfileCard(p) {
    const isRunning   = p.status === 'running';
    const statusClass = isRunning ? 'status-running' : 'status-stopped';
    const statusText  = isRunning ? 'Running' : 'Stopped';

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
                    ? `<button class="btn btn-danger  btn-sm" data-action="stop"   data-id="${p.id}">Parar</button>`
                    : `<button class="btn btn-success btn-sm" data-action="launch" data-id="${p.id}">Abrir</button>`
                }
                <button class="btn btn-secondary btn-sm btn-config" data-action="config" data-id="${p.id}">Config</button>
            </div>
        </div>
    `;
}

function renderTabs() {
    const container = document.getElementById('tabs-list');
    if (!container) return '';

    const isDefault = id => appConfig.default_tab_id === id;

    let html = `
        <div class="tab-item ${activeTab === 'all' ? 'active' : ''} ${isDefault('all') ? 'tab-default' : ''}" 
             data-tab="all">
            Todos
            <button class="tab-options-btn" data-group-action="options" data-id="all" data-name="Todos">⋮</button>
        </div>
    `;

    if (appConfig.show_unassigned_tab) {
        html += `
            <div class="tab-item ${activeTab === 'unassigned' ? 'active' : ''} ${isDefault('unassigned') ? 'tab-default' : ''}" 
                 data-tab="unassigned">
                Sem Grupo
                <button class="tab-options-btn" data-group-action="options" data-id="unassigned" data-name="Sem Grupo">⋮</button>
            </div>
        `;
    }

    groups.forEach(g => {
        html += `
            <div class="tab-item ${activeTab === g.id ? 'active' : ''} ${isDefault(g.id) ? 'tab-default' : ''}" 
                 data-tab="${g.id}" draggable="true">
                ${g.name}
                <button class="tab-options-btn" data-group-action="options" data-id="${g.id}" data-name="${g.name}">⋮</button>
            </div>
        `;
    });

    html += `<button class="tab-add" id="btn-add-group" title="Criar nova aba">+</button>`;
    
    container.innerHTML = html;

    // Events attachment
    container.querySelectorAll('.tab-item').forEach(el => {
        const tabId = el.dataset.tab;

        el.onclick = (e) => {
            if (e.target.closest('[data-group-action]')) return;
            activeTab = tabId;
            render();
        };

        // Drag and drop logic
        el.ondragstart = (e) => {
            if (!el.getAttribute('draggable')) {
                e.preventDefault();
                return;
            }
            draggedTabId = tabId;
            isDragging = true;
            el.classList.add('tab-dragging');
            e.dataTransfer.effectAllowed = 'move';
            e.dataTransfer.setData('text/plain', tabId); // Required by some browsers
        };

        el.ondragend = () => {
            el.classList.remove('tab-dragging');
            container.querySelectorAll('.tab-item').forEach(t => t.classList.remove('tab-drag-over'));
            isDragging = false;
            draggedTabId = null;
        };

        el.ondragover = (e) => {
            e.preventDefault();
            const isCustom = groups.some(g => g.id === tabId);
            if (isDragging && draggedTabId !== tabId && isCustom) {
                el.classList.add('tab-drag-over');
            }
            e.dataTransfer.dropEffect = 'move';
        };

        el.ondragleave = () => {
            el.classList.remove('tab-drag-over');
        };

        el.ondrop = (e) => {
            e.preventDefault();
            const targetId = tabId;
            const isCustomTarget = groups.some(g => g.id === targetId);
            
            if (draggedTabId && draggedTabId !== targetId && isCustomTarget) {
                const ids = groups.map(g => g.id);
                const fromIdx = ids.indexOf(draggedTabId);
                const toIdx = ids.indexOf(targetId);
                if (fromIdx !== -1 && toIdx !== -1) {
                    ids.splice(toIdx, 0, ids.splice(fromIdx, 1)[0]);
                    handleReorderGroups(ids);
                }
            }
        };
    });

    const addBtn = container.querySelector('#btn-add-group');
    if (addBtn) addBtn.onclick = handleCreateGroup;

    container.querySelectorAll('[data-group-action="options"]').forEach(btn => {
        btn.onclick = (e) => {
            e.stopPropagation();
            const { id, name } = btn.dataset;
            showTabMenu(e, id, name);
        };
    });
}

function showTabMenu(event, id, name) {
    const existing = document.querySelector('.context-menu');
    if (existing) existing.remove();

    const menu = document.createElement('div');
    menu.className = 'context-menu';
    
    const isSpecial = id === 'all' || id === 'unassigned';
    const isDefault = appConfig.default_tab_id === id;

    menu.innerHTML = `
        ${!isSpecial ? `<button class="cmenu-item" id="cmenu-members">Gerenciar Membros</button>` : ''}
        ${!isSpecial ? `<button class="cmenu-item" id="cmenu-rename">Renomear</button>` : ''}
        <button class="cmenu-item ${isDefault ? 'disabled' : ''}" id="cmenu-default">
            ${isDefault ? '✓ Aba Padrao' : 'Tornar Padrao'}
        </button>
        ${!isSpecial ? `<div class="cmenu-divider"></div>` : ''}
        ${!isSpecial ? `<button class="cmenu-item cmenu-danger" id="cmenu-delete">Deletar</button>` : ''}
    `;

    document.body.appendChild(menu);
    
    // Position menu
    const rect = event.target.getBoundingClientRect();
    menu.style.top = `${rect.bottom + 5}px`;
    menu.style.left = `${rect.left}px`;

    // Click outside to close
    const close = () => { menu.remove(); document.removeEventListener('click', close); };
    setTimeout(() => document.addEventListener('click', close), 10);

    // Handlers
    if (!isSpecial) {
        menu.querySelector('#cmenu-members').onclick = () => showGroupMembersModal(id, name);
        menu.querySelector('#cmenu-rename').onclick = () => handleRenameGroup(id, name);
        menu.querySelector('#cmenu-delete').onclick = () => handleDeleteGroup(id, name);
    }
    if (!isDefault) {
        menu.querySelector('#cmenu-default').onclick = () => handleSetDefaultTab(id);
    }
}

function showGroupMembersModal(groupId, groupName) {
    const overlay = document.createElement('div');
    overlay.className = 'modal-overlay';
    
    overlay.innerHTML = `
        <div class="modal modal-members">
            <div class="batch-header">
                <h2>Membros da Aba: ${groupName}</h2>
                <p class="ps-subtitle">Selecione os perfis que deseja incluir ou remover desta aba.</p>
            </div>

            <div class="batch-search-wrap">
                <div class="batch-search-container">
                    <span class="batch-search-icon">🔍</span>
                    <input type="text" class="batch-search" id="batch-search-input" placeholder="Buscar perfis...">
                </div>
            </div>
            
            <div class="batch-list" id="batch-profile-list">
                <!-- Perfis renderizados dinamicamente -->
            </div>

            <div class="batch-footer">
                <div class="batch-stats">
                    <span id="batch-selected-count">0</span> perfis selecionados
                </div>
                <div class="modal-actions">
                    <button class="btn btn-secondary" id="batch-cancel">Cancelar</button>
                    <button class="btn btn-primary" id="batch-save">Salvar</button>
                </div>
            </div>
        </div>
    `;

    document.body.appendChild(overlay);

    const listContainer = overlay.querySelector('#batch-profile-list');
    const searchInput = overlay.querySelector('#batch-search-input');
    const statsCount = overlay.querySelector('#batch-selected-count');

    function updateStats() {
        const count = overlay.querySelectorAll('.batch-checkbox:checked').length;
        statsCount.textContent = count;
    }

    function renderBatchList(filter = '') {
        const filtered = profiles.filter(p => p.name.toLowerCase().includes(filter.toLowerCase()));
        
        listContainer.innerHTML = filtered.map(p => {
            const isMember = p.group_id === groupId;
            const otherGroupName = p.group_id && !isMember ? (groups.find(g => g.id === p.group_id)?.name || 'Outro') : null;
            let metaText = 'Sem grupo';
            if (isMember) metaText = 'Já neste grupo';
            else if (otherGroupName) metaText = `Atualmente em: ${otherGroupName}`;

            return `
                <div class="batch-item ${isMember ? 'selected' : ''}" data-profile-id="${p.id}">
                    <div class="batch-checkbox-wrap">
                        <input type="checkbox" class="batch-checkbox" data-id="${p.id}" ${isMember ? 'checked' : ''}>
                        <div class="batch-custom-check"></div>
                    </div>
                    <div class="batch-icon">👤</div>
                    <div class="batch-info">
                        <span class="batch-name">${p.name}</span>
                        <span class="batch-meta">${metaText}</span>
                    </div>
                </div>
            `;
        }).join('');

        // Re-attach item click events
        listContainer.querySelectorAll('.batch-item').forEach(item => {
            item.onclick = (e) => {
                if (e.target.classList.contains('batch-checkbox')) return;
                const cb = item.querySelector('.batch-checkbox');
                cb.checked = !cb.checked;
                item.classList.toggle('selected', cb.checked);
                updateStats();
            };
            item.querySelector('.batch-checkbox').onclick = (e) => {
                e.stopPropagation();
                item.classList.toggle('selected', e.target.checked);
                updateStats();
            };
        });
    }

    renderBatchList();
    updateStats();

    searchInput.oninput = (e) => renderBatchList(e.target.value);

    overlay.querySelector('#batch-cancel').onclick = () => overlay.remove();
    overlay.querySelector('#batch-save').onclick = async () => {
        const selectedIds = [];
        const removedIds = [];
        
        // Verificamos todos os perfis, não apenas os visíveis no filtro
        // Para isso, precisamos olhar o estado atualizado no DOM ou manter um estado local
        // Usaremos o DOM mas iterando sobre todos os perfis
        overlay.querySelectorAll('.batch-checkbox').forEach(cb => {
            const id = cb.dataset.id;
            const profile = profiles.find(p => p.id === id);
            if (cb.checked && profile.group_id !== groupId) {
                selectedIds.push(id);
            } else if (!cb.checked && profile.group_id === groupId) {
                removedIds.push(id);
            }
        });

        overlay.remove();
        if (selectedIds.length > 0 || removedIds.length > 0) {
            await handleBatchAssign(groupId, selectedIds, removedIds);
        }
    };
}

function render() {
    renderTabs();

    const container = document.getElementById('profiles-list');

    let filteredProfiles = profiles;
    if (activeTab === 'unassigned') {
        filteredProfiles = profiles.filter(p => !p.group_id);
    } else if (activeTab !== 'all') {
        filteredProfiles = profiles.filter(p => p.group_id === activeTab);
    }

    if (!filteredProfiles || filteredProfiles.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <p>Nenhum perfil nesta categoria</p>
                ${activeTab === 'all' ? `<button class="btn btn-primary" id="empty-create">Criar Perfil</button>` : ''}
            </div>
        `;
        const btn = document.getElementById('empty-create');
        if (btn) btn.onclick = handleCreate;
        return;
    }

    container.innerHTML = filteredProfiles.map(renderProfileCard).join('');
}

// ─── Event delegation ─────────────────────────────────────────────────────────

document.addEventListener('click', e => {
    const btn = e.target.closest('[data-action]');
    if (!btn) return;
    const { action, id, name } = btn.dataset;
    switch (action) {
        case 'launch': handleLaunch(id); break;
        case 'stop':   handleStop(id); break;
        case 'config': handleProfileSettings(id); break;
    }
});

// ─── Init ─────────────────────────────────────────────────────────────────────

async function init() {
    let browserName = '', browserPath = '', hasBrowser = false;
    try {
        browserName = await GetBrowserName();
        browserPath = await GetBrowserPath();
        hasBrowser  = await HasBrowser();
    } catch (e) {}

    const browserInfoHtml = hasBrowser
        ? `<span class="browser-name">${browserName}</span><span class="chromium-path">${browserPath}</span>`
        : `<span class="browser-warning">Nenhum navegador configurado</span>`;

    document.querySelector('#app').innerHTML = `
        <header>
            <div class="header-left">
                <div class="hamburger-wrap">
                    <button class="hamburger-btn" id="hamburger-btn" title="Menu">
                        <span></span><span></span><span></span>
                    </button>
                    <nav class="hamburger-menu" id="hamburger-menu">
                        <button class="hmenu-item" id="hmenu-browser">Configuracoes globais</button>
                        <button class="hmenu-item" id="hmenu-import">Importar backup</button>
                        <div class="hmenu-divider"></div>
                        <button class="hmenu-item" id="hmenu-about">Sobre</button>
                    </nav>
                </div>
                <div class="title-group">
                    <h1>MultiBrowser</h1>
                    <span class="browser-info">${browserInfoHtml}</span>
                </div>
            </div>
            <div class="toolbar">
                <button class="btn btn-primary" id="btn-create">+ Novo Perfil</button>
            </div>
        </header>
        <div class="tabs-container" id="tabs-list"></div>
        <div class="profiles-container" id="profiles-list"></div>
    `;

    document.getElementById('btn-create').onclick = handleCreate;
    setupHamburger();

    await refreshProfiles();
    pollInterval = setInterval(refreshProfiles, 3000);
}

init();
