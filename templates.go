package main

const HOME_HTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SSH Tunnel Manager</title>
    <style>
        :root {
            --primary: #4f46e5;
            --primary-hover: #4338ca;
            --danger: #ef4444;
            --danger-hover: #dc2626;
            --secondary: #6b7280;
            --secondary-hover: #4b5563;
            --bg-color: #f3f4f6;
            --card-bg: #ffffff;
            --text-main: #111827;
            --text-muted: #6b7280;
            --border-color: #e5e7eb;
            --success: #10b981;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--bg-color);
            color: var(--text-main);
            padding: 2rem 1rem;
            line-height: 1.5;
        }
        
        .container { max-width: 900px; margin: 0 auto; }
        
        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 2rem;
        }
        
        h1 { font-size: 1.8rem; font-weight: 600; color: var(--text-main); }
        
        /* Buttons */
        .btn {
            display: inline-flex; align-items: center; justify-content: center;
            padding: 0.5rem 1rem; border: none; border-radius: 0.375rem;
            font-size: 0.875rem; font-weight: 500; cursor: pointer;
            transition: all 0.2s;
            gap: 0.5rem;
        }
        .btn:disabled { opacity: 0.6; cursor: not-allowed; }
        .btn-primary { background: var(--primary); color: white; box-shadow: 0 1px 2px rgba(0,0,0,0.05); }
        .btn-primary:hover:not(:disabled) { background: var(--primary-hover); }
        .btn-danger { background: white; color: var(--danger); border: 1px solid var(--danger); }
        .btn-danger:hover:not(:disabled) { background: #fef2f2; }
        .btn-secondary { background: white; color: var(--text-main); border: 1px solid var(--border-color); }
        .btn-secondary:hover:not(:disabled) { background: #f9fafb; }
        .btn-icon { padding: 0.4rem; }

        /* Cards and Lists */
        .card {
            background: var(--card-bg); border-radius: 0.75rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1), 0 1px 2px rgba(0,0,0,0.06);
            overflow: hidden;
        }
        
        .tunnel-list { display: flex; flex-direction: column; }
        .tunnel-item {
            display: flex; justify-content: space-between; align-items: center;
            padding: 1.25rem 1.5rem; border-bottom: 1px solid var(--border-color);
            transition: background-color 0.2s;
        }
        .tunnel-item:last-child { border-bottom: none; }
        .tunnel-item:hover { background-color: #f9fafb; }
        
        .tunnel-info { flex: 1; display: flex; flex-direction: column; gap: 0.5rem; }
        .tunnel-header {
            display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap;
        }
        .tunnel-name { font-weight: 600; font-size: 1.1rem; color: var(--text-main); }
        
        .tunnel-details { 
            color: var(--text-muted); font-size: 0.875rem; 
            display: flex; gap: 1.5rem; flex-wrap: wrap; 
        }
        .detail-item { display: flex; align-items: center; gap: 0.375rem; }
        
        .tunnel-actions { display: flex; gap: 0.5rem; align-items: center; }

        /* Badges */
        .badge {
            padding: 0.125rem 0.625rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 500;
            display: inline-flex; align-items: center; gap: 0.375rem;
        }
        .badge::before { content: ''; display: inline-block; width: 6px; height: 6px; border-radius: 50%; }
        .status-running { background: #d1fae5; color: #065f46; }
        .status-running::before { background: #10b981; }
        .status-stopped { background: #f3f4f6; color: #4b5563; }
        .status-stopped::before { background: #9ca3af; }
        .status-starting { background: #fef3c7; color: #92400e; }
        .status-starting::before { background: #f59e0b; animation: pulse 1s infinite; }
        
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }
        
        .type-badge { background: #e0e7ff; color: #3730a3; padding: 0.125rem 0.5rem; border-radius: 0.25rem; font-size: 0.75rem; font-weight: 500; display: inline-block;}

        /* Empty State */
        .empty-state { text-align: center; padding: 4rem 1rem; color: var(--text-muted); }
        .empty-state svg { width: 3rem; height: 3rem; margin-bottom: 1rem; opacity: 0.5; color: var(--text-muted); }
        .empty-state p { margin-bottom: 1.5rem; }

        /* Modal */
        .modal {
            display: none; position: fixed; inset: 0; background: rgba(0,0,0,0.5);
            backdrop-filter: blur(2px); justify-content: center; align-items: center;
            z-index: 50; padding: 1rem; opacity: 0; transition: opacity 0.2s;
        }
        .modal.active { display: flex; opacity: 1; }
        .modal-content {
            background: var(--card-bg); border-radius: 0.75rem; width: 100%;
            max-width: 600px; max-height: 90vh; overflow-y: auto;
            box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1), 0 10px 10px -5px rgba(0,0,0,0.04);
            transform: scale(0.95); transition: transform 0.2s;
            display: flex; flex-direction: column;
        }
        .modal.active .modal-content { transform: scale(1); }
        
        .modal-header {
            padding: 1.25rem 1.5rem; border-bottom: 1px solid var(--border-color);
            display: flex; justify-content: space-between; align-items: center;
            position: sticky; top: 0; background: var(--card-bg); z-index: 10;
        }
        .modal-header h2 { font-size: 1.25rem; font-weight: 600; }
        .close-btn { background: none; border: none; font-size: 1.5rem; cursor: pointer; color: var(--text-muted); line-height: 1; }
        .close-btn:hover { color: var(--text-main); }
        
        .modal-body { padding: 1.5rem; }
        
        .modal-footer {
            padding: 1.25rem 1.5rem; border-top: 1px solid var(--border-color);
            display: flex; justify-content: flex-end; gap: 0.75rem;
            position: sticky; bottom: 0; background: var(--card-bg); z-index: 10;
        }

        /* Forms */
        .form-section { margin-bottom: 1.5rem; background: #f9fafb; padding: 1.25rem; border-radius: 0.5rem; border: 1px solid var(--border-color); }
        .form-section-title { font-size: 0.875rem; font-weight: 600; color: var(--text-main); margin-bottom: 1rem; }
        .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
        .form-group { display: flex; flex-direction: column; margin-bottom: 1rem; }
        .form-group:last-child { margin-bottom: 0; }
        .form-group.full-width { grid-column: span 2; }
        label { font-size: 0.875rem; font-weight: 500; margin-bottom: 0.375rem; color: var(--text-main); }
        input, select {
            padding: 0.625rem 0.75rem; border: 1px solid var(--border-color);
            border-radius: 0.375rem; font-size: 0.875rem; transition: all 0.2s;
            background: white; color: var(--text-main); width: 100%;
        }
        input:focus, select:focus { outline: none; border-color: var(--primary); box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1); }
        input[type="checkbox"] { width: auto; margin-right: 0.5rem; cursor: pointer; }
        .checkbox-label { display: flex; align-items: center; cursor: pointer; }
        .help-text { font-size: 0.75rem; color: var(--text-muted); margin-top: 0.25rem; }

        /* Toast */
        #toast-container { position: fixed; bottom: 1.5rem; right: 1.5rem; z-index: 100; display: flex; flex-direction: column; gap: 0.5rem; }
        .toast {
            background: white; color: var(--text-main); padding: 1rem 1.25rem;
            border-radius: 0.5rem; font-size: 0.875rem; font-weight: 500;
            box-shadow: 0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -2px rgba(0,0,0,0.05);
            border-left: 4px solid var(--primary);
            animation: slideIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
            display: flex; align-items: center; justify-content: space-between; gap: 1rem;
            min-width: 300px;
        }
        .toast.error { border-left-color: var(--danger); }
        .toast.success { border-left-color: var(--success); }
        .toast-close { background: none; border: none; color: var(--text-muted); cursor: pointer; }

        @keyframes slideIn { from { transform: translateX(100%); opacity: 0; } to { transform: translateX(0); opacity: 1; } }
        @keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
        .animate-spin { animation: spin 1s linear infinite; }
        
        @media (max-width: 640px) { 
            .form-grid { grid-template-columns: 1fr; } 
            .form-group.full-width { grid-column: span 1; } 
            .tunnel-item { flex-direction: column; align-items: flex-start; gap: 1rem; } 
            .tunnel-actions { width: 100%; justify-content: flex-end; border-top: 1px solid var(--border-color); padding-top: 1rem; margin-top: 0.5rem; } 
            .header { flex-direction: column; align-items: flex-start; gap: 1rem; }
            .header .btn { width: 100%; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div style="display:flex;align-items:center;gap:1rem;">
                <h1>SSH Tunnel Manager</h1>
                <span id="connectionStatus" style="font-size:0.75rem;font-weight:500;display:flex;align-items:center;">
                    <span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:#10b981;margin-right:6px;"></span>已连接
                </span>
            </div>
            <button class="btn btn-primary" onclick="showAddModal()">
                <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
                添加新隧道
            </button>
        </div>
        
        <div class="card">
            <div id="tunnelList" class="tunnel-list">
                <!-- Tunnels will be rendered here -->
            </div>
        </div>
    </div>

    <!-- Add Modal -->
    <div id="addModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h2>添加新隧道</h2>
                <button type="button" class="close-btn" onclick="hideAddModal()">&times;</button>
            </div>
            <form id="addForm">
                <div class="modal-body">
                    <div class="form-group">
                        <label>隧道名称</label>
                        <input type="text" name="name" required placeholder="例如: MySQL 转发">
                    </div>

                    <div class="form-section">
                        <div class="form-section-title">转发设置</div>
                        <div class="form-grid">
                            <div class="form-group full-width">
                                <label>类型</label>
                                <select name="type" required>
                                    <option value="local">本地端口转发 (-L)</option>
                                    <option value="remote">远程端口转发 (-R)</option>
                                </select>
                            </div>
                            <div class="form-group">
                                <label>本地端口</label>
                                <input type="number" name="local_port" required placeholder="例如: 3306">
                            </div>
                            <div class="form-group">
                                <label>远程端口</label>
                                <input type="number" name="remote_port" required placeholder="例如: 3306">
                            </div>
                            <div class="form-group full-width">
                                <label>目标主机 (Remote Host)</label>
                                <input type="text" name="remote_host" placeholder="例如: localhost 或 127.0.0.1" value="localhost">
                                <div class="help-text">通常为 localhost (相对于 SSH 服务器)</div>
                            </div>
                        </div>
                    </div>

                    <div class="form-section">
                        <div class="form-section-title">SSH 服务器认证</div>
                        <div class="form-grid">
                            <div class="form-group">
                                <label>SSH 主机</label>
                                <input type="text" name="ssh_host" required placeholder="例如: 192.168.1.100">
                            </div>
                            <div class="form-group">
                                <label>SSH 端口</label>
                                <input type="number" name="ssh_port" placeholder="22" value="22">
                            </div>
                            <div class="form-group full-width">
                                <label>SSH 用户名</label>
                                <input type="text" name="ssh_user" required placeholder="例如: root">
                            </div>
                            <div class="form-group full-width">
                                <label>SSH 密码</label>
                                <input type="password" name="ssh_pass" placeholder="如果使用密码验证则填写">
                            </div>
                            <div class="form-group full-width">
                                <label>SSH 密钥路径 (可选)</label>
                                <input type="text" name="ssh_key" placeholder="例如: ~/.ssh/id_rsa">
                                <div class="help-text">如果配置了密码，密钥将被优先使用。可留空。</div>
                            </div>
                        </div>
                    </div>

                    <div class="form-section">
                        <div class="form-section-title">高级设置</div>
                        <div class="form-grid">
                            <div class="form-group full-width">
                                <label class="checkbox-label">
                                    <input type="checkbox" name="auto_reconnect" id="add_auto_reconnect">
                                    <span>断线后自动重连</span>
                                </label>
                            </div>
                            <div class="form-group" id="add_reconnect_delay_group" style="display: none;">
                                <label>重连延迟 (秒)</label>
                                <input type="number" name="reconnect_delay" id="add_reconnect_delay" placeholder="5" value="5" min="1" max="60">
                            </div>
                        </div>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" onclick="hideAddModal()">取消</button>
                    <button type="submit" class="btn btn-primary" id="submitBtn">保存隧道</button>
                </div>
            </form>
        </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div id="deleteModal" class="modal">
        <div class="modal-content" style="max-width: 400px;">
            <div class="modal-header">
                <h2>确认删除</h2>
                <button type="button" class="close-btn" onclick="hideDeleteModal()">&times;</button>
            </div>
            <div class="modal-body">
                <p>确定要删除隧道 <strong id="deleteTunnelName"></strong> 吗？此操作无法撤销。</p>
            </div>
            <div class="modal-footer">
                <button type="button" class="btn btn-secondary" onclick="hideDeleteModal()">取消</button>
                <button type="button" class="btn btn-danger" id="confirmDeleteBtn">确认删除</button>
            </div>
        </div>
    </div>

    <!-- Edit Modal -->
    <div id="editModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h2>编辑隧道</h2>
                <button type="button" class="close-btn" onclick="hideEditModal()">&times;</button>
            </div>
            <form id="editForm">
                <div class="modal-body">
                    <div class="form-group">
                        <label>隧道名称</label>
                        <input type="text" name="name" required placeholder="例如: MySQL 转发">
                    </div>

                    <div class="form-section">
                        <div class="form-section-title">转发设置</div>
                        <div class="form-grid">
                            <div class="form-group full-width">
                                <label>类型</label>
                                <select name="type" required>
                                    <option value="local">本地端口转发 (-L)</option>
                                    <option value="remote">远程端口转发 (-R)</option>
                                </select>
                            </div>
                            <div class="form-group">
                                <label>本地端口</label>
                                <input type="number" name="local_port" required placeholder="例如: 3306">
                            </div>
                            <div class="form-group">
                                <label>远程端口</label>
                                <input type="number" name="remote_port" required placeholder="例如: 3306">
                            </div>
                            <div class="form-group full-width">
                                <label>目标主机 (Remote Host)</label>
                                <input type="text" name="remote_host" placeholder="例如: localhost 或 127.0.0.1">
                                <div class="help-text">通常为 localhost (相对于 SSH 服务器)</div>
                            </div>
                        </div>
                    </div>

                    <div class="form-section">
                        <div class="form-section-title">SSH 服务器认证</div>
                        <div class="form-grid">
                            <div class="form-group">
                                <label>SSH 主机</label>
                                <input type="text" name="ssh_host" required placeholder="例如: 192.168.1.100">
                            </div>
                            <div class="form-group">
                                <label>SSH 端口</label>
                                <input type="number" name="ssh_port" placeholder="22" value="22">
                            </div>
                            <div class="form-group full-width">
                                <label>SSH 用户名</label>
                                <input type="text" name="ssh_user" required placeholder="例如: root">
                            </div>
                            <div class="form-group full-width">
                                <label>SSH 密码</label>
                                <input type="password" name="ssh_pass" placeholder="留空则不修改密码">
                            </div>
                            <div class="form-group full-width">
                                <label>SSH 密钥路径 (可选)</label>
                                <input type="text" name="ssh_key" placeholder="例如: ~/.ssh/id_rsa">
                                <div class="help-text">如果配置了密码，密钥将被优先使用。</div>
                            </div>
                        </div>
                    </div>

                    <div class="form-section">
                        <div class="form-section-title">高级设置</div>
                        <div class="form-grid">
                            <div class="form-group full-width">
                                <label class="checkbox-label">
                                    <input type="checkbox" name="auto_reconnect" id="edit_auto_reconnect">
                                    <span>断线后自动重连</span>
                                </label>
                            </div>
                            <div class="form-group" id="edit_reconnect_delay_group" style="display: none;">
                                <label>重连延迟 (秒)</label>
                                <input type="number" name="reconnect_delay" id="edit_reconnect_delay" placeholder="5" value="5" min="1" max="60">
                            </div>
                        </div>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" onclick="hideEditModal()">取消</button>
                    <button type="submit" class="btn btn-primary" id="editSubmitBtn">保存更改</button>
                </div>
            </form>
        </div>
    </div>

    <div id="toast-container"></div>

    <script>
        let tunnels = [];
        let tunnelToDelete = null;
        let tunnelToEdit = null;
        let connectionStatus = 'connected';
        let lastPongTime = Date.now();

        // --- Toast System ---
        function showToast(message, type) {
            type = type || 'success';
            const container = document.getElementById('toast-container');
            const toast = document.createElement('div');
            toast.className = 'toast ' + type;
            
            const icon = type === 'success' 
                ? '<svg width="20" height="20" fill="none" stroke="currentColor" viewBox="0 0 24 24" color="var(--success)"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>'
                : '<svg width="20" height="20" fill="none" stroke="currentColor" viewBox="0 0 24 24" color="var(--danger)"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>';

            toast.innerHTML = 
                '<div style="display: flex; align-items: center; gap: 0.5rem;">' +
                    icon +
                    '<span>' + escapeHtml(message) + '</span>' +
                '</div>' +
                '<button type="button" class="toast-close" onclick="this.parentElement.remove()">' +
                    '<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>' +
                '</button>';
            
            container.appendChild(toast);
            
            setTimeout(() => {
                if (toast.parentElement) {
                    toast.style.opacity = '0';
                    toast.style.transform = 'translateX(100%)';
                    toast.style.transition = 'all 0.3s ease-out';
                    setTimeout(() => toast.remove(), 300);
                }
            }, 3000);
        }

        // --- Connection Status ---
        function updateConnectionStatus(status) {
            connectionStatus = status;
            const statusDot = document.getElementById('connectionStatus');
            if (statusDot) {
                if (status === 'connected') {
                    statusDot.innerHTML = '<span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:#10b981;margin-right:6px;"></span>已连接';
                    statusDot.style.color = '#065f46';
                } else if (status === 'disconnected') {
                    statusDot.innerHTML = '<span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:#ef4444;margin-right:6px;"></span>连接断开';
                    statusDot.style.color = '#dc2626';
                }
            }
        }

        // --- Data Fetching & Rendering ---
        async function loadTunnels() {
            try {
                const res = await fetch('/api/tunnels');
                if (!res.ok) {
                    throw new Error('Failed to load tunnels: ' + res.status);
                }
                const data = await res.json();
                tunnels = Array.isArray(data) ? data : [];
                renderTunnels();
            } catch (err) {
                console.error('Failed to load tunnels:', err);
                showToast('加载隧道列表失败', 'error');
                tunnels = [];
                renderTunnels();
            }
        }

        async function ping() {
            try {
                const res = await fetch('/api/ping');
                if (res.ok) {
                    lastPongTime = Date.now();
                    updateConnectionStatus('connected');
                } else {
                    updateConnectionStatus('disconnected');
                }
            } catch (err) {
                console.error('Ping failed:', err);
                updateConnectionStatus('disconnected');
            }
        }

        function renderTunnels() {
            const list = document.getElementById('tunnelList');
            if (!Array.isArray(tunnels) || tunnels.length === 0) {
                list.innerHTML = 
                    '<div class="empty-state">' +
                        '<svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg>' +
                        '<p>暂无配置的 SSH 隧道</p>' +
                        '<button class="btn btn-primary" onclick="showAddModal()">立即添加</button>' +
                    '</div>';
                return;
            }
            
            list.innerHTML = tunnels.map(t => {
                const isStarting = t.status === 'starting';
                const isRunning = t.status === 'running';
                const isStopped = t.status === 'stopped';
                
                let statusClass, statusLabel;
                if (isStarting) {
                    statusClass = 'status-starting';
                    statusLabel = '启动中';
                } else if (isRunning) {
                    statusClass = 'status-running';
                    statusLabel = '运行中';
                } else {
                    statusClass = 'status-stopped';
                    statusLabel = '已停止';
                }
                
                const typeLabel = t.type === 'local' ? '本地转发 (L)' : '远程转发 (R)';
                const sshPort = t.ssh_port || '22';
                const remoteHost = t.remote_host || 'localhost';
                
                let actionBtn;
                if (isStarting) {
                    actionBtn = '<button class="btn btn-secondary" disabled title="启动中">' +
                        '<svg class="animate-spin" width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path></svg>' +
                        '启动中' +
                       '</button>';
                } else if (isRunning) {
                    actionBtn = '<button class="btn btn-secondary" onclick="stopTunnel(\'' + t.id + '\', this)" title="停止">' +
                        '<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"></path></svg>' +
                        '停止' +
                       '</button>';
                } else {
                    actionBtn = '<button class="btn btn-primary" onclick="startTunnel(\'' + t.id + '\', this)" title="启动">' +
                        '<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>' +
                        '启动' +
                       '</button>';
                }
                       
                return '<div class="tunnel-item">' +
                    '<div class="tunnel-info">' +
                        '<div class="tunnel-header">' +
                            '<span class="tunnel-name">' + escapeHtml(t.name) + '</span>' +
                            '<span class="badge ' + statusClass + '">' + statusLabel + '</span>' +
                            '<span class="type-badge">' + typeLabel + '</span>' +
                        '</div>' +
                        '<div class="tunnel-details">' +
                            '<div class="detail-item">' +
                                '<svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"></path></svg>' +
                                '<span>' + escapeHtml(t.local_port) + ' &rarr; ' + escapeHtml(remoteHost) + ':' + escapeHtml(t.remote_port) + '</span>' +
                            '</div>' +
                            '<div class="detail-item">' +
                                '<svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"></path></svg>' +
                                '<span>' + escapeHtml(t.ssh_user) + '@' + escapeHtml(t.ssh_host) + ':' + escapeHtml(sshPort) + '</span>' +
                            '</div>' +
                        '</div>' +
                    '</div>' +
                    '<div class="tunnel-actions">' +
                        actionBtn +
                        '<button class="btn btn-secondary btn-icon" onclick="showEditModal(\'' + t.id + '\')" title="编辑" ' + (isRunning || isStarting ? 'disabled' : '') + '>' +
                            '<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path></svg>' +
                        '</button>' +
                        '<button class="btn btn-danger btn-icon" onclick="promptDelete(\'' + t.id + '\', \'' + escapeHtml(t.name) + '\')" title="删除">' +
                            '<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>' +
                        '</button>' +
                    '</div>' +
                '</div>';
            }).join('');
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text || '';
            return div.innerHTML;
        }

        // --- Modals ---
        function showAddModal() {
            document.getElementById('addModal').classList.add('active');
        }

        function hideAddModal() {
            document.getElementById('addModal').classList.remove('active');
            document.getElementById('addForm').reset();
        }

        function promptDelete(id, name) {
            tunnelToDelete = id;
            document.getElementById('deleteTunnelName').innerText = name;
            document.getElementById('deleteModal').classList.add('active');
        }

        function hideDeleteModal() {
            tunnelToDelete = null;
            document.getElementById('deleteModal').classList.remove('active');
        }

        function showEditModal(id) {
            const tunnel = tunnels.find(t => t.id === id);
            if (!tunnel) return;

            tunnelToEdit = id;
            const form = document.getElementById('editForm');

            form.name.value = tunnel.name || '';
            form.type.value = tunnel.type || 'local';
            form.local_port.value = tunnel.local_port || '';
            form.remote_port.value = tunnel.remote_port || '';
            form.remote_host.value = tunnel.remote_host || 'localhost';
            form.ssh_host.value = tunnel.ssh_host || '';
            form.ssh_port.value = tunnel.ssh_port || '22';
            form.ssh_user.value = tunnel.ssh_user || '';
            form.ssh_pass.value = '';
            form.ssh_key.value = tunnel.ssh_key || '';
            form.auto_reconnect.checked = tunnel.auto_reconnect || false;
            form.reconnect_delay.value = tunnel.reconnect_delay || 5;
            toggleReconnectDelay('edit');

            document.getElementById('editModal').classList.add('active');
        }

        function hideEditModal() {
            tunnelToEdit = null;
            document.getElementById('editModal').classList.remove('active');
            document.getElementById('editForm').reset();
            document.getElementById('edit_reconnect_delay_group').style.display = 'none';
        }

        function toggleReconnectDelay(formPrefix) {
            const checkbox = document.getElementById(formPrefix + '_auto_reconnect');
            const delayGroup = document.getElementById(formPrefix + '_reconnect_delay_group');
            if (checkbox && delayGroup) {
                delayGroup.style.display = checkbox.checked ? 'block' : 'none';
            }
        }

        document.getElementById('add_auto_reconnect').addEventListener('change', function() {
            toggleReconnectDelay('add');
        });
        document.getElementById('edit_auto_reconnect').addEventListener('change', function() {
            toggleReconnectDelay('edit');
        });

        // --- Actions ---
        document.getElementById('addForm').onsubmit = async (e) => {
            e.preventDefault();
            const form = e.target;
            const submitBtn = document.getElementById('submitBtn');
            const originalText = submitBtn.innerText;
            
            submitBtn.disabled = true;
            submitBtn.innerText = '保存中...';

            const data = {
                name: form.name.value,
                type: form.type.value,
                local_port: form.local_port.value,
                remote_host: form.remote_host.value || 'localhost',
                remote_port: form.remote_port.value,
                ssh_host: form.ssh_host.value,
                ssh_port: form.ssh_port.value || '22',
                ssh_user: form.ssh_user.value,
                ssh_key: form.ssh_key.value,
                ssh_pass: form.ssh_pass.value,
                auto_reconnect: form.auto_reconnect.checked,
                reconnect_delay: parseInt(form.reconnect_delay.value) || 5
            };

            try {
                const res = await fetch('/api/tunnels', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(data)
                });
                if (res.ok) {
                    hideAddModal();
                    loadTunnels();
                    showToast('隧道添加成功');
                } else {
                    const err = await res.json();
                    showToast('添加失败: ' + (err.error || '未知错误'), 'error');
                }
            } catch (err) {
                showToast('添加失败: ' + err, 'error');
            } finally {
                submitBtn.disabled = false;
                submitBtn.innerText = originalText;
            }
        };

        document.getElementById('editForm').onsubmit = async (e) => {
            e.preventDefault();
            if (!tunnelToEdit) return;

            const form = e.target;
            const submitBtn = document.getElementById('editSubmitBtn');
            const originalText = submitBtn.innerText;

            submitBtn.disabled = true;
            submitBtn.innerText = '保存中...';

            const data = {
                name: form.name.value,
                type: form.type.value,
                local_port: form.local_port.value,
                remote_host: form.remote_host.value || 'localhost',
                remote_port: form.remote_port.value,
                ssh_host: form.ssh_host.value,
                ssh_port: form.ssh_port.value || '22',
                ssh_user: form.ssh_user.value,
                ssh_key: form.ssh_key.value,
                ssh_pass: form.ssh_pass.value,
                auto_reconnect: form.auto_reconnect.checked,
                reconnect_delay: parseInt(form.reconnect_delay.value) || 5
            };

            try {
                const res = await fetch('/api/tunnels/' + tunnelToEdit, {
                    method: 'PUT',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(data)
                });
                if (res.ok) {
                    hideEditModal();
                    loadTunnels();
                    showToast('隧道已更新');
                } else {
                    const err = await res.json();
                    showToast('更新失败: ' + (err.error || '未知错误'), 'error');
                }
            } catch (err) {
                showToast('更新失败: ' + err, 'error');
            } finally {
                submitBtn.disabled = false;
                submitBtn.innerText = originalText;
            }
        };

        async function startTunnel(id, btnElement) {
            btnElement.disabled = true;
            btnElement.innerHTML = '启动中...';
            try {
                const res = await fetch('/api/tunnels/' + id + '/start', {method: 'POST'});
                if (res.ok) {
                    showToast('隧道已启动');
                    loadTunnels();
                } else {
                    const err = await res.json().catch(()=>({}));
                    showToast('启动失败: ' + (err.error || '服务器错误'), 'error');
                    loadTunnels();
                }
            } catch (err) {
                showToast('启动失败: ' + err, 'error');
                loadTunnels();
            }
        }

        async function stopTunnel(id, btnElement) {
            btnElement.disabled = true;
            btnElement.innerHTML = '停止中...';
            try {
                const res = await fetch('/api/tunnels/' + id + '/stop', {method: 'POST'});
                if (res.ok) {
                    showToast('隧道已停止');
                    loadTunnels();
                } else {
                    const err = await res.json().catch(()=>({}));
                    showToast('停止失败: ' + (err.error || '服务器错误'), 'error');
                    loadTunnels();
                }
            } catch (err) {
                showToast('停止失败: ' + err, 'error');
                loadTunnels();
            }
        }

        document.getElementById('confirmDeleteBtn').onclick = async () => {
            if (!tunnelToDelete) return;
            
            const btn = document.getElementById('confirmDeleteBtn');
            const originalText = btn.innerText;
            btn.disabled = true;
            btn.innerText = '删除中...';

            try {
                const res = await fetch('/api/tunnels/' + tunnelToDelete, {method: 'DELETE'});
                if (res.ok) {
                    showToast('隧道已删除');
                    hideDeleteModal();
                    loadTunnels();
                } else {
                    const err = await res.json().catch(()=>({}));
                    showToast('删除失败: ' + (err.error || '服务器错误'), 'error');
                }
            } catch (err) {
                showToast('删除失败: ' + err, 'error');
            } finally {
                btn.disabled = false;
                btn.innerText = originalText;
            }
        };

        // Close modals on clicking outside
        document.getElementById('addModal').addEventListener('click', (e) => {
            if (e.target === document.getElementById('addModal')) hideAddModal();
        });
        document.getElementById('deleteModal').addEventListener('click', (e) => {
            if (e.target === document.getElementById('deleteModal')) hideDeleteModal();
        });
        document.getElementById('editModal').addEventListener('click', (e) => {
            if (e.target === document.getElementById('editModal')) hideEditModal();
        });

        // Heartbeat and auto refresh
        setInterval(() => {
            ping();
            loadTunnels();
        }, 5000);
        
        // Initial load
        document.addEventListener('DOMContentLoaded', () => {
            ping();
            loadTunnels();
        });
    </script>
</body>
</html>`
