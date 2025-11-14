// Package config berisi konfigurasi untuk WhatsApp Bot
// File ini mengatur semua pengaturan dasar bot seperti database path, log level, dll
package config

import (
	"os"
	"strconv"
)

// Config adalah struktur yang menyimpan semua konfigurasi bot
type Config struct {
	// DatabasePath adalah lokasi file database SQLite untuk menyimpan session WhatsApp
	DatabasePath string
	
	// LogLevel menentukan level logging (DEBUG, INFO, WARN, ERROR)
	LogLevel string
	
	// QRCodePath adalah lokasi file QR code PNG akan disimpan
	QRCodePath string
	
	// AutoReplyPersonal menentukan apakah bot otomatis membalas chat personal
	AutoReplyPersonal bool
	
	// AutoReplyGroup menentukan apakah bot otomatis membalas chat grup
	// PENTING: Set false jika Anda ada di banyak grup untuk menghindari spam
	AutoReplyGroup bool
	
	// EnableAntiSpamMode mengaktifkan mode anti-spam yang lebih ketat
	EnableAntiSpamMode bool
	
	// ConnectionRetryDelay waktu tunggu antar percobaan koneksi (detik)
	ConnectionRetryDelay int
	
	// MaxConnectionRetries jumlah maksimal percobaan koneksi
	MaxConnectionRetries int
}

// NewConfig membuat konfigurasi default untuk bot
// Fungsi ini akan dipanggil saat bot pertama kali dijalankan
func NewConfig() *Config {
	return &Config{
		// Database akan disimpan di folder data
		DatabasePath: getEnvOrDefault("DB_PATH", "data/session.db"),
		
		// Log level default adalah INFO (tidak terlalu verbose)
		LogLevel: getEnvOrDefault("LOG_LEVEL", "INFO"),
		
		// QR code akan disimpan di folder data
		QRCodePath: getEnvOrDefault("QR_PATH", "data/qrcode.png"),
		
		// Auto reply untuk chat personal diaktifkan
		AutoReplyPersonal: getEnvBoolOrDefault("AUTO_REPLY_PERSONAL", true),
		
		// Auto reply untuk grup DIMATIKAN untuk menghindari spam
		// Anda bisa mengubah ini ke true jika ingin bot membalas di grup
		AutoReplyGroup: getEnvBoolOrDefault("AUTO_REPLY_GROUP", false),
		
		// Mode anti-spam aktif secara default untuk mencegah deteksi spam
		EnableAntiSpamMode: getEnvBoolOrDefault("ENABLE_ANTISPAM", true),
		
		// Delay 10 detik antar percobaan koneksi (default aman)
		ConnectionRetryDelay: getEnvIntOrDefault("CONNECTION_RETRY_DELAY", 10),
		
		// Maksimal 3 percobaan koneksi untuk mencegah spam
		MaxConnectionRetries: getEnvIntOrDefault("MAX_CONNECTION_RETRIES", 3),
	}
}

// getEnvOrDefault mengambil nilai dari environment variable atau menggunakan default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBoolOrDefault mengambil nilai boolean dari environment variable atau menggunakan default
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

// getEnvIntOrDefault mengambil nilai integer dari environment variable atau menggunakan default
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}