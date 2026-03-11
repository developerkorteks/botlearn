// Package handlers berisi semua handler untuk menangani pesan dan event WhatsApp
// File ini khusus menangani pesan masuk dari chat personal dan grup
package handlers

import (
	"context"
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"github.com/nabilulilalbab/promote/utils"
)

// MessageHandler adalah struktur yang menangani semua pesan masuk
type MessageHandler struct {
	// client adalah instance WhatsApp client untuk mengirim pesan
	client *whatsmeow.Client

	// autoReplyPersonal menentukan apakah bot membalas chat personal
	autoReplyPersonal bool

	// autoReplyGroup menentukan apakah bot membalas chat grup
	autoReplyGroup bool

	// Typing simulator untuk human-like responses
	typingSimulator *utils.TypingSimulator
	logger          *utils.Logger
}

// NewMessageHandler membuat handler baru untuk pesan
// Parameter:
// - client: WhatsApp client yang sudah terhubung
// - autoReplyPersonal: true jika ingin auto reply di chat personal
// - autoReplyGroup: true jika ingin auto reply di grup (hati-hati spam!)
// - logger: Logger instance untuk debugging
func NewMessageHandler(client *whatsmeow.Client, autoReplyPersonal, autoReplyGroup bool, logger *utils.Logger) *MessageHandler {
	return &MessageHandler{
		client:            client,
		autoReplyPersonal: autoReplyPersonal,
		autoReplyGroup:    autoReplyGroup,
		typingSimulator:   utils.NewTypingSimulator(client, logger),
		logger:            logger,
	}
}

// SetAutoPromoteHandlers dihapus karena fitur promote dicabut

// sendMessageWithTyping mengirim pesan dengan simulasi typing yang natural
func (h *MessageHandler) sendMessageWithTyping(chatJID types.JID, message string) error {
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

	h.logger.Debugf("✅ Pesan terkirim dengan typing delay ke %s", chatJID)
	return nil
}

// sendQuickResponse mengirim response cepat dengan delay minimal
func (h *MessageHandler) sendQuickResponse(chatJID types.JID, message string) error {
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
// Fungsi ini akan dipanggil setiap kali ada pesan baru
func (h *MessageHandler) HandleMessage(evt *events.Message) {
	// STEP 1: Skip pesan dari diri sendiri
	// Ini penting untuk menghindari bot membalas pesannya sendiri (infinite loop)
	if evt.Info.IsFromMe {
		return
	}

	// STEP 2: Ambil teks dari pesan
	// WhatsApp memiliki beberapa tipe pesan, kita hanya proses yang teks
	messageText := h.getMessageText(evt.Message)
	if messageText == "" {
		// Jika bukan pesan teks (misal gambar, voice note), skip
		return
	}

	// STEP 3: Identifikasi jenis chat (personal atau grup)
	isGroup := evt.Info.Chat.Server == types.GroupServer
	chatType := "personal"
	if isGroup {
		chatType = "group"
	}

	// STEP 4: Log informasi pesan untuk debugging
	sender := evt.Info.Sender.User // Nomor pengirim (tanpa @s.whatsapp.net)
	fmt.Printf("📨 Pesan masuk [%s]: %s\n", chatType, messageText)
	fmt.Printf("👤 Dari: %s\n", sender)

	// Jika grup, tampilkan nama grup juga
	if isGroup {
		fmt.Printf("👥 Grup: %s\n", evt.Info.Chat.User)
	}

	// STEP 5: Proses pesan berdasarkan jenis chat
	if isGroup {
		h.handleGroupMessage(evt, messageText)
	} else {
		h.handlePersonalMessage(evt, messageText)
	}
}

// handlePersonalMessage menangani pesan dari chat personal (1 on 1)
func (h *MessageHandler) handlePersonalMessage(evt *events.Message, messageText string) {
	fmt.Println("💬 Memproses pesan personal...")

	// Cek apakah ini adalah command (dimulai dengan / atau .)
	if strings.HasPrefix(messageText, "/") || strings.HasPrefix(messageText, ".") {
		h.handleCommand(evt, messageText)
		return
	}

	// Bot tidak memberikan auto reply untuk non-admin
	// Hanya merespon command auto promote dari admin
}

// handleGroupMessage menangani pesan dari grup
func (h *MessageHandler) handleGroupMessage(evt *events.Message, messageText string) {
	fmt.Println("👥 Memproses pesan grup...")

	// BOT DIAM TOTAL DI GRUP - TIDAK ADA RESPONSE APAPUN
	// Bot hanya akan mengirim auto promote sesuai scheduler
	// Semua kontrol dilakukan melalui chat personal dengan admin

	// Log untuk monitoring (tanpa response)
	fmt.Printf("👥 Grup: %s | Pesan: %s | Action: IGNORED\n",
		evt.Info.Chat.User, h.truncateString(messageText, 30))

	// Bot tidak memberikan response apapun di grup
	return
}

// handleCommand menangani command yang dimulai dengan /
func (h *MessageHandler) handleCommand(evt *events.Message, messageText string) {
	var response string

	// Semua fitur promote command sudah dihapus
	// Tidak ada response untuk command yang tidak dikenal
	return
	if response != "" {
		h.sendMessageWithTyping(evt.Info.Chat, response)
	}
}

// sendAutoReply mengirim balasan otomatis
func (h *MessageHandler) sendAutoReply(chatJID types.JID, originalMessage string, isGroup bool) {
	var response string

	if isGroup {
		// Response untuk grup lebih formal dan tidak terlalu sering
		response = `👋 *AUTO-REPLY*

Terima kasih! Saya adalah bot otomatis.
Ketik */help* untuk bantuan.`
	} else {
		// Response untuk personal bisa lebih personal
		response = `👋 *AUTO-REPLY*

✅ Terima kasih atas pesannya!
Saya adalah bot otomatis yang siap membantu.

Ketik */help* untuk melihat command yang tersedia.`
	}

	h.sendMessageWithTyping(chatJID, response)
}

// getMessageText mengekstrak teks dari berbagai tipe pesan WhatsApp
func (h *MessageHandler) getMessageText(msg *waProto.Message) string {
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

// isBotMentioned mengecek apakah bot di-mention dalam pesan grup
func (h *MessageHandler) isBotMentioned(msg *waProto.Message) bool {
	// Cek di extended text message (yang biasanya berisi mention)
	if msg.GetExtendedTextMessage() != nil && msg.GetExtendedTextMessage().GetContextInfo() != nil {
		mentions := msg.GetExtendedTextMessage().GetContextInfo().GetMentionedJID()
		botJID := h.client.Store.ID.String()

		// Cek apakah JID bot ada dalam daftar mention
		for _, mention := range mentions {
			if mention == botJID {
				return true
			}
		}
	}

	return false
}

// sendMessage mengirim pesan ke chat tertentu (DEPRECATED - use sendMessageWithTyping)
// Function ini tetap ada untuk backward compatibility, tapi sekarang menggunakan typing delay
func (h *MessageHandler) sendMessage(chatJID types.JID, text string) {
	// Redirect ke sendMessageWithTyping untuk consistency
	h.sendMessageWithTyping(chatJID, text)
}

// Helper functions untuk pesan informatif

func (h *MessageHandler) getHelpMessage() string {
	return `📋 *BANTUAN WHATSAPP BOT*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
          *COMMAND TERSEDIA*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🤖 *BASIC COMMANDS*

• */start*
  _Mulai bot_

• */help*
  _Bantuan lengkap_

• */ping*
  _Test koneksi bot_

• */info*
  _Informasi tentang bot_

• */status*
  _Status bot saat ini_

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *TIPS PENGGUNAAN*

• *Chat Personal:* Bot membalas semua pesan
• *Di Grup:* Bot hanya respon command/mention
• *Command:* Ketik tanpa parameter untuk info

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📞 *SUPPORT:* Hubungi admin jika ada masalah`
}

func (h *MessageHandler) getInfoMessage() string {
	return `ℹ️ *INFORMASI BOT*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *DETAIL SISTEM*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🤖 *SPESIFIKASI*
📝 *Nama:* WhatsApp Bot
💻 *Bahasa:* Go (Golang)
📚 *Library:* whatsmeow + go-qrcode
✨ *Versi:* 1.0.0
🎯 *Fitur:* Visual QR, Auto-reply, Commands

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔧 *KONFIGURASI AKTIF*
• *Auto Reply Personal:* Aktif
• *Auto Reply Group:* Tidak aktif
• *Session:* Tersimpan otomatis
• *QR Code:* Visual display

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 Bot ini dibuat untuk pembelajaran dan automasi WhatsApp`
}

func (h *MessageHandler) getStatusMessage() string {
	return fmt.Sprintf(`📊 *STATUS BOT*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *SISTEM STATUS*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔋 *STATUS UTAMA*
✅ *Status:* Online dan aktif
🔗 *Koneksi:* Terhubung ke WhatsApp
💾 *Session:* Tersimpan di database
🤖 *Bot ID:* %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚙️ *KONFIGURASI*
📱 *Auto Reply Personal:* %v
👥 *Auto Reply Group:* %v

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🟢 *Semua sistem berjalan normal!*`,
		h.client.Store.ID.User,
		h.autoReplyPersonal,
		h.autoReplyGroup)
}

// truncateString memotong string jika terlalu panjang untuk logging
func (h *MessageHandler) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Fungsi-fungsi isAutoPromoteCommand dan handleAutoPromoteCommand dihapus
