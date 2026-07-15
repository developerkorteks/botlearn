package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"go.mau.fi/whatsmeow"

	"github.com/nabilulilalbab/promote/utils"
)

func TestHandlePhonePairing_InvalidPhone(t *testing.T) {
	s := &DashboardServer{
		logger:      utils.NewLogger("TEST", false),
		waManager:   &utils.WAManager{},
		whatsappClient: &whatsmeow.Client{},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handlePhonePairing(w, r)
	})

	tests := []struct {
		name       string
		phone      string
		wantSuccess bool
		wantSubstring string
	}{
		{"too short", "0812", false, "Format nomor telepon tidak valid"},
		{"too long", strings.Repeat("1", 16), false, "Format nomor telepon tidak valid"},
		{"with plus sanitized", "+6281111111111", true, "Pairing code generation started"},
		{"with spaces sanitized", "6281 1111 1111", true, "Pairing code generation started"},
		{"valid format", "6281111111111", true, "Pairing code generation started"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"phone_number":"` + tt.phone + `"}`
			req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/phone", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			var res map[string]interface{}
			if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if res["success"] != tt.wantSuccess {
				t.Fatalf("success=%v, want %v for phone %q", res["success"], tt.wantSuccess, tt.phone)
			}
			msg, _ := res["message"].(string)
			if !strings.Contains(msg, tt.wantSubstring) {
				t.Fatalf("message=%q, want substring %q", msg, tt.wantSubstring)
			}
		})
	}
}

func TestHandlePhonePairing_MissingWhatsAppClient(t *testing.T) {
	s := &DashboardServer{
		logger:      utils.NewLogger("TEST", false),
		waManager:   &utils.WAManager{},
		whatsappClient: nil,
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handlePhonePairing(w, r)
	})

	body := `{"phone_number":"085117557905"}`
	req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/phone", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var res map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res["success"] != false {
		t.Fatalf("expected success=false when whatsappClient is nil")
	}
	msg, _ := res["message"].(string)
	if !strings.Contains(msg, "WhatsApp client tidak tersedia") {
		t.Fatalf("expected client unavailable message, got %q", msg)
	}
}

func TestHandlePhonePairing_MissingWAManager(t *testing.T) {
	s := &DashboardServer{
		logger:      utils.NewLogger("TEST", false),
		waManager:   nil,
		whatsappClient: &whatsmeow.Client{},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handlePhonePairing(w, r)
	})

	body := `{"phone_number":"085117557905"}`
	req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/phone", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var res map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res["success"] != true {
		t.Fatalf("expected success=true when request accepted but background goroutine handles nil waManager")
	}
	msg, _ := res["message"].(string)
	if !strings.Contains(msg, "Pairing code generation started") {
		t.Fatalf("expected started message, got %q", msg)
	}
}

func TestHandleLogout_ReturnsAcceptedResponse(t *testing.T) {
	s := &DashboardServer{
		logger:      utils.NewLogger("TEST", false),
		whatsappClient: &whatsmeow.Client{},
		waManager:   &utils.WAManager{},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handleLogout(w, r)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/logout", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var res map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res["success"] != true {
		t.Fatalf("expected success=true for logout initiation")
	}
	msg, _ := res["message"].(string)
	if !strings.Contains(msg, "Safe logout initiated") {
		t.Fatalf("expected safe logout message, got %q", msg)
	}
}

func TestHandleLogout_NilClient(t *testing.T) {
	s := &DashboardServer{
		logger:      utils.NewLogger("TEST", false),
		whatsappClient: nil,
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handleLogout(w, r)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/logout", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var res map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res["success"] != false {
		t.Fatalf("expected success=false when whatsappClient is nil")
	}
}

func TestHandleFullReset_CleansSessionDB(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	if err := os.MkdirAll("data", 0o755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}
	if err := os.WriteFile("data/session.db", []byte("fake"), 0o644); err != nil {
		t.Fatalf("write session.db: %v", err)
	}
	if err := os.WriteFile("data/session.db-wal", []byte("wal"), 0o644); err != nil {
		t.Fatalf("write wal: %v", err)
	}
	if err := os.WriteFile("data/session.db-shm", []byte("shm"), 0o644); err != nil {
		t.Fatalf("write shm: %v", err)
	}

	s := &DashboardServer{
		logger:      utils.NewLogger("TEST", false),
		dashboardQR: &utils.DashboardQRHandler{},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/full_reset", nil)
	rec := httptest.NewRecorder()
	s.handleFullReset(rec, req)

	var res map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res["success"] != true {
		t.Fatalf("expected success=true for full reset, got %v", res["success"])
	}

	if _, err := os.Stat("data/session.db"); !os.IsNotExist(err) {
		t.Fatalf("expected session.db to be removed after full reset")
	}
	if _, err := os.Stat("data/session.db-wal"); !os.IsNotExist(err) {
		t.Fatalf("expected session.db-wal to be removed after full reset")
	}
	if _, err := os.Stat("data/session.db-shm"); !os.IsNotExist(err) {
		t.Fatalf("expected session.db-shm to be removed after full reset")
	}
}
