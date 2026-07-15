package web

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupSessionDB_RenamesExisting(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("data", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("data/session.db", []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := &DashboardServer{}
	backupPath, err := s.backupSessionDB()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backupPath == "" {
		t.Fatal("expected non-empty backup path")
	}
	if filepath.Base(backupPath) == "session.db" {
		t.Fatalf("backup path should be a .old file, got %q", backupPath)
	}

	if _, statErr := os.Stat(backupPath); statErr != nil {
		t.Fatalf("backup file missing: %v", statErr)
	}
	if _, statErr := os.Stat("data/session.db"); !os.IsNotExist(statErr) {
		t.Fatalf("original session.db should be removed after rename: %v", statErr)
	}
}

func TestBackupSessionDB_NoFileReturnsEmpty(t *testing.T) {
	t.Chdir(t.TempDir())

	s := &DashboardServer{}
	backupPath, err := s.backupSessionDB()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backupPath != "" {
		t.Fatalf("expected empty backup path when no session, got %q", backupPath)
	}
}

func TestBackupSessionDB_ReadsContent(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("data", 0o755); err != nil {
		t.Fatal(err)
	}
	want := []byte("secret-session-data")
	if err := os.WriteFile("data/session.db", want, 0o644); err != nil {
		t.Fatal(err)
	}

	s := &DashboardServer{}
	backupPath, err := s.backupSessionDB()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("cannot read backup: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("backup content mismatch: got %q want %q", got, want)
	}
}
