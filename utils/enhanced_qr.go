package utils

import (
	"context"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
)

// EnhancedQRManager manages QR codes with dashboard integration
type EnhancedQRManager struct {
	client         *whatsmeow.Client
	logger         *Logger
	currentQRCode  string
	qrMutex        sync.RWMutex
	onQRCodeUpdate func(string) // Callback when QR code updates
}

// NewEnhancedQRManager creates a new enhanced QR manager
func NewEnhancedQRManager(client *whatsmeow.Client, logger *Logger) *EnhancedQRManager {
	return &EnhancedQRManager{
		client: client,
		logger: logger,
	}
}

// SetQRUpdateCallback sets callback function for QR code updates
func (eq *EnhancedQRManager) SetQRUpdateCallback(callback func(string)) {
	eq.onQRCodeUpdate = callback
}

// GetCurrentQRCode returns the current QR code
func (eq *EnhancedQRManager) GetCurrentQRCode() string {
	eq.qrMutex.RLock()
	defer eq.qrMutex.RUnlock()
	return eq.currentQRCode
}

// StartQRPairingWithCallback starts QR pairing with real-time updates
func (eq *EnhancedQRManager) StartQRPairingWithCallback() error {
	// Get QR channel
	qrChan, err := eq.client.GetQRChannel(context.Background())
	if err != nil {
		return err
	}

	// Start connection in background
	go func() {
		if err := eq.client.Connect(); err != nil {
			eq.logger.Errorf("Failed to start connect: %v", err)
		}
	}()

	// Handle QR events
	go func() {
		timeout := time.After(3 * time.Minute)
		
		for {
			select {
			case evt, ok := <-qrChan:
				if !ok {
					eq.logger.Error("QR channel closed")
					return
				}
				
				switch evt.Event {
				case "code":
					eq.qrMutex.Lock()
					eq.currentQRCode = evt.Code
					eq.qrMutex.Unlock()
					
					eq.logger.Info("🔄 QR code updated")
					
					// Call callback if set
					if eq.onQRCodeUpdate != nil {
						eq.onQRCodeUpdate(evt.Code)
					}
					
				case "success":
					eq.logger.Success("✅ QR code pairing berhasil!")
					eq.qrMutex.Lock()
					eq.currentQRCode = ""
					eq.qrMutex.Unlock()
					return
					
				case "timeout":
					eq.logger.Warning("⏰ QR code timeout, generating new...")
					
				case "error":
					eq.logger.Error("❌ Error dalam QR pairing")
					return
				}
				
			case <-timeout:
				eq.logger.Error("⏰ QR pairing timeout (3 menit)")
				return
			}
		}
	}()

	return nil
}