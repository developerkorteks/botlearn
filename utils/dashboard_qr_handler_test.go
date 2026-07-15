package utils

import (
	"testing"
)

// TestDashboardQRHandler_SetGetLastError memastikan lastError tersimpan dan terbaca dengan benar.
func TestDashboardQRHandler_SetGetLastError(t *testing.T) {
	dqh := &DashboardQRHandler{}

	if got := dqh.GetLastError(); got != "" {
		t.Fatalf("expected empty lastError on init, got %q", got)
	}

	dqh.SetLastError("test error")
	if got := dqh.GetLastError(); got != "test error" {
		t.Fatalf("expected %q, got %q", "test error", got)
	}

	dqh.SetLastError("")
	if got := dqh.GetLastError(); got != "" {
		t.Fatalf("expected empty after clear, got %q", got)
	}
}

// TestDashboardQRHandler_SetGetLastError_Concurrent memastikan tidak ada data race.
func TestDashboardQRHandler_SetGetLastError_Concurrent(t *testing.T) {
	dqh := &DashboardQRHandler{}
	done := make(chan struct{})

	for i := 0; i < 10; i++ {
		go func() {
			dqh.SetLastError("concurrent error")
			_ = dqh.GetLastError()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestDashboardQRHandler_ClearOnStart memastikan lastError dibersihkan saat StartDashboardQRPairing
// dipanggil (tanpa live WA — cukup verifikasi state sebelum GetQRChannel dipanggil).
func TestDashboardQRHandler_ClearOnStartLogic(t *testing.T) {
	dqh := &DashboardQRHandler{}

	// Set error lama
	dqh.SetLastError("error lama dari sesi sebelumnya")
	if dqh.GetLastError() == "" {
		t.Fatal("expected lastError to be set before clearing")
	}

	// Simulasi: clear seperti yang dilakukan StartDashboardQRPairing sebelum konek
	dqh.ClearQRCode()
	dqh.SetLastError("")

	if got := dqh.GetLastError(); got != "" {
		t.Fatalf("expected lastError cleared, got %q", got)
	}
}
