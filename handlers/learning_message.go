// Package handlers - Learning message handler untuk bot pembelajaran
package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"github.com/nabilulilalbab/promote/database"
	"github.com/nabilulilalbab/promote/services"
	"github.com/nabilulilalbab/promote/utils"
)

// LearningMessageHandler menangani pesan untuk bot pembelajaran
type LearningMessageHandler struct {
	client             *whatsmeow.Client
	learningService    *services.LearningService
	xrayConverterService *services.XRayConverterService
	logger             *utils.Logger
	adminNumbers       []string // Daftar nomor admin
	
	// Rate limiting: map[userJID]lastCommandTime
	commandCooldown    map[string]time.Time
	
	// Typing simulator untuk human-like responses
	typingSimulator    *utils.TypingSimulator
}

// NewLearningMessageHandler membuat handler baru untuk learning bot
func NewLearningMessageHandler(
	client *whatsmeow.Client,
	learningService *services.LearningService,
	xrayConverterService *services.XRayConverterService,
	logger *utils.Logger,
	adminNumbers []string,
) *LearningMessageHandler {
	return &LearningMessageHandler{
		client:             client,
		learningService:    learningService,
		xrayConverterService: xrayConverterService,
		logger:             logger,
		adminNumbers:       adminNumbers,
		commandCooldown:    make(map[string]time.Time),
		typingSimulator:    utils.NewTypingSimulator(client, logger),
	}
}

// sendMessageWithTyping mengirim pesan dengan simulasi typing yang natural
func (h *LearningMessageHandler) sendMessageWithTyping(chatJID types.JID, message string) error {
	// Simulasi typing berdasarkan kompleksitas pesan
	h.typingSimulator.SmartTypingDelay(chatJID, message)
	
	// Kirim pesan setelah simulasi typing selesai
	_, err := h.client.SendMessage(context.Background(), chatJID, &waProto.Message{
		Conversation: &message,
	})
	
	if err != nil {
		h.logger.Errorf("Gagal kirim pesan ke %s: %v", chatJID, err)
		return err
	}
	
	h.logger.Debugf("✅ Pesan learning terkirim dengan typing delay ke %s", chatJID)
	return nil
}

// sendQuickResponse mengirim response cepat dengan delay minimal
func (h *LearningMessageHandler) sendQuickResponse(chatJID types.JID, message string) error {
	// Quick delay untuk response singkat
	h.typingSimulator.QuickDelay(chatJID)
	
	// Kirim pesan
	_, err := h.client.SendMessage(context.Background(), chatJID, &waProto.Message{
		Conversation: &message,
	})
	
	if err != nil {
		h.logger.Errorf("Gagal kirim quick response ke %s: %v", chatJID, err)
		return err
	}
	
	h.logger.Debugf("⚡ Quick response terkirim ke %s", chatJID)
	return nil
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
		return // Skip jika bukan pesan teks
	}

	// STEP 3: Identifikasi chat type dan IDs
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
	// Rate limiting: 1 command per 3 seconds per user
	cooldownKey := fmt.Sprintf("%s:%s", userJID, groupJID)
	now := time.Now()
	
	if lastTime, exists := h.commandCooldown[cooldownKey]; exists {
		if now.Sub(lastTime) < 3*time.Second {
			h.logger.Debugf("🕒 Rate limit: User %s in cooldown, ignoring command: %s", userJID, command)
			return
		}
	}
	
	// Update cooldown time
	h.commandCooldown[cooldownKey] = now
	
	h.logger.Infof("🔧 Processing learning command: %s | Group: %s | User: %s",
		command, groupJID, userJID)

	// Cek apakah ini XRay converter command
	if h.isXRayConverterCommand(command) {
		h.handleXRayConverterCommand(groupJID, userJID, command)
		return
	}

	// Process normal learning command
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

// handleAdminCommand menangani command admin dari personal chat
func (h *LearningMessageHandler) handleAdminCommand(evt *events.Message, userJID, command string) {
	// Rate limiting untuk admin: 1 command per 2 seconds
	cooldownKey := fmt.Sprintf("admin:%s", userJID)
	now := time.Now()
	
	if lastTime, exists := h.commandCooldown[cooldownKey]; exists {
		if now.Sub(lastTime) < 2*time.Second {
			h.logger.Debugf("🕒 Admin rate limit: User %s in cooldown, ignoring command: %s", userJID, command)
			return
		}
	}
	
	// Update cooldown time
	h.commandCooldown[cooldownKey] = now
	
	h.logger.Infof("🔧 Processing admin command: %s | User: %s", command, userJID)
	
	// Cek apakah ini XRay converter command
	if h.isXRayConverterCommand(command) {
		h.handleXRayConverterCommand(evt.Info.Chat.String(), userJID, command)
		return
	}
	
	// Command untuk mengelola grup pembelajaran
	switch {
	case strings.HasPrefix(command, ".addgroup"):
		h.handleAddGroupCommand(evt, userJID, command)
	case strings.HasPrefix(command, ".removegroup"):
		h.handleRemoveGroupCommand(evt, userJID, command)
	case strings.HasPrefix(command, ".listgroups"):
		h.handleListGroupsCommand(evt, userJID)
	case strings.HasPrefix(command, ".stats"):
		h.handleStatsCommand(evt, userJID)
	case strings.HasPrefix(command, ".logs"):
		h.handleLogsCommand(evt, userJID)
	case command == ".help":
		h.sendAdminHelp(evt.Info.Chat)
	default:
		// Try processing as learning command
		err := h.learningService.ProcessCommand(evt.Info.Chat.String(), userJID, command)
		if err != nil {
			h.logger.Errorf("Failed to process admin command %s: %v", command, err)
			h.sendAdminMessage(evt.Info.Chat, fmt.Sprintf("❌ Command tidak dikenali: %s\n\nKetik .help untuk bantuan.", command))
		}
	}
}

// === XRAY CONVERTER HANDLERS ===

// isXRayConverterCommand cek apakah command adalah XRay converter
func (h *LearningMessageHandler) isXRayConverterCommand(command string) bool {
	// Parse command untuk extract nama converter dan XRay link
	parts := strings.Fields(command)
	if len(parts) < 2 {
		return false
	}
	
	commandName := parts[0]
	
	// Cek apakah command dimulai dengan .convert atau custom command yang ada di database
	if strings.HasPrefix(commandName, ".convert") {
		return true
	}
	
	// Cek di database apakah ada converter dengan nama ini
	converterName := strings.TrimPrefix(commandName, ".")
	converter, err := h.xrayConverterService.GetAllConverters()
	if err != nil {
		return false
	}
	
	for _, conv := range converter {
		if conv.CommandName == converterName && conv.IsActive {
			return true
		}
	}
	
	return false
}

// handleXRayConverterCommand menangani XRay converter command
func (h *LearningMessageHandler) handleXRayConverterCommand(groupJID, userJID, command string) {
	// Parse command: .convertbizz vmess://xxx
	parts := strings.Fields(command)
	if len(parts) < 2 {
		h.sendErrorMessage(groupJID, "❌ Format salah!\n\nContoh: .convertbizz vmess://xxx")
		return
	}
	
	commandName := strings.TrimPrefix(parts[0], ".")
	xrayLink := parts[1]
	
	h.logger.Infof("🔄 Processing XRay conversion: %s | Link: %s", commandName, h.truncateString(xrayLink, 50))
	
	// Process conversion
	result, err := h.xrayConverterService.ProcessConversion(commandName, xrayLink, userJID, groupJID)
	if err != nil {
		h.logger.Errorf("XRay conversion failed: %v", err)
		
		errorMsg := fmt.Sprintf("❌ **Conversion Failed!**\n\n🔧 **Command:** %s\n📝 **Error:** %s\n\n💡 **Tips:**\n• Pastikan link XRay valid\n• Cek format: vmess://, vless://, trojan://\n• Command tersedia: %s", 
			commandName, err.Error(), h.getAvailableConverters())
		
		h.sendErrorMessage(groupJID, errorMsg)
		return
	}
	
	// Send success response
	h.sendConversionResult(groupJID, result, commandName)
}

// sendConversionResult mengirim hasil conversion ke grup (2 pesan terpisah)
func (h *LearningMessageHandler) sendConversionResult(groupJID string, result *database.ModifiedXRayConfig, commandName string) {
	// Parse JID untuk chat target
	chatJID, err := types.ParseJID(groupJID)
	if err != nil {
		h.logger.Errorf("Failed to parse group JID: %v", err)
		return
	}
	
	// Get converter info
	converter, _ := h.xrayConverterService.GetAllConverters()
	var displayName string
	for _, conv := range converter {
		if conv.CommandName == commandName {
			displayName = conv.DisplayName
			break
		}
	}
	if displayName == "" {
		displayName = strings.ToUpper(commandName)
	}
	
	// === PESAN 1: INFO & DETAILS ===
	var infoBuilder strings.Builder
	
	// Header dengan emoji dan info
	infoBuilder.WriteString("✅ *Conversion Success!*\n\n")
	infoBuilder.WriteString(fmt.Sprintf("🏷️ *Converter:* %s\n", displayName))
	infoBuilder.WriteString(fmt.Sprintf("🔧 *Type:* %s\n", strings.ToUpper(result.ModifyType)))
	infoBuilder.WriteString(fmt.Sprintf("📡 *Protocol:* %s | *Network:* %s | *TLS:* %s\n\n", 
		strings.ToUpper(result.DetectedConfig.Protocol), 
		strings.ToUpper(result.DetectedConfig.Network),
		func() string { if result.DetectedConfig.TLS { return "Yes" } else { return "No" } }()))
	
	// Modification details dengan format rapi
	infoBuilder.WriteString("🔍 *Modification Details:*\n")
	infoBuilder.WriteString(fmt.Sprintf("• Original Server: %s\n", result.DetectedConfig.Server))
	infoBuilder.WriteString(fmt.Sprintf("• Bug Host: %s\n", result.BugHost))
	
	switch result.ModifyType {
	case "wildcard":
		infoBuilder.WriteString(fmt.Sprintf("• Modified Server: %s\n", result.ModifiedServer))
		infoBuilder.WriteString(fmt.Sprintf("• Modified Host: %s\n", result.ModifiedHost))
		if result.DetectedConfig.TLS {
			infoBuilder.WriteString(fmt.Sprintf("• Modified SNI: %s\n", result.ModifiedSNI))
		}
	case "sni":
		infoBuilder.WriteString(fmt.Sprintf("• Modified SNI: %s\n", result.ModifiedSNI))
		infoBuilder.WriteString("• Server & Host: _unchanged_\n")
	case "ws", "grpc":
		infoBuilder.WriteString(fmt.Sprintf("• Modified Server: %s\n", result.ModifiedServer))
		infoBuilder.WriteString("• Host & SNI: _unchanged_\n")
	}
	
	// YAML Configuration dengan format rapi
	infoBuilder.WriteString("\n📁 *YAML Configuration:*\n")
	infoBuilder.WriteString("```yaml\n")
	infoBuilder.WriteString(result.YAMLConfig)
	infoBuilder.WriteString("```\n\n")
	
	infoBuilder.WriteString("💡 *Usage Instructions:*\n")
	infoBuilder.WriteString("1. Copy modified link untuk V2Ray/Xray\n")
	infoBuilder.WriteString("2. Copy YAML config untuk Clash/OpenClash\n")
	infoBuilder.WriteString("3. Restart aplikasi setelah config\n\n")
	infoBuilder.WriteString("📱 _Modified link akan dikirim di pesan berikutnya untuk kemudahan copy..._")
	
	// Kirim pesan 1 dengan typing delay
	infoText := infoBuilder.String()
	err = h.sendMessageWithTyping(chatJID, infoText)
	if err != nil {
		h.logger.Errorf("Failed to send conversion info: %v", err)
		return
	}
	
	// === PESAN 2: MODIFIED LINK ONLY ===
	linkText := result.ModifiedLink
	
	// Kirim pesan 2 dengan typing delay
	err = h.sendMessageWithTyping(chatJID, linkText)
	if err != nil {
		h.logger.Errorf("Failed to send conversion link: %v", err)
	} else {
		h.logger.Infof("✅ Conversion result sent to %s (2 messages)", groupJID)
	}
}

// sendErrorMessage mengirim pesan error
func (h *LearningMessageHandler) sendErrorMessage(groupJID, errorMsg string) {
	chatJID, err := types.ParseJID(groupJID)
	if err != nil {
		h.logger.Errorf("Failed to parse group JID: %v", err)
		return
	}
	
	h.sendMessageWithTyping(chatJID, errorMsg)
}

// getAvailableConverters mendapatkan daftar converter yang tersedia
func (h *LearningMessageHandler) getAvailableConverters() string {
	converters, err := h.xrayConverterService.GetActiveConverters()
	if err != nil || len(converters) == 0 {
		return "Tidak ada converter aktif"
	}
	
	var available []string
	for _, conv := range converters {
		available = append(available, fmt.Sprintf(".%s", conv.CommandName))
	}
	
	return strings.Join(available, ", ")
}

// === ADMIN COMMAND HANDLERS ===

// sendAdminHelp mengirim bantuan untuk admin
func (h *LearningMessageHandler) sendAdminHelp(chatJID types.JID) {
	helpText := `🤖 **BANTUAN ADMIN BOT PEMBELAJARAN** 🤖

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           **COMMAND MANAGEMENT GRUP**
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 **Group Management:**
• .addgroup [JID] [Nama] - Tambah grup ke whitelist
• .removegroup [JID] - Hapus grup dari whitelist
• .listgroups - List semua grup yang diizinkan

📊 **Statistics:**
• .stats - Statistik penggunaan bot
• .logs - Log aktivitas terakhir

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           **XRAY CONVERTER COMMANDS**
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 **Converter Commands:**
• .convertbizz [vmess://xxx] - XL-Line-WC (Wildcard)
• .convertinsta [vmess://xxx] - XL-Instagram-SNI 
• .convertnetflix [vmess://xxx] - XL-Netflix-WS
• .convertgopay [vmess://xxx] - XL-Gopay-Midtrans-WC
• .convertgrpc [vmess://xxx] - Generic-gRPC

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           **LEARNING COMMANDS**
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📚 **Default Commands:**
• .help - Bantuan umum
• .info - Info tentang bot
• .listbugs - List bug server VPN

💡 **Dashboard:** http://localhost:1462
🌐 **Manage via web:** Groups, Commands, Auto Response

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

**Bot siap melayani!** 🚀`

	h.sendAdminMessage(chatJID, helpText)
}

// sendAdminMessage mengirim pesan ke admin dengan typing delay
func (h *LearningMessageHandler) sendAdminMessage(chatJID types.JID, message string) {
	err := h.sendMessageWithTyping(chatJID, message)
	if err != nil {
		h.logger.Errorf("Failed to send admin message: %v", err)
	}
}

// handleAddGroupCommand menangani command untuk menambah grup
func (h *LearningMessageHandler) handleAddGroupCommand(evt *events.Message, userJID, command string) {
	// Parse: .addgroup 120363420243864186@g.us Grup Test
	parts := strings.Fields(command)
	if len(parts) < 3 {
		h.sendAdminMessage(evt.Info.Chat, "❌ Format salah!\n\nContoh: .addgroup 120363420243864186@g.us Grup Pembelajaran")
		return
	}
	
	groupJID := parts[1]
	groupName := strings.Join(parts[2:], " ")
	
	err := h.learningService.AddAllowedGroup(groupJID, groupName, userJID)
	if err != nil {
		h.logger.Errorf("Failed to add group: %v", err)
		h.sendAdminMessage(evt.Info.Chat, fmt.Sprintf("❌ Gagal menambah grup: %v", err))
		return
	}
	
	h.sendAdminMessage(evt.Info.Chat, fmt.Sprintf("✅ Grup berhasil ditambahkan!\n\n📋 **Grup:** %s\n🆔 **JID:** %s\n\nBot sekarang aktif di grup tersebut.", groupName, groupJID))
}

// handleRemoveGroupCommand menangani command untuk menghapus grup
func (h *LearningMessageHandler) handleRemoveGroupCommand(evt *events.Message, userJID, command string) {
	// Parse: .removegroup 120363420243864186@g.us
	parts := strings.Fields(command)
	if len(parts) < 2 {
		h.sendAdminMessage(evt.Info.Chat, "❌ Format salah!\n\nContoh: .removegroup 120363420243864186@g.us")
		return
	}
	
	groupJID := parts[1]
	
	err := h.learningService.RemoveAllowedGroup(groupJID)
	if err != nil {
		h.logger.Errorf("Failed to remove group: %v", err)
		h.sendAdminMessage(evt.Info.Chat, fmt.Sprintf("❌ Gagal menghapus grup: %v", err))
		return
	}
	
	h.sendAdminMessage(evt.Info.Chat, fmt.Sprintf("✅ Grup berhasil dihapus!\n\n🆔 **JID:** %s\n\nBot tidak lagi aktif di grup tersebut.", groupJID))
}

// handleListGroupsCommand menangani command untuk list grup
func (h *LearningMessageHandler) handleListGroupsCommand(evt *events.Message, userJID string) {
	groups, err := h.learningService.GetAllowedGroups()
	if err != nil {
		h.logger.Errorf("Failed to get groups: %v", err)
		h.sendAdminMessage(evt.Info.Chat, "❌ Gagal mengambil daftar grup")
		return
	}
	
	if len(groups) == 0 {
		h.sendAdminMessage(evt.Info.Chat, "📋 Belum ada grup yang diizinkan.\n\nGunakan .addgroup untuk menambah grup.")
		return
	}
	
	var response strings.Builder
	response.WriteString("📋 **DAFTAR GRUP YANG DIIZINKAN**\n\n")
	
	activeCount := 0
	for i, group := range groups {
		status := "✅ Aktif"
		if !group.IsActive {
			status = "❌ Nonaktif"
		} else {
			activeCount++
		}
		
		response.WriteString(fmt.Sprintf("**%d. %s**\n", i+1, group.GroupName))
		response.WriteString(fmt.Sprintf("🆔 JID: `%s`\n", group.GroupJID))
		response.WriteString(fmt.Sprintf("📊 Status: %s\n", status))
		response.WriteString(fmt.Sprintf("👤 Ditambah: %s\n\n", group.CreatedBy))
	}
	
	response.WriteString(fmt.Sprintf("📊 **Total:** %d grup | **Aktif:** %d grup", len(groups), activeCount))
	
	h.sendAdminMessage(evt.Info.Chat, response.String())
}

// handleStatsCommand menangani command statistik
func (h *LearningMessageHandler) handleStatsCommand(evt *events.Message, userJID string) {
	statsText := `📊 **STATISTIK BOT PEMBELAJARAN** 📊

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 **XRay Converter:**
Gunakan dashboard untuk stats lengkap
🌐 http://localhost:1462

📚 **Learning Commands:**
Lihat usage stats di dashboard

🔧 **System Status:**
✅ Learning System: Running
✅ XRay Converter: Running  
✅ Web Dashboard: Running

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

	h.sendAdminMessage(evt.Info.Chat, statsText)
}

// handleLogsCommand menangani command logs
func (h *LearningMessageHandler) handleLogsCommand(evt *events.Message, userJID string) {
	logsText := `📋 **BOT LOGS** 📋

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 **Recent Activity:**
Lihat logs real-time di console

🌐 **Detailed Logs:**
Dashboard: http://localhost:1462

💡 **Tips:**
- Monitor console untuk real-time logs
- Dashboard menyediakan logs terstruktur
- XRay conversion logs tersimpan otomatis

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

	h.sendAdminMessage(evt.Info.Chat, logsText)
}

// === UTILITY FUNCTIONS ===

// getMessageText mengekstrak teks dari pesan WhatsApp
func (h *LearningMessageHandler) getMessageText(message *waProto.Message) string {
	if message.Conversation != nil {
		return *message.Conversation
	}
	
	if message.ExtendedTextMessage != nil && message.ExtendedTextMessage.Text != nil {
		return *message.ExtendedTextMessage.Text
	}
	
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