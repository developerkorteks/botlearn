// Package web - Additional handlers for dashboard
package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nabilulilalbab/promote/database"
)

// handleAutoResponses handles auto response management API
func (s *DashboardServer) handleAutoResponses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getAutoResponses(w, r)
	case "POST":
		s.createAutoResponse(w, r)
	case "PUT":
		s.updateAutoResponse(w, r)
	case "DELETE":
		s.deleteAutoResponse(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getAutoResponses returns all auto responses
func (s *DashboardServer) getAutoResponses(w http.ResponseWriter, r *http.Request) {
	responses, err := s.repository.GetAllAutoResponses()
	if err != nil {
		s.logger.Errorf("Failed to get auto responses: %v", err)
		http.Error(w, "Failed to get auto responses", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// createAutoResponse creates a new auto response
func (s *DashboardServer) createAutoResponse(w http.ResponseWriter, r *http.Request) {
	var response database.AutoResponse
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Set default values
	response.IsActive = true
	response.CreatedBy = "admin"
	
	if err := s.repository.CreateAutoResponse(&response); err != nil {
		s.logger.Errorf("Failed to create auto response: %v", err)
		http.Error(w, "Failed to create auto response", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// updateAutoResponse updates an auto response
func (s *DashboardServer) updateAutoResponse(w http.ResponseWriter, r *http.Request) {
	var response database.AutoResponse
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if err := s.repository.UpdateAutoResponse(&response); err != nil {
		s.logger.Errorf("Failed to update auto response: %v", err)
		http.Error(w, "Failed to update auto response", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// deleteAutoResponse deletes an auto response
func (s *DashboardServer) deleteAutoResponse(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		http.Error(w, "Keyword required", http.StatusBadRequest)
		return
	}
	
	if err := s.repository.DeleteAutoResponse(keyword); err != nil {
		s.logger.Errorf("Failed to delete auto response: %v", err)
		http.Error(w, "Failed to delete auto response", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleUpload handles file uploads
func (s *DashboardServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse multipart form
	err := r.ParseMultipartForm(50 << 20) // 50MB limit
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	// Get file type from form
	fileType := r.FormValue("type")
	if fileType == "" {
		fileType = "files" // default
	}
	
	// Create filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s", timestamp, header.Filename)
	
	// Determine upload directory
	uploadDir := filepath.Join(s.mediaPath, fileType)
	os.MkdirAll(uploadDir, 0755)
	
	// Create destination file
	filePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		s.logger.Errorf("Failed to create file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	
	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		s.logger.Errorf("Failed to copy file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	
	s.logger.Infof("File uploaded: %s", filePath)
	
	// Return file path
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "success",
		"filepath": filePath,
		"filename": filename,
	})
}

// handleStats handles statistics API
func (s *DashboardServer) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Get days parameter
	daysStr := r.URL.Query().Get("days")
	days := 7 // default
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}
	
	// Get command usage stats
	stats, err := s.repository.GetCommandUsageStats(days)
	if err != nil {
		s.logger.Errorf("Failed to get stats: %v", err)
		http.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}
	
	// Get recent logs
	logs, err := s.repository.GetCommandUsageLogs(20)
	if err != nil {
		s.logger.Errorf("Failed to get logs: %v", err)
		logs = []database.CommandUsageLog{} // empty slice if error
	}
	
	// Get group count
	groups, err := s.repository.GetAllLearningGroups()
	if err != nil {
		s.logger.Errorf("Failed to get groups: %v", err)
		groups = []database.LearningGroup{}
	}
	
	// Get command count
	commands, err := s.repository.GetAllLearningCommands()
	if err != nil {
		s.logger.Errorf("Failed to get commands: %v", err)
		commands = []database.LearningCommand{}
	}
	
	// Get auto response count
	autoResponses, err := s.repository.GetAllAutoResponses()
	if err != nil {
		s.logger.Errorf("Failed to get auto responses: %v", err)
		autoResponses = []database.AutoResponse{}
	}
	
	// Build response
	response := map[string]interface{}{
		"usage_stats": stats,
		"recent_logs": logs,
		"counts": map[string]int{
			"groups":         len(groups),
			"commands":       len(commands),
			"auto_responses": len(autoResponses),
		},
		"days": days,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}