// Dashboard JavaScript for Learning Bot Management

// Global variables
let currentGroups = [];
let currentCommands = [];
let currentAutoResponses = [];

// Initialize dashboard
document.addEventListener('DOMContentLoaded', function() {
    showTab('groups');
    refreshGroups();
    createModals();
});

// Tab management
function showTab(tabName) {
    // Hide all tabs
    const tabs = document.querySelectorAll('.tab-content');
    tabs.forEach(tab => tab.style.display = 'none');
    
    // Remove active class from all nav links
    const navLinks = document.querySelectorAll('.nav-link');
    navLinks.forEach(link => link.classList.remove('active'));
    
    // Show selected tab
    document.getElementById(tabName + '-tab').style.display = 'block';
    
    // Add active class to clicked nav link
    event.target.classList.add('active');
    
    // Load data for selected tab
    switch(tabName) {
        case 'groups':
            refreshGroups();
            break;
        case 'commands':
            refreshCommands();
            break;
        case 'autoresponses':
            refreshAutoResponses();
            break;
        case 'stats':
            refreshStats();
            break;
    }
}

// === GROUPS MANAGEMENT ===

function refreshGroups() {
    fetch('/api/groups')
        .then(response => response.json())
        .then(data => {
            currentGroups = data || [];
            displayGroups();
        })
        .catch(error => {
            console.error('Error fetching groups:', error);
            showAlert('danger', 'Gagal memuat data grup');
        });
}

function displayGroups() {
    const container = document.getElementById('groups-list');
    if (currentGroups.length === 0) {
        container.innerHTML = `
            <div class="alert alert-info">
                <i class="fas fa-info-circle"></i> Belum ada grup yang terdaftar.
                Gunakan personal chat bot dengan command .addgroup untuk menambah grup.
            </div>
        `;
        return;
    }
    
    let html = '<div class="row">';
    currentGroups.forEach(group => {
        const statusBadge = group.is_active ? 
            '<span class="badge bg-success">Aktif</span>' : 
            '<span class="badge bg-secondary">Tidak Aktif</span>';
        
        const toggleButton = group.is_active ? 
            `<button class="btn btn-sm btn-warning" onclick="toggleGroup('${group.group_jid}', false)">
                <i class="fas fa-pause"></i> Nonaktifkan
            </button>` :
            `<button class="btn btn-sm btn-success" onclick="toggleGroup('${group.group_jid}', true)">
                <i class="fas fa-play"></i> Aktifkan
            </button>`;
        
        html += `
            <div class="col-md-6 mb-3">
                <div class="card">
                    <div class="card-body">
                        <h6 class="card-title">
                            <i class="fas fa-users"></i> ${group.group_name}
                            ${statusBadge}
                        </h6>
                        <p class="card-text text-muted small">
                            <strong>JID:</strong> ${group.group_jid}<br>
                            <strong>Dibuat:</strong> ${formatDate(group.created_at)}<br>
                            <strong>Oleh:</strong> ${group.created_by}
                        </p>
                        <div class="btn-group" role="group">
                            ${toggleButton}
                            <button class="btn btn-sm btn-danger" onclick="deleteGroup('${group.group_jid}', '${group.group_name}')">
                                <i class="fas fa-trash"></i> Hapus
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        `;
    });
    html += '</div>';
    
    container.innerHTML = html;
}

function toggleGroup(groupJID, isActive) {
    const group = currentGroups.find(g => g.group_jid === groupJID);
    if (!group) return;
    
    group.is_active = isActive;
    
    fetch('/api/groups', {
        method: 'PUT',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(group)
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            showAlert('success', `Grup ${isActive ? 'diaktifkan' : 'dinonaktifkan'}`);
            refreshGroups();
        } else {
            showAlert('danger', 'Gagal mengubah status grup');
        }
    })
    .catch(error => {
        console.error('Error toggling group:', error);
        showAlert('danger', 'Gagal mengubah status grup');
    });
}

function deleteGroup(groupJID, groupName) {
    if (!confirm(`Hapus grup "${groupName}"?`)) return;
    
    fetch(`/api/groups?jid=${encodeURIComponent(groupJID)}`, {
        method: 'DELETE'
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            showAlert('success', 'Grup berhasil dihapus');
            refreshGroups();
        } else {
            showAlert('danger', 'Gagal menghapus grup');
        }
    })
    .catch(error => {
        console.error('Error deleting group:', error);
        showAlert('danger', 'Gagal menghapus grup');
    });
}

// === COMMANDS MANAGEMENT ===

function refreshCommands() {
    fetch('/api/commands')
        .then(response => response.json())
        .then(data => {
            currentCommands = data || [];
            displayCommands();
        })
        .catch(error => {
            console.error('Error fetching commands:', error);
            showAlert('danger', 'Gagal memuat data command');
        });
}

function displayCommands() {
    const container = document.getElementById('commands-list');
    if (currentCommands.length === 0) {
        container.innerHTML = `
            <div class="alert alert-info">
                <i class="fas fa-info-circle"></i> Belum ada command yang terdaftar.
                Klik "Tambah Command" untuk menambah command baru.
            </div>
        `;
        return;
    }
    
    let html = `
        <div class="table-responsive">
            <table class="table table-striped">
                <thead>
                    <tr>
                        <th>Command</th>
                        <th>Judul</th>
                        <th>Tipe Response</th>
                        <th>Kategori</th>
                        <th>Penggunaan</th>
                        <th>Status</th>
                        <th>Aksi</th>
                    </tr>
                </thead>
                <tbody>
    `;
    
    currentCommands.forEach(cmd => {
        const statusBadge = cmd.is_active ? 
            '<span class="badge bg-success">Aktif</span>' : 
            '<span class="badge bg-secondary">Tidak Aktif</span>';
        
        const responseTypeIcon = getResponseTypeIcon(cmd.response_type);
        
        html += `
            <tr>
                <td><code>${cmd.command}</code></td>
                <td>${cmd.title}</td>
                <td>${responseTypeIcon} ${cmd.response_type}</td>
                <td><span class="badge bg-info">${cmd.category}</span></td>
                <td>${cmd.usage_count || 0}x</td>
                <td>${statusBadge}</td>
                <td>
                    <div class="btn-group" role="group">
                        <button class="btn btn-sm btn-primary" onclick="editCommand('${cmd.command}')">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button class="btn btn-sm btn-danger" onclick="deleteCommand('${cmd.command}')">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </td>
            </tr>
        `;
    });
    
    html += '</tbody></table></div>';
    container.innerHTML = html;
}

function getResponseTypeIcon(type) {
    const icons = {
        'text': '<i class="fas fa-font"></i>',
        'image': '<i class="fas fa-image"></i>',
        'video': '<i class="fas fa-video"></i>',
        'audio': '<i class="fas fa-music"></i>',
        'sticker': '<i class="fas fa-smile"></i>',
        'file': '<i class="fas fa-file"></i>'
    };
    return icons[type] || '<i class="fas fa-question"></i>';
}

function editCommand(command) {
    const cmd = currentCommands.find(c => c.command === command);
    if (!cmd) return;
    
    // Fill edit form with current data
    document.getElementById('editCommand').value = cmd.command;
    document.getElementById('editTitle').value = cmd.title;
    document.getElementById('editDescription').value = cmd.description || '';
    document.getElementById('editResponseType').value = cmd.response_type;
    document.getElementById('editTextContent').value = cmd.text_content || '';
    document.getElementById('editCaption').value = cmd.caption || '';
    document.getElementById('editCategory').value = cmd.category;
    document.getElementById('editIsActive').checked = cmd.is_active;
    
    // Show edit modal
    new bootstrap.Modal(document.getElementById('editCommandModal')).show();
}

function deleteCommand(command) {
    if (!confirm(`Hapus command "${command}"?`)) return;
    
    fetch(`/api/commands?command=${encodeURIComponent(command)}`, {
        method: 'DELETE'
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            showAlert('success', 'Command berhasil dihapus');
            refreshCommands();
        } else {
            showAlert('danger', 'Gagal menghapus command');
        }
    })
    .catch(error => {
        console.error('Error deleting command:', error);
        showAlert('danger', 'Gagal menghapus command');
    });
}

// === AUTO RESPONSES MANAGEMENT ===

function refreshAutoResponses() {
    fetch('/api/autoresponses')
        .then(response => response.json())
        .then(data => {
            currentAutoResponses = data || [];
            displayAutoResponses();
        })
        .catch(error => {
            console.error('Error fetching auto responses:', error);
            showAlert('danger', 'Gagal memuat data auto response');
        });
}

function displayAutoResponses() {
    const container = document.getElementById('autoresponses-list');
    if (currentAutoResponses.length === 0) {
        container.innerHTML = `
            <div class="alert alert-info">
                <i class="fas fa-info-circle"></i> Belum ada auto response yang terdaftar.
                Klik "Tambah Auto Response" untuk menambah auto response baru.
            </div>
        `;
        return;
    }
    
    let html = `
        <div class="table-responsive">
            <table class="table table-striped">
                <thead>
                    <tr>
                        <th>Keyword</th>
                        <th>Tipe Response</th>
                        <th>Response Content</th>
                        <th>Penggunaan</th>
                        <th>Status</th>
                        <th>Aksi</th>
                    </tr>
                </thead>
                <tbody>
    `;
    
    currentAutoResponses.forEach(resp => {
        const statusBadge = resp.is_active ? 
            '<span class="badge bg-success">Aktif</span>' : 
            '<span class="badge bg-secondary">Tidak Aktif</span>';
        
        let responseContent = '';
        if (resp.text_response) {
            responseContent = resp.text_response.substring(0, 50) + '...';
        } else if (resp.sticker_path) {
            responseContent = `<i class="fas fa-smile"></i> Sticker: ${resp.sticker_path.split('/').pop()}`;
        } else if (resp.audio_path) {
            responseContent = `<i class="fas fa-music"></i> Audio: ${resp.audio_path.split('/').pop()}`;
        }
        
        html += `
            <tr>
                <td><code>${resp.keyword}</code></td>
                <td><span class="badge bg-info">${resp.response_type}</span></td>
                <td>${responseContent}</td>
                <td>${resp.usage_count || 0}x</td>
                <td>${statusBadge}</td>
                <td>
                    <div class="btn-group" role="group">
                        <button class="btn btn-sm btn-primary" onclick="editAutoResponse('${resp.keyword}')">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button class="btn btn-sm btn-danger" onclick="deleteAutoResponse('${resp.keyword}')">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </td>
            </tr>
        `;
    });
    
    html += '</tbody></table></div>';
    container.innerHTML = html;
}

function editAutoResponse(keyword) {
    const resp = currentAutoResponses.find(r => r.keyword === keyword);
    if (!resp) return;
    
    // Fill edit form with current data
    document.getElementById('editAutoKeyword').value = resp.keyword;
    document.getElementById('editAutoResponseType').value = resp.response_type;
    document.getElementById('editAutoTextResponse').value = resp.text_response || '';
    document.getElementById('editAutoIsActive').checked = resp.is_active;
    
    // Show edit modal
    new bootstrap.Modal(document.getElementById('editAutoResponseModal')).show();
}

function deleteAutoResponse(keyword) {
    if (!confirm(`Hapus auto response "${keyword}"?`)) return;
    
    fetch(`/api/autoresponses?keyword=${encodeURIComponent(keyword)}`, {
        method: 'DELETE'
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            showAlert('success', 'Auto response berhasil dihapus');
            refreshAutoResponses();
        } else {
            showAlert('danger', 'Gagal menghapus auto response');
        }
    })
    .catch(error => {
        console.error('Error deleting auto response:', error);
        showAlert('danger', 'Gagal menghapus auto response');
    });
}

// === STATISTICS ===

function refreshStats() {
    fetch('/api/stats?days=7')
        .then(response => response.json())
        .then(data => {
            displayStats(data);
        })
        .catch(error => {
            console.error('Error fetching stats:', error);
            showAlert('danger', 'Gagal memuat statistik');
        });
}

function displayStats(data) {
    const container = document.getElementById('stats-content');
    
    let html = `
        <div class="row mb-4">
            <div class="col-md-3">
                <div class="card card-stats">
                    <div class="card-body">
                        <h3 class="text-primary">${data.counts.groups}</h3>
                        <p class="mb-0"><i class="fas fa-users"></i> Total Grup</p>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card card-stats">
                    <div class="card-body">
                        <h3 class="text-success">${data.counts.commands}</h3>
                        <p class="mb-0"><i class="fas fa-terminal"></i> Total Command</p>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card card-stats">
                    <div class="card-body">
                        <h3 class="text-info">${data.counts.auto_responses}</h3>
                        <p class="mb-0"><i class="fas fa-magic"></i> Auto Response</p>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card card-stats">
                    <div class="card-body">
                        <h3 class="text-warning">${Object.keys(data.usage_stats || {}).length}</h3>
                        <p class="mb-0"><i class="fas fa-chart-line"></i> Command Digunakan</p>
                    </div>
                </div>
            </div>
        </div>
    `;
    
    // Usage stats
    if (data.usage_stats && Object.keys(data.usage_stats).length > 0) {
        html += `
            <div class="row">
                <div class="col-md-6">
                    <div class="card">
                        <div class="card-header">
                            <h5><i class="fas fa-chart-bar"></i> Command Populer (${data.days} hari)</h5>
                        </div>
                        <div class="card-body">
        `;
        
        Object.entries(data.usage_stats)
            .sort(([,a], [,b]) => b - a)
            .slice(0, 10)
            .forEach(([command, count]) => {
                html += `
                    <div class="d-flex justify-content-between mb-2">
                        <span><code>${command}</code></span>
                        <span class="badge bg-primary">${count}x</span>
                    </div>
                `;
            });
        
        html += `
                        </div>
                    </div>
                </div>
                <div class="col-md-6">
                    <div class="card">
                        <div class="card-header">
                            <h5><i class="fas fa-history"></i> Activity Log Terbaru</h5>
                        </div>
                        <div class="card-body">
        `;
        
        if (data.recent_logs && data.recent_logs.length > 0) {
            data.recent_logs.slice(0, 10).forEach(log => {
                const statusIcon = log.success ? 
                    '<i class="fas fa-check text-success"></i>' : 
                    '<i class="fas fa-times text-danger"></i>';
                
                html += `
                    <div class="d-flex justify-content-between mb-2 small">
                        <span>${statusIcon} <code>${log.command_value}</code></span>
                        <span class="text-muted">${formatDateTime(log.used_at)}</span>
                    </div>
                `;
            });
        } else {
            html += '<p class="text-muted">Belum ada aktivitas</p>';
        }
        
        html += `
                        </div>
                    </div>
                </div>
            </div>
        `;
    } else {
        html += `
            <div class="alert alert-info">
                <i class="fas fa-info-circle"></i> Belum ada data penggunaan dalam ${data.days} hari terakhir.
            </div>
        `;
    }
    
    container.innerHTML = html;
}

// === UTILITY FUNCTIONS ===

function showAlert(type, message) {
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type} alert-dismissible fade show`;
    alertDiv.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;
    
    // Insert at top of content area
    const contentArea = document.querySelector('.content-area');
    contentArea.insertBefore(alertDiv, contentArea.firstChild);
    
    // Auto dismiss after 5 seconds
    setTimeout(() => {
        alertDiv.remove();
    }, 5000);
}

function formatDate(dateString) {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleDateString('id-ID');
}

function formatDateTime(dateString) {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleString('id-ID');
}

function uploadFile(fileInput, fileType, callback) {
    const file = fileInput.files[0];
    if (!file) {
        showAlert('warning', 'Pilih file terlebih dahulu');
        return;
    }
    
    const formData = new FormData();
    formData.append('file', file);
    formData.append('type', fileType);
    
    fetch('/api/upload', {
        method: 'POST',
        body: formData
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            showAlert('success', `File berhasil diupload: ${data.filename}`);
            if (callback) callback(data.filepath);
        } else {
            showAlert('danger', 'Gagal mengupload file');
        }
    })
    .catch(error => {
        console.error('Error uploading file:', error);
        showAlert('danger', 'Gagal mengupload file');
    });
}

// === MODAL MANAGEMENT ===

function createModals() {
    const modalsContainer = document.getElementById('modals-container');
    modalsContainer.innerHTML = `
        <!-- Add Command Modal -->
        <div class="modal fade" id="addCommandModal" tabindex="-1">
            <div class="modal-dialog modal-lg">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title"><i class="fas fa-plus"></i> Tambah Command Baru</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                    </div>
                    <div class="modal-body">
                        <form id="addCommandForm">
                            <div class="row">
                                <div class="col-md-6">
                                    <div class="mb-3">
                                        <label class="form-label">Command *</label>
                                        <input type="text" class="form-control" id="newCommand" placeholder=".listbugs" required>
                                        <small class="text-muted">Harus dimulai dengan titik (.)</small>
                                    </div>
                                </div>
                                <div class="col-md-6">
                                    <div class="mb-3">
                                        <label class="form-label">Kategori</label>
                                        <select class="form-control" id="newCategory">
                                            <option value="injec">💉 Injec</option>
                                            <option value="pembelajaran">📚 Pembelajaran</option>
                                            <option value="informasi">ℹ️ Informasi</option>
                                            <option value="tools">🛠️ Tools</option>
                                        </select>
                                    </div>
                                </div>
                            </div>
                            <div class="mb-3">
                                <label class="form-label">Judul *</label>
                                <input type="text" class="form-control" id="newTitle" placeholder="List Bug VPN" required>
                            </div>
                            <div class="mb-3">
                                <label class="form-label">Deskripsi</label>
                                <input type="text" class="form-control" id="newDescription" placeholder="Daftar bug server VPN untuk pembelajaran">
                            </div>
                            <div class="mb-3">
                                <label class="form-label">Tipe Response *</label>
                                <select class="form-control" id="newResponseType" onchange="toggleResponseInputs(this.value)" required>
                                    <option value="text">📝 Text</option>
                                    <option value="image">🖼️ Gambar</option>
                                    <option value="video">🎥 Video</option>
                                    <option value="audio">🎵 Audio</option>
                                    <option value="sticker">😄 Sticker</option>
                                    <option value="file">📁 File/APK</option>
                                </select>
                            </div>
                            <div id="textResponse" class="mb-3">
                                <label class="form-label">Text Content</label>
                                <textarea class="form-control" id="newTextContent" rows="5" placeholder="Masukkan text response..."></textarea>
                            </div>
                            <div id="mediaResponse" class="mb-3" style="display:none;">
                                <label class="form-label">Upload File</label>
                                <input type="file" class="form-control" id="newMediaFile" accept="*/*">
                                <small class="text-muted">Max 50MB. Supported: gambar, video, audio, file</small>
                            </div>
                            <div class="mb-3">
                                <label class="form-label">Caption (untuk media)</label>
                                <input type="text" class="form-control" id="newCaption" placeholder="Caption untuk video/gambar">
                            </div>
                        </form>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Batal</button>
                        <button type="button" class="btn btn-primary" onclick="saveNewCommand()">
                            <i class="fas fa-save"></i> Simpan
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Edit Command Modal -->
        <div class="modal fade" id="editCommandModal" tabindex="-1">
            <div class="modal-dialog modal-lg">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title"><i class="fas fa-edit"></i> Edit Command</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                    </div>
                    <div class="modal-body">
                        <form id="editCommandForm">
                            <div class="row">
                                <div class="col-md-6">
                                    <div class="mb-3">
                                        <label class="form-label">Command</label>
                                        <input type="text" class="form-control" id="editCommand" readonly>
                                    </div>
                                </div>
                                <div class="col-md-6">
                                    <div class="mb-3">
                                        <label class="form-label">Kategori</label>
                                        <select class="form-control" id="editCategory">
                                            <option value="injec">💉 Injec</option>
                                            <option value="pembelajaran">📚 Pembelajaran</option>
                                            <option value="informasi">ℹ️ Informasi</option>
                                            <option value="tools">🛠️ Tools</option>
                                        </select>
                                    </div>
                                </div>
                            </div>
                            <div class="mb-3">
                                <label class="form-label">Judul</label>
                                <input type="text" class="form-control" id="editTitle" required>
                            </div>
                            <div class="mb-3">
                                <label class="form-label">Deskripsi</label>
                                <input type="text" class="form-control" id="editDescription">
                            </div>
                            <div class="mb-3">
                                <label class="form-label">Tipe Response</label>
                                <select class="form-control" id="editResponseType" onchange="toggleEditResponseInputs(this.value)">
                                    <option value="text">📝 Text</option>
                                    <option value="image">🖼️ Gambar</option>
                                    <option value="video">🎥 Video</option>
                                    <option value="audio">🎵 Audio</option>
                                    <option value="sticker">😄 Sticker</option>
                                    <option value="file">📁 File/APK</option>
                                </select>
                            </div>
                            <div id="editTextResponse" class="mb-3">
                                <label class="form-label">Text Content</label>
                                <textarea class="form-control" id="editTextContent" rows="5"></textarea>
                            </div>
                            <div id="editMediaResponse" class="mb-3" style="display:none;">
                                <label class="form-label">Upload File Baru (opsional)</label>
                                <input type="file" class="form-control" id="editMediaFile" accept="*/*">
                            </div>
                            <div class="mb-3">
                                <label class="form-label">Caption</label>
                                <input type="text" class="form-control" id="editCaption">
                            </div>
                            <div class="mb-3">
                                <div class="form-check">
                                    <input class="form-check-input" type="checkbox" id="editIsActive">
                                    <label class="form-check-label" for="editIsActive">Aktif</label>
                                </div>
                            </div>
                        </form>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Batal</button>
                        <button type="button" class="btn btn-primary" onclick="saveEditCommand()">
                            <i class="fas fa-save"></i> Update
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Add Auto Response Modal -->
        <div class="modal fade" id="addAutoResponseModal" tabindex="-1">
            <div class="modal-dialog">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title"><i class="fas fa-plus"></i> Tambah Auto Response</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                    </div>
                    <div class="modal-body">
                        <form id="addAutoResponseForm">
                            <div class="mb-3">
                                <label class="form-label">Keyword *</label>
                                <input type="text" class="form-control" id="newAutoKeyword" placeholder="cape" required>
                                <small class="text-muted">Kata kunci yang akan trigger response</small>
                            </div>
                            <div class="mb-3">
                                <label class="form-label">Tipe Response *</label>
                                <select class="form-control" id="newAutoResponseType" onchange="toggleAutoResponseInputs(this.value)" required>
                                    <option value="text">📝 Text</option>
                                    <option value="sticker">😄 Sticker</option>
                                    <option value="audio">🎵 Audio</option>
                                    <option value="mixed">🎭 Mixed</option>
                                </select>
                            </div>
                            <div id="newAutoTextResponse" class="mb-3">
                                <label class="form-label">Text Response</label>
                                <textarea class="form-control" id="newAutoTextContent" rows="3" placeholder="Response text..."></textarea>
                            </div>
                            <div id="newAutoMediaResponse" class="mb-3" style="display:none;">
                                <label class="form-label">Upload File</label>
                                <input type="file" class="form-control" id="newAutoMediaFile" accept="audio/*,.webp">
                                <small class="text-muted">Audio atau sticker (.webp)</small>
                            </div>
                        </form>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Batal</button>
                        <button type="button" class="btn btn-primary" onclick="saveNewAutoResponse()">
                            <i class="fas fa-save"></i> Simpan
                        </button>
                    </div>
                </div>
            </div>
        </div>
    `;
}

// === FORM HANDLERS ===

function toggleResponseInputs(responseType) {
    const textDiv = document.getElementById('textResponse');
    const mediaDiv = document.getElementById('mediaResponse');
    
    if (responseType === 'text') {
        textDiv.style.display = 'block';
        mediaDiv.style.display = 'none';
    } else {
        textDiv.style.display = 'none';
        mediaDiv.style.display = 'block';
    }
}

function toggleEditResponseInputs(responseType) {
    const textDiv = document.getElementById('editTextResponse');
    const mediaDiv = document.getElementById('editMediaResponse');
    
    if (responseType === 'text') {
        textDiv.style.display = 'block';
        mediaDiv.style.display = 'none';
    } else {
        textDiv.style.display = 'none';
        mediaDiv.style.display = 'block';
    }
}

function toggleAutoResponseInputs(responseType) {
    const textDiv = document.getElementById('newAutoTextResponse');
    const mediaDiv = document.getElementById('newAutoMediaResponse');
    
    if (responseType === 'text') {
        textDiv.style.display = 'block';
        mediaDiv.style.display = 'none';
    } else {
        textDiv.style.display = 'block'; // Mixed juga bisa pakai text
        mediaDiv.style.display = 'block';
    }
}

function saveNewCommand() {
    const command = document.getElementById('newCommand').value;
    const title = document.getElementById('newTitle').value;
    const description = document.getElementById('newDescription').value;
    const responseType = document.getElementById('newResponseType').value;
    const category = document.getElementById('newCategory').value;
    const caption = document.getElementById('newCaption').value;
    
    if (!command || !title) {
        showAlert('warning', 'Command dan title harus diisi');
        return;
    }
    
    let commandData = {
        command: command,
        title: title,
        description: description,
        response_type: responseType,
        category: category,
        caption: caption || null,
        is_active: true
    };
    
    if (responseType === 'text') {
        const textContent = document.getElementById('newTextContent').value;
        if (!textContent) {
            showAlert('warning', 'Text content harus diisi untuk response text');
            return;
        }
        commandData.text_content = textContent;
        
        // Save command directly
        saveCommandData(commandData);
    } else {
        // Upload file first, then save command
        const fileInput = document.getElementById('newMediaFile');
        if (!fileInput.files[0]) {
            showAlert('warning', 'File harus diupload untuk response media');
            return;
        }
        
        uploadFile(fileInput, getFileTypeFromResponseType(responseType), function(filepath) {
            commandData.media_file_path = filepath;
            saveCommandData(commandData);
        });
    }
}

function saveCommandData(commandData) {
    fetch('/api/commands', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(commandData)
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            showAlert('success', 'Command berhasil ditambahkan');
            bootstrap.Modal.getInstance(document.getElementById('addCommandModal')).hide();
            refreshCommands();
            
            // Reset form
            document.getElementById('addCommandForm').reset();
        } else {
            showAlert('danger', 'Gagal menambahkan command');
        }
    })
    .catch(error => {
        console.error('Error saving command:', error);
        showAlert('danger', 'Gagal menambahkan command');
    });
}

function saveEditCommand() {
    const command = document.getElementById('editCommand').value;
    const title = document.getElementById('editTitle').value;
    const description = document.getElementById('editDescription').value;
    const responseType = document.getElementById('editResponseType').value;
    const category = document.getElementById('editCategory').value;
    const caption = document.getElementById('editCaption').value;
    const isActive = document.getElementById('editIsActive').checked;
    
    let commandData = {
        command: command,
        title: title,
        description: description,
        response_type: responseType,
        category: category,
        caption: caption || null,
        is_active: isActive
    };
    
    if (responseType === 'text') {
        commandData.text_content = document.getElementById('editTextContent').value;
    }
    
    const fileInput = document.getElementById('editMediaFile');
    if (fileInput.files[0]) {
        // Upload new file
        uploadFile(fileInput, getFileTypeFromResponseType(responseType), function(filepath) {
            commandData.media_file_path = filepath;
            updateCommandData(commandData);
        });
    } else {
        updateCommandData(commandData);
    }
}

function updateCommandData(commandData) {
    fetch('/api/commands', {
        method: 'PUT',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(commandData)
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            showAlert('success', 'Command berhasil diupdate');
            bootstrap.Modal.getInstance(document.getElementById('editCommandModal')).hide();
            refreshCommands();
        } else {
            showAlert('danger', 'Gagal mengupdate command');
        }
    })
    .catch(error => {
        console.error('Error updating command:', error);
        showAlert('danger', 'Gagal mengupdate command');
    });
}

function saveNewAutoResponse() {
    const keyword = document.getElementById('newAutoKeyword').value;
    const responseType = document.getElementById('newAutoResponseType').value;
    const textContent = document.getElementById('newAutoTextContent').value;
    
    if (!keyword) {
        showAlert('warning', 'Keyword harus diisi');
        return;
    }
    
    let responseData = {
        keyword: keyword,
        response_type: responseType,
        text_response: textContent || null,
        is_active: true
    };
    
    const fileInput = document.getElementById('newAutoMediaFile');
    if (fileInput.files[0]) {
        const fileType = responseType === 'sticker' ? 'stickers' : 'audios';
        uploadFile(fileInput, fileType, function(filepath) {
            if (responseType === 'sticker') {
                responseData.sticker_path = filepath;
            } else {
                responseData.audio_path = filepath;
            }
            saveAutoResponseData(responseData);
        });
    } else {
        saveAutoResponseData(responseData);
    }
}

function saveAutoResponseData(responseData) {
    fetch('/api/autoresponses', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(responseData)
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            showAlert('success', 'Auto response berhasil ditambahkan');
            bootstrap.Modal.getInstance(document.getElementById('addAutoResponseModal')).hide();
            refreshAutoResponses();
            
            // Reset form
            document.getElementById('addAutoResponseForm').reset();
        } else {
            showAlert('danger', 'Gagal menambahkan auto response');
        }
    })
    .catch(error => {
        console.error('Error saving auto response:', error);
        showAlert('danger', 'Gagal menambahkan auto response');
    });
}

function getFileTypeFromResponseType(responseType) {
    const typeMap = {
        'image': 'images',
        'video': 'videos',
        'audio': 'audios',
        'sticker': 'stickers',
        'file': 'files'
    };
    return typeMap[responseType] || 'files';
}