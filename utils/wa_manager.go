// Package utils - Enhanced WhatsApp Manager dengan anti-spam protection
// File ini berisi WhatsApp manager yang lebih aman dari deteksi spam
package utils

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// WAManager adalah struktur yang mengelola WhatsApp client dengan anti-spam protection
type WAManager struct {
	// Container untuk session database
	Container *sqlstore.Container
	
	// Client WhatsApp
	Client *whatsmeow.Client
	
	// Logger
	Logger *Logger
	
	// Anti-spam protection
	pairingMu     sync.Mutex
	pairingActive bool
	lastConnect   time.Time
	connectCount  int

	// Phone pairing protection
	phonePairMu sync.Mutex
	
	// Configuration
	DBPath          string
	MinConnectDelay time.Duration
	MaxConnectCount int
	ResetInterval   time.Duration
}

// NewWAManager membuat WhatsApp manager baru dengan anti-spam protection
func NewWAManager(dbPath string, logger *Logger) (*WAManager, error) {
	// Setup database logger dengan level yang lebih rendah untuk mengurangi noise
	dbLog := waLog.Stdout("Database", "ERROR", true)
	
	container, err := sqlstore.New(context.Background(), "sqlite3", "file:"+dbPath+"?_foreign_keys=on", dbLog)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat database container: %w", err)
	}

	manager := &WAManager{
		Container:       container,
		Logger:          logger,
		DBPath:          dbPath,
		MinConnectDelay: 5 * time.Second,  // Minimum 5 detik antar koneksi
		MaxConnectCount: 3,                // Maksimal 3 koneksi per interval
		ResetInterval:   5 * time.Minute,  // Reset counter setiap 5 menit
	}

	return manager, nil
}

// CreateClient membuat WhatsApp client dengan konfigurasi anti-spam
func (w *WAManager) CreateClient() error {
	// Dapatkan device store
	deviceStore, err := w.Container.GetFirstDevice(context.Background())
	if err != nil {
		return fmt.Errorf("gagal mendapatkan device store: %w", err)
	}

	// Setup client logger dengan level INFO untuk observabilitas yang baik
	clientLog := waLog.Stdout("WhatsApp", "INFO", true)
	
	// Buat client baru
	w.Client = whatsmeow.NewClient(deviceStore, clientLog)
	
	// Setup event handlers untuk monitoring
	w.setupEventHandlers()
	
	w.Logger.Success("WhatsApp client berhasil dibuat")
	return nil
}

// setupEventHandlers mengatur event handlers untuk monitoring koneksi
func (w *WAManager) setupEventHandlers() {
	w.Client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Connected:
			w.Logger.Success("✅ WhatsApp terhubung dengan sukses")
			// Re-enable auto reconnect after successful connection
			w.Client.EnableAutoReconnect = true
			w.Client.DisableLoginAutoReconnect = false
			w.resetConnectCounter()
			
		case *events.Disconnected:
			w.Logger.Warning("⚠️ WhatsApp terputus, akan reconnect otomatis...")
			
		case *events.LoggedOut:
			w.Logger.Error("🚪 WhatsApp logged out - perlu scan QR lagi")
			w.Logger.Infof("Alasan logout: %v", v.Reason)
			
		case *events.StreamReplaced:
			w.Logger.Warning("🔄 Session digantikan oleh login dari device lain")
		}
	})
}

// ConnectSafely melakukan koneksi dengan proteksi anti-spam
func (w *WAManager) ConnectSafely() error {
	// Cek rate limiting
	if err := w.checkRateLimit(); err != nil {
		return err
	}
	
	// Lock untuk mencegah race condition
	w.pairingMu.Lock()
	defer w.pairingMu.Unlock()
	
	// Cek apakah sudah login sebelumnya
	if w.Client.Store.ID == nil {
		w.Logger.Warning("Belum login, memerlukan QR code...")
		return fmt.Errorf("belum login - perlu QR code")
	}
	
	w.Logger.Info("Melakukan koneksi ke WhatsApp...")
	
	// Update counter dan timestamp
	w.connectCount++
	w.lastConnect = time.Now()
	
	// Koneksi dengan timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Buat channel untuk menangani hasil koneksi
	resultChan := make(chan error, 1)
	
	go func() {
		err := w.Client.Connect()
		if err != nil {
			// Toleransi untuk error "already connected"
			if strings.Contains(strings.ToLower(err.Error()), "already") {
				resultChan <- nil
				return
			}
		}
		resultChan <- err
	}()
	
	// Tunggu hasil atau timeout
	select {
	case err := <-resultChan:
		if err != nil {
			w.Logger.Errorf("Gagal connect: %v", err)
			return err
		}
		w.Logger.Success("Berhasil terhubung ke WhatsApp!")
		return nil
		
	case <-ctx.Done():
		w.Logger.Error("Timeout saat mencoba connect")
		return fmt.Errorf("timeout koneksi")
	}
}

// ConnectWithQRSafely melakukan pairing QR dengan proteksi anti-spam  
func (w *WAManager) ConnectWithQRSafely(qrGen *QRCodeGenerator) error {
	// Cek rate limiting
	if err := w.checkRateLimit(); err != nil {
		return err
	}
	
	// Lock untuk mencegah race condition
	w.pairingMu.Lock()
	if w.pairingActive {
		w.pairingMu.Unlock()
		return fmt.Errorf("pairing sudah sedang berlangsung")
	}
	w.pairingActive = true
	w.pairingMu.Unlock()
	
	defer func() {
		w.pairingMu.Lock()
		w.pairingActive = false
		w.pairingMu.Unlock()
	}()
	
	w.Logger.Info("Memulai proses pairing dengan QR code...")
	
	// Update counter dan timestamp
	w.connectCount++
	w.lastConnect = time.Now()
	
	// Dapatkan QR channel dengan background context agar tidak timeout
	qrChan, err := w.Client.GetQRChannel(context.Background())
	if err != nil {
		return fmt.Errorf("gagal membuat QR channel: %w", err)
	}
	
	// Mulai koneksi di goroutine terpisah
	go func() {
		if err := w.Client.Connect(); err != nil {
			w.Logger.Errorf("Gagal start connect: %v", err)
		}
	}()
	
	// Tunggu dan handle QR events dengan timeout yang lebih panjang
	timeout := time.After(3 * time.Minute) // 3 menit untuk scan QR
	
	for {
		select {
		case evt, ok := <-qrChan:
			if !ok {
				return fmt.Errorf("QR channel tertutup")
			}
			
			switch evt.Event {
			case "code":
				w.Logger.Info("QR code diterima, menampilkan...")
				err = qrGen.GenerateAndDisplay(evt.Code)
				if err != nil {
					w.Logger.Errorf("Gagal menampilkan QR: %v", err)
					w.Logger.Infof("QR Code (text): %s", evt.Code)
				}
				
			case "success":
				w.Logger.Success("QR code berhasil di-scan! Login berhasil.")
				w.resetConnectCounter()
				return nil
				
			case "timeout":
				w.Logger.Warning("QR code timeout, generating QR code baru...")
				
			case "error":
				w.Logger.Error("Error dalam proses login QR code")
				return fmt.Errorf("QR code login error")
				
			default:
				w.Logger.Debugf("QR event: %s", evt.Event)
			}
			
		case <-timeout:
			w.Logger.Error("Timeout menunggu scan QR code (3 menit)")
			return fmt.Errorf("timeout QR scan")
		}
	}
}

// checkRateLimit mengecek apakah kita sudah terlalu sering mencoba connect
func (w *WAManager) checkRateLimit() error {
	now := time.Now()
	
	// Reset counter jika sudah melewati interval reset
	if now.Sub(w.lastConnect) > w.ResetInterval {
		w.resetConnectCounter()
		return nil
	}
	
	// Cek apakah sudah mencapai batas maksimal
	if w.connectCount >= w.MaxConnectCount {
		waitTime := w.ResetInterval - now.Sub(w.lastConnect)
		return fmt.Errorf("terlalu banyak percobaan koneksi, tunggu %v lagi", waitTime.Round(time.Second))
	}
	
	// Cek minimum delay antar koneksi
	if now.Sub(w.lastConnect) < w.MinConnectDelay {
		waitTime := w.MinConnectDelay - now.Sub(w.lastConnect)
		w.Logger.Infof("Menunggu %v sebelum koneksi berikutnya...", waitTime.Round(time.Second))
		time.Sleep(waitTime)
	}
	
	return nil
}

// resetConnectCounter mereset counter koneksi
func (w *WAManager) resetConnectCounter() {
	w.connectCount = 0
	w.lastConnect = time.Time{}
}

// IsLoggedIn mengecek apakah sudah login
func (w *WAManager) IsLoggedIn() bool {
	return w.Client != nil && w.Client.Store.ID != nil
}

// IsConnected mengecek apakah sedang terhubung
func (w *WAManager) IsConnected() bool {
	return w.Client != nil && w.Client.IsConnected()
}

// Disconnect memutuskan koneksi dengan aman
func (w *WAManager) Disconnect() {
	if w.Client != nil {
		w.Logger.Info("Memutuskan koneksi WhatsApp...")
		w.Client.Disconnect()
	}
}

// PairByPhone melakukan pairing via nomor telepon dengan aman
func (w *WAManager) PairByPhone(ctx context.Context, client *whatsmeow.Client, phone string) (string, error) {
	w.phonePairMu.Lock()
	defer w.phonePairMu.Unlock()

	// Pastikan tidak ada koneksi aktif
	if client != nil && client.IsConnected() {
		w.Logger.Info("Memutuskan koneksi aktif sebelum phone pairing...")
		client.Disconnect()
		// Tunggu sedikit agar socket benar-benar tertutup
		time.Sleep(2 * time.Second)
	}

	// Coba logout jika masih ada session login (ignore error jika websocket sudah off)
	if client != nil && client.Store != nil && client.Store.ID != nil {
		w.Logger.Info("Melakukan logout sebelum phone pairing...")
		_ = client.Logout(ctx)
		// Tunggu sejenak
		time.Sleep(1 * time.Second)
	}

	// Bangun client baru dari store agar state fresh
	deviceStore, err := w.Container.GetFirstDevice(ctx)
	if err != nil {
		return "", fmt.Errorf("gagal mengambil device store: %w", err)
	}
	
	clientLog := waLog.Stdout("WhatsApp", "INFO", true)
	freshClient := whatsmeow.NewClient(deviceStore, clientLog)

	// Setup QR channel BEFORE connect, then connect
	qrChan, err := freshClient.GetQRChannel(ctx)
	if err != nil {
		w.Logger.Errorf("Gagal membuat QR channel sebelum connect: %v", err)
	}
	go func() {
		_ = freshClient.Connect()
	}()
	// Wait for first QR event or timeout (pre-login ready)
	preloginReady := make(chan struct{}, 1)
	go func() {
		select {
		case <-qrChan:
			preloginReady <- struct{}{}
		case <-time.After(2 * time.Second):
			preloginReady <- struct{}{}
		}
	}()
	<-preloginReady

	// Request pairing code
	w.Logger.Infof("Meminta pairing code untuk nomor: %s", phone)
	code, err := freshClient.PairPhone(ctx, phone, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		return "", fmt.Errorf("gagal request pairing code: %w", err)
	}

	w.Logger.Successf("Pairing code diterima: %s", code)
	return code, nil
}

// GetStats mengembalikan statistik koneksi
func (w *WAManager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"logged_in":     w.IsLoggedIn(),
		"connected":     w.IsConnected(),
		"connect_count": w.connectCount,
		"last_connect":  w.lastConnect,
		"pairing_active": w.pairingActive,
	}
}