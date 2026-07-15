package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nabilulilalbab/promote/utils"
)

// TestHandleQRImage_ReportsLastError memastikan endpoint qr-image mengembalikan
// error JSON (bukan "belum tersedia") ketika ada lastError di dashboardQR.
func TestHandleQRImage_ReportsLastError(t *testing.T) {
	dqh := &utils.DashboardQRHandler{}
	dqh.SetLastError("❌ QR pairing gagal: err-client-outdated")

	s := &DashboardServer{
		logger:      utils.NewLogger("TEST", false),
		dashboardQR: dqh,
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handleQRImage(w, r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/whatsapp/qr-image", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("cannot decode response: %v", err)
	}
	if body["success"] != false {
		t.Fatalf("expected success=false, got %v", body["success"])
	}
	if body["message"] == "" || body["message"] == "QR code belum tersedia, mulai pairing dulu" {
		t.Fatalf("expected error message forwarded, got %q", body["message"])
	}
}

// TestHandleQRImage_NoError_NoQR memastikan ketika tidak ada error dan tidak ada QR,
// response tetap "belum tersedia" (bukan crash).
func TestHandleQRImage_NoError_NoQR(t *testing.T) {
	dqh := &utils.DashboardQRHandler{} // lastError kosong, currentQRCode kosong
	s := &DashboardServer{
		logger:      utils.NewLogger("TEST", false),
		dashboardQR: dqh,
		qrGenerator: utils.NewQRCodeGenerator("/dev/null"),
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handleQRImage(w, r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/whatsapp/qr-image", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("cannot decode response: %v", err)
	}
	if body["success"] != false {
		t.Fatalf("expected success=false when no QR, got %v", body["success"])
	}
}

// TestHandleGetPairingCode_ReportsPhoneError memastikan endpoint pairing-code
// mengembalikan error dari phone pairing (bukan "belum tersedia" generic).
func TestHandleGetPairingCode_ReportsPhoneError(t *testing.T) {
	s := &DashboardServer{
		logger:                utils.NewLogger("TEST", false),
		lastPhonePairingError: "❌ Gagal pairing: websocket not connected",
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handleGetPairingCode(w, r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/whatsapp/pairing-code", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("cannot decode response: %v", err)
	}
	if body["success"] != false {
		t.Fatalf("expected success=false, got %v", body["success"])
	}
	if body["error"] != true {
		t.Fatalf("expected error=true, got %v", body["error"])
	}
	if body["message"] == "Pairing code belum tersedia" {
		t.Fatalf("expected specific phone error, got generic message")
	}
}

// TestHandleGetPairingCode_NoError_NoCode memastikan generic "belum tersedia"
// ketika tidak ada error dan belum ada code.
func TestHandleGetPairingCode_NoError_NoCode(t *testing.T) {
	s := &DashboardServer{logger: utils.NewLogger("TEST", false)}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handleGetPairingCode(w, r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/whatsapp/pairing-code", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("cannot decode: %v", err)
	}
	if body["success"] != false {
		t.Fatalf("expected success=false")
	}
	if msg, _ := body["message"].(string); msg != "Pairing code belum tersedia" {
		t.Fatalf("expected generic message, got %q", msg)
	}
}

// TestHandleGetPairingCode_CodeAvailable memastikan ketika code tersedia, dikembalikan.
func TestHandleGetPairingCode_CodeAvailable(t *testing.T) {
	s := &DashboardServer{
		logger:             utils.NewLogger("TEST", false),
		currentPairingCode: "ABCD-1234",
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handleGetPairingCode(w, r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/whatsapp/pairing-code", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("cannot decode: %v", err)
	}
	if body["success"] != true {
		t.Fatalf("expected success=true, got %v", body["success"])
	}
	if body["pairing_code"] != "ABCD-1234" {
		t.Fatalf("expected code ABCD-1234, got %v", body["pairing_code"])
	}
}
