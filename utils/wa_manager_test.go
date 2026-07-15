package utils

import (
	"context"
	"testing"
	"time"

	"go.mau.fi/whatsmeow"
)

// TestPairByPhone_Timeout_EarlyExit memastikan PairByPhone mengembalikan error
// secara langsung jika context sudah di-cancel, mencegah proses hang yang tidak perlu.
func TestPairByPhone_Timeout_EarlyExit(t *testing.T) {
	// Buat manager sederhana
	logger := NewLogger("TEST", false)
	// Kita tidak butuh koneksi database asli karena context akan di-cancel duluan
	// Tetapi w.Container diperlukan oleh PairByPhone (walau early exit ada sebelum itu)
	// karena struct dereference. Oh tunggu, early exit diletakkan sebelum lock.
	
	manager := &WAManager{
		Logger: logger,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	cancel() // Langsung cancel

	// Panggil
	client := &whatsmeow.Client{}
	code, err := manager.PairByPhone(ctx, client, "628111")
	
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	if code != "" {
		t.Fatalf("expected empty code, got %q", code)
	}
	if err != context.Canceled && err != context.DeadlineExceeded {
		t.Fatalf("expected context canceled error, got: %v", err)
	}
}
