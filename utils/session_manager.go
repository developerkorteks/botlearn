package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SessionManager manages WhatsApp session databases with corruption prevention
type SessionManager struct {
	sessionPath   string
	backupPath    string
	logger        *Logger
	lastBackup    time.Time
	backupInterval time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(sessionPath string, logger *Logger) *SessionManager {
	backupPath := filepath.Dir(sessionPath) + "/session_backups"
	return &SessionManager{
		sessionPath:    sessionPath,
		backupPath:     backupPath,
		logger:         logger,
		backupInterval: 1 * time.Hour, // Backup every hour
	}
}

// InitSession initializes session with safety checks
func (sm *SessionManager) InitSession() error {
	// Create backup directory if not exists
	if err := os.MkdirAll(sm.backupPath, 0755); err != nil {
		sm.logger.Errorf("Failed to create backup directory: %v", err)
	}
	
	// Create session directory if not exists
	sessionDir := filepath.Dir(sm.sessionPath)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}
	
	// Check if session exists and validate it
	if sm.sessionExists() {
		if sm.isSessionHealthy() {
			sm.logger.Info("✅ Existing session is healthy")
			return nil
		} else {
			sm.logger.Warning("⚠️ Session appears corrupted, backing up and resetting...")
			return sm.resetCorruptedSession()
		}
	}
	
	sm.logger.Info("🆕 Creating fresh session database")
	return nil
}

// sessionExists checks if session database exists
func (sm *SessionManager) sessionExists() bool {
	_, err := os.Stat(sm.sessionPath)
	return err == nil
}

// isSessionHealthy performs basic health checks on session database
func (sm *SessionManager) isSessionHealthy() bool {
	// Check if file is readable
	file, err := os.Open(sm.sessionPath)
	if err != nil {
		sm.logger.Debugf("Session file not readable: %v", err)
		return false
	}
	defer file.Close()
	
	// Check file size (corrupted sessions often have unusual sizes)
	info, err := file.Stat()
	if err != nil {
		return false
	}
	
	// Session file too small (less than 1KB) is suspicious
	if info.Size() < 1024 {
		sm.logger.Debug("Session file too small, might be corrupted")
		return false
	}
	
	// Session file too large (more than 100MB) is also suspicious
	if info.Size() > 100*1024*1024 {
		sm.logger.Debug("Session file too large, might be corrupted")
		return false
	}
	
	return true
}

// resetCorruptedSession backs up and resets corrupted session
func (sm *SessionManager) resetCorruptedSession() error {
	// Backup corrupted session for analysis
	backupFile := fmt.Sprintf("%s/corrupted_session_%s.db", 
		sm.backupPath, time.Now().Format("20060102_150405"))
	
	if err := sm.copyFile(sm.sessionPath, backupFile); err != nil {
		sm.logger.Errorf("Failed to backup corrupted session: %v", err)
	} else {
		sm.logger.Infof("💾 Corrupted session backed up to: %s", backupFile)
	}
	
	// Remove corrupted session
	if err := os.Remove(sm.sessionPath); err != nil {
		return fmt.Errorf("failed to remove corrupted session: %w", err)
	}
	
	sm.logger.Info("🗑️ Corrupted session removed, fresh session will be created")
	return nil
}

// BackupSessionIfNeeded creates backup if enough time has passed
func (sm *SessionManager) BackupSessionIfNeeded() {
	if !sm.sessionExists() {
		return
	}
	
	if time.Since(sm.lastBackup) < sm.backupInterval {
		return
	}
	
	backupFile := fmt.Sprintf("%s/session_backup_%s.db", 
		sm.backupPath, time.Now().Format("20060102_150405"))
	
	if err := sm.copyFile(sm.sessionPath, backupFile); err != nil {
		sm.logger.Errorf("Failed to backup session: %v", err)
		return
	}
	
	sm.lastBackup = time.Now()
	sm.logger.Debugf("💾 Session backed up to: %s", backupFile)
	
	// Clean old backups (keep only last 5)
	sm.cleanOldBackups()
}

// cleanOldBackups removes old backup files, keeping only the most recent ones
func (sm *SessionManager) cleanOldBackups() {
	pattern := sm.backupPath + "/session_backup_*.db"
	backups, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	
	// Keep only last 5 backups
	if len(backups) > 5 {
		// Sort by modification time and remove oldest
		for i := 0; i < len(backups)-5; i++ {
			if err := os.Remove(backups[i]); err == nil {
				sm.logger.Debugf("🗑️ Removed old backup: %s", backups[i])
			}
		}
	}
}

// copyFile copies a file from source to destination
func (sm *SessionManager) copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	return os.WriteFile(dst, input, 0644)
}

// GetSessionStats returns session file statistics
func (sm *SessionManager) GetSessionStats() map[string]interface{} {
	stats := map[string]interface{}{
		"exists": sm.sessionExists(),
		"healthy": false,
		"size": 0,
		"last_modified": "",
	}
	
	if sm.sessionExists() {
		info, err := os.Stat(sm.sessionPath)
		if err == nil {
			stats["size"] = info.Size()
			stats["last_modified"] = info.ModTime().Format("2006-01-02 15:04:05")
			stats["healthy"] = sm.isSessionHealthy()
		}
	}
	
	return stats
}