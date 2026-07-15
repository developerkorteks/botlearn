package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nabilulilalbab/promote/utils"
)

func TestHandleQRImage_OptionsAllowed(t *testing.T) {
	s := &DashboardServer{logger: utils.NewLogger("TEST", false)}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handleQRImage(w, r)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/whatsapp/qr-image", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for OPTIONS, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET" {
		t.Fatalf("expected Allow-Methods GET, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected Allow-Origin *, got %q", got)
	}
}

func TestHandleQRImage_NonGetRejected(t *testing.T) {
	s := &DashboardServer{logger: utils.NewLogger("TEST", false)}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handleQRImage(w, r)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/qr-image", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for POST, got %d", rec.Code)
	}
}
