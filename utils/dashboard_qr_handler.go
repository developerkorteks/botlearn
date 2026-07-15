package utils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
)

// DashboardQRHandler handles QR pairing for dashboard
type DashboardQRHandler struct {
	client       *whatsmeow.Client
	logger       *Logger
	qrGenerator  *QRCodeGenerator
	
	// State management
	currentQRCode string
	qrMutex       sync.RWMutex
	isActive      bool
	activeMutex   sync.RWMutex

	// lastError menyimpan pesan error terakhir dari proses pairing
	// (mis. "err-client-outdated" / 405) agar bisa dilaporkan ke dashboard,
	// bukan hanya dicetak ke log.
	lastError string
	lastErrMu sync.RWMutex

	// cancel context for QR pairing
	pairCtx    context.Context
	pairCancel context.CancelFunc
}

// NewDashboardQRHandler creates new dashboard QR handler
func NewDashboardQRHandler(client *whatsmeow.Client, logger *Logger, qrGen *QRCodeGenerator) *DashboardQRHandler {
	return &DashboardQRHandler{
		client:      client,
		logger:      logger,
		qrGenerator: qrGen,
	}
}

// GetCurrentQRCode returns current QR code
func (dqh *DashboardQRHandler) GetCurrentQRCode() string {
	dqh.qrMutex.RLock()
	defer dqh.qrMutex.RUnlock()
	return dqh.currentQRCode
}

// SetLastError menyimpan pesan error pairing terakhir.
func (dqh *DashboardQRHandler) SetLastError(msg string) {
	dqh.lastErrMu.Lock()
	defer dqh.lastErrMu.Unlock()
	dqh.lastError = msg
}

// GetLastError mengembalikan pesan error pairing terakhir (kosong bila tidak ada).
func (dqh *DashboardQRHandler) GetLastError() string {
	dqh.lastErrMu.RLock()
	defer dqh.lastErrMu.RUnlock()
	return dqh.lastError
}

// SetQRCode sets current QR code
func (dqh *DashboardQRHandler) SetQRCode(qrCode string) {
	dqh.qrMutex.Lock()
	defer dqh.qrMutex.Unlock()
	dqh.currentQRCode = qrCode
}

// ClearQRCode clears current QR code
func (dqh *DashboardQRHandler) ClearQRCode() {
	dqh.qrMutex.Lock()
	defer dqh.qrMutex.Unlock()
	dqh.currentQRCode = ""
}

// IsActive returns if QR pairing is active
func (dqh *DashboardQRHandler) IsActive() bool {
	dqh.activeMutex.RLock()
	defer dqh.activeMutex.RUnlock()
	return dqh.isActive
}

// Cancel cancels active QR pairing
func (dqh *DashboardQRHandler) Cancel() {
	dqh.activeMutex.Lock()
	defer dqh.activeMutex.Unlock()
	if dqh.isActive {
		if dqh.pairCancel != nil {
			dqh.pairCancel()
		}
		dqh.isActive = false
		dqh.ClearQRCode()
		dqh.logger.Info("🛑 QR pairing dibatalkan via dashboard")
	}
}

// StartDashboardQRPairing starts QR pairing optimized for dashboard
func (dqh *DashboardQRHandler) StartDashboardQRPairing() error {
	dqh.activeMutex.Lock()
	if dqh.isActive {
		dqh.activeMutex.Unlock()
		return fmt.Errorf("QR pairing sudah aktif")
	}
	dqh.isActive = true
	dqh.activeMutex.Unlock()

	// Clear any previous QR code and error state
	dqh.ClearQRCode()
	dqh.SetLastError("")

	dqh.logger.Info("🔄 Memulai QR pairing untuk dashboard...")

	// Get QR channel first (before connecting)
	qrChan, err := dqh.client.GetQRChannel(context.Background())
	if err != nil {
		dqh.activeMutex.Lock()
		dqh.isActive = false
		dqh.activeMutex.Unlock()
		return fmt.Errorf("gagal membuat QR channel: %w", err)
	}

	// Start connection in background AFTER getting QR channel
	go func() {
		if err := dqh.client.Connect(); err != nil {
			dqh.logger.Errorf("Gagal start connect: %v", err)
		}
	}()

	// Handle QR events
	go func() {
		defer func() {
			dqh.activeMutex.Lock()
			dqh.isActive = false
			dqh.activeMutex.Unlock()
			dqh.ClearQRCode()
			if dqh.pairCancel != nil {
				dqh.pairCancel()
				dqh.pairCancel = nil
			}
		}()

		timeout := time.After(3 * time.Minute) // 3 menit timeout
		
		for {
			select {
			case evt, ok := <-qrChan:
					if !ok {
						// Channel ditutup — bisa karena error (mis. 405) atau selesai
						if dqh.GetLastError() == "" {
							dqh.SetLastError("QR channel tertutup tanpa berhasil — cek koneksi / versi library")
						}
						dqh.logger.Warning("QR channel tertutup")
						return
					}

					switch evt.Event {
					case "code":
						// QR code baru → reset error state, tampilkan ke dashboard & terminal
						dqh.SetLastError("")
						dqh.SetQRCode(evt.Code)
						dqh.logger.Infof("📱 QR code baru tersedia di dashboard: %s", evt.Code[:50]+"...")

						err := dqh.qrGenerator.GenerateAndDisplay(evt.Code)
						if err != nil {
							dqh.logger.Errorf("Gagal tampilkan QR di terminal: %v", err)
						} else {
							dqh.logger.Info("✅ QR code juga disimpan untuk dashboard")
						}

					case "success":
						dqh.SetLastError("")
						dqh.logger.Success("✅ QR pairing berhasil!")
						return

					case "timeout":
						// Timeout WA: QR kadaluarsa, whatsmeow akan kirim code baru
						dqh.logger.Warning("⏰ QR code timeout, generating QR baru...")

					case "error":
						dqh.SetLastError("❌ QR pairing error dari server WhatsApp")
						dqh.logger.Error("❌ Error dalam QR pairing")
						return

					default:
						// Tangkap event tidak dikenal — mis. "err-client-outdated" (405)
						errMsg := fmt.Sprintf("❌ QR pairing gagal: %s", evt.Event)
						dqh.SetLastError(errMsg)
						dqh.logger.Debugf("QR event: %s", evt.Event)
					}
				
			case <-timeout:
					dqh.SetLastError("⏰ QR pairing timeout (3 menit) — coba lagi")
					dqh.logger.Error("⏰ QR pairing timeout (3 menit)")
					return
			}
		}
	}()

	return nil
}