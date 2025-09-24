// Package web - Dashboard server untuk manage learning bot
package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/nabilulilalbab/promote/database"
	"github.com/nabilulilalbab/promote/utils"
)

// DashboardServer manages the web dashboard
type DashboardServer struct {
	repository     database.Repository
	logger         *utils.Logger
	adminNumbers   []string
	mediaPath      string
	whatsappClient interface{} // WhatsApp client untuk akses grup
}

// NewDashboardServer creates a new dashboard server
func NewDashboardServer(repo database.Repository, logger *utils.Logger, adminNumbers []string) *DashboardServer {
	return &DashboardServer{
		repository:   repo,
		logger:       logger,
		adminNumbers: adminNumbers,
		mediaPath:    "media", // Default media path
	}
}

// SetWhatsAppClient sets the WhatsApp client for group access
func (s *DashboardServer) SetWhatsAppClient(client interface{}) {
	s.whatsappClient = client
}

// StartServer starts the web dashboard server
func (s *DashboardServer) StartServer(port int) error {
	// Setup routes
	http.HandleFunc("/", s.handleDashboard)
	http.HandleFunc("/api/groups", s.handleGroups)
	http.HandleFunc("/api/groups/whatsapp", s.handleWhatsAppGroups)
	http.HandleFunc("/api/commands", s.handleCommands)
	http.HandleFunc("/api/autoresponses", s.handleAutoResponses)
	http.HandleFunc("/api/upload", s.handleUpload)
	http.HandleFunc("/api/stats", s.handleStats)
	
	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))
	
	// Create media directories
	s.createMediaDirectories()
	
	addr := fmt.Sprintf(":%d", port)
	s.logger.Infof("Dashboard server starting on http://localhost%s", addr)
	
	return http.ListenAndServe(addr, nil)
}

// createMediaDirectories creates necessary media directories
func (s *DashboardServer) createMediaDirectories() {
	dirs := []string{
		"media/images",
		"media/videos", 
		"media/audios",
		"media/stickers",
		"media/files",
	}
	
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}
}

// handleDashboard serves the main dashboard page
func (s *DashboardServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	html := `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Bot Pembelajaran Dashboard</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    <style>
        .sidebar { background: #2c3e50; min-height: 100vh; }
        .sidebar .nav-link { color: #ecf0f1; }
        .sidebar .nav-link:hover { background: #34495e; color: white; }
        .sidebar .nav-link.active { background: #3498db; color: white; }
        .content-area { padding: 20px; }
        .card-stats { border-left: 4px solid #3498db; }
        .media-preview { max-width: 200px; max-height: 150px; }
    </style>
</head>
<body>
    <div class="container-fluid">
        <div class="row">
            <!-- Sidebar -->
            <div class="col-md-2 sidebar">
                <div class="p-3">
                    <h4 class="text-white"><i class="fas fa-robot"></i> Bot Dashboard</h4>
                </div>
                <nav class="nav flex-column">
                    <a class="nav-link active" href="#" onclick="showTab('groups')">
                        <i class="fas fa-users"></i> Kelola Grup
                    </a>
                    <a class="nav-link" href="#" onclick="showTab('commands')">
                        <i class="fas fa-terminal"></i> Command
                    </a>
                    <a class="nav-link" href="#" onclick="showTab('autoresponses')">
                        <i class="fas fa-magic"></i> Auto Response
                    </a>
                    <a class="nav-link" href="#" onclick="showTab('stats')">
                        <i class="fas fa-chart-bar"></i> Statistik
                    </a>
                </nav>
            </div>
            
            <!-- Main Content -->
            <div class="col-md-10 content-area">
                <!-- Groups Tab -->
                <div id="groups-tab" class="tab-content">
                    <h2><i class="fas fa-users"></i> Kelola Grup Pembelajaran</h2>
                    <div class="row mb-3">
                        <div class="col-md-12">
                            <button class="btn btn-success" onclick="showWhatsAppGroupsModal()">
                                <i class="fas fa-plus"></i> Tambah dari WhatsApp
                            </button>
                            <button class="btn btn-primary" onclick="refreshGroups()">
                                <i class="fas fa-sync"></i> Refresh
                            </button>
                        </div>
                    </div>
                    <div id="groups-list"></div>
                </div>
                
                <!-- Commands Tab -->
                <div id="commands-tab" class="tab-content" style="display:none;">
                    <h2><i class="fas fa-terminal"></i> Kelola Command Pembelajaran</h2>
                    <div class="row mb-3">
                        <div class="col-md-12">
                            <button class="btn btn-success" onclick="showAddCommandModal()">
                                <i class="fas fa-plus"></i> Tambah Command
                            </button>
                            <button class="btn btn-primary" onclick="refreshCommands()">
                                <i class="fas fa-sync"></i> Refresh
                            </button>
                        </div>
                    </div>
                    <div id="commands-list"></div>
                </div>
                
                <!-- Auto Responses Tab -->
                <div id="autoresponses-tab" class="tab-content" style="display:none;">
                    <h2><i class="fas fa-magic"></i> Kelola Auto Response</h2>
                    <div class="row mb-3">
                        <div class="col-md-12">
                            <button class="btn btn-success" onclick="showAddAutoResponseModal()">
                                <i class="fas fa-plus"></i> Tambah Auto Response
                            </button>
                            <button class="btn btn-primary" onclick="refreshAutoResponses()">
                                <i class="fas fa-sync"></i> Refresh
                            </button>
                        </div>
                    </div>
                    <div id="autoresponses-list"></div>
                </div>
                
                <!-- Stats Tab -->
                <div id="stats-tab" class="tab-content" style="display:none;">
                    <h2><i class="fas fa-chart-bar"></i> Statistik Penggunaan</h2>
                    <div id="stats-content"></div>
                </div>
            </div>
        </div>
    </div>
    
    <!-- Add Command Modal -->
    <div class="modal fade" id="addCommandModal" tabindex="-1">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Tambah Command Baru</h5>
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
                            <select class="form-control" id="newResponseType" onchange="toggleResponseInputs()" required>
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
                            <small class="text-muted">Max 50MB</small>
                        </div>
                        <div class="mb-3">
                            <label class="form-label">Caption (untuk media)</label>
                            <input type="text" class="form-control" id="newCaption" placeholder="Caption untuk video/gambar">
                        </div>
                    </form>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Batal</button>
                    <button type="button" class="btn btn-primary" onclick="saveNewCommand()">Simpan</button>
                </div>
            </div>
        </div>
    </div>

    <!-- Add Auto Response Modal -->
    <div class="modal fade" id="addAutoResponseModal" tabindex="-1">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Tambah Auto Response</h5>
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
                            <select class="form-control" id="newAutoResponseType" onchange="toggleAutoResponseInputs()" required>
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
                    <button type="button" class="btn btn-primary" onclick="saveNewAutoResponse()">Simpan</button>
                </div>
            </div>
        </div>
    </div>

    <!-- Edit Command Modal -->
    <div class="modal fade" id="editCommandModal" tabindex="-1">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Edit Command</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <form id="editCommandForm">
                        <input type="hidden" id="editOriginalCommand">
                        <div class="row">
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">Command *</label>
                                    <input type="text" class="form-control" id="editCommand" required>
                                    <small class="text-muted">Harus dimulai dengan titik (.)</small>
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
                            <label class="form-label">Judul *</label>
                            <input type="text" class="form-control" id="editTitle" required>
                        </div>
                        <div class="mb-3">
                            <label class="form-label">Deskripsi</label>
                            <input type="text" class="form-control" id="editDescription">
                        </div>
                        <div class="mb-3">
                            <label class="form-label">Tipe Response *</label>
                            <select class="form-control" id="editResponseType" onchange="toggleEditResponseInputs()" required>
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
                            <div class="mb-2">
                                <label class="form-label">File Saat Ini</label>
                                <div id="currentMediaInfo" class="text-muted small"></div>
                            </div>
                            <label class="form-label">Upload File Baru (Opsional)</label>
                            <input type="file" class="form-control" id="editMediaFile" accept="*/*">
                            <small class="text-muted">Kosongkan jika tidak ingin mengubah file</small>
                        </div>
                        <div class="mb-3">
                            <label class="form-label">Caption (untuk media)</label>
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
                    <button type="button" class="btn btn-primary" onclick="saveEditCommand()">Update</button>
                </div>
            </div>
        </div>
    </div>

    <!-- Edit Auto Response Modal -->
    <div class="modal fade" id="editAutoResponseModal" tabindex="-1">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Edit Auto Response</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <form id="editAutoResponseForm">
                        <input type="hidden" id="editOriginalKeyword">
                        <div class="mb-3">
                            <label class="form-label">Keyword *</label>
                            <input type="text" class="form-control" id="editAutoKeyword" required>
                        </div>
                        <div class="mb-3">
                            <label class="form-label">Tipe Response *</label>
                            <select class="form-control" id="editAutoResponseType" onchange="toggleEditAutoResponseInputs()" required>
                                <option value="text">📝 Text</option>
                                <option value="sticker">😄 Sticker</option>
                                <option value="audio">🎵 Audio</option>
                                <option value="mixed">🎭 Mixed</option>
                            </select>
                        </div>
                        <div id="editAutoTextResponse" class="mb-3">
                            <label class="form-label">Text Response</label>
                            <textarea class="form-control" id="editAutoTextContent" rows="3"></textarea>
                        </div>
                        <div id="editAutoMediaResponse" class="mb-3" style="display:none;">
                            <div class="mb-2">
                                <label class="form-label">File Saat Ini</label>
                                <div id="currentAutoMediaInfo" class="text-muted small"></div>
                            </div>
                            <label class="form-label">Upload File Baru (Opsional)</label>
                            <input type="file" class="form-control" id="editAutoMediaFile" accept="audio/*,.webp">
                            <small class="text-muted">Kosongkan jika tidak ingin mengubah file</small>
                        </div>
                        <div class="mb-3">
                            <div class="form-check">
                                <input class="form-check-input" type="checkbox" id="editAutoIsActive">
                                <label class="form-check-label" for="editAutoIsActive">Aktif</label>
                            </div>
                        </div>
                    </form>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Batal</button>
                    <button type="button" class="btn btn-primary" onclick="saveEditAutoResponse()">Update</button>
                </div>
            </div>
        </div>
    </div>

    <!-- WhatsApp Groups Modal -->
    <div class="modal fade" id="whatsappGroupsModal" tabindex="-1">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Pilih Grup dari WhatsApp</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <div id="whatsapp-groups-loading" class="text-center">
                        <div class="spinner-border" role="status">
                            <span class="visually-hidden">Loading...</span>
                        </div>
                        <p>Mengambil daftar grup dari WhatsApp...</p>
                    </div>
                    <div id="whatsapp-groups-list" style="display:none;">
                        <div class="table-responsive">
                            <table class="table table-hover">
                                <thead>
                                    <tr>
                                        <th>Nama Grup</th>
                                        <th>Member</th>
                                        <th>Status</th>
                                        <th>Aksi</th>
                                    </tr>
                                </thead>
                                <tbody id="whatsapp-groups-tbody">
                                </tbody>
                            </table>
                        </div>
                    </div>
                    <div id="whatsapp-groups-error" style="display:none;" class="alert alert-danger">
                        <i class="fas fa-exclamation-triangle"></i>
                        <span id="error-message">Gagal mengambil daftar grup</span>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Tutup</button>
                </div>
            </div>
        </div>
    </div>
    
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        let currentGroups = [];
        let currentCommands = [];
        let currentAutoResponses = [];

        document.addEventListener('DOMContentLoaded', function() {
            showTab('groups');
            refreshGroups();
        });

        function showTab(tabName) {
            const tabs = document.querySelectorAll('.tab-content');
            tabs.forEach(tab => tab.style.display = 'none');
            
            const navLinks = document.querySelectorAll('.nav-link');
            navLinks.forEach(link => link.classList.remove('active'));
            
            document.getElementById(tabName + '-tab').style.display = 'block';
            
            if (event && event.target) {
                event.target.classList.add('active');
            }
            
            switch(tabName) {
                case 'groups': refreshGroups(); break;
                case 'commands': refreshCommands(); break;
                case 'autoresponses': refreshAutoResponses(); break;
                case 'stats': refreshStats(); break;
            }
        }

        function refreshGroups() {
            fetch('/api/groups')
                .then(response => response.json())
                .then(data => {
                    currentGroups = data || [];
                    displayGroups();
                })
                .catch(error => {
                    console.error('Error:', error);
                    showAlert('danger', 'Gagal memuat data grup');
                });
        }

        function displayGroups() {
            const container = document.getElementById('groups-list');
            if (currentGroups.length === 0) {
                container.innerHTML = '<div class="alert alert-info">Belum ada grup. Gunakan .addgroup di chat personal.</div>';
                return;
            }
            
            let html = '<div class="row">';
            currentGroups.forEach(group => {
                const status = group.is_active ? 'Aktif' : 'Tidak Aktif';
                const badge = group.is_active ? 'bg-success' : 'bg-secondary';
                
                html += '<div class="col-md-6 mb-3"><div class="card"><div class="card-body">';
                html += '<h6>' + (group.group_name || 'Tanpa Nama') + ' <span class="badge ' + badge + '">' + status + '</span></h6>';
                html += '<p class="small text-muted">JID: ' + group.group_jid + '</p>';
                html += '<div class="mt-2">';
                html += '<button class="btn btn-sm btn-danger" onclick="removeLearningGroup(\'' + group.group_jid + '\', \'' + (group.group_name || 'Tanpa Nama') + '\')">Hapus</button>';
                html += '</div>';
                html += '</div></div></div>';
            });
            html += '</div>';
            
            container.innerHTML = html;
        }

        function refreshCommands() {
            fetch('/api/commands')
                .then(response => response.json())
                .then(data => {
                    currentCommands = data || [];
                    displayCommands();
                })
                .catch(error => showAlert('danger', 'Gagal memuat commands'));
        }

        function displayCommands() {
            const container = document.getElementById('commands-list');
            if (currentCommands.length === 0) {
                container.innerHTML = '<div class="alert alert-info">Belum ada command.</div>';
                return;
            }
            
            let html = '<table class="table table-striped"><thead><tr><th>Command</th><th>Judul</th><th>Tipe</th><th>Status</th><th>Aksi</th></tr></thead><tbody>';
            
            currentCommands.forEach(cmd => {
                const status = cmd.is_active ? 'Aktif' : 'Tidak Aktif';
                const badge = cmd.is_active ? 'bg-success' : 'bg-secondary';
                
                html += '<tr>';
                html += '<td><code>' + cmd.command + '</code></td>';
                html += '<td>' + cmd.title + '</td>';
                html += '<td>' + cmd.response_type + '</td>';
                html += '<td><span class="badge ' + badge + '">' + status + '</span></td>';
                html += '<td>';
                html += '<button class="btn btn-sm btn-primary me-1" onclick="editCommand(\'' + cmd.command + '\')">Edit</button>';
                html += '<button class="btn btn-sm btn-danger" onclick="deleteCommand(\'' + cmd.command + '\')">Hapus</button>';
                html += '</td>';
                html += '</tr>';
            });
            
            html += '</tbody></table>';
            container.innerHTML = html;
        }

        function refreshAutoResponses() {
            fetch('/api/autoresponses')
                .then(response => response.json())
                .then(data => {
                    currentAutoResponses = data || [];
                    displayAutoResponses();
                })
                .catch(error => showAlert('danger', 'Gagal memuat auto responses'));
        }

        function displayAutoResponses() {
            const container = document.getElementById('autoresponses-list');
            if (currentAutoResponses.length === 0) {
                container.innerHTML = '<div class="alert alert-info">Belum ada auto response.</div>';
                return;
            }
            
            let html = '<table class="table table-striped"><thead><tr><th>Keyword</th><th>Tipe</th><th>Status</th><th>Aksi</th></tr></thead><tbody>';
            
            currentAutoResponses.forEach(resp => {
                const status = resp.is_active ? 'Aktif' : 'Tidak Aktif';
                const badge = resp.is_active ? 'bg-success' : 'bg-secondary';
                
                html += '<tr>';
                html += '<td><code>' + resp.keyword + '</code></td>';
                html += '<td>' + resp.response_type + '</td>';
                html += '<td><span class="badge ' + badge + '">' + status + '</span></td>';
                html += '<td>';
                html += '<button class="btn btn-sm btn-primary me-1" onclick="editAutoResponse(\'' + resp.keyword + '\')">Edit</button>';
                html += '<button class="btn btn-sm btn-danger" onclick="deleteAutoResponse(\'' + resp.keyword + '\')">Hapus</button>';
                html += '</td>';
                html += '</tr>';
            });
            
            html += '</tbody></table>';
            container.innerHTML = html;
        }

        function refreshStats() {
            fetch('/api/stats?days=7')
                .then(response => response.json())
                .then(data => displayStats(data))
                .catch(error => showAlert('danger', 'Gagal memuat statistik'));
        }

        function displayStats(data) {
            const container = document.getElementById('stats-content');
            
            let html = '<div class="row mb-4">';
            html += '<div class="col-md-3"><div class="card"><div class="card-body text-center">';
            html += '<h3 class="text-primary">' + (data.counts ? data.counts.groups : 0) + '</h3>';
            html += '<p class="mb-0">Total Grup</p></div></div></div>';
            
            html += '<div class="col-md-3"><div class="card"><div class="card-body text-center">';
            html += '<h3 class="text-success">' + (data.counts ? data.counts.commands : 0) + '</h3>';
            html += '<p class="mb-0">Total Command</p></div></div></div>';
            
            html += '<div class="col-md-3"><div class="card"><div class="card-body text-center">';
            html += '<h3 class="text-info">' + (data.counts ? data.counts.auto_responses : 0) + '</h3>';
            html += '<p class="mb-0">Auto Response</p></div></div></div>';
            
            html += '<div class="col-md-3"><div class="card"><div class="card-body text-center">';
            html += '<h3 class="text-warning">' + (data.usage_stats ? Object.keys(data.usage_stats).length : 0) + '</h3>';
            html += '<p class="mb-0">Command Aktif</p></div></div></div>';
            html += '</div>';
            
            container.innerHTML = html;
        }

        function showAlert(type, message) {
            const alertDiv = document.createElement('div');
            alertDiv.className = 'alert alert-' + type + ' alert-dismissible fade show';
            alertDiv.innerHTML = message + ' <button type="button" class="btn-close" data-bs-dismiss="alert"></button>';
            
            const contentArea = document.querySelector('.content-area');
            contentArea.insertBefore(alertDiv, contentArea.firstChild);
            
            setTimeout(() => alertDiv.remove(), 5000);
        }

        function formatDate(dateString) {
            if (!dateString) return '-';
            return new Date(dateString).toLocaleDateString('id-ID');
        }

        // Modal Functions
        function showAddCommandModal() {
            document.getElementById('addCommandForm').reset();
            toggleResponseInputs();
            new bootstrap.Modal(document.getElementById('addCommandModal')).show();
        }

        function showAddAutoResponseModal() {
            document.getElementById('addAutoResponseForm').reset();
            toggleAutoResponseInputs();
            new bootstrap.Modal(document.getElementById('addAutoResponseModal')).show();
        }

        // Toggle input visibility
        function toggleResponseInputs() {
            const responseType = document.getElementById('newResponseType').value;
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

        function toggleAutoResponseInputs() {
            const responseType = document.getElementById('newAutoResponseType').value;
            const textDiv = document.getElementById('newAutoTextResponse');
            const mediaDiv = document.getElementById('newAutoMediaResponse');
            
            if (responseType === 'text') {
                textDiv.style.display = 'block';
                mediaDiv.style.display = 'none';
            } else {
                textDiv.style.display = 'block';
                mediaDiv.style.display = 'block';
            }
        }

        // Save Functions
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
            
            if (!command.startsWith('.')) {
                showAlert('warning', 'Command harus dimulai dengan titik (.)');
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
                saveCommandData(commandData);
            } else {
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
                } else {
                    showAlert('danger', 'Gagal menambahkan command: ' + (data.error || 'Unknown error'));
                }
            })
            .catch(error => {
                console.error('Error saving command:', error);
                showAlert('danger', 'Gagal menambahkan command');
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
                if (responseType !== 'text' && !textContent) {
                    showAlert('warning', 'Text response atau file harus diisi');
                    return;
                }
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
                } else {
                    showAlert('danger', 'Gagal menambahkan auto response: ' + (data.error || 'Unknown error'));
                }
            })
            .catch(error => {
                console.error('Error saving auto response:', error);
                showAlert('danger', 'Gagal menambahkan auto response');
            });
        }

        // File upload function
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
                    if (callback) callback(data.filepath);
                } else {
                    showAlert('danger', 'Gagal mengupload file: ' + (data.error || 'Unknown error'));
                }
            })
            .catch(error => {
                console.error('Error uploading file:', error);
                showAlert('danger', 'Gagal mengupload file');
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

        // Edit/Delete Functions
        function editCommand(command) {
            const cmd = currentCommands.find(c => c.command === command);
            if (!cmd) return;
            
            // Fill form dengan data existing
            document.getElementById('editOriginalCommand').value = cmd.command;
            document.getElementById('editCommand').value = cmd.command;
            document.getElementById('editTitle').value = cmd.title;
            document.getElementById('editDescription').value = cmd.description || '';
            document.getElementById('editCategory').value = cmd.category || 'informasi';
            document.getElementById('editResponseType').value = cmd.response_type;
            document.getElementById('editTextContent').value = cmd.text_content || '';
            document.getElementById('editCaption').value = cmd.caption || '';
            document.getElementById('editIsActive').checked = cmd.is_active;
            
            // Show current media info jika ada
            if (cmd.media_file_path) {
                const fileName = cmd.media_file_path.split('/').pop();
                document.getElementById('currentMediaInfo').innerHTML = 
                    '<i class="fas fa-file"></i> ' + fileName;
            } else {
                document.getElementById('currentMediaInfo').innerHTML = 'Tidak ada file';
            }
            
            // Toggle input visibility
            toggleEditResponseInputs();
            
            // Show modal
            new bootstrap.Modal(document.getElementById('editCommandModal')).show();
        }

        function toggleEditResponseInputs() {
            const responseType = document.getElementById('editResponseType').value;
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

        function saveEditCommand() {
            const originalCommand = document.getElementById('editOriginalCommand').value;
            const command = document.getElementById('editCommand').value;
            const title = document.getElementById('editTitle').value;
            const description = document.getElementById('editDescription').value;
            const responseType = document.getElementById('editResponseType').value;
            const category = document.getElementById('editCategory').value;
            const caption = document.getElementById('editCaption').value;
            const isActive = document.getElementById('editIsActive').checked;
            
            if (!command || !title) {
                showAlert('warning', 'Command dan title harus diisi');
                return;
            }
            
            if (!command.startsWith('.')) {
                showAlert('warning', 'Command harus dimulai dengan titik (.)');
                return;
            }
            
            let cmdData = {
                original_command: originalCommand,
                command: command,
                title: title,
                description: description,
                response_type: responseType,
                category: category,
                caption: caption || null,
                is_active: isActive
            };
            
            if (responseType === 'text') {
                cmdData.text_content = document.getElementById('editTextContent').value;
                saveEditCommandData(cmdData);
            } else {
                const fileInput = document.getElementById('editMediaFile');
                if (fileInput.files[0]) {
                    // Upload file baru
                    uploadFile(fileInput, getFileTypeFromResponseType(responseType), function(filepath) {
                        cmdData.media_file_path = filepath;
                        saveEditCommandData(cmdData);
                    });
                } else {
                    // Tidak ada file baru, gunakan yang lama
                    saveEditCommandData(cmdData);
                }
            }
        }

        function saveEditCommandData(cmdData) {
            fetch('/api/commands', {
                method: 'PUT',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(cmdData)
            })
            .then(response => response.json())
            .then(data => {
                if (data.status === 'success') {
                    showAlert('success', 'Command berhasil diupdate');
                    bootstrap.Modal.getInstance(document.getElementById('editCommandModal')).hide();
                    refreshCommands();
                } else {
                    showAlert('danger', 'Gagal mengupdate command: ' + (data.error || 'Unknown error'));
                }
            })
            .catch(error => {
                console.error('Error updating command:', error);
                showAlert('danger', 'Gagal mengupdate command');
            });
        }

        function deleteCommand(command) {
            if (!confirm('Hapus command "' + command + '"?')) return;
            
            fetch('/api/commands?command=' + encodeURIComponent(command), {
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

        function editAutoResponse(keyword) {
            const resp = currentAutoResponses.find(r => r.keyword === keyword);
            if (!resp) return;
            
            // Fill form dengan data existing
            document.getElementById('editOriginalKeyword').value = resp.keyword;
            document.getElementById('editAutoKeyword').value = resp.keyword;
            document.getElementById('editAutoResponseType').value = resp.response_type;
            document.getElementById('editAutoTextContent').value = resp.text_response || '';
            document.getElementById('editAutoIsActive').checked = resp.is_active;
            
            // Show current media info jika ada
            let mediaInfo = '';
            if (resp.sticker_path) {
                const fileName = resp.sticker_path.split('/').pop();
                mediaInfo += '<i class="fas fa-smile"></i> Sticker: ' + fileName + '<br>';
            }
            if (resp.audio_path) {
                const fileName = resp.audio_path.split('/').pop();
                mediaInfo += '<i class="fas fa-music"></i> Audio: ' + fileName;
            }
            if (!mediaInfo) {
                mediaInfo = 'Tidak ada file media';
            }
            document.getElementById('currentAutoMediaInfo').innerHTML = mediaInfo;
            
            // Toggle input visibility
            toggleEditAutoResponseInputs();
            
            // Show modal
            new bootstrap.Modal(document.getElementById('editAutoResponseModal')).show();
        }

        function toggleEditAutoResponseInputs() {
            const responseType = document.getElementById('editAutoResponseType').value;
            const textDiv = document.getElementById('editAutoTextResponse');
            const mediaDiv = document.getElementById('editAutoMediaResponse');
            
            if (responseType === 'text') {
                textDiv.style.display = 'block';
                mediaDiv.style.display = 'none';
            } else {
                textDiv.style.display = 'block';
                mediaDiv.style.display = 'block';
            }
        }

        function saveEditAutoResponse() {
            const originalKeyword = document.getElementById('editOriginalKeyword').value;
            const keyword = document.getElementById('editAutoKeyword').value;
            const responseType = document.getElementById('editAutoResponseType').value;
            const textContent = document.getElementById('editAutoTextContent').value;
            const isActive = document.getElementById('editAutoIsActive').checked;
            
            if (!keyword) {
                showAlert('warning', 'Keyword harus diisi');
                return;
            }
            
            let respData = {
                original_keyword: originalKeyword,
                keyword: keyword,
                response_type: responseType,
                text_response: textContent || null,
                is_active: isActive
            };
            
            const fileInput = document.getElementById('editAutoMediaFile');
            if (fileInput.files[0]) {
                // Upload file baru
                const fileType = responseType === 'sticker' ? 'stickers' : 'audios';
                uploadFile(fileInput, fileType, function(filepath) {
                    if (responseType === 'sticker') {
                        respData.sticker_path = filepath;
                    } else {
                        respData.audio_path = filepath;
                    }
                    saveEditAutoResponseData(respData);
                });
            } else {
                // Tidak ada file baru, gunakan yang lama
                saveEditAutoResponseData(respData);
            }
        }

        function saveEditAutoResponseData(respData) {
            fetch('/api/autoresponses', {
                method: 'PUT',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(respData)
            })
            .then(response => response.json())
            .then(data => {
                if (data.status === 'success') {
                    showAlert('success', 'Auto response berhasil diupdate');
                    bootstrap.Modal.getInstance(document.getElementById('editAutoResponseModal')).hide();
                    refreshAutoResponses();
                } else {
                    showAlert('danger', 'Gagal mengupdate auto response: ' + (data.error || 'Unknown error'));
                }
            })
            .catch(error => {
                console.error('Error updating auto response:', error);
                showAlert('danger', 'Gagal mengupdate auto response');
            });
        }

        // WhatsApp Groups Functions
        function showWhatsAppGroupsModal() {
            // Reset modal state
            document.getElementById('whatsapp-groups-loading').style.display = 'block';
            document.getElementById('whatsapp-groups-list').style.display = 'none';
            document.getElementById('whatsapp-groups-error').style.display = 'none';
            
            // Show modal
            new bootstrap.Modal(document.getElementById('whatsappGroupsModal')).show();
            
            // Fetch WhatsApp groups
            fetch('/api/groups/whatsapp')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('whatsapp-groups-loading').style.display = 'none';
                    if (data.status === 'success') {
                        displayWhatsAppGroups(data.groups || []);
                        document.getElementById('whatsapp-groups-list').style.display = 'block';
                    } else {
                        document.getElementById('error-message').textContent = data.error || 'Gagal mengambil daftar grup';
                        document.getElementById('whatsapp-groups-error').style.display = 'block';
                    }
                })
                .catch(error => {
                    document.getElementById('whatsapp-groups-loading').style.display = 'none';
                    document.getElementById('error-message').textContent = 'Error: ' + error.message;
                    document.getElementById('whatsapp-groups-error').style.display = 'block';
                    console.error('Error fetching WhatsApp groups:', error);
                });
        }

        function displayWhatsAppGroups(whatsappGroups) {
            const tbody = document.getElementById('whatsapp-groups-tbody');
            if (whatsappGroups.length === 0) {
                tbody.innerHTML = '<tr><td colspan="4" class="text-center">Tidak ada grup WhatsApp yang ditemukan</td></tr>';
                return;
            }

            let html = '';
            whatsappGroups.forEach(group => {
                // Check if group is already added
                const isAdded = currentGroups.some(lg => lg.group_jid === group.jid);
                const statusBadge = isAdded ? 
                    '<span class="badge bg-success">Sudah Ditambahkan</span>' : 
                    '<span class="badge bg-secondary">Belum Ditambahkan</span>';
                
                const actionButton = isAdded ? 
                    '<button class="btn btn-sm btn-warning" onclick="removeLearningGroup(\'' + group.jid + '\', \'' + group.name + '\')">Hapus</button>' :
                    '<button class="btn btn-sm btn-success" onclick="addGroupFromWhatsApp(\'' + group.jid + '\', \'' + group.name + '\')">Tambah</button>';

                html += '<tr>';
                html += '<td><strong>' + (group.name || 'Tanpa Nama') + '</strong></td>';
                html += '<td>' + (group.participant_count || 0) + ' member</td>';
                html += '<td>' + statusBadge + '</td>';
                html += '<td>' + actionButton + '</td>';
                html += '</tr>';
            });
            
            tbody.innerHTML = html;
        }

        function addGroupFromWhatsApp(jid, name) {
            const groupData = {
                group_jid: jid,
                group_name: name,
                is_active: true,
                description: 'Ditambahkan dari WhatsApp via dashboard',
                created_by: 'admin'
            };

            fetch('/api/groups', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(groupData)
            })
            .then(response => response.json())
            .then(data => {
                if (data.status === 'success') {
                    showAlert('success', 'Grup "' + name + '" berhasil ditambahkan');
                    refreshGroups();
                    // Refresh modal content
                    showWhatsAppGroupsModal();
                } else {
                    showAlert('danger', 'Gagal menambahkan grup: ' + (data.error || 'Unknown error'));
                }
            })
            .catch(error => {
                console.error('Error adding group:', error);
                showAlert('danger', 'Gagal menambahkan grup');
            });
        }

        function removeLearningGroup(jid, name) {
            if (!confirm('Hapus grup "' + name + '" dari daftar pembelajaran?')) return;
            
            fetch('/api/groups?jid=' + encodeURIComponent(jid), {
                method: 'DELETE'
            })
            .then(response => response.json())
            .then(data => {
                if (data.status === 'success') {
                    showAlert('success', 'Grup "' + name + '" berhasil dihapus');
                    refreshGroups();
                    // Refresh modal if open
                    if (document.getElementById('whatsappGroupsModal').classList.contains('show')) {
                        showWhatsAppGroupsModal();
                    }
                } else {
                    showAlert('danger', 'Gagal menghapus grup');
                }
            })
            .catch(error => {
                console.error('Error removing group:', error);
                showAlert('danger', 'Gagal menghapus grup');
            });
        }

        function deleteAutoResponse(keyword) {
            if (!confirm('Hapus auto response "' + keyword + '"?')) return;
            
            fetch('/api/autoresponses?keyword=' + encodeURIComponent(keyword), {
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
    </script>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// === API HANDLERS ===

// handleGroups handles group management API
func (s *DashboardServer) handleGroups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getGroups(w, r)
	case "POST":
		s.createGroup(w, r)
	case "PUT":
		s.updateGroup(w, r)
	case "DELETE":
		s.deleteGroup(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getGroups returns all learning groups
func (s *DashboardServer) getGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := s.repository.GetAllLearningGroups()
	if err != nil {
		s.logger.Errorf("Failed to get groups: %v", err)
		http.Error(w, "Failed to get groups", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

// createGroup creates a new learning group
func (s *DashboardServer) createGroup(w http.ResponseWriter, r *http.Request) {
	var group database.LearningGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if err := s.repository.CreateLearningGroup(&group); err != nil {
		s.logger.Errorf("Failed to create group: %v", err)
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// updateGroup updates a learning group
func (s *DashboardServer) updateGroup(w http.ResponseWriter, r *http.Request) {
	var group database.LearningGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if err := s.repository.UpdateLearningGroup(&group); err != nil {
		s.logger.Errorf("Failed to update group: %v", err)
		http.Error(w, "Failed to update group", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// deleteGroup deletes a learning group
func (s *DashboardServer) deleteGroup(w http.ResponseWriter, r *http.Request) {
	groupJID := r.URL.Query().Get("jid")
	if groupJID == "" {
		http.Error(w, "Group JID required", http.StatusBadRequest)
		return
	}
	
	if err := s.repository.DeleteLearningGroup(groupJID); err != nil {
		s.logger.Errorf("Failed to delete group: %v", err)
		http.Error(w, "Failed to delete group", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleCommands handles command management API
func (s *DashboardServer) handleCommands(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getCommands(w, r)
	case "POST":
		s.createCommand(w, r)
	case "PUT":
		s.updateCommand(w, r)
	case "DELETE":
		s.deleteCommand(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getCommands returns all learning commands
func (s *DashboardServer) getCommands(w http.ResponseWriter, r *http.Request) {
	commands, err := s.repository.GetAllLearningCommands()
	if err != nil {
		s.logger.Errorf("Failed to get commands: %v", err)
		http.Error(w, "Failed to get commands", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commands)
}

// createCommand creates a new learning command
func (s *DashboardServer) createCommand(w http.ResponseWriter, r *http.Request) {
	var cmd database.LearningCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Set default values
	cmd.IsActive = true
	cmd.CreatedBy = "admin"
	
	if err := s.repository.CreateLearningCommand(&cmd); err != nil {
		s.logger.Errorf("Failed to create command: %v", err)
		http.Error(w, "Failed to create command", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// updateCommand updates a learning command
func (s *DashboardServer) updateCommand(w http.ResponseWriter, r *http.Request) {
	var reqData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Get original command to update
	originalCommand, ok := reqData["original_command"].(string)
	if !ok || originalCommand == "" {
		originalCommand = reqData["command"].(string) // fallback untuk backward compatibility
	}
	
	// Get existing command
	existingCmd, err := s.repository.GetLearningCommand(originalCommand)
	if err != nil || existingCmd == nil {
		http.Error(w, "Command not found", http.StatusNotFound)
		return
	}
	
	// Update fields
	if cmd, ok := reqData["command"].(string); ok {
		existingCmd.Command = cmd
	}
	if title, ok := reqData["title"].(string); ok {
		existingCmd.Title = title
	}
	if desc, ok := reqData["description"].(string); ok {
		existingCmd.Description = desc
	}
	if respType, ok := reqData["response_type"].(string); ok {
		existingCmd.ResponseType = respType
	}
	if category, ok := reqData["category"].(string); ok {
		existingCmd.Category = category
	}
	if caption, ok := reqData["caption"].(string); ok {
		existingCmd.Caption = &caption
	}
	if isActive, ok := reqData["is_active"].(bool); ok {
		existingCmd.IsActive = isActive
	}
	if textContent, ok := reqData["text_content"].(string); ok {
		existingCmd.TextContent = &textContent
	}
	if mediaPath, ok := reqData["media_file_path"].(string); ok {
		existingCmd.MediaFilePath = &mediaPath
	}
	
	// Jika command berubah, hapus yang lama dan buat yang baru
	if originalCommand != existingCmd.Command {
		// Delete old command
		if err := s.repository.DeleteLearningCommand(originalCommand); err != nil {
			s.logger.Errorf("Failed to delete old command: %v", err)
			http.Error(w, "Failed to update command", http.StatusInternalServerError)
			return
		}
		// Create new command
		if err := s.repository.CreateLearningCommand(existingCmd); err != nil {
			s.logger.Errorf("Failed to create new command: %v", err)
			http.Error(w, "Failed to update command", http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing command
		if err := s.repository.UpdateLearningCommand(existingCmd); err != nil {
			s.logger.Errorf("Failed to update command: %v", err)
			http.Error(w, "Failed to update command", http.StatusInternalServerError)
			return
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// deleteCommand deletes a learning command
func (s *DashboardServer) deleteCommand(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Query().Get("command")
	if command == "" {
		http.Error(w, "Command required", http.StatusBadRequest)
		return
	}
	
	if err := s.repository.DeleteLearningCommand(command); err != nil {
		s.logger.Errorf("Failed to delete command: %v", err)
		http.Error(w, "Failed to delete command", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}