// Package web - Dashboard server untuk manage learning bot
package web

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	waLog "go.mau.fi/whatsmeow/util/log"

	"github.com/nabilulilalbab/promote/database"
	"github.com/nabilulilalbab/promote/utils"
)

// DashboardServer manages the web dashboard
type DashboardServer struct {
	repository     database.Repository
	logger         *utils.Logger
	adminNumbers   []string
	mediaPath      string
	whatsappClient *whatsmeow.Client // WhatsApp client untuk akses grup
	waManager      *utils.WAManager  // Enhanced WhatsApp Manager
	qrGenerator    *utils.QRCodeGenerator
	dashboardQR    *utils.DashboardQRHandler

	// QR and Pairing code state
	currentQRCode      string
	currentPairingCode string
	qrMutex            sync.RWMutex
	pairingMutex       sync.RWMutex

	// Callback functions for dynamic DB swap
	onReloadSessionDB func() error
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
func (s *DashboardServer) SetWhatsAppClient(client *whatsmeow.Client) {
	s.whatsappClient = client
	// Setup dashboard QR handler when both client and QR generator available
	if s.qrGenerator != nil {
		s.dashboardQR = utils.NewDashboardQRHandler(client, s.logger, s.qrGenerator)
	}
}

// SetWAManager sets the enhanced WhatsApp manager for pairing
func (s *DashboardServer) SetWAManager(waManager *utils.WAManager) {
	s.waManager = waManager
}

// SetQRGenerator sets the QR code generator
func (s *DashboardServer) SetQRGenerator(qrGen *utils.QRCodeGenerator) {
	s.qrGenerator = qrGen
	// Setup dashboard QR handler when both client and QR generator available
	if s.whatsappClient != nil {
		s.dashboardQR = utils.NewDashboardQRHandler(s.whatsappClient, s.logger, qrGen)
	}
}

// SetOnReloadSessionDB sets the callback function to handle WA session reload
func (s *DashboardServer) SetOnReloadSessionDB(callback func() error) {
	s.onReloadSessionDB = callback
}

// SetRepository updates the dashboard's repository dynamically
func (s *DashboardServer) SetRepository(repo database.Repository) {
	s.repository = repo
}

// StartServer starts the web dashboard server
func (s *DashboardServer) StartServer(port int) error {
	// Setup routes
	http.HandleFunc("/", s.handleDashboard)
	http.HandleFunc("/api/groups", s.handleGroups)
	http.HandleFunc("/api/groups/whatsapp", s.handleWhatsAppGroups)
	http.HandleFunc("/api/commands", s.handleCommands)
	http.HandleFunc("/api/autoresponses", s.handleAutoResponses)
	http.HandleFunc("/api/forbidden_words", s.handleForbiddenWords)
	http.HandleFunc("/api/upload", s.handleUpload)
	http.HandleFunc("/api/stats", s.handleStats)
	http.HandleFunc("/api/xray_converters", s.handleXRayConverters)
	http.HandleFunc("/api/xray_converters/test", s.handleXRayConverterTest)

	// WhatsApp Pairing endpoints
	http.HandleFunc("/api/whatsapp/status", s.handleWhatsAppStatus)
	http.HandleFunc("/api/whatsapp/qr", s.handleQRPairing)
	http.HandleFunc("/api/whatsapp/qr/cancel", s.handleQRCancel)
	http.HandleFunc("/api/whatsapp/qr-image", s.handleQRImage)
	http.HandleFunc("/api/whatsapp/phone", s.handlePhonePairing)
	http.HandleFunc("/api/whatsapp/disconnect", s.handleDisconnect)
	http.HandleFunc("/api/whatsapp/reconnect", s.handleReconnect)
	http.HandleFunc("/api/whatsapp/logout", s.handleLogout)
	http.HandleFunc("/api/whatsapp/pairing-code", s.handleGetPairingCode)
	http.HandleFunc("/api/whatsapp/full_reset", s.handleFullReset)

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
                    <a class="nav-link" href="#" onclick="showTab('autoremove')">
                        <i class="fas fa-trash-alt"></i> Auto Remove
                    </a>
                    <a class="nav-link" href="#" onclick="showTab('stats')">
                        <i class="fas fa-chart-bar"></i> Statistik
                    </a>
                    <a class="nav-link" href="#" onclick="showTab('whatsapp')">
                        <i class="fab fa-whatsapp"></i> WhatsApp Pairing
                    </a>
                    <a class="nav-link" href="#" onclick="showTab('xray-tab')">
                        <i class="fas fa-exchange-alt"></i> XRay Converter
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

                <!-- Auto Remove Tab -->
                <div id="autoremove-tab" class="tab-content" style="display:none;">
                    <h2><i class="fas fa-trash-alt"></i> Kelola Auto Remove Chat</h2>
                    <div id="autoremove-group-list"></div>
                </div>
                
                <!-- WhatsApp Pairing Tab -->
                <div id="whatsapp" class="tab-content" style="display: none;">
                    <h3><i class="fab fa-whatsapp text-success"></i> WhatsApp Pairing Management</h3>
                    
                    <!-- Status Card -->
                    <div class="row mb-4">
                        <div class="col-12">
                            <div class="card">
                                <div class="card-header bg-success text-white">
                                    <h5><i class="fas fa-info-circle"></i> Status Koneksi WhatsApp</h5>
                                </div>
                                <div class="card-body">
                                    <div class="row" id="whatsapp-status">
                                        <div class="col-md-3">
                                            <div class="text-center">
                                                <div class="status-indicator mb-2" id="connection-status">
                                                    <i class="fas fa-circle text-secondary fa-2x"></i>
                                                </div>
                                                <small>Koneksi</small>
                                            </div>
                                        </div>
                                        <div class="col-md-3">
                                            <div class="text-center">
                                                <div class="status-indicator mb-2" id="login-status">
                                                    <i class="fas fa-circle text-secondary fa-2x"></i>
                                                </div>
                                                <small>Login Status</small>
                                            </div>
                                        </div>
                                        <div class="col-md-6">
                                            <h6>Informasi Device:</h6>
                                            <p class="mb-1"><strong>Nomor:</strong> <span id="phone-number">-</span></p>
                                            <p class="mb-1"><strong>Device:</strong> <span id="device-name">-</span></p>
                                            <button class="btn btn-outline-primary btn-sm" onclick="refreshWhatsAppStatus()">
                                                <i class="fas fa-sync-alt"></i> Refresh Status
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Pairing Methods -->
                    <div class="row">
                        <!-- QR Code Pairing -->
                        <div class="col-md-6 mb-4">
                            <div class="card h-100">
                                <div class="card-header bg-primary text-white">
                                    <h5><i class="fas fa-qrcode"></i> QR Code Pairing</h5>
                                </div>
                                <div class="card-body text-center">
                                    <p>Scan QR code dengan WhatsApp untuk pairing cepat</p>
                                    
                                    <!-- QR Code Display Area -->
                                    <div class="mb-3" id="qr-display-area">
                                        <div id="qr-placeholder" style="display: block;">
                                            <i class="fas fa-qrcode fa-4x text-primary"></i>
                                            <p class="mt-2 text-muted">QR code akan muncul di sini</p>
                                        </div>
                                        <div id="qr-image-container" style="display: none;">
                                            <img id="qr-image" src="" alt="QR Code" class="img-fluid" style="max-width: 300px; max-height: 300px;">
                                            <div class="mt-2">
                                                <small class="text-success">
                                                    <i class="fas fa-mobile-alt"></i> Scan dengan WhatsApp Anda
                                                </small>
                                            </div>
                                        </div>
                                        <div id="qr-loading" style="display: none;">
                                            <div class="spinner-border text-primary" role="status">
                                                <span class="visually-hidden">Loading...</span>
                                            </div>
                                            <p class="mt-2 text-muted">Generating QR code...</p>
                                        </div>
                                    </div>
                                    
                                    <button class="btn btn-primary btn-lg" onclick="startQRPairing()" id="qr-pairing-btn">
                                        <i class="fas fa-qrcode"></i> Start QR Pairing
                                    </button>
                                    
                                    <button class="btn btn-secondary ms-2" onclick="refreshQRCode()" id="qr-refresh-btn" style="display: none;">
                                        <i class="fas fa-sync-alt"></i> Refresh QR
                                    </button>
                                    <button class="btn btn-outline-danger ms-2" onclick="cancelQRPairing()" id="qr-cancel-btn">
                                        <i class="fas fa-ban"></i> Cancel
                                    </button>
                                </div>
                            </div>
                        </div>

                        <!-- Phone Number Pairing -->
                        <div class="col-md-6 mb-4">
                            <div class="card h-100">
                                <div class="card-header bg-success text-white">
                                    <h5><i class="fas fa-phone"></i> Phone Number Pairing</h5>
                                </div>
                                <div class="card-body">
                                    <p>Gunakan nomor telepon untuk mendapatkan pairing code</p>
                                    <div class="mb-3">
                                        <label class="form-label">Nomor Telepon:</label>
                                        <input type="tel" class="form-control" id="phone-input" 
                                               placeholder="Contoh: 6281234567890" maxlength="15">
                                        <small class="text-muted">Format: Kode negara + nomor (tanpa +)</small>
                                    </div>
                                    <button class="btn btn-success btn-lg w-100" onclick="startPhonePairing()" id="phone-pairing-btn">
                                        <i class="fas fa-phone"></i> Get Pairing Code
                                    </button>
                                    
                                    <!-- Pairing Code Display Area -->
                                    <div class="mt-3" id="pairing-code-area" style="display: none;">
                                        <div class="alert alert-success text-center">
                                            <h4><i class="fas fa-key"></i> Pairing Code:</h4>
                                            <h1 id="pairing-code-display" class="font-monospace"></h1>
                                            <small>Masukkan code ini di WhatsApp > Linked Devices</small>
                                        </div>
                                    </div>
                                    
                                    <div class="mt-3">
                                        <small class="text-muted">
                                            Pairing code akan muncul di atas setelah nomor diverifikasi
                                        </small>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Connection Controls -->
                    <div class="row">
                        <div class="col-12">
                            <div class="card">
                                <div class="card-header bg-warning text-dark">
                                    <h5><i class="fas fa-cog"></i> Connection Controls</h5>
                                </div>
                                <div class="card-body">
                                    <div class="btn-group me-2" role="group">
                                        <button class="btn btn-outline-primary" onclick="reconnectWhatsApp()">
                                            <i class="fas fa-sync-alt"></i> Reconnect
                                        </button>
                                        <button class="btn btn-outline-warning" onclick="disconnectWhatsApp()">
                                            <i class="fas fa-unlink"></i> Disconnect
                                        </button>
                                        <button class="btn btn-outline-danger" onclick="safeLogoutWhatsApp()">
                                            <i class="fas fa-sign-out-alt"></i> Safe Logout
                                        </button>
                                        <button class="btn btn-danger" onclick="fullResetWhatsApp()">
                                            <i class="fas fa-skull-crossbones"></i> Full Reset (Hard)
                                        </button>
                                    </div>
                                    <div class="mt-3">
                                        <small class="text-muted">
                                            <strong>Reconnect:</strong> Hubungkan ulang jika sudah login<br>
                                            <strong>Disconnect:</strong> Putus koneksi WhatsApp<br>
                                            <strong>Safe Logout:</strong> Logout aman dari WhatsApp (perlu QR lagi)
                                        </small>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Instructions -->
                    <div class="row mt-4">
                        <div class="col-12">
                            <div class="card">
                                <div class="card-header bg-info text-white">
                                    <h5><i class="fas fa-info-circle"></i> Petunjuk Penggunaan</h5>
                                </div>
                                <div class="card-body">
                                    <h6><i class="fas fa-qrcode"></i> QR Code Pairing:</h6>
                                    <ol>
                                        <li>Klik "Start QR Pairing"</li>
                                        <li>QR code akan muncul di terminal/console</li>
                                        <li>Buka WhatsApp > Menu > Linked Devices</li>
                                        <li>Scan QR code yang muncul di terminal</li>
                                    </ol>
                                    
                                    <h6 class="mt-3"><i class="fas fa-phone"></i> Phone Number Pairing:</h6>
                                    <ol>
                                        <li>Masukkan nomor telepon (tanpa +, contoh: 6281234567890)</li>
                                        <li>Klik "Get Pairing Code"</li>
                                        <li>Pairing code akan muncul di terminal</li>
                                        <li>Buka WhatsApp > Menu > Linked Devices > Link a Device</li>
                                        <li>Pilih "Link with phone number instead"</li>
                                        <li>Masukkan pairing code yang muncul di terminal</li>
                                    </ol>
                                    
                                    <div class="alert alert-warning mt-3">
                                        <i class="fas fa-exclamation-triangle"></i>
                                        <strong>Penting:</strong> Pastikan terminal/console terlihat untuk melihat QR code atau pairing code!
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Stats Tab -->
                <div id="stats-tab" class="tab-content" style="display:none;">
                    <h2><i class="fas fa-chart-bar"></i> Statistik Bot</h2>
                    <div id="stats-content">
                        <div class="row">
                            <div class="col-md-12">
                                <div class="card">
                                    <div class="card-body">
                                        <h5>Statistik akan dimuat...</h5>
                                        <div class="spinner-border" role="status">
                                            <span class="visually-hidden">Loading...</span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- XRay Converter Tab -->
                <div id="xray-tab" class="tab-content" style="display:none;">
                    <h2><i class="fas fa-exchange-alt"></i> Kelola XRay Converter</h2>
                    <div class="d-flex justify-content-between align-items-center mb-3">
                        <div>
                            <button class="btn btn-success" data-bs-toggle="modal" data-bs-target="#addXRayConverterModal">
                                <i class="fas fa-plus"></i> Tambah Converter
                            </button>
                            <button class="btn btn-primary" onclick="refreshXRayConverters()">
                                <i class="fas fa-sync"></i> Refresh
                            </button>
                        </div>
                    </div>
                    <div id="xray-converters-list" class="row">
                        <!-- XRay converters will be loaded here -->
                    </div>
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

    <!-- Edit XRay Converter Modal -->
    <div class="modal fade" id="editXRayConverterModal" tabindex="-1">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Edit XRay Converter</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <form id="editXRayConverterForm">
                        <input type="hidden" id="editConverterOriginalCommand">
                        <div class="row">
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">Command Name *</label>
                                    <input type="text" class="form-control" id="editConverterCommand" readonly>
                                    <small class="text-muted">Command name tidak bisa diubah</small>
                                </div>
                            </div>
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">Display Name *</label>
                                    <input type="text" class="form-control" id="editConverterDisplayName" required>
                                </div>
                            </div>
                        </div>
                        <div class="row">
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">Bug Host *</label>
                                    <input type="text" class="form-control" id="editConverterBugHost" required>
                                </div>
                            </div>
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">Modify Type *</label>
                                    <select class="form-control" id="editConverterModifyType" onchange="toggleEditAdvancedSettings()" required>
                                        <option value="wildcard">🌐 Wildcard</option>
                                        <option value="sni">🔐 SNI Only</option>
                                        <option value="ws">📡 WebSocket</option>
                                        <option value="grpc">⚡ gRPC</option>
                                        <option value="custom">🎛️ Custom (Advanced)</option>
                                    </select>
                                </div>
                            </div>
                        </div>
                        
                        <!-- Advanced Template Settings -->
                        <div id="editAdvancedSettings" style="display:none;">
                            <h6 class="text-primary mb-3">🎛️ Advanced Template Settings</h6>
                            <div class="alert alert-info">
                                <strong>Available Placeholders:</strong><br>
                                <code>{bug_host}</code> - Bug host domain<br>
                                <code>{bug_ip}</code> - Bug host IP<br>
                                <code>{original_server}</code> - Original server<br>
                                <code>{original_host}</code> - Original host<br>
                                <code>{original_sni}</code> - Original SNI<br>
                                <small class="text-muted">Leave empty to use original value</small>
                            </div>
                            <div class="row">
                                <div class="col-md-4">
                                    <div class="mb-3">
                                        <label class="form-label">Server Template</label>
                                        <input type="text" class="form-control" id="editConverterServerTemplate">
                                    </div>
                                </div>
                                <div class="col-md-4">
                                    <div class="mb-3">
                                        <label class="form-label">Host Template</label>
                                        <input type="text" class="form-control" id="editConverterHostTemplate">
                                    </div>
                                </div>
                                <div class="col-md-4">
                                    <div class="mb-3">
                                        <label class="form-label">SNI Template</label>
                                        <input type="text" class="form-control" id="editConverterSNITemplate">
                                    </div>
                                </div>
                            </div>
                        </div>
                        
                        <div class="row">
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">Path Template</label>
                                    <input type="text" class="form-control" id="editConverterPathTemplate">
                                    <small class="text-muted">Untuk WS/HTTPUpgrade, kosongkan untuk keep original</small>
                                </div>
                            </div>
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">gRPC Service Name</label>
                                    <input type="text" class="form-control" id="editConverterGrpcService">
                                    <small class="text-muted">Hanya untuk gRPC modify type</small>
                                </div>
                            </div>
                        </div>
                        <div class="row">
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">Port Override</label>
                                    <input type="number" class="form-control" id="editConverterPortOverride">
                                    <small class="text-muted">Kosongkan untuk gunakan port original</small>
                                </div>
                            </div>
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">Status</label>
                                    <select class="form-control" id="editConverterIsActive">
                                        <option value="true">✅ Aktif</option>
                                        <option value="false">❌ Nonaktif</option>
                                    </select>
                                </div>
                            </div>
                        </div>
                    </form>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Batal</button>
                    <button type="button" class="btn btn-primary" onclick="saveEditXRayConverter()">Update</button>
                </div>
            </div>
        </div>
    </div>

    <!-- Add XRay Converter Modal -->
    <div class="modal fade" id="addXRayConverterModal" tabindex="-1">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Tambah XRay Converter</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <form id="addXRayConverterForm">
                        <div class="row">
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">Command Name *</label>
                                    <input type="text" class="form-control" id="newConverterCommand" placeholder="convertbizz" required>
                                    <small class="text-muted">Tanpa titik, contoh: convertbizz</small>
                                </div>
                            </div>
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">Display Name *</label>
                                    <input type="text" class="form-control" id="newConverterDisplayName" placeholder="XL-Line-WC" required>
                                </div>
                            </div>
                        </div>
                        <div class="row">
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">Bug Host *</label>
                                    <input type="text" class="form-control" id="newConverterBugHost" placeholder="ava.game.naver.com" required>
                                </div>
                            </div>
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">Modify Type *</label>
                                    <select class="form-control" id="newConverterModifyType" onchange="toggleAdvancedSettings()" required>
                                        <option value="wildcard">🌐 Wildcard</option>
                                        <option value="sni">🔐 SNI Only</option>
                                        <option value="ws">📡 WebSocket</option>
                                        <option value="grpc">⚡ gRPC</option>
                                        <option value="custom">🎛️ Custom (Advanced)</option>
                                    </select>
                                </div>
                            </div>
                        </div>
                        
                        <!-- Advanced Template Settings -->
                        <div id="advancedSettings" style="display:none;">
                            <h6 class="text-primary mb-3">🎛️ Advanced Template Settings</h6>
                            <div class="alert alert-info">
                                <strong>Available Placeholders:</strong><br>
                                <code>{bug_host}</code> - Bug host domain<br>
                                <code>{bug_ip}</code> - Bug host IP<br>
                                <code>{original_server}</code> - Original server<br>
                                <code>{original_host}</code> - Original host<br>
                                <code>{original_sni}</code> - Original SNI<br>
                                <small class="text-muted">Leave empty to use original value</small>
                            </div>
                            <div class="row">
                                <div class="col-md-4">
                                    <div class="mb-3">
                                        <label class="form-label">Server Template</label>
                                        <input type="text" class="form-control" id="newConverterServerTemplate" placeholder="{bug_host}">
                                        <small class="text-muted">e.g., {bug_host} or {bug_ip}</small>
                                    </div>
                                </div>
                                <div class="col-md-4">
                                    <div class="mb-3">
                                        <label class="form-label">Host Template</label>
                                        <input type="text" class="form-control" id="newConverterHostTemplate" placeholder="{bug_host}.{original_server}">
                                        <small class="text-muted">e.g., {bug_host}.{original_server}</small>
                                    </div>
                                </div>
                                <div class="col-md-4">
                                    <div class="mb-3">
                                        <label class="form-label">SNI Template</label>
                                        <input type="text" class="form-control" id="newConverterSNITemplate" placeholder="{bug_host}.{original_server}">
                                        <small class="text-muted">e.g., {bug_host}.{original_server}</small>
                                    </div>
                                </div>
                            </div>
                        </div>
                        
                        <div class="row">
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">Path Template</label>
                                    <input type="text" class="form-control" id="newConverterPathTemplate" placeholder="/rsv">
                                    <small class="text-muted">Untuk WS/HTTPUpgrade, kosongkan untuk keep original</small>
                                </div>
                            </div>
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label class="form-label">gRPC Service Name</label>
                                    <input type="text" class="form-control" id="newConverterGrpcService" placeholder="vmess-grpc">
                                    <small class="text-muted">Hanya untuk gRPC modify type</small>
                                </div>
                            </div>
                        </div>
                        <div class="mb-3">
                            <label class="form-label">Port Override</label>
                            <input type="number" class="form-control" id="newConverterPortOverride" placeholder="443">
                            <small class="text-muted">Kosongkan untuk gunakan port original</small>
                        </div>
                    </form>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Batal</button>
                    <button type="button" class="btn btn-primary" onclick="saveNewXRayConverter()">Simpan</button>
                </div>
            </div>
        </div>
    </div>
    
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        let currentGroups = [];
        let currentCommands = [];
        let currentAutoResponses = [];
        let currentXRayConverters = [];

        document.addEventListener('DOMContentLoaded', function() {
            showTab('groups');
            refreshGroups();
        });

        function showTab(tabName) {
            const tabs = document.querySelectorAll('.tab-content');
            tabs.forEach(tab => tab.style.display = 'none');
            
            const navLinks = document.querySelectorAll('.nav-link');
            navLinks.forEach(link => link.classList.remove('active'));
            
            // Handle different tab naming conventions
            let targetElement;
            
            // Map tab names to their actual IDs
            const tabIdMap = {
                'groups': 'groups-tab',
                'commands': 'commands-tab', 
                'autoresponses': 'autoresponses-tab',
                'autoremove': 'autoremove-tab',
                'stats': 'stats-tab',
                'whatsapp': 'whatsapp',
                'xray-tab': 'xray-tab'
            };
            
            const actualId = tabIdMap[tabName] || tabName;
            targetElement = document.getElementById(actualId);
            
            if (targetElement) {
                targetElement.style.display = 'block';
            }
            
            // Update active tab styling
            const activeTab = document.querySelector('.nav-link.active');
            if (activeTab) {
                activeTab.classList.remove('active');
            }
            
            // Find and activate the clicked tab
            const clickedTab = document.querySelector('a[onclick*="' + tabName + '"]');
            if (clickedTab) {
                clickedTab.classList.add('active');
            }
            
            switch(tabName) {
                case 'groups': refreshGroups(); break;
                case 'commands': refreshCommands(); break;
                case 'autoresponses': refreshAutoResponses(); break;
                case 'xray-tab': refreshXRayConverters(); break;
                case 'autoremove': refreshAutoRemoveTab(); break;
                case 'stats': refreshStats(); break;
                case 'whatsapp': refreshWhatsAppStatus(); break;
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

        // === AUTO REMOVE FUNCTIONS ===

        function refreshAutoRemoveTab() {
            const container = document.getElementById('autoremove-group-list');
            container.innerHTML = '<div class="spinner-border" role="status"><span class="visually-hidden">Loading...</span></div>';

            fetch('/api/groups')
                .then(response => response.json())
                .then(groups => {
                    if (!groups || groups.length === 0) {
                        container.innerHTML = '<div class="alert alert-info">Tidak ada grup yang dikelola. Tambahkan grup di tab Kelola Grup.</div>';
                        return;
                    }

                    let html = '<div class="accordion" id="autoRemoveAccordion">';
                    let promises = groups.map((group, index) => {
                        return fetch('/api/forbidden_words?group_jid=' + encodeURIComponent(group.group_jid))
                            .then(res => res.json())
                            .then(words => {
                                return getGroupAccordionItem(group, words || [], index);
                            });
                    });

                    Promise.all(promises).then(items => {
                        html += items.join('');
                        html += '</div>';
                        container.innerHTML = html;
                    });
                })
                .catch(error => {
                    console.error('Error:', error);
                    container.innerHTML = '<div class="alert alert-danger">Gagal memuat data grup.</div>';
                });
        }

        function getGroupAccordionItem(group, words, index) {
            let itemHtml = '<div class="accordion-item">';
            itemHtml += '<h2 class="accordion-header" id="heading' + index + '">';
            itemHtml += '<button class="accordion-button collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#collapse' + index + '" aria-expanded="false" aria-controls="collapse' + index + '">';
            itemHtml += group.group_name + ' <span class="badge bg-secondary ms-2">' + words.length + ' kata</span>';
            itemHtml += '</button></h2>';
            itemHtml += '<div id="collapse' + index + '" class="accordion-collapse collapse" aria-labelledby="heading' + index + '" data-bs-parent="#autoRemoveAccordion">';
            itemHtml += '<div class="accordion-body">';

            itemHtml += '<form class="row g-3 mb-3">' +
                '<div class="col-auto">' +
                '<input type="text" class="form-control" id="newForbiddenWord-' + group.group_jid + '" placeholder="Kata baru" required>' +
                '</div>' +
                '<div class="col-auto">' +
                '<button type="button" class="btn btn-success" onclick="saveNewForbiddenWord(\'' + group.group_jid + '\')">Tambah</button>' +
                '</div>' +
                '</form>';

            if (words.length > 0) {
                itemHtml += '<ul class="list-group">';
                words.forEach(word => {
                    itemHtml += '<li class="list-group-item d-flex justify-content-between align-items-center">';
                    itemHtml += word.word;
                    itemHtml += '<button class="btn btn-sm btn-danger" onclick="deleteForbiddenWord(' + word.id + ')">Hapus</button>';
                    itemHtml += '</li>';
                });
                itemHtml += '</ul>';
            } else {
                itemHtml += '<p class="text-muted">Belum ada kata terlarang untuk grup ini.</p>';
            }

            itemHtml += '</div></div></div>';
            return itemHtml;
        }

        function saveNewForbiddenWord(groupJID) {
            const newWord = document.getElementById('newForbiddenWord-' + groupJID).value;
            if (!newWord) {
                showAlert('warning', 'Isi kata terlarang');
                return;
            }

            const wordData = {
                group_jid: groupJID,
                word: newWord,
                created_by: 'admin'
            };

            fetch('/api/forbidden_words', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(wordData)
            })
            .then(response => response.json())
            .then(data => {
                if (data.status === 'success') {
                    showAlert('success', 'Kata terlarang berhasil ditambahkan');
                    refreshAutoRemoveTab();
                } else {
                    showAlert('danger', 'Gagal menambahkan kata terlarang: ' + (data.error || 'Unknown error'));
                }
            })
            .catch(error => {
                console.error('Error saving forbidden word:', error);
                showAlert('danger', 'Gagal menambahkan kata terlarang');
            });
        }

        function deleteForbiddenWord(id) {
            if (!confirm('Hapus kata terlarang ini?')) return;

            fetch('/api/forbidden_words?id=' + id, {
                method: 'DELETE'
            })
            .then(response => response.json())
            .then(data => {
                if (data.status === 'success') {
                    showAlert('success', 'Kata terlarang berhasil dihapus');
                    refreshAutoRemoveTab();
                } else {
                    showAlert('danger', 'Gagal menghapus kata terlarang');
                }
            })
            .catch(error => {
                console.error('Error deleting forbidden word:', error);
                showAlert('danger', 'Gagal menghapus kata terlarang');
            });
        }

        // === XRAY CONVERTER FUNCTIONS ===

        function refreshXRayConverters() {
            fetch('/api/xray_converters')
                .then(response => response.json())
                .then(data => {
                    currentXRayConverters = data.converters || [];
                    displayXRayConverters();
                })
                .catch(error => console.error('Error:', error));
        }

        function displayXRayConverters() {
            const container = document.getElementById('xray-converters-list');
            if (currentXRayConverters.length === 0) {
                container.innerHTML = '<div class="col-12"><div class="alert alert-info"><i class="fas fa-info-circle"></i> Belum ada XRay converter. Klik "Tambah Converter" untuk membuat yang pertama.</div></div>';
                return;
            }

            let html = '';
            currentXRayConverters.forEach(converter => {
                const statusBadge = converter.is_active ? 
                    '<span class="badge bg-success">Aktif</span>' : 
                    '<span class="badge bg-secondary">Nonaktif</span>';
                
                const typeIcon = {
                    'wildcard': '🌐',
                    'sni': '🔐',
                    'ws': '📡',
                    'grpc': '⚡'
                }[converter.modify_type] || '🔧';

                html += ` + "`" + `
                    <div class="col-md-6 col-lg-4 mb-3">
                        <div class="card h-100">
                            <div class="card-body">
                                <div class="d-flex justify-content-between align-items-start mb-2">
                                    <h6 class="card-title mb-0">${converter.display_name}</h6>
                                    ${statusBadge}
                                </div>
                                <p class="card-text mb-1">
                                    <strong>Command:</strong> .${converter.command_name}<br>
                                    <strong>Type:</strong> ${typeIcon} ${converter.modify_type}<br>
                                    <strong>Bug Host:</strong> ${converter.bug_host}<br>
                                    <strong>Usage:</strong> ${converter.usage_count || 0}x
                                </p>
                                ${converter.path_template ? ` + "`" + `<small class="text-muted">Path: ${converter.path_template}</small><br>` + "`" + ` : ''}
                                ${converter.grpc_service_name ? ` + "`" + `<small class="text-muted">gRPC: ${converter.grpc_service_name}</small><br>` + "`" + ` : ''}
                                <small class="text-muted">Created by: ${converter.created_by}</small>
                            </div>
                            <div class="card-footer">
                                <div class="btn-group w-100" role="group">
                                    <button class="btn btn-outline-primary btn-sm" onclick="editXRayConverter('${converter.command_name}')">
                                        <i class="fas fa-edit"></i>
                                    </button>
                                    <button class="btn btn-outline-danger btn-sm" onclick="deleteXRayConverter('${converter.command_name}')">
                                        <i class="fas fa-trash"></i>
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                ` + "`" + `;
            });

            container.innerHTML = html;
        }

        function toggleAdvancedSettings() {
            const modifyType = document.getElementById('newConverterModifyType').value;
            const advancedSettings = document.getElementById('advancedSettings');
            
            if (modifyType === 'custom') {
                advancedSettings.style.display = 'block';
            } else {
                advancedSettings.style.display = 'none';
                // Clear template fields when not using custom
                document.getElementById('newConverterServerTemplate').value = '';
                document.getElementById('newConverterHostTemplate').value = '';
                document.getElementById('newConverterSNITemplate').value = '';
            }
        }

        function saveNewXRayConverter() {
            const converterData = {
                command_name: document.getElementById('newConverterCommand').value,
                display_name: document.getElementById('newConverterDisplayName').value,
                bug_host: document.getElementById('newConverterBugHost').value,
                modify_type: document.getElementById('newConverterModifyType').value,
                server_template: document.getElementById('newConverterServerTemplate').value,
                host_template: document.getElementById('newConverterHostTemplate').value,
                sni_template: document.getElementById('newConverterSNITemplate').value,
                path_template: document.getElementById('newConverterPathTemplate').value,
                grpc_service_name: document.getElementById('newConverterGrpcService').value,
                port_override: document.getElementById('newConverterPortOverride').value ? 
                    parseInt(document.getElementById('newConverterPortOverride').value) : null
            };

            // Validation
            if (!converterData.command_name || !converterData.display_name || !converterData.bug_host || !converterData.modify_type) {
                alert('Mohon isi semua field yang required (*)');
                return;
            }

            fetch('/api/xray_converters', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(converterData)
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('✅ XRay Converter berhasil ditambahkan!');
                    document.getElementById('addXRayConverterForm').reset();
                    bootstrap.Modal.getInstance(document.getElementById('addXRayConverterModal')).hide();
                    refreshXRayConverters();
                } else {
                    alert('❌ Gagal menambahkan converter: ' + (data.message || 'Unknown error'));
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('❌ Error: ' + error.message);
            });
        }

        function deleteXRayConverter(commandName) {
            if (!confirm(` + "`" + `Yakin ingin menghapus converter "${commandName}"?` + "`" + `)) return;

            fetch(` + "`" + `/api/xray_converters?command=${commandName}` + "`" + `, {
                method: 'DELETE'
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('✅ Converter berhasil dihapus!');
                    refreshXRayConverters();
                } else {
                    alert('❌ Gagal menghapus converter: ' + (data.message || 'Unknown error'));
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('❌ Error: ' + error.message);
            });
        }

        function toggleEditAdvancedSettings() {
            const modifyType = document.getElementById('editConverterModifyType').value;
            const advancedSettings = document.getElementById('editAdvancedSettings');
            
            if (modifyType === 'custom') {
                advancedSettings.style.display = 'block';
            } else {
                advancedSettings.style.display = 'none';
            }
        }

        function editXRayConverter(commandName) {
            // Debug: Log available converters
            console.log('Looking for converter:', commandName);
            console.log('Available converters:', currentXRayConverters);
            
            // Find converter data
            const converter = currentXRayConverters.find(c => c.command_name === commandName);
            if (!converter) {
                alert('❌ Converter tidak ditemukan! Command: ' + commandName + '\nAvailable: ' + currentXRayConverters.map(c => c.command_name).join(', '));
                return;
            }

            // Populate form
            document.getElementById('editConverterOriginalCommand').value = converter.command_name;
            document.getElementById('editConverterCommand').value = converter.command_name;
            document.getElementById('editConverterDisplayName').value = converter.display_name;
            document.getElementById('editConverterBugHost').value = converter.bug_host;
            document.getElementById('editConverterModifyType').value = converter.modify_type;
            document.getElementById('editConverterServerTemplate').value = converter.server_template || '';
            document.getElementById('editConverterHostTemplate').value = converter.host_template || '';
            document.getElementById('editConverterSNITemplate').value = converter.sni_template || '';
            document.getElementById('editConverterPathTemplate').value = converter.path_template || '';
            document.getElementById('editConverterGrpcService').value = converter.grpc_service_name || '';
            document.getElementById('editConverterPortOverride').value = converter.port_override || '';
            document.getElementById('editConverterIsActive').value = converter.is_active ? 'true' : 'false';

            // Toggle advanced settings if needed
            toggleEditAdvancedSettings();

            // Show modal
            new bootstrap.Modal(document.getElementById('editXRayConverterModal')).show();
        }

        function saveEditXRayConverter() {
            const converterData = {
                command_name: document.getElementById('editConverterOriginalCommand').value,
                display_name: document.getElementById('editConverterDisplayName').value,
                bug_host: document.getElementById('editConverterBugHost').value,
                modify_type: document.getElementById('editConverterModifyType').value,
                server_template: document.getElementById('editConverterServerTemplate').value,
                host_template: document.getElementById('editConverterHostTemplate').value,
                sni_template: document.getElementById('editConverterSNITemplate').value,
                path_template: document.getElementById('editConverterPathTemplate').value,
                grpc_service_name: document.getElementById('editConverterGrpcService').value,
                port_override: document.getElementById('editConverterPortOverride').value ? 
                    parseInt(document.getElementById('editConverterPortOverride').value) : null,
                is_active: document.getElementById('editConverterIsActive').value === 'true'
            };

            // Validation
            if (!converterData.display_name || !converterData.bug_host || !converterData.modify_type) {
                alert('Mohon isi semua field yang required (*)');
                return;
            }

            fetch('/api/xray_converters', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(converterData)
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('✅ XRay Converter berhasil diupdate!');
                    bootstrap.Modal.getInstance(document.getElementById('editXRayConverterModal')).hide();
                    refreshXRayConverters();
                } else {
                    alert('❌ Gagal mengupdate converter: ' + (data.message || 'Unknown error'));
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('❌ Error: ' + error.message);
            });
        }

        // === WHATSAPP PAIRING FUNCTIONS ===
        
        function refreshWhatsAppStatus() {
            fetch('/api/whatsapp/status')
                .then(response => response.json())
                .then(data => {
                    updateWhatsAppStatusUI(data);
                })
                .catch(error => {
                    console.error('Error fetching WhatsApp status:', error);
                });
        }

        function updateWhatsAppStatusUI(status) {
            const connectionStatus = document.getElementById('connection-status');
            const loginStatus = document.getElementById('login-status');
            const phoneNumber = document.getElementById('phone-number');
            const deviceName = document.getElementById('device-name');

            // Update connection status
            if (status.connected) {
                connectionStatus.innerHTML = '<i class="fas fa-circle text-success fa-2x"></i>';
            } else {
                connectionStatus.innerHTML = '<i class="fas fa-circle text-danger fa-2x"></i>';
            }

            // Update login status
            if (status.logged_in) {
                loginStatus.innerHTML = '<i class="fas fa-circle text-success fa-2x"></i>';
            } else {
                loginStatus.innerHTML = '<i class="fas fa-circle text-warning fa-2x"></i>';
            }

            // Update device info
            phoneNumber.textContent = status.phone_number || '-';
            deviceName.textContent = status.device_name || '-';
        }

        let qrPairingInProgress = false;

        function startQRPairing() {
            const btn = document.getElementById('qr-pairing-btn');
            const refreshBtn = document.getElementById('qr-refresh-btn');
            if (qrPairingInProgress) { return; }
            qrPairingInProgress = true;
            
            btn.disabled = true;
            btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Starting QR Pairing...';

            // Show loading state
            showQRLoading();

            fetch('/api/whatsapp/qr', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    showAlert('success', data.message);
                    // Start polling for QR code
                    setTimeout(() => pollForQRCode(), 2000);
                    refreshBtn.style.display = 'inline-block';
                } else {
                    showAlert('error', data.message);
                    showQRPlaceholder();
                }
            })
            .catch(error => {
                console.error('Error starting QR pairing:', error);
                showAlert('error', 'Error starting QR pairing: ' + error.message);
                showQRPlaceholder();
            })
            .finally(() => {
                btn.disabled = false;
                btn.innerHTML = '<i class="fas fa-qrcode"></i> Start QR Pairing';
            });
        }

        function pollForQRCode() {
            fetch('/api/whatsapp/qr-image')
                .then(response => response.json())
                .then(data => {
                    console.log('QR polling response:', data);
                    
                    if (data.success && data.qr_image) {
                        showQRImage(data.qr_image);
                        showAlert('info', 'QR code siap! Scan dengan WhatsApp Anda.');
                        // Continue polling untuk check if success
                        setTimeout(() => pollForStatus(), 5000);
                    } else {
                        // Continue polling if QR not ready
                        console.log('QR not ready, continuing polling...');
                        setTimeout(() => pollForQRCode(), 2000);
                    }
                })
                .catch(error => {
                    console.error('Error polling QR code:', error);
                    setTimeout(() => pollForQRCode(), 3000); // Retry in 3 seconds
                });
        }

        function pollForStatus() {
            if (pairingInProgress) { return; }
            fetch('/api/whatsapp/status')
                .then(response => response.json())
                .then(data => {
                    if (data.logged_in && data.connected) {
                        showQRPlaceholder();
                        showAlert('success', 'QR pairing berhasil! WhatsApp terhubung.');
                        refreshWhatsAppStatus();
                    } else {
                        // Continue polling
                        setTimeout(() => pollForStatus(), 3000);
                    }
                })
                .catch(error => {
                    console.error('Error polling status:', error);
                });
        }

        function cancelQRPairing() {
            fetch('/api/whatsapp/qr/cancel', {
                method: 'POST', headers: { 'Content-Type': 'application/json' }
            }).then(r => r.json()).then(data => {
                if (data.success) {
                    showAlert('warning', data.message);
                    qrPairingInProgress = false;
                    showQRPlaceholder();
                } else {
                    showAlert('error', data.message || 'Gagal membatalkan QR pairing');
                }
            }).catch(err => {
                console.error('Cancel QR error:', err);
            });
        }

        function refreshQRCode() {
            showQRLoading();
            pollForQRCode();
        }

        function showQRPlaceholder() {
            document.getElementById('qr-placeholder').style.display = 'block';
            document.getElementById('qr-image-container').style.display = 'none';
            document.getElementById('qr-loading').style.display = 'none';
            document.getElementById('qr-refresh-btn').style.display = 'none';
        }

        function showQRLoading() {
            document.getElementById('qr-placeholder').style.display = 'none';
            document.getElementById('qr-image-container').style.display = 'none';
            document.getElementById('qr-loading').style.display = 'block';
        }

        function showQRImage(base64Image) {
            document.getElementById('qr-placeholder').style.display = 'none';
            document.getElementById('qr-loading').style.display = 'none';
            document.getElementById('qr-image-container').style.display = 'block';
            document.getElementById('qr-image').src = base64Image;
        }

        function showPairingCode(code) {
            document.getElementById('pairing-code-display').textContent = code;
            document.getElementById('pairing-code-area').style.display = 'block';
            showAlert('success', 'Pairing code siap! Masukkan di WhatsApp Anda.');
        }

        function hidePairingCode() {
            document.getElementById('pairing-code-area').style.display = 'none';
        }

        function pollForPairingCode() {
            fetch('/api/whatsapp/pairing-code')
                .then(response => response.json())
                .then(data => {
                    console.log('Pairing code polling response:', data);
                    
                    if (data.success && data.pairing_code) {
                        showPairingCode(data.pairing_code);
                        showAlert('success', 'Pairing code siap! Masukkan di WhatsApp Anda.');
                        // Continue polling untuk check if success
                        setTimeout(() => pollForStatus(), 5000);
                    } else {
                        // Continue polling if code not ready
                        console.log('Pairing code not ready, continuing polling...');
                        setTimeout(() => pollForPairingCode(), 2000);
                    }
                })
                .catch(error => {
                    console.error('Error polling pairing code:', error);
                    setTimeout(() => pollForPairingCode(), 3000);
                });
        }

        let pairingInProgress = false;

        function startPhonePairing() {
            const phoneInput = document.getElementById('phone-input');
            const btn = document.getElementById('phone-pairing-btn');
            const phoneNumber = phoneInput.value.trim();

            if (!phoneNumber) {
                showAlert('error', 'Masukkan nomor telepon!');
                return;
            }

            btn.disabled = true;
            btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Getting Pairing Code...';
            pairingInProgress = true;

            fetch('/api/whatsapp/phone', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    phone_number: phoneNumber
                })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    showAlert('success', data.message);
                    // Start polling for pairing code
                    setTimeout(() => pollForPairingCode(), 3000);
                } else {
                    showAlert('error', data.message);
                }
            })
            .catch(error => {
                console.error('Error starting phone pairing:', error);
                showAlert('error', 'Error starting phone pairing: ' + error.message);
            })
            .finally(() => {
                btn.disabled = false;
                btn.innerHTML = '<i class="fas fa-phone"></i> Get Pairing Code';
            });
        }

        function reconnectWhatsApp() {
            if (!confirm('Reconnect WhatsApp connection?')) {
                return;
            }

            fetch('/api/whatsapp/reconnect', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    showAlert('success', data.message);
                    setTimeout(refreshWhatsAppStatus, 3000);
                } else {
                    showAlert('error', data.message);
                }
            })
            .catch(error => {
                console.error('Error reconnecting WhatsApp:', error);
                showAlert('error', 'Error reconnecting: ' + error.message);
            });
        }

        function disconnectWhatsApp() {
            if (!confirm('Disconnect WhatsApp? Bot akan berhenti menerima pesan.')) {
                return;
            }

            fetch('/api/whatsapp/disconnect', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    showAlert('success', data.message);
                    setTimeout(refreshWhatsAppStatus, 2000);
                } else {
                    showAlert('error', data.message);
                }
            })
            .catch(error => {
                console.error('Error disconnecting WhatsApp:', error);
                showAlert('error', 'Error disconnecting: ' + error.message);
            });
        }

        function safeLogoutWhatsApp() {
            if (!confirm('⚠️ SAFE LOGOUT dari WhatsApp?\\n\\nIni akan:\\n• Logout aman dari WhatsApp\\n• Hapus session dengan benar\\n• Perlu QR pairing lagi\\n\\nGunakan ini untuk mencegah database corruption.')) {
                return;
            }

            fetch('/api/whatsapp/logout', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    showAlert('success', data.message + '\\n\\nGunakan QR pairing untuk login ulang.');
                    // Reset UI state
                    showQRPlaceholder();
                    hidePairingCode();
                    // Refresh status after logout completes
                    setTimeout(refreshWhatsAppStatus, 5000);
                } else {
                    showAlert('error', data.message);
                }
            })
            .catch(error => {
                console.error('Error during safe logout:', error);
                showAlert('error', 'Error during logout: ' + error.message);
            });
        }

        function fullResetWhatsApp() {
            if (!confirm('⚠️ FULL RESET akan menghapus file session.db dan QR file.\n\nLanjutkan?')) { return; }
            fetch('/api/whatsapp/full_reset', { method: 'POST', headers: { 'Content-Type': 'application/json' }})
              .then(r => r.json())
              .then(data => {
                  if (data.success) {
                      showAlert('warning', data.message);
                      showQRPlaceholder(); hidePairingCode();
                      setTimeout(refreshWhatsAppStatus, 1000);
                  } else { showAlert('error', data.message || 'Full reset gagal'); }
              }).catch(err => {
                  console.error('Full reset error:', err);
                  showAlert('error', 'Full reset error: ' + err.message);
              });
        }

        // Auto-refresh WhatsApp status when tab is shown
        document.addEventListener('DOMContentLoaded', function() {
            // Refresh status when WhatsApp tab is shown
            const originalShowTab = window.showTab;
            window.showTab = function(tabName) {
                originalShowTab(tabName);
                if (tabName === 'whatsapp') {
                    refreshWhatsAppStatus();
                }
            };
        });

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

// handleForbiddenWords handles forbidden word management API
func (s *DashboardServer) handleForbiddenWords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getForbiddenWords(w, r)
	case "POST":
		s.createForbiddenWord(w, r)
	case "DELETE":
		s.deleteForbiddenWord(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getForbiddenWords returns all forbidden words for a group
func (s *DashboardServer) getForbiddenWords(w http.ResponseWriter, r *http.Request) {
	groupJID := r.URL.Query().Get("group_jid")
	if groupJID == "" {
		http.Error(w, "Group JID required", http.StatusBadRequest)
		return
	}

	words, err := s.repository.GetForbiddenWordsByGroup(groupJID)
	if err != nil {
		s.logger.Errorf("Failed to get forbidden words: %v", err)
		http.Error(w, "Failed to get forbidden words", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(words)
}

// createForbiddenWord creates a new forbidden word
func (s *DashboardServer) createForbiddenWord(w http.ResponseWriter, r *http.Request) {
	var word database.ForbiddenWord
	if err := json.NewDecoder(r.Body).Decode(&word); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := s.repository.CreateForbiddenWord(&word); err != nil {
		s.logger.Errorf("Failed to create forbidden word: %v", err)
		http.Error(w, "Failed to create forbidden word", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// deleteForbiddenWord deletes a forbidden word
func (s *DashboardServer) deleteForbiddenWord(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	var idInt int
	fmt.Sscanf(id, "%d", &idInt)

	if err := s.repository.DeleteForbiddenWord(idInt); err != nil {
		s.logger.Errorf("Failed to delete forbidden word: %v", err)
		http.Error(w, "Failed to delete forbidden word", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// === XRAY CONVERTER HANDLERS ===

// handleXRayConverters handles CRUD operations for XRay converters
func (s *DashboardServer) handleXRayConverters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	switch r.Method {
	case "GET":
		s.handleGetXRayConverters(w, r)
	case "POST":
		s.handleCreateXRayConverter(w, r)
	case "PUT":
		s.handleUpdateXRayConverter(w, r)
	case "DELETE":
		s.handleDeleteXRayConverter(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetXRayConverters returns all XRay converters
func (s *DashboardServer) handleGetXRayConverters(w http.ResponseWriter, r *http.Request) {
	converters, err := s.repository.GetAllXRayConverters()
	if err != nil {
		s.logger.Errorf("Failed to get XRay converters: %v", err)
		http.Error(w, "Failed to get converters", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"converters": converters,
		"count":      len(converters),
	}

	json.NewEncoder(w).Encode(response)
}

// handleCreateXRayConverter creates a new XRay converter
func (s *DashboardServer) handleCreateXRayConverter(w http.ResponseWriter, r *http.Request) {
	var converter database.XRayConverter
	if err := json.NewDecoder(r.Body).Decode(&converter); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if converter.CommandName == "" || converter.DisplayName == "" || converter.BugHost == "" || converter.ModifyType == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Set default values
	converter.IsActive = true
	converter.CreatedBy = "admin" // TODO: Get from session/auth

	// Create converter
	err := s.repository.CreateXRayConverter(&converter)
	if err != nil {
		s.logger.Errorf("Failed to create XRay converter: %v", err)
		http.Error(w, "Failed to create converter", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"message":   "Converter created successfully",
		"converter": converter,
	}

	json.NewEncoder(w).Encode(response)
}

// handleUpdateXRayConverter updates an existing XRay converter
func (s *DashboardServer) handleUpdateXRayConverter(w http.ResponseWriter, r *http.Request) {
	var converter database.XRayConverter
	if err := json.NewDecoder(r.Body).Decode(&converter); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if converter.CommandName == "" {
		http.Error(w, "Command name is required", http.StatusBadRequest)
		return
	}

	err := s.repository.UpdateXRayConverter(&converter)
	if err != nil {
		s.logger.Errorf("Failed to update XRay converter: %v", err)
		http.Error(w, "Failed to update converter", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Converter updated successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// handleDeleteXRayConverter deletes an XRay converter
func (s *DashboardServer) handleDeleteXRayConverter(w http.ResponseWriter, r *http.Request) {
	commandName := r.URL.Query().Get("command")
	if commandName == "" {
		http.Error(w, "Command name is required", http.StatusBadRequest)
		return
	}

	err := s.repository.DeleteXRayConverter(commandName)
	if err != nil {
		s.logger.Errorf("Failed to delete XRay converter: %v", err)
		http.Error(w, "Failed to delete converter", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Converter deleted successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// handleXRayConverterTest tests an XRay converter with sample input
func (s *DashboardServer) handleXRayConverterTest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var testRequest struct {
		ConverterName string `json:"converter_name"`
		XRayLink      string `json:"xray_link"`
	}

	if err := json.NewDecoder(r.Body).Decode(&testRequest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implement XRay converter test logic
	// This would use the XRayConverterService to test conversion

	response := map[string]interface{}{
		"success": true,
		"message": "Test functionality will be implemented with XRayConverterService integration",
		"input":   testRequest,
	}

	json.NewEncoder(w).Encode(response)
}

// === WHATSAPP PAIRING HANDLERS ===

// handleWhatsAppStatus mengembalikan status koneksi WhatsApp
func (s *DashboardServer) handleWhatsAppStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	status := map[string]interface{}{
		"connected":       false,
		"logged_in":       false,
		"phone_number":    "",
		"device_name":     "",
		"last_seen":       "",
		"session_healthy": true,
	}

	if s.waManager != nil {
		status["connected"] = s.waManager.IsConnected()
		status["logged_in"] = s.waManager.IsLoggedIn()

		if s.whatsappClient != nil && s.whatsappClient.Store.ID != nil {
			status["phone_number"] = s.whatsappClient.Store.ID.User
			status["device_name"] = fmt.Sprintf("Device-%d", s.whatsappClient.Store.ID.Device)
		}

		// Get manager stats
		stats := s.waManager.GetStats()
		status["stats"] = stats
	}

	json.NewEncoder(w).Encode(status)
}

// handleQRCancel membatalkan proses QR pairing aktif
func (s *DashboardServer) handleQRCancel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.dashboardQR != nil {
		s.dashboardQR.Cancel()
	}
	// Clear QR file to avoid stale
	_ = os.Remove("data/qrcode.png")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "QR pairing dibatalkan",
	})
}

// handleQRPairing menangani pairing via QR code
func (s *DashboardServer) handleQRPairing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.waManager == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "WhatsApp Manager tidak tersedia",
		})
		return
	}

	// Check if already logged in
	if s.waManager.IsLoggedIn() {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Sudah login, logout dulu untuk QR pairing baru",
		})
		return
	}

	// Remove old QR file so dashboard won't show stale image
	_ = os.Remove("data/qrcode.png")

	// Start QR pairing using dashboard QR handler
	go func() {
		if s.dashboardQR == nil {
			s.logger.Error("❌ Dashboard QR handler tidak tersedia")
			return
		}

		err := s.dashboardQR.StartDashboardQRPairing()
		if err != nil {
			s.logger.Errorf("❌ QR pairing gagal: %v", err)
		}
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "QR pairing dimulai, QR code akan muncul di bawah",
		"qr_ready": false,
	})
}

// handleQRImage mengembalikan QR code sebagai base64 image
func (s *DashboardServer) handleQRImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// QUICK FIX: Check existing QR file first
	qrPath := "data/qrcode.png"

	// Check if QR file exists (from terminal generation)
	if _, err := os.Stat(qrPath); err == nil {
		s.logger.Debug("📱 Using existing QR file for dashboard")

		// Read existing QR file
		imageData, err := ioutil.ReadFile(qrPath)
		if err == nil {
			// Convert to base64
			base64Image := base64.StdEncoding.EncodeToString(imageData)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":  true,
				"qr_image": "data:image/png;base64," + base64Image,
				"source":   "existing_file",
				"message":  "QR code loaded from file",
			})
			return
		}
	}

	// Fallback: Get QR code from dashboard QR handler
	var qrCode string
	if s.dashboardQR != nil {
		qrCode = s.dashboardQR.GetCurrentQRCode()
	}

	if qrCode == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   false,
			"message":   "QR code belum tersedia, mulai pairing dulu",
			"qr_active": s.dashboardQR != nil && s.dashboardQR.IsActive(),
		})
		return
	}

	// Generate QR code image using custom path method
	qrPath = "data/qr_temp.png"
	err := s.qrGenerator.GenerateQRToFile(qrCode, qrPath)
	if err != nil {
		s.logger.Errorf("Gagal generate QR image: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Gagal generate QR image: " + err.Error(),
		})
		return
	}

	// Read image file and convert to base64
	imageData, err := ioutil.ReadFile(qrPath)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Gagal membaca QR image",
		})
		return
	}

	// Clean up temp file
	os.Remove(qrPath)

	// Convert to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"qr_code":  qrCode,
		"qr_image": "data:image/png;base64," + base64Image,
	})
}

// handlePhonePairing menangani pairing via nomor telepon
func (s *DashboardServer) handlePhonePairing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		PhoneNumber string `json:"phone_number"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request format",
		})
		return
	}

	if s.whatsappClient == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "WhatsApp client tidak tersedia",
		})
		return
	}

	// Validate phone number format
	phoneNumber := strings.ReplaceAll(request.PhoneNumber, " ", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")

	if len(phoneNumber) < 10 || len(phoneNumber) > 15 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Format nomor telepon tidak valid",
		})
		return
	}

	// Start phone pairing process
	go func() {
		s.logger.Infof("📱 Memulai pairing dengan nomor: %s", phoneNumber)

		// Request pairing code using WAManager (safe flow)
		if s.waManager == nil {
			s.logger.Error("WAManager tidak tersedia")
			return
		}
		code, err := s.waManager.PairByPhone(context.Background(), s.whatsappClient, phoneNumber)
		if err != nil {
			s.logger.Errorf("❌ Gagal request pairing code: %v", err)
			return
		}

		s.logger.Successf("🔑 Pairing code: %s", code)
		s.logger.Info("📋 Masukkan code ini di WhatsApp > Linked Devices > Link a Device")

		// Store pairing code untuk dashboard
		s.pairingMutex.Lock()
		s.currentPairingCode = code
		s.pairingMutex.Unlock()

		s.logger.Info("✅ Pairing code tersedia di dashboard")

		// Refresh main client and reconnect so dashboard sees logged-in state
		go func() {
			defer func() { recover() }()
			if s.waManager == nil {
				return
			}
			deviceStore, err := s.waManager.Container.GetFirstDevice(context.Background())
			if err != nil {
				s.logger.Errorf("Gagal ambil device store setelah pairing: %v", err)
				return
			}
			clientLog := waLog.Stdout("WhatsApp", "INFO", true)
			newClient := whatsmeow.NewClient(deviceStore, clientLog)
			s.whatsappClient = newClient
			s.waManager.Client = newClient
			newClient.EnableAutoReconnect = true
			newClient.DisableLoginAutoReconnect = false
			if err := s.waManager.ConnectSafely(); err != nil {
				s.logger.Errorf("Reconnect setelah phone pairing gagal: %v", err)
			} else {
				s.logger.Success("Reconnect setelah phone pairing berhasil")
			}
		}()
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Pairing code generation started",
		"phone":   phoneNumber,
	})
}

// handleDisconnect menangani disconnect WhatsApp
func (s *DashboardServer) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.waManager == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "WhatsApp Manager tidak tersedia",
		})
		return
	}

	s.waManager.Disconnect()
	s.logger.Info("🔌 WhatsApp disconnected via dashboard")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "WhatsApp berhasil di-disconnect",
	})
}

// handleReconnect menangani reconnect WhatsApp
func (s *DashboardServer) handleReconnect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.waManager == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "WhatsApp Manager tidak tersedia",
		})
		return
	}

	if !s.waManager.IsLoggedIn() {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Belum login, gunakan QR atau phone pairing dulu",
		})
		return
	}

	// Reconnect in background
	go func() {
		s.logger.Info("🔄 Reconnecting WhatsApp via dashboard...")
		// Re-enable auto reconnect before connecting
		if s.whatsappClient != nil {
			s.whatsappClient.EnableAutoReconnect = true
			s.whatsappClient.DisableLoginAutoReconnect = false
		}
		err := s.waManager.ConnectSafely()
		if err != nil {
			s.logger.Errorf("❌ Reconnect gagal: %v", err)
		} else {
			s.logger.Success("✅ Reconnect berhasil!")
		}
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Reconnecting WhatsApp...",
	})
}

// handleLogout menangani true logout (benar-benar hapus session)
func (s *DashboardServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.whatsappClient == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "WhatsApp client tidak tersedia",
		})
		return
	}

	// True logout process
	go func() {
		s.logger.Info("🔐 Starting true logout process...")
		// Disable auto reconnect to avoid background connects
		if s.whatsappClient != nil {
			s.whatsappClient.EnableAutoReconnect = false
			s.whatsappClient.DisableLoginAutoReconnect = true
		}
		// Disconnect gracefully
		if s.waManager != nil {
			s.waManager.Disconnect()
		}
		s.logger.Info("🔌 Disconnected, draining...")
		time.Sleep(2 * time.Second)

		// Try server-side logout (if socket up it's fine; else ignore)
		if s.whatsappClient != nil && s.whatsappClient.Store != nil {
			_ = s.whatsappClient.Logout(context.Background())
		}

		// Backup session.db lama agar file handle dilepas secara aman dan tidak readonly
		timestamp := time.Now().Format("20060102_150405")
		_ = os.Rename("data/session.db", fmt.Sprintf("data/session_%s.db.old", timestamp))
		_ = os.Remove("data/session.db-wal")
		_ = os.Remove("data/session.db-shm")

		s.logger.Success("📦 WA Session lama di-backup untuk sesi baru.")

		// Reset dashboard states
		s.pairingMutex.Lock()
		s.currentPairingCode = ""
		s.pairingMutex.Unlock()
		s.qrMutex.Lock()
		s.currentQRCode = ""
		s.qrMutex.Unlock()
		s.logger.Success("🔐 True logout completed - fresh state ready. Sesi db diperbarui.")

		// Tunggu sedikit agar client web dapat menangkap respon JSON
		time.Sleep(3 * time.Second)
		s.logger.Info("🔄 Re-initializing WA Session (Hot Swap)...")
		if s.onReloadSessionDB != nil {
			if err := s.onReloadSessionDB(); err != nil {
				s.logger.Errorf("❌ Failed to hot-swap session: %v", err)
			} else {
				s.logger.Success("✅ Hot swap session completed successfully without app restart.")
			}
		}
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Safe logout initiated - Sesi baru sedang disiapkan tanpa mematikan bot, silakan refresh halaman sebentar lagi.",
	})
}

// handleFullReset melakukan reset penuh session (hapus file DB)
func (s *DashboardServer) handleFullReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Batalkan pairing jika aktif
	if s.dashboardQR != nil {
		s.dashboardQR.Cancel()
	}
	// Matikan auto-reconnect dan disconnect
	if s.whatsappClient != nil {
		s.whatsappClient.EnableAutoReconnect = false
		s.whatsappClient.DisableLoginAutoReconnect = true
		s.whatsappClient.Disconnect()
	}
	// Hapus file session DB (termasuk WAL/SHM) dan db pembelajaran
	_ = os.Remove("data/session.db")
	_ = os.Remove("data/session.db-wal")
	_ = os.Remove("data/session.db-shm")

	// Rename session.db yang lama ke .old supaya tidak readonly
	timestamp := time.Now().Format("20060102_150405")
	_ = os.Rename("data/session.db", fmt.Sprintf("data/session_%s.db.old", timestamp))
	_ = os.Remove("data/session.db-wal")
	_ = os.Remove("data/session.db-shm")

	_ = os.Remove("data/qrcode.png")

	// Pastikan folder data ada dan writable
	_ = os.MkdirAll("data", 0o775)

	// Reset in-memory states
	s.pairingMutex.Lock()
	s.currentPairingCode = ""
	s.pairingMutex.Unlock()
	s.qrMutex.Lock()
	s.currentQRCode = ""
	s.qrMutex.Unlock()

	// Kirim response ke client sebelum proses re-init
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Full reset completed. Database baru telah disiapkan tanpa mematikan bot. Silakan refresh halaman.",
	})

	// Jalankan reload Session di background agar response sempat terkirim
	go func() {
		time.Sleep(3 * time.Second)
		s.logger.Info("🔄 Re-initializing WA Session akibat Full Reset (Hot Swap)...")
		if s.onReloadSessionDB != nil {
			if err := s.onReloadSessionDB(); err != nil {
				s.logger.Errorf("❌ Failed to hot-swap session: %v", err)
			} else {
				s.logger.Success("✅ Hot swap session completed successfully without app restart.")
			}
		}
	}()
}

// handleGetPairingCode mengembalikan pairing code yang tersimpan
func (s *DashboardServer) handleGetPairingCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	s.pairingMutex.RLock()
	pairingCode := s.currentPairingCode
	s.pairingMutex.RUnlock()

	if pairingCode == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Pairing code belum tersedia",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"pairing_code": pairingCode,
		"message":      "Pairing code ready",
	})
}
