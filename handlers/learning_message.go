// Package handlers - Learning message handler untuk bot pembelajaran
package handlers

import (
	"context"
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"github.com/nabilulilalbab/promote/services"
	"github.com/nabilulilalbab/promote/utils"
)

// LearningMessageHandler menangani pesan untuk bot pembelajaran
type LearningMessageHandler struct {
	client          *whatsmeow.Client
	learningService *services.LearningService
	logger          *utils.Logger
	adminNumbers    []string // Daftar nomor admin
}

// NewLearningMessageHandler membuat handler baru untuk learning bot
func NewLearningMessageHandler(
	client *whatsmeow.Client,
	learningService *services.LearningService,
	logger *utils.Logger,
	adminNumbers []string,
) *LearningMessageHandler {
	return &LearningMessageHandler{
		client:          client,
		learningService: learningService,
		logger:          logger,
		adminNumbers:    adminNumbers,
	}
}

// HandleMessage adalah fungsi utama untuk menangani pesan masuk
func (h *LearningMessageHandler) HandleMessage(evt *events.Message) {
	// STEP 1: Skip pesan dari diri sendiri
	if evt.Info.IsFromMe {
		return
	}

	// STEP 2: Ambil teks dari pesan
	messageText := h.getMessageText(evt.Message)
	if messageText == "" {
		return // Bukan pesan teks, skip
	}

	// STEP 3: Identifikasi jenis chat dan info
	isGroup := evt.Info.Chat.Server == types.GroupServer
	groupJID := evt.Info.Chat.String()
	userJID := evt.Info.Sender.String()

	// Log pesan untuk debugging
	chatType := "personal"
	if isGroup {
		chatType = "group"
	}

	h.logger.Debugf("📨 Message [%s]: %s | From: %s | Group: %s",
		chatType, h.truncateString(messageText, 50), userJID, groupJID)

	// STEP 4: Proses berdasarkan jenis chat
	if isGroup {
		h.handleGroupMessage(evt, groupJID, userJID, messageText)
	} else {
		h.handlePersonalMessage(evt, userJID, messageText)
	}
}

// handleGroupMessage menangani pesan dari grup
func (h *LearningMessageHandler) handleGroupMessage(evt *events.Message, groupJID, userJID, messageText string) {
	// Cek apakah grup diizinkan untuk menggunakan bot
	if !h.learningService.IsGroupAllowed(groupJID) {
		// BOT DIAM TOTAL - tidak ada response apapun
		h.logger.Debugf("👥 Group not allowed: %s | Message ignored", groupJID)
		return
	}

	// Grup diizinkan, proses pesan
	h.logger.Debugf("👥 Processing group message: %s", groupJID)

	// Cek dan tendang pengguna jika mengirim kata terlarang
	if err := h.learningService.CheckAndHandleForbiddenWord(evt); err != nil {
		h.logger.Errorf("Error handling forbidden word: %v", err)
		// Lanjutkan proses meskipun gagal menendang
	}

	// Cek apakah ini command (.command)
	if strings.HasPrefix(messageText, ".") {
		h.handleLearningCommand(groupJID, userJID, messageText)
		return
	}

	// Cek auto response untuk kata kunci
	h.handleAutoResponse(groupJID, userJID, messageText)
}

// handlePersonalMessage menangani pesan personal (admin only)
func (h *LearningMessageHandler) handlePersonalMessage(evt *events.Message, userJID, messageText string) {
	// Cek apakah user adalah admin
	if !h.isAdmin(userJID) {
		h.logger.Debugf("💬 Non-admin personal message ignored: %s", userJID)
		return // Bot diam untuk non-admin
	}

	h.logger.Debugf("💬 Processing admin personal message: %s", userJID)

	// Admin command processing
	if strings.HasPrefix(messageText, ".") {
		h.handleAdminCommand(evt, userJID, messageText)
		return
	}

	// Admin bisa ngobrol biasa, bot kasih response sederhana
	if strings.Contains(strings.ToLower(messageText), "bot") ||
		strings.Contains(strings.ToLower(messageText), "help") {
		h.sendAdminHelp(evt.Info.Chat)
	}
}

// handleLearningCommand memproses command pembelajaran
func (h *LearningMessageHandler) handleLearningCommand(groupJID, userJID, command string) {
	h.logger.Infof("🔧 Processing learning command: %s | Group: %s | User: %s",
		command, groupJID, userJID)

	err := h.learningService.ProcessCommand(groupJID, userJID, command)
	if err != nil {
		h.logger.Errorf("Failed to process command %s: %v", command, err)
	}
}

// handleAutoResponse memproses auto response berdasarkan kata kunci
func (h *LearningMessageHandler) handleAutoResponse(groupJID, userJID, messageText string) {
	// Cek kata kunci dalam pesan
	lowerText := strings.ToLower(messageText)

	err := h.learningService.ProcessAutoResponse(groupJID, userJID, lowerText)
	if err != nil {
		h.logger.Errorf("Failed to process auto response: %v", err)
	}
}

// handleAdminCommand memproses command admin via personal chat
func (h *LearningMessageHandler) handleAdminCommand(evt *events.Message, userJID, command string) {
	h.logger.Infof("👑 Processing admin command: %s | Admin: %s", command, userJID)

	lowerCommand := strings.ToLower(strings.TrimSpace(command))

	switch {
	case strings.HasPrefix(lowerCommand, ".addgroup"):
		h.handleAddGroup(evt, command)

	case strings.HasPrefix(lowerCommand, ".removegroup"):
		h.handleRemoveGroup(evt, command)

	case strings.HasPrefix(lowerCommand, ".listgroups"):
		h.handleListGroups(evt)

	case strings.HasPrefix(lowerCommand, ".getgroups"):
		h.handleGetGroups(evt)

	case strings.HasPrefix(lowerCommand, ".groups") || strings.HasPrefix(lowerCommand, ".allgroups"):
		h.handleGetAllGroups(evt)

	case strings.HasPrefix(lowerCommand, ".stats"):
		h.handleStats(evt, command)

	case strings.HasPrefix(lowerCommand, ".logs"):
		h.handleLogs(evt, command)

	case lowerCommand == ".getgroups" || lowerCommand == ".allgroups":
		h.handleGetAllGroups(evt)

	case lowerCommand == ".adminhelp" || lowerCommand == ".help":
		h.sendAdminHelp(evt.Info.Chat)

	default:
		h.sendUnknownCommand(evt.Info.Chat)
	}
}

// === ADMIN COMMAND HANDLERS ===

// handleAddGroup menangani command .addgroup
func (h *LearningMessageHandler) handleAddGroup(evt *events.Message, command string) {
	// Format: .addgroup <group_jid> <group_name>
	parts := strings.Fields(command)
	if len(parts) < 3 {
		h.sendTextToChat(evt.Info.Chat, `❌ *FORMAT SALAH*

Format: .addgroup <group_jid> <group_name>

Contoh:
.addgroup 120363123456789@g.us Grup Belajar Coding`)
		return
	}

	groupJID := parts[1]
	groupName := strings.Join(parts[2:], " ")

	err := h.learningService.AddAllowedGroup(groupJID, groupName, evt.Info.Sender.User)
	if err != nil {
		h.sendTextToChat(evt.Info.Chat, fmt.Sprintf("❌ *GAGAL*\n\nError: %v", err))
		return
	}

	response := fmt.Sprintf(`✅ *GRUP BERHASIL DITAMBAHKAN*

📱 *JID:* %s
👥 *Nama:* %s
🎯 *Status:* Aktif

Bot sekarang bisa digunakan di grup tersebut!`, groupJID, groupName)

	h.sendTextToChat(evt.Info.Chat, response)
}

// handleRemoveGroup menangani command .removegroup
func (h *LearningMessageHandler) handleRemoveGroup(evt *events.Message, command string) {
	// Format: .removegroup <group_jid>
	parts := strings.Fields(command)
	if len(parts) < 2 {
		h.sendTextToChat(evt.Info.Chat, `❌ *FORMAT SALAH*

Format: .removegroup <group_jid>

Contoh:
.removegroup 120363123456789@g.us`)
		return
	}

	groupJID := parts[1]

	err := h.learningService.RemoveAllowedGroup(groupJID)
	if err != nil {
		h.sendTextToChat(evt.Info.Chat, fmt.Sprintf("❌ *GAGAL*\n\nError: %v", err))
		return
	}

	response := fmt.Sprintf(`✅ *GRUP BERHASIL DINONAKTIFKAN*

📱 *JID:* %s
🎯 *Status:* Tidak aktif

Bot tidak akan merespon di grup tersebut lagi.`, groupJID)

	h.sendTextToChat(evt.Info.Chat, response)
}

// handleListGroups menangani command .listgroups
func (h *LearningMessageHandler) handleListGroups(evt *events.Message) {
	groups, err := h.learningService.GetAllowedGroups()
	if err != nil {
		h.sendTextToChat(evt.Info.Chat, fmt.Sprintf("❌ *GAGAL*\n\nError: %v", err))
		return
	}

	if len(groups) == 0 {
		h.sendTextToChat(evt.Info.Chat, `ℹ️ *TIDAK ADA GRUP*

Belum ada grup yang diaktifkan untuk bot pembelajaran.
Gunakan .addgroup untuk menambahkan grup.`)
		return
	}

	response := `📋 *DAFTAR GRUP PEMBELAJARAN*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`

	for i, group := range groups {
		status := "✅ Aktif"
		if !group.IsActive {
			status = "❌ Tidak aktif"
		}

		response += fmt.Sprintf(`
%d. *%s*
   📱 JID: %s
   🎯 Status: %s
   📅 Dibuat: %s`,
			i+1, group.GroupName, group.GroupJID, status,
			group.CreatedAt.Format("02/01/2006 15:04"))
	}

	response += `

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total: ` + fmt.Sprintf("%d grup", len(groups))

	h.sendTextToChat(evt.Info.Chat, response)
}

// handleGetGroups menangani command .getgroups (sama seperti .listgroups)
func (h *LearningMessageHandler) handleGetGroups(evt *events.Message) {
	h.handleListGroups(evt) // Redirect ke handleListGroups
}

// handleGetAllGroups menampilkan semua grup yang diikuti WhatsApp bot (langsung dari WhatsApp)
func (h *LearningMessageHandler) handleGetAllGroups(evt *events.Message) {
	// Hanya admin yang bisa
	if !h.isAdmin(evt.Info.Sender.User) {
		h.sendTextToChat(evt.Info.Chat, "❌ Akses ditolak: hanya admin")
		return
	}

	text, err := h.learningService.ListJoinedGroups()
	if err != nil {
		h.sendTextToChat(evt.Info.Chat, "❌ Gagal mengambil daftar grup: "+err.Error())
		return
	}

	h.sendTextToChat(evt.Info.Chat, text)
}

// handleStats menangani command .stats
func (h *LearningMessageHandler) handleStats(evt *events.Message, command string) {
	// Format: .stats [days] (default 7 hari)
	days := 7
	parts := strings.Fields(command)
	if len(parts) > 1 {
		if d, err := fmt.Sscanf(parts[1], "%d", &days); err != nil || d != 1 {
			days = 7
		}
	}

	stats, err := h.learningService.GetUsageStats(days)
	if err != nil {
		h.sendTextToChat(evt.Info.Chat, fmt.Sprintf("❌ *GAGAL*\n\nError: %v", err))
		return
	}

	if len(stats) == 0 {
		h.sendTextToChat(evt.Info.Chat, fmt.Sprintf(`ℹ️ *TIDAK ADA DATA*

Belum ada aktivitas command dalam %d hari terakhir.`, days))
		return
	}

	response := fmt.Sprintf(`📊 *STATISTIK PENGGUNAAN* (%d hari)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`, days)

	i := 1
	for command, count := range stats {
		response += fmt.Sprintf("%d. %s: *%d kali*\n", i, command, count)
		i++
		if i > 10 { // Batasi 10 teratas
			break
		}
	}

	response += `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

	h.sendTextToChat(evt.Info.Chat, response)
}

// handleLogs menangani command .logs
func (h *LearningMessageHandler) handleLogs(evt *events.Message, command string) {
	// Format: .logs [limit] (default 10)
	limit := 10
	parts := strings.Fields(command)
	if len(parts) > 1 {
		if l, err := fmt.Sscanf(parts[1], "%d", &limit); err != nil || l != 1 {
			limit = 10
		}
	}

	logs, err := h.learningService.GetUsageLogs(limit)
	if err != nil {
		h.sendTextToChat(evt.Info.Chat, fmt.Sprintf("❌ *GAGAL*\n\nError: %v", err))
		return
	}

	if len(logs) == 0 {
		h.sendTextToChat(evt.Info.Chat, `ℹ️ *TIDAK ADA LOG*

Belum ada aktivitas yang tercatat.`)
		return
	}

	response := fmt.Sprintf(`📋 *LOG AKTIVITAS* (%d terakhir)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`, limit)

	for i, log := range logs {
		status := "✅"
		if !log.Success {
			status = "❌"
		}

		response += fmt.Sprintf(`
%d. %s %s (%s)
   ⏰ %s`,
			i+1, status, log.CommandValue, log.ResponseType,
			log.UsedAt.Format("02/01 15:04"))
	}

	response += `

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

	h.sendTextToChat(evt.Info.Chat, response)
}

// === HELPER FUNCTIONS ===

// sendAdminHelp mengirim bantuan untuk admin
func (h *LearningMessageHandler) sendAdminHelp(chatJID types.JID) {
	help := `👑 *BANTUAN ADMIN BOT PEMBELAJARAN*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
            *COMMAND ADMIN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔧 *KELOLA GRUP:*
• .groups/.allgroups - Lihat semua grup yang diikuti bot
• .addgroup <jid> <nama> - Aktifkan grup untuk learning
• .removegroup <jid> - Nonaktifkan grup
• .listgroups/.getgroups - Daftar grup learning yang aktif

📊 *MONITORING:*
• .stats [days] - Statistik penggunaan
• .logs [limit] - Log aktivitas terbaru

ℹ️ *INFORMASI:*
• .help / .adminhelp - Bantuan ini

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *CARA KERJA BOT:*
1. Admin aktifkan grup dengan .addgroup
2. Bot hanya merespon di grup yang diaktifkan
3. User di grup bisa pakai command pembelajaran
4. Bot diam total di grup yang tidak diaktifkan

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`

	h.sendTextToChat(chatJID, help)
}

// sendUnknownCommand mengirim pesan command tidak dikenal
func (h *LearningMessageHandler) sendUnknownCommand(chatJID types.JID) {
	response := `❓ *COMMAND TIDAK DIKENAL*

Ketik .help untuk melihat daftar command yang tersedia.`

	h.sendTextToChat(chatJID, response)
}

// sendTextToChat mengirim pesan teks ke chat
func (h *LearningMessageHandler) sendTextToChat(chatJID types.JID, text string) {
	msg := &waProto.Message{
		Conversation: &text,
	}

	_, err := h.client.SendMessage(context.Background(), chatJID, msg)
	if err != nil {
		h.logger.Errorf("Failed to send text message: %v", err)
	}
}

// getMessageText mengekstrak teks dari berbagai tipe pesan WhatsApp
func (h *LearningMessageHandler) getMessageText(msg *waProto.Message) string {
	// Pesan teks biasa
	if msg.GetConversation() != "" {
		return msg.GetConversation()
	}

	// Pesan teks dengan format (bold, italic, dll) atau reply
	if msg.GetExtendedTextMessage() != nil {
		return msg.GetExtendedTextMessage().GetText()
	}

	// Jika bukan teks, return empty string
	return ""
}

// isAdmin mengecek apakah user adalah admin
func (h *LearningMessageHandler) isAdmin(userJID string) bool {
	// Extract nomor dari berbagai format JID
	userNumber := strings.Replace(userJID, "@s.whatsapp.net", "", 1)
	userNumber = strings.Replace(userNumber, "@c.us", "", 1)

	// Handle format dengan :angka (seperti 6287817739901:8@s.whatsapp.net)
	if strings.Contains(userNumber, ":") {
		userNumber = strings.Split(userNumber, ":")[0]
	}

	h.logger.Debugf("Checking admin: userJID=%s, extracted=%s", userJID, userNumber)

	for _, admin := range h.adminNumbers {
		h.logger.Debugf("Comparing with admin: %s", admin)
		if admin == userNumber {
			h.logger.Debugf("Admin match found: %s", userNumber)
			return true
		}
	}

	h.logger.Debugf("No admin match for: %s", userNumber)
	return false
}

// truncateString memotong string jika terlalu panjang untuk logging
func (h *LearningMessageHandler) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
