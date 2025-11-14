// Package utils - Human-like Typing Delay System
// File ini berisi sistem delay yang mensimulasikan typing manusia
package utils

import (
	"context"
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

// TypingSimulator mensimulasikan typing behavior manusia
type TypingSimulator struct {
	client *whatsmeow.Client
	logger *Logger
	
	// Konfigurasi timing
	CharsPerSecond    float64 // Karakter per detik (rata-rata manusia: 3-5)
	TypingVariation   float64 // Variasi kecepatan typing (0.0-1.0)
	PauseChance       float64 // Kemungkinan jeda saat typing (0.0-1.0)
	PauseDuration     time.Duration // Durasi jeda
	MinTypingTime     time.Duration // Minimum waktu typing
	MaxTypingTime     time.Duration // Maximum waktu typing
	
	// State management
	activeChats map[string]bool // Track typing status per chat
}

// NewTypingSimulator membuat typing simulator baru
func NewTypingSimulator(client *whatsmeow.Client, logger *Logger) *TypingSimulator {
	return &TypingSimulator{
		client: client,
		logger: logger,
		
		// Konfigurasi realistis manusia
		CharsPerSecond:    3.5, // Rata-rata 3.5 karakter per detik
		TypingVariation:   0.3, // 30% variasi kecepatan
		PauseChance:       0.15, // 15% chance untuk jeda
		PauseDuration:     800 * time.Millisecond, // Jeda 0.8 detik
		MinTypingTime:     1 * time.Second, // Minimum 1 detik
		MaxTypingTime:     15 * time.Second, // Maximum 15 detik
		
		activeChats: make(map[string]bool),
	}
}

// SimulateTyping mensimulasikan proses typing untuk pesan
func (ts *TypingSimulator) SimulateTyping(chatJID types.JID, message string) {
	chatKey := chatJID.String()
	
	// Prevent concurrent typing untuk chat yang sama
	if ts.activeChats[chatKey] {
		ts.logger.Debug("Typing sudah active untuk chat ini, skip...")
		return
	}
	
	ts.activeChats[chatKey] = true
	defer func() {
		ts.activeChats[chatKey] = false
	}()
	
	// Hitung durasi typing berdasarkan panjang pesan
	typingDuration := ts.calculateTypingDuration(message)
	
	ts.logger.Debugf("Simulasi typing untuk %d karakter (durasi: %v)", utf8.RuneCountInString(message), typingDuration)
	
	// Start typing indicator
	ts.startTyping(chatJID)
	
	// Simulasi typing dengan variasi dan jeda
	ts.simulateHumanTyping(chatJID, typingDuration)
	
	// Stop typing indicator
	ts.stopTyping(chatJID)
}

// calculateTypingDuration menghitung durasi typing berdasarkan panjang pesan
func (ts *TypingSimulator) calculateTypingDuration(message string) time.Duration {
	// Hitung jumlah karakter (termasuk emoji)
	charCount := float64(utf8.RuneCountInString(message))
	
	// Base duration berdasarkan kecepatan typing
	baseDuration := time.Duration(charCount/ts.CharsPerSecond) * time.Second
	
	// Tambahkan variasi random
	variation := 1.0 + (rand.Float64()-0.5)*ts.TypingVariation
	finalDuration := time.Duration(float64(baseDuration) * variation)
	
	// Apply min/max limits
	if finalDuration < ts.MinTypingTime {
		finalDuration = ts.MinTypingTime
	}
	if finalDuration > ts.MaxTypingTime {
		finalDuration = ts.MaxTypingTime
	}
	
	// Pesan panjang memerlukan waktu ekstra untuk "berpikir"
	if charCount > 100 {
		thinkingTime := time.Duration(charCount/50) * time.Millisecond
		finalDuration += thinkingTime
	}
	
	// Tambahkan waktu ekstra untuk pesan kompleks
	if strings.Contains(message, "http") || strings.Contains(message, "@") {
		finalDuration += 500 * time.Millisecond
	}
	
	return finalDuration
}

// simulateHumanTyping mensimulasikan typing behavior manusia dengan jeda
func (ts *TypingSimulator) simulateHumanTyping(chatJID types.JID, totalDuration time.Duration) {
	startTime := time.Now()
	segments := ts.calculateTypingSegments(totalDuration)
	
	for i, segment := range segments {
		// Cek apakah masih dalam durasi total
		elapsed := time.Since(startTime)
		if elapsed >= totalDuration {
			break
		}
		
		// Typing segment
		ts.logger.Debugf("Typing segment %d/%d (%v)", i+1, len(segments), segment.Duration)
		time.Sleep(segment.Duration)
		
		// Random pause (seperti berpikir atau koreksi typo)
		if segment.ShouldPause && i < len(segments)-1 {
			ts.logger.Debugf("Jeda berpikir (%v)", ts.PauseDuration)
			ts.stopTyping(chatJID) // Stop typing saat jeda
			time.Sleep(ts.PauseDuration)
			ts.startTyping(chatJID) // Resume typing
		}
	}
}

// TypingSegment represents a segment of typing activity
type TypingSegment struct {
	Duration    time.Duration
	ShouldPause bool
}

// calculateTypingSegments membagi durasi typing menjadi segmen-segmen natural
func (ts *TypingSimulator) calculateTypingSegments(totalDuration time.Duration) []TypingSegment {
	// Bagi menjadi 2-4 segmen untuk typing yang natural
	segmentCount := 2 + rand.Intn(3) // 2-4 segmen
	segmentDuration := totalDuration / time.Duration(segmentCount)
	
	segments := make([]TypingSegment, segmentCount)
	
	for i := 0; i < segmentCount; i++ {
		// Variasi durasi setiap segmen
		variation := 0.7 + rand.Float64()*0.6 // 0.7x - 1.3x
		duration := time.Duration(float64(segmentDuration) * variation)
		
		// Chance untuk pause setelah segmen (kecuali segmen terakhir)
		shouldPause := i < segmentCount-1 && rand.Float64() < ts.PauseChance
		
		segments[i] = TypingSegment{
			Duration:    duration,
			ShouldPause: shouldPause,
		}
	}
	
	return segments
}

// startTyping memulai typing indicator
func (ts *TypingSimulator) startTyping(chatJID types.JID) {
	err := ts.client.SendChatPresence(context.Background(), chatJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	if err != nil {
		ts.logger.Debugf("Gagal start typing indicator: %v", err)
	}
}

// stopTyping menghentikan typing indicator  
func (ts *TypingSimulator) stopTyping(chatJID types.JID) {
	err := ts.client.SendChatPresence(context.Background(), chatJID, types.ChatPresencePaused, types.ChatPresenceMediaText)
	if err != nil {
		ts.logger.Debugf("Gagal stop typing indicator: %v", err)
	}
}

// QuickDelay memberikan delay singkat untuk response cepat
func (ts *TypingSimulator) QuickDelay(chatJID types.JID) {
	// Delay 0.5 - 2 detik untuk response cepat
	duration := 500*time.Millisecond + time.Duration(rand.Intn(1500))*time.Millisecond
	
	ts.startTyping(chatJID)
	time.Sleep(duration)
	ts.stopTyping(chatJID)
}

// SmartTypingDelay memberikan delay berdasarkan kompleksitas pesan
func (ts *TypingSimulator) SmartTypingDelay(chatJID types.JID, message string) {
	// Analisis kompleksitas pesan
	complexity := ts.analyzeMessageComplexity(message)
	
	// Sesuaikan parameter berdasarkan kompleksitas
	originalSpeed := ts.CharsPerSecond
	
	switch complexity {
	case "simple":
		ts.CharsPerSecond = 4.5 // Typing lebih cepat untuk pesan sederhana
	case "medium":
		ts.CharsPerSecond = 3.5 // Normal speed
	case "complex":
		ts.CharsPerSecond = 2.5 // Typing lebih lambat untuk pesan kompleks
	}
	
	// Simulasi typing
	ts.SimulateTyping(chatJID, message)
	
	// Restore original speed
	ts.CharsPerSecond = originalSpeed
}

// analyzeMessageComplexity menganalisis kompleksitas pesan
func (ts *TypingSimulator) analyzeMessageComplexity(message string) string {
	charCount := utf8.RuneCountInString(message)
	wordCount := len(strings.Fields(message))
	
	// Faktor kompleksitas
	hasLinks := strings.Contains(message, "http")
	hasMentions := strings.Contains(message, "@")
	hasNumbers := strings.ContainsAny(message, "0123456789")
	hasSpecialChars := strings.ContainsAny(message, "!@#$%^&*()[]{}|;:,.<>?")
	
	// Scoring
	complexityScore := 0
	
	// Panjang pesan
	if charCount > 100 {
		complexityScore += 2
	} else if charCount > 50 {
		complexityScore += 1
	}
	
	// Jumlah kata
	if wordCount > 20 {
		complexityScore += 2
	} else if wordCount > 10 {
		complexityScore += 1
	}
	
	// Konten khusus
	if hasLinks {
		complexityScore += 2
	}
	if hasMentions {
		complexityScore += 1
	}
	if hasNumbers {
		complexityScore += 1
	}
	if hasSpecialChars {
		complexityScore += 1
	}
	
	// Kategorisasi
	if complexityScore <= 2 {
		return "simple"
	} else if complexityScore <= 5 {
		return "medium"
	} else {
		return "complex"
	}
}

// SetTypingSpeed mengatur kecepatan typing (1.0 = normal, 2.0 = 2x cepat, 0.5 = 2x lambat)
func (ts *TypingSimulator) SetTypingSpeed(speed float64) {
	if speed > 0 {
		ts.CharsPerSecond = 3.5 * speed
		ts.logger.Infof("Typing speed diatur ke %.1fx (%.1f chars/sec)", speed, ts.CharsPerSecond)
	}
}

// GetStats mengembalikan statistik typing simulator
func (ts *TypingSimulator) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"chars_per_second": ts.CharsPerSecond,
		"variation":        ts.TypingVariation,
		"pause_chance":     ts.PauseChance,
		"active_chats":     len(ts.activeChats),
		"min_typing_time":  ts.MinTypingTime.String(),
		"max_typing_time":  ts.MaxTypingTime.String(),
	}
}