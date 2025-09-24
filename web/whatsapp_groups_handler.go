// Package web - WhatsApp groups handler
package web

import (
	"encoding/json"
	"net/http"

	"go.mau.fi/whatsmeow"
)

// WhatsAppGroupInfo represents a WhatsApp group for API response
type WhatsAppGroupInfo struct {
	JID              string `json:"jid"`
	Name             string `json:"name"`
	ParticipantCount int    `json:"participant_count"`
}

// handleWhatsAppGroups handles WhatsApp groups API
func (s *DashboardServer) handleWhatsAppGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Check if WhatsApp client is available
	if s.whatsappClient == nil {
		s.logger.Errorf("WhatsApp client not available")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "WhatsApp client not available",
		})
		return
	}

	// Cast to whatsmeow client
	client, ok := s.whatsappClient.(*whatsmeow.Client)
	if !ok {
		s.logger.Errorf("Invalid WhatsApp client type")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Invalid WhatsApp client type",
		})
		return
	}

	// Check if client is connected
	if !client.IsConnected() {
		s.logger.Errorf("WhatsApp client not connected")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "WhatsApp client not connected",
		})
		return
	}

	// Get joined groups
	groups, err := client.GetJoinedGroups()
	if err != nil {
		s.logger.Errorf("Failed to get joined groups: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Failed to get WhatsApp groups: " + err.Error(),
		})
		return
	}

	// Convert to API response format
	var groupInfos []WhatsAppGroupInfo
	for _, group := range groups {
		groupInfo := WhatsAppGroupInfo{
			JID:              group.JID.String(),
			Name:             group.Name,
			ParticipantCount: len(group.Participants),
		}
		groupInfos = append(groupInfos, groupInfo)
	}

	s.logger.Debugf("Retrieved %d WhatsApp groups", len(groupInfos))

	// Return response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"groups": groupInfos,
		"count":  len(groupInfos),
	})
}