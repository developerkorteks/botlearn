// Package services - Learning service untuk mengelola bot pembelajaran
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"

	"github.com/nabilulilalbab/promote/database"
	"github.com/nabilulilalbab/promote/utils"
)

// LearningService mengelola sistem pembelajaran bot
type LearningService struct {
	client     *whatsmeow.Client
	repository database.Repository
	logger     *utils.Logger

	// Cooldown map untuk mencegah spam auto response
	// key: "groupJID:keyword", value: last response time
	responseCooldown map[string]time.Time
}

// NewLearningService membuat service baru untuk learning bot
func NewLearningService(client *whatsmeow.Client, repo database.Repository, logger *utils.Logger) *LearningService {
	return &LearningService{
		client:           client,
		repository:       repo,
		logger:           logger,
		responseCooldown: make(map[string]time.Time),
	}
}

// === GROUP ACCESS CONTROL ===

// IsGroupAllowed mengecek apakah grup diizinkan menggunakan bot
func (s *LearningService) IsGroupAllowed(groupJID string) bool {
	group, err := s.repository.GetLearningGroup(groupJID)
	if err != nil || group == nil {
		s.logger.Debugf("Group %s not in allowed list", groupJID)
		return false
	}

	if !group.IsActive {
		s.logger.Debugf("Group %s is inactive", groupJID)
		return false
	}

	s.logger.Debugf("Group %s is allowed and active", groupJID)
	return true
}

// AddAllowedGroup menambahkan grup ke daftar yang diizinkan
func (s *LearningService) AddAllowedGroup(groupJID, groupName, createdBy string) error {
	// Cek apakah grup sudah ada
	existingGroup, err := s.repository.GetLearningGroup(groupJID)
	if err != nil {
		return fmt.Errorf("failed to check existing group: %v", err)
	}

	if existingGroup != nil {
		// Update grup yang sudah ada
		existingGroup.IsActive = true
		existingGroup.GroupName = groupName
		return s.repository.UpdateLearningGroup(existingGroup)
	}

	// Buat grup baru
	newGroup := &database.LearningGroup{
		GroupJID:    groupJID,
		GroupName:   groupName,
		IsActive:    true,
		Description: "Bot pembelajaran dan injec",
		CreatedBy:   createdBy,
	}

	err = s.repository.CreateLearningGroup(newGroup)
	if err != nil {
		return fmt.Errorf("failed to create learning group: %v", err)
	}

	s.logger.Infof("Added learning group: %s (%s)", groupName, groupJID)
	return nil
}

// RemoveAllowedGroup menghapus grup dari daftar yang diizinkan
func (s *LearningService) RemoveAllowedGroup(groupJID string) error {
	err := s.repository.ToggleLearningGroup(groupJID, false)
	if err != nil {
		return fmt.Errorf("failed to deactivate learning group: %v", err)
	}

	s.logger.Infof("Deactivated learning group: %s", groupJID)
	return nil
}

// GetAllowedGroups mendapatkan semua grup yang diizinkan
func (s *LearningService) GetAllowedGroups() ([]database.LearningGroup, error) {
	return s.repository.GetAllLearningGroups()
}

// === COMMAND PROCESSING ===

// ProcessCommand memproses command pembelajaran
func (s *LearningService) ProcessCommand(groupJID, userJID, command string) error {
	// Cek apakah grup diizinkan
	if !s.IsGroupAllowed(groupJID) {
		s.logger.Debugf("Command %s blocked - group %s not allowed", command, groupJID)
		return nil // Diam saja, tidak ada response
	}

	// Ambil command dari database
	cmd, err := s.repository.GetLearningCommand(command)
	if err != nil {
		s.logCommandUsage("learning_command", command, groupJID, userJID, "", false, err.Error())
		return fmt.Errorf("failed to get command: %v", err)
	}

	if cmd == nil {
		s.logger.Debugf("Command %s not found", command)
		return nil // Command tidak ditemukan, diam saja
	}

	// Kirim response berdasarkan tipe
	err = s.sendResponse(groupJID, cmd)
	if err != nil {
		s.logCommandUsage("learning_command", command, groupJID, userJID, cmd.ResponseType, false, err.Error())
		return fmt.Errorf("failed to send response: %v", err)
	}

	// Update usage count dan log
	s.repository.IncrementCommandUsage(command)
	s.logCommandUsage("learning_command", command, groupJID, userJID, cmd.ResponseType, true, "")

	s.logger.Infof("Command %s processed successfully for group %s", command, groupJID)
	return nil
}

// ProcessAutoResponse memproses auto response berdasarkan kata kunci
func (s *LearningService) ProcessAutoResponse(groupJID, userJID, messageText string) error {
	// Cek apakah grup diizinkan
	if !s.IsGroupAllowed(groupJID) {
		return nil // Diam saja
	}

	// Cari auto response yang match
	responses, err := s.repository.GetAutoResponsesByKeyword(strings.ToLower(messageText))
	if err != nil {
		return fmt.Errorf("failed to get auto responses: %v", err)
	}

	if len(responses) == 0 {
		return nil // Tidak ada response yang match
	}

	// Kirim hanya 1 response pertama yang match dengan cooldown check
	for _, response := range responses {
		// Cek cooldown untuk mencegah spam
		cooldownKey := fmt.Sprintf("%s:%s", groupJID, response.Keyword)
		lastResponseTime, exists := s.responseCooldown[cooldownKey]

		// Cooldown 10 detik untuk mencegah spam
		if exists && time.Since(lastResponseTime) < 10*time.Second {
			s.logger.Debugf("Auto response '%s' on cooldown for group %s", response.Keyword, groupJID)
			continue
		}

		// Set cooldown time
		s.responseCooldown[cooldownKey] = time.Now()

		// Kirim response
		err = s.sendAutoResponse(groupJID, &response)
		if err != nil {
			s.logCommandUsage("auto_response", response.Keyword, groupJID, userJID, response.ResponseType, false, err.Error())
			continue
		}

		// Update usage count dan log
		s.repository.IncrementAutoResponseUsage(response.Keyword)
		s.logCommandUsage("auto_response", response.Keyword, groupJID, userJID, response.ResponseType, true, "")

		s.logger.Infof("Auto response '%s' triggered for group %s", response.Keyword, groupJID)

		// Hanya kirim 1 response pertama yang match, lalu break
		break
	}

	return nil
}

// === ADMIN/UTILITY QUERIES ===

// ListJoinedGroups mengembalikan daftar grup yang diikuti bot dalam bentuk teks
func (s *LearningService) ListJoinedGroups() (string, error) {
	groups, err := s.client.GetJoinedGroups()
	if err != nil {
		return "", fmt.Errorf("failed to get joined groups: %v", err)
	}
	if len(groups) == 0 {
		return "ℹ️ TIDAK ADA GRUP\n\nBot belum tergabung di grup manapun.", nil
	}

	result := "📋 DAFTAR GRUP YANG DIIKUTI BOT\n\n"
	for i, g := range groups {
		name := g.Name
		if name == "" {
			name = "(Tanpa Nama)"
		}
		memberCount := len(g.Participants)
		result += fmt.Sprintf("%d. %s\n   📱 JID: %s\n   👥 Member: %d\n", i+1, name, g.JID.String(), memberCount)
	}
	return result, nil
}

// === RESPONSE SENDERS ===

// sendResponse mengirim response command pembelajaran
func (s *LearningService) sendResponse(chatJID string, cmd *database.LearningCommand) error {
	jid, err := types.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid JID: %v", err)
	}

	// Special handling untuk command dinamis
	if cmd.Command == ".help" || cmd.Command == ".info" {
		dynamicHelp, err := s.GenerateDynamicHelp()
		if err != nil {
			s.logger.Errorf("Failed to generate dynamic help: %v", err)
			// Fallback ke text content asli jika ada error
			if cmd.TextContent != nil {
				return s.sendTextMessage(jid, *cmd.TextContent)
			}
			return fmt.Errorf("failed to generate dynamic help: %v", err)
		}
		return s.sendTextMessage(jid, dynamicHelp)
	}

	switch cmd.ResponseType {
	case "text":
		return s.sendTextMessage(jid, *cmd.TextContent)

	case "image":
		caption := ""
		if cmd.Caption != nil {
			caption = *cmd.Caption
		}
		return s.sendImageMessage(jid, *cmd.MediaFilePath, caption)

	case "video":
		caption := ""
		if cmd.Caption != nil {
			caption = *cmd.Caption
		}
		return s.sendVideoMessage(jid, *cmd.MediaFilePath, caption)

	case "audio":
		return s.sendAudioMessage(jid, *cmd.MediaFilePath)

	case "sticker":
		return s.sendStickerMessage(jid, *cmd.MediaFilePath)

	case "file":
		caption := ""
		if cmd.Caption != nil {
			caption = *cmd.Caption
		}
		return s.sendFileMessage(jid, *cmd.MediaFilePath, caption)

	default:
		return fmt.Errorf("unsupported response type: %s", cmd.ResponseType)
	}
}

// sendAutoResponse mengirim auto response
func (s *LearningService) sendAutoResponse(chatJID string, response *database.AutoResponse) error {
	jid, err := types.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid JID: %v", err)
	}

	switch response.ResponseType {
	case "text":
		if response.TextResponse != nil {
			return s.sendTextMessage(jid, *response.TextResponse)
		}

	case "sticker":
		if response.StickerPath != nil {
			return s.sendStickerMessage(jid, *response.StickerPath)
		}

	case "audio":
		if response.AudioPath != nil {
			return s.sendAudioMessage(jid, *response.AudioPath)
		}

	case "mixed":
		// Kirim text dulu jika ada
		if response.TextResponse != nil {
			s.sendTextMessage(jid, *response.TextResponse)
		}
		// Kirim sticker jika ada
		if response.StickerPath != nil {
			s.sendStickerMessage(jid, *response.StickerPath)
		}
		// Kirim audio jika ada
		if response.AudioPath != nil {
			s.sendAudioMessage(jid, *response.AudioPath)
		}
	}

	return nil
}

// === MEDIA SENDERS ===

// sendTextMessage mengirim pesan teks
func (s *LearningService) sendTextMessage(jid types.JID, text string) error {
	msg := &waProto.Message{
		Conversation: &text,
	}

	_, err := s.client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send text: %v", err)
	}

	s.logger.Debugf("Text message sent to %s", jid.String())
	return nil
}

// sendImageMessage mengirim gambar
func (s *LearningService) sendImageMessage(jid types.JID, imagePath, caption string) error {
	s.logger.Debugf("Attempting to send image: %s to %s", imagePath, jid.String())

	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		s.logger.Errorf("Failed to read image file %s: %v", imagePath, err)
		return fmt.Errorf("failed to read image: %v", err)
	}

	s.logger.Debugf("Image file read successfully, size: %d bytes", len(imageData))

	uploaded, err := s.client.Upload(context.Background(), imageData, whatsmeow.MediaImage)
	if err != nil {
		s.logger.Errorf("Failed to upload image to WhatsApp: %v", err)
		return fmt.Errorf("failed to upload image: %v", err)
	}

	s.logger.Debugf("Image uploaded successfully, URL: %s", uploaded.URL)

	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Caption:       &caption,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			Mimetype:      &[]string{"image/jpeg"}[0],
		},
	}

	_, err = s.client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		s.logger.Errorf("Failed to send image message: %v", err)
		return fmt.Errorf("failed to send image: %v", err)
	}

	s.logger.Infof("Image sent successfully to %s: %s", jid.String(), filepath.Base(imagePath))
	return nil
}

// sendVideoMessage mengirim video
func (s *LearningService) sendVideoMessage(jid types.JID, videoPath, caption string) error {
	s.logger.Debugf("Attempting to send video: %s to %s", videoPath, jid.String())

	videoData, err := os.ReadFile(videoPath)
	if err != nil {
		s.logger.Errorf("Failed to read video file %s: %v", videoPath, err)
		return fmt.Errorf("failed to read video: %v", err)
	}

	s.logger.Debugf("Video file read successfully, size: %d bytes", len(videoData))

	uploaded, err := s.client.Upload(context.Background(), videoData, whatsmeow.MediaVideo)
	if err != nil {
		s.logger.Errorf("Failed to upload video to WhatsApp: %v", err)
		return fmt.Errorf("failed to upload video: %v", err)
	}

	s.logger.Debugf("Video uploaded successfully, URL: %s", uploaded.URL)

	msg := &waProto.Message{
		VideoMessage: &waProto.VideoMessage{
			Caption:       &caption,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			Mimetype:      &[]string{"video/mp4"}[0],
		},
	}

	_, err = s.client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		s.logger.Errorf("Failed to send video message: %v", err)
		return fmt.Errorf("failed to send video: %v", err)
	}

	s.logger.Infof("Video sent successfully to %s: %s", jid.String(), filepath.Base(videoPath))
	return nil
}

// sendAudioMessage mengirim audio/voice note
func (s *LearningService) sendAudioMessage(jid types.JID, audioPath string) error {
	s.logger.Debugf("Attempting to send audio: %s to %s", audioPath, jid.String())

	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		s.logger.Errorf("Failed to read audio file %s: %v", audioPath, err)
		return fmt.Errorf("failed to read audio: %v", err)
	}

	s.logger.Debugf("Audio file read successfully, size: %d bytes", len(audioData))

	uploaded, err := s.client.Upload(context.Background(), audioData, whatsmeow.MediaAudio)
	if err != nil {
		s.logger.Errorf("Failed to upload audio to WhatsApp: %v", err)
		return fmt.Errorf("failed to upload audio: %v", err)
	}

	s.logger.Debugf("Audio uploaded successfully, URL: %s", uploaded.URL)

	msg := &waProto.Message{
		AudioMessage: &waProto.AudioMessage{
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			PTT:           &[]bool{true}[0], // Voice note
			Mimetype:      &[]string{"audio/ogg; codecs=opus"}[0],
		},
	}

	_, err = s.client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		s.logger.Errorf("Failed to send audio message: %v", err)
		return fmt.Errorf("failed to send audio: %v", err)
	}

	s.logger.Infof("Audio sent successfully to %s: %s", jid.String(), filepath.Base(audioPath))
	return nil
}

// sendStickerMessage mengirim sticker
func (s *LearningService) sendStickerMessage(jid types.JID, stickerPath string) error {
	s.logger.Debugf("Attempting to send sticker: %s to %s", stickerPath, jid.String())

	stickerData, err := os.ReadFile(stickerPath)
	if err != nil {
		s.logger.Errorf("Failed to read sticker file %s: %v", stickerPath, err)
		return fmt.Errorf("failed to read sticker: %v", err)
	}

	s.logger.Debugf("Sticker file read successfully, size: %d bytes", len(stickerData))

	// Sticker harus menggunakan MediaImage untuk upload
	uploaded, err := s.client.Upload(context.Background(), stickerData, whatsmeow.MediaImage)
	if err != nil {
		s.logger.Errorf("Failed to upload sticker to WhatsApp: %v", err)
		return fmt.Errorf("failed to upload sticker: %v", err)
	}

	s.logger.Debugf("Sticker uploaded successfully, URL: %s", uploaded.URL)

	msg := &waProto.Message{
		StickerMessage: &waProto.StickerMessage{
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			Mimetype:      &[]string{"image/webp"}[0],
		},
	}

	_, err = s.client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		s.logger.Errorf("Failed to send sticker message: %v", err)
		return fmt.Errorf("failed to send sticker: %v", err)
	}

	s.logger.Infof("Sticker sent successfully to %s: %s", jid.String(), filepath.Base(stickerPath))
	return nil
}

// sendFileMessage mengirim file/dokumen
func (s *LearningService) sendFileMessage(jid types.JID, filePath, caption string) error {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	fileName := filepath.Base(filePath)
	uploaded, err := s.client.Upload(context.Background(), fileData, whatsmeow.MediaDocument)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	msg := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			Caption:       &caption,
			FileName:      &fileName,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
		},
	}

	_, err = s.client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send file: %v", err)
	}

	s.logger.Debugf("File sent to %s: %s", jid.String(), fileName)
	return nil
}

// === UTILITY FUNCTIONS ===

// logCommandUsage mencatat log penggunaan command
func (s *LearningService) logCommandUsage(commandType, commandValue, groupJID, userJID, responseType string, success bool, errorMsg string) {
	log := &database.CommandUsageLog{
		CommandType:  commandType,
		CommandValue: commandValue,
		GroupJID:     groupJID,
		UserJID:      userJID,
		ResponseType: responseType,
		Success:      success,
		ErrorMessage: nil,
	}

	if errorMsg != "" {
		log.ErrorMessage = &errorMsg
	}

	err := s.repository.LogCommandUsage(log)
	if err != nil {
		s.logger.Errorf("Failed to log command usage: %v", err)
	}
}

// GetUsageStats mendapatkan statistik penggunaan command
func (s *LearningService) GetUsageStats(days int) (map[string]int, error) {
	return s.repository.GetCommandUsageStats(days)
}

// GetUsageLogs mendapatkan log penggunaan command
func (s *LearningService) GetUsageLogs(limit int) ([]database.CommandUsageLog, error) {
	return s.repository.GetCommandUsageLogs(limit)
}

// GenerateDynamicHelp membuat response help dinamis berdasarkan command yang ada
func (s *LearningService) GenerateDynamicHelp() (string, error) {
	// Ambil semua command aktif dari database
	commands, err := s.repository.GetAllLearningCommands()
	if err != nil {
		return "", fmt.Errorf("failed to get commands: %v", err)
	}

	// Filter hanya command aktif
	activeCommands := make([]database.LearningCommand, 0)
	for _, cmd := range commands {
		if cmd.IsActive {
			activeCommands = append(activeCommands, cmd)
		}
	}

	if len(activeCommands) == 0 {
		return `📚 *BANTUAN BOT PEMBELAJARAN* 📚

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ℹ️ Belum ada command yang tersedia.
Admin dapat menambahkan command melalui dashboard.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *Bot ini untuk pembelajaran saja*
🚫 *Gunakan dengan bijak*`, nil
	}

	// Kelompokkan command berdasarkan kategori
	categories := make(map[string][]database.LearningCommand)
	for _, cmd := range activeCommands {
		categories[cmd.Category] = append(categories[cmd.Category], cmd)
	}

	// Buat response dinamis
	response := `📚 *BANTUAN BOT PEMBELAJARAN* 📚

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *COMMAND TERSEDIA*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

`

	// Kategori dengan icon
	categoryIcons := map[string]string{
		"injec":        "🔧",
		"pembelajaran": "📚",
		"informasi":    "ℹ️",
		"tools":        "🛠️",
		"general":      "📝",
	}

	categoryNames := map[string]string{
		"injec":        "INJEC & VPN",
		"pembelajaran": "PEMBELAJARAN",
		"informasi":    "INFORMASI",
		"tools":        "TOOLS",
		"general":      "UMUM",
	}

	// Urutkan kategori
	orderedCategories := []string{"injec", "pembelajaran", "tools", "informasi", "general"}

	for _, catKey := range orderedCategories {
		if commands, exists := categories[catKey]; exists && len(commands) > 0 {
			icon := categoryIcons[catKey]
			name := categoryNames[catKey]
			if icon == "" {
				icon = "📝"
			}
			if name == "" {
				name = strings.ToUpper(catKey)
			}

			response += fmt.Sprintf("%s *%s:*\n", icon, name)

			for _, cmd := range commands {
				// Tambahkan deskripsi jika ada
				desc := ""
				if cmd.Description != "" {
					desc = fmt.Sprintf(" - %s", cmd.Description)
				}
				response += fmt.Sprintf("• %s%s\n", cmd.Command, desc)
			}
			response += "\n"
		}
	}

	// Tambahkan kategori lain yang tidak terdefinisi
	for catKey, commands := range categories {
		found := false
		for _, orderedCat := range orderedCategories {
			if orderedCat == catKey {
				found = true
				break
			}
		}

		if !found && len(commands) > 0 {
			response += fmt.Sprintf("📝 *%s:*\n", strings.ToUpper(catKey))
			for _, cmd := range commands {
				desc := ""
				if cmd.Description != "" {
					desc = fmt.Sprintf(" - %s", cmd.Description)
				}
				response += fmt.Sprintf("• %s%s\n", cmd.Command, desc)
			}
			response += "\n"
		}
	}

	response += `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *Bot ini untuk pembelajaran saja*
🚫 *Gunakan dengan bijak*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`

	return response, nil
}
