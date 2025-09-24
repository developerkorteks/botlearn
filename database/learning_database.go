// Package database - learning database initialization
package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// InitializeLearningDatabase menginisialisasi database untuk learning bot
func InitializeLearningDatabase(dbPath string) (*sql.DB, Repository, error) {
	// Buka koneksi database
	db, err := sql.Open("sqlite3", "file:"+dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open learning database: %v", err)
	}

	// Test koneksi
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to ping learning database: %v", err)
	}

	// Jalankan migrasi learning
	if err := RunLearningMigrations(db); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to run learning migrations: %v", err)
	}

	// Buat repository
	repo := NewSQLiteRepository(db)

	return db, repo, nil
}