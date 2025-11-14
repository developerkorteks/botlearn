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

	// Clear any previous QR code
	dqh.ClearQRCode()

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
					dqh.logger.Warning("QR channel tertutup")
					return
				}
				
				switch evt.Event {
				case "code":
					// Set QR code untuk dashboard
					dqh.SetQRCode(evt.Code)
					dqh.logger.Infof("📱 QR code baru tersedia di dashboard: %s", evt.Code[:50]+"...")
					
					// Juga tampilkan di terminal sebagai backup
					err := dqh.qrGenerator.GenerateAndDisplay(evt.Code)
					if err != nil {
						dqh.logger.Errorf("Gagal tampilkan QR di terminal: %v", err)
					} else {
						dqh.logger.Info("✅ QR code juga disimpan untuk dashboard")
					}
					
				case "success":
					dqh.logger.Success("✅ QR pairing berhasil!")
					return
					
				case "timeout":
					dqh.logger.Warning("⏰ QR code timeout, generating QR baru...")
					
				case "error":
					dqh.logger.Error("❌ Error dalam QR pairing")
					return
					
				default:
					dqh.logger.Debugf("QR event: %s", evt.Event)
				}
				
			case <-timeout:
				dqh.logger.Error("⏰ QR pairing timeout (3 menit)")
				return
			}
		}
	}()

	return nil
}