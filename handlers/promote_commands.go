// Package handlers - Command handlers untuk fitur auto promote
package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"github.com/nabilulilalbab/promote/services"
	"github.com/nabilulilalbab/promote/utils"
)

// PromoteCommandHandler menangani command-command auto promote
type PromoteCommandHandler struct {
	autoPromoteService *services.AutoPromoteService
	templateService    *services.TemplateService
	logger             *utils.Logger
	client             *whatsmeow.Client
	typingSimulator    *utils.TypingSimulator
}

// NewPromoteCommandHandler membuat handler baru
func NewPromoteCommandHandler(
	autoPromoteService *services.AutoPromoteService,
	templateService *services.TemplateService,
	logger *utils.Logger,
	client *whatsmeow.Client,
) *PromoteCommandHandler {
	return &PromoteCommandHandler{
		autoPromoteService: autoPromoteService,
		templateService:    templateService,
		logger:             logger,
		client:             client,
		typingSimulator:    utils.NewTypingSimulator(client, logger),
	}
}

// sendMessageWithTyping mengirim pesan dengan simulasi typing yang natural
func (h *PromoteCommandHandler) sendMessageWithTyping(chatJID types.JID, message string) error {
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
	
	h.logger.Debugf("✅ Pesan promote terkirim dengan typing delay ke %s", chatJID)
	return nil
}

// HandleAcaCommand menangani command .aca (dulu .promote)
func (h *PromoteCommandHandler) HandleAcaCommand(evt *events.Message) string {
	// Hanya bisa digunakan di grup
	if evt.Info.Chat.Server != types.GroupServer {
		return `❌ *COMMAND TIDAK VALID*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *HANYA UNTUK GRUP*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Command .aca hanya bisa digunakan di grup

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Command ini khusus untuk grup WhatsApp
• Tidak bisa digunakan di chat personal
• Gunakan /help untuk command umum

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba gunakan di grup yang sesuai*`
	}

	groupJID := evt.Info.Chat.String()

	// Aktifkan auto promote
	err := h.autoPromoteService.StartAutoPromote(groupJID)
	if err != nil {
		h.logger.Errorf("Failed to start auto promote for %s: %v", groupJID, err)
		return fmt.Sprintf(`❌ *GAGAL MENGAKTIFKAN AUTO PROMOTE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mengaktifkan auto promote: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Masalah koneksi database
• Template promosi belum tersedia
• Grup sudah aktif sebelumnya
• Masalah konfigurasi sistem

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi admin*`, err.Error())
	}

	return `✅ *AUTO PROMOTE DIAKTIFKAN!*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *BERHASIL AKTIF*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 Auto promote telah diaktifkan untuk grup ini
⏰ Bot akan mengirim promosi sesuai interval
🎲 Template akan dipilih secara random
📊 Status dapat dipantau dengan .statuspromo

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚀 *Sistem auto promote siap bekerja!*`
}

// HandleDisableAcaCommand menangani command .disableaca (dulu .disablepromote)
func (h *PromoteCommandHandler) HandleDisableAcaCommand(evt *events.Message) string {
	// Hanya bisa digunakan di grup
	if evt.Info.Chat.Server != types.GroupServer {
		return `❌ *COMMAND TIDAK VALID*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *HANYA UNTUK GRUP*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Command .disableaca hanya bisa digunakan di grup

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Command ini khusus untuk grup WhatsApp
• Tidak bisa digunakan di chat personal
• Gunakan /help untuk command umum

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba gunakan di grup yang sesuai*`
	}

	groupJID := evt.Info.Chat.String()

	// Nonaktifkan auto promote
	err := h.autoPromoteService.StopAutoPromote(groupJID)
	if err != nil {
		h.logger.Errorf("Failed to stop auto promote for %s: %v", groupJID, err)
		return fmt.Sprintf(`❌ *GAGAL MENONAKTIFKAN AUTO PROMOTE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal menonaktifkan auto promote: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Masalah koneksi database
• Auto promote sudah tidak aktif
• Grup tidak terdaftar di sistem
• Masalah konfigurasi sistem

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi admin*`, err.Error())
	}

	return `🛑 *AUTO PROMOTE DINONAKTIFKAN!*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *BERHASIL DINONAKTIFKAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 Auto promote telah dinonaktifkan untuk grup ini
⏰ Bot tidak akan mengirim promosi lagi
📊 Data grup tetap tersimpan di sistem
🔄 Dapat diaktifkan kembali kapan saja

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ *Auto promote berhasil dihentikan!*`
}

// HandleStatusPromoCommand menangani command .statuspromo
func (h *PromoteCommandHandler) HandleStatusPromoCommand(evt *events.Message) string {
	// Hanya bisa digunakan di grup
	if evt.Info.Chat.Server != types.GroupServer {
		return `❌ *COMMAND TIDAK VALID*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *HANYA UNTUK GRUP*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Command .statuspromo hanya bisa digunakan di grup

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Command ini khusus untuk grup WhatsApp
• Tidak bisa digunakan di chat personal
• Gunakan /help untuk command umum

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba gunakan di grup yang sesuai*`
	}

	groupJID := evt.Info.Chat.String()

	// Ambil status grup
	group, err := h.autoPromoteService.GetGroupStatus(groupJID)
	if err != nil {
		h.logger.Errorf("Failed to get group status for %s: %v", groupJID, err)
		return `❌ *GAGAL MENDAPATKAN STATUS GRUP*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mendapatkan status grup

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Masalah koneksi database
• Grup belum terdaftar di sistem
• Masalah konfigurasi sistem
• Error internal server

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi admin*`
	}

	if group == nil {
		return `📊 *STATUS AUTO PROMOTE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
            *TIDAK TERDAFTAR*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

❌ *Status:* Tidak terdaftar
💡 *Info:* Grup ini belum pernah menggunakan auto promote

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚀 *Gunakan .promote untuk mengaktifkan auto promote*`
	}

	// Format status
	status := "❌ Tidak Aktif"
	if group.IsActive {
		status = "✅ Aktif"
	}

	var startedInfo string
	if group.StartedAt != nil {
		startedInfo = group.StartedAt.Format("2006-01-02 15:04")
	} else {
		startedInfo = "Belum pernah"
	}

	var lastPromoteInfo string
	if group.LastPromoteAt != nil {
		lastPromoteInfo = group.LastPromoteAt.Format("2006-01-02 15:04")
	} else {
		lastPromoteInfo = "Belum pernah"
	}

	// Ambil jumlah template aktif
	templates, _ := h.templateService.GetActiveTemplates()
	templateCount := len(templates)

	return fmt.Sprintf(`📊 *STATUS AUTO PROMOTE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *STATUS GRUP INI*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *Status:* %s
📅 *Dimulai:* %s
⏰ *Promosi Terakhir:* %s
📝 *Template Tersedia:* %d template

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *COMMANDS TERSEDIA*
• *.promote* - Aktifkan auto promote
• *.disablepromote* - Nonaktifkan auto promote
• *.testpromo* - Test kirim promosi
• *.listtemplates* - Lihat template

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *Status diperbarui real-time*`, status, startedInfo, lastPromoteInfo, templateCount)
}

// HandleTestPromoCommand menangani command .testpromo
func (h *PromoteCommandHandler) HandleTestPromoCommand(evt *events.Message) string {
	// Hanya bisa digunakan di grup
	if evt.Info.Chat.Server != types.GroupServer {
		return `❌ *COMMAND TIDAK VALID*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *HANYA UNTUK GRUP*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Command .testpromo hanya bisa digunakan di grup

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Command ini khusus untuk grup WhatsApp
• Tidak bisa digunakan di chat personal
• Gunakan /help untuk command umum

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba gunakan di grup yang sesuai*`
	}

	groupJID := evt.Info.Chat.String()

	// Kirim promosi manual
	err := h.autoPromoteService.SendManualPromote(groupJID)
	if err != nil {
		h.logger.Errorf("Failed to send manual promote for %s: %v", groupJID, err)
		return fmt.Sprintf(`❌ *GAGAL MENGIRIM TEST PROMOSI*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mengirim test promosi: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Template promosi tidak tersedia
• Masalah koneksi WhatsApp
• Grup tidak terdaftar di sistem
• Bot tidak memiliki izin kirim pesan

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi admin*`, err.Error())
	}

	return `🚀 *PROMOSI BERHASIL DIKIRIM!*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
          *BERHASIL TERKIRIM*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Promosi telah dikirim ke grup ini
🎲 Template dipilih secara random
📝 Contoh bagaimana auto promote bekerja

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Tidak mempengaruhi jadwal auto promote
• Auto promote tetap berjalan sesuai interval
• Gunakan *.statuspromo* untuk cek status

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *Cek hasil promosi di grup!*`
}

// HandleListTemplatesCommand menangani command .listtemplates
func (h *PromoteCommandHandler) HandleListTemplatesCommand(evt *events.Message) string {
	templates, err := h.templateService.GetActiveTemplates()
	if err != nil {
		h.logger.Errorf("Failed to get templates: %v", err)
		return `❌ *GAGAL MENDAPATKAN TEMPLATE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		         *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mendapatkan daftar template aktif.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Masalah koneksi database.
• Service template tidak tersedia.
• Error internal sistem.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi admin*`
	}

	if len(templates) == 0 {
		return `📝 *DAFTAR TEMPLATE PROMOSI*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
            *TIDAK ADA TEMPLATE*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

❌ Tidak ada template aktif yang tersedia

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Admin belum menambahkan template promosi
• Gunakan *.addtemplate* untuk menambah template
• Contoh: *.addtemplate* "Promo Hari Ini" "diskon" "🔥 Diskon 50%!"
• Gunakan *.alltemplates* untuk melihat semua template

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *Hubungi admin untuk menambah template*`
	}

	var result strings.Builder
	result.WriteString("📝 *DAFTAR TEMPLATE PROMOSI*\n\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString(fmt.Sprintf("        *TOTAL: %d TEMPLATE AKTIF*\n", len(templates)))
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, template := range templates {
		if i >= 15 { // Batasi tampilan maksimal 15 template
			result.WriteString(fmt.Sprintf("... dan %d template lainnya\n\n", len(templates)-15))
			break
		}

		result.WriteString(fmt.Sprintf("🆔 *ID: %d* - %s\n", template.ID, template.Title))
		result.WriteString(fmt.Sprintf("📂 *Kategori:* %s\n", template.Category))
		result.WriteString(fmt.Sprintf("📅 *Dibuat:* %s\n", template.CreatedAt.Format("2006-01-02")))
		result.WriteString(fmt.Sprintf("✅ *Status:* %s\n", getTemplateStatusText(template.IsActive)))

		if i < len(templates)-1 && i < 14 {
			result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
		}
	}

	result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("            *COMMANDS*\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString("• *.previewtemplate [ID]*\n")
	result.WriteString("  _Preview template_\n\n")
	result.WriteString("• *.alltemplates*\n")
	result.WriteString("  _Lihat semua template_\n\n")
	result.WriteString("• *.addtemplate* (admin)\n")
	result.WriteString("  _Tambah template_\n\n")
	result.WriteString("• *.edittemplate [ID]* (admin)\n")
	result.WriteString("  _Edit template_\n\n")
	result.WriteString("• *.deletetemplate [ID]* (admin)\n")
	result.WriteString("  _Hapus template_\n\n")
	result.WriteString("💡 *Contoh:* .previewtemplate 1 atau .deletetemplate 5")

	return result.String()
}

// getTemplateStatusText helper function untuk status template
func getTemplateStatusText(isActive bool) string {
	if isActive {
		return "Aktif ✅"
	}
	return "Tidak Aktif ❌"
}

// HandlePreviewTemplateCommand menangani command .previewtemplate [ID]
func (h *PromoteCommandHandler) HandlePreviewTemplateCommand(evt *events.Message, args []string) string {
	if len(args) < 2 {
		return `❌ *FORMAT SALAH*

📝 **Format:** .previewtemplate [ID]
📋 **Contoh:** .previewtemplate 1

💡 Gunakan .listtemplates untuk melihat daftar template`
	}

	// Parse ID template
	templateID, err := strconv.Atoi(args[1])
	if err != nil {
		return `❌ *ID TIDAK VALID*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		         *FORMAT ID SALAH*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 ID template harus berupa angka.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *CONTOH PENGGUNAAN*
• .previewtemplate 1
• .previewtemplate 5

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *Gunakan .listtemplates untuk melihat ID*`
	}

	// Preview template
	preview, err := h.templateService.PreviewTemplate(templateID)
	if err != nil {
		h.logger.Errorf("Failed to preview template %d: %v", templateID, err)
		return fmt.Sprintf(`❌ *GAGAL PREVIEW TEMPLATE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		         *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mendapatkan preview template: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Template dengan ID tersebut tidak ditemukan.
• Masalah koneksi database.
• Error internal sistem.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Periksa ID atau hubungi admin*`, err.Error())
	}

	return preview
}

// HandleAllTemplatesCommand menangani command .alltemplates
func (h *PromoteCommandHandler) HandleAllTemplatesCommand(evt *events.Message) string {
	templates, err := h.templateService.GetAllTemplates()
	if err != nil {
		h.logger.Errorf("Failed to get all templates: %v", err)
		return `❌ *GAGAL MENDAPATKAN TEMPLATE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		         *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mendapatkan semua template.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Masalah koneksi database.
• Service template tidak tersedia.
• Error internal sistem.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi admin*`
	}

	if len(templates) == 0 {
		return `📝 *SEMUA TEMPLATE PROMOSI*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
            *DATABASE KOSONG*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

❌ Database template masih kosong

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *CARA MENAMBAH TEMPLATE (ADMIN)*
• *.addtemplate* "Judul" "Kategori" "Konten"
• Contoh: *.addtemplate* "Flash Sale" "diskon" "🔥 FLASH SALE! Diskon 70%!"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 *KATEGORI YANG DISARANKAN*
• produk, diskon, testimoni, flashsale
• bundle, ongkir, cashback, limited, contact

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *Hubungi admin untuk menambah template*`
	}

	var result strings.Builder
	result.WriteString("📝 *SEMUA TEMPLATE PROMOSI*\n\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString(fmt.Sprintf("        *TOTAL: %d TEMPLATE*\n", len(templates)))
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	activeCount := 0
	inactiveCount := 0

	for i, template := range templates {
		if template.IsActive {
			activeCount++
		} else {
			inactiveCount++
		}

		statusIcon := "✅"
		if !template.IsActive {
			statusIcon = "❌"
		}

		result.WriteString(fmt.Sprintf("%s *ID: %d* - %s\n", statusIcon, template.ID, template.Title))
		result.WriteString(fmt.Sprintf("📂 *Kategori:* %s\n", template.Category))
		result.WriteString(fmt.Sprintf("📅 *Dibuat:* %s\n", template.CreatedAt.Format("2006-01-02")))
		result.WriteString(fmt.Sprintf("✅ *Status:* %s\n", getTemplateStatusText(template.IsActive)))

		if i < len(templates)-1 {
			result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
		}
	}

	result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("            *RINGKASAN*\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString(fmt.Sprintf("✅ *Aktif:* %d template\n", activeCount))
	result.WriteString(fmt.Sprintf("❌ *Tidak Aktif:* %d template\n", inactiveCount))

	result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("            *COMMANDS ADMIN*\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString("• *.deletetemplate [ID]* - Hapus template\n")
	result.WriteString("• *.edittemplate [ID]* - Edit template\n")
	result.WriteString("• *.previewtemplate [ID]* - Preview template")

	return result.String()
}

// HandleHelpCommand menangani command .help
func (h *PromoteCommandHandler) HandleHelpCommand(evt *events.Message) string {
	return `🤖 *PANDUAN AUTO PROMOTE SYSTEM*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
          *ADMIN COMMANDS*
        _(Personal Chat Only)_
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🏠 *GROUP MANAGEMENT*

• *.listgroups*
  _Lihat semua grup yang diikuti_

• *.enablegroup* [ID]
  _Aktifkan auto promote grup_
  Contoh: .enablegroup 3

• *.disablegroup* [ID]
  _Nonaktifkan auto promote grup_
  Contoh: .disablegroup 3

• *.groupstatus* [ID]
  _Status detail grup_
  Contoh: .groupstatus 3

• *.testgroup* [ID]
  _Kirim promosi ke grup_
  Contoh: .testgroup 3

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *TEMPLATE MANAGEMENT*

• *.listtemplates*
  _Lihat template aktif_

• *.alltemplates*
  _Lihat semua template_

• *.previewtemplate* [ID]
  _Preview template_
  Contoh: .previewtemplate 5

• *.help*
  _Bantuan lengkap_

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚙️ *ADVANCED COMMANDS*

• *.addtemplate* "Judul" "Kategori"
  _Tambah template baru_

• *.edittemplate* [ID] "Judul"
  _Edit template existing_

• *.deletetemplate* [ID]
  _Hapus template_

• *.deleteall*
  _Hapus semua template_

• *.deletemulti* [ID1,ID2,ID3]
  _Hapus multiple template_

• *.templatestats*
  _Statistik template_

• *.promotestats*
  _Statistik auto promote_

• *.activegroups*
  _Grup aktif auto promote_

• *.fetchproducts*
  _Ambil produk dari API_

• *.productstats*
  _Statistik produk API_

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📖 *QUICK START GUIDE*

1️⃣ Ketik: *.listgroups*
   _Lihat semua grup_

2️⃣ Ketik: *.enablegroup 1*
   _Aktifkan auto promote_

3️⃣ Ketik: *.testgroup 1*
   _Test kirim promosi_

4️⃣ Ketik: *.groupstatus 1*
   _Monitor status grup_

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️ *PENTING*
• Bot *TIDAK* merespon di grup
• Semua kontrol via *personal chat*
• Hanya *admin* yang bisa menggunakan
• Mulai dengan *.listgroups*

🚀 *Selamat Mempromosikan!*`
}

// HandlePromoteHelpCommand menangani command .promotehelp
func (h *PromoteCommandHandler) HandlePromoteHelpCommand(evt *events.Message) string {
	return `📋 *BANTUAN AUTO PROMOTE*

🤖 **Fitur Auto Promote:**
Sistem otomatis untuk mengirim promosi bisnis sesuai interval yang dikonfigurasi

🎯 **Commands Utama:**
• .promote - Aktifkan auto promote di grup
• .disablepromote - Nonaktifkan auto promote
• .statuspromo - Cek status auto promote
• .testpromo - Test kirim promosi manual

📝 **Commands Template:**
• .listtemplates - Lihat daftar template
• .previewtemplate [ID] - Preview template
• .addtemplate - Tambah template (admin only)
• .edittemplate [ID] - Edit template (admin only)
• .deletetemplate [ID] - Hapus template (admin only)

⚙️ **Commands Admin:**
• .templatestats - Statistik template
• .promotestats - Statistik auto promote
• .activegroups - Lihat grup aktif

💡 **Cara Kerja:**
1. Aktifkan dengan .promote di grup
2. Bot akan kirim promosi sesuai interval yang dikonfigurasi
3. Template dipilih random dari yang tersedia
4. Nonaktifkan kapan saja dengan .disablepromote

🎲 **Template System:**
• 10+ template promosi bisnis siap pakai
• Random selection untuk variasi
• Admin bisa tambah/edit template
• Support variables: {DATE}, {TIME}, dll

❓ **Butuh bantuan?**
Hubungi admin atau gunakan command di atas`
}

// IsPromoteCommand mengecek apakah pesan adalah command auto promote
func (h *PromoteCommandHandler) IsPromoteCommand(messageText string) bool {
	lowerText := strings.ToLower(strings.TrimSpace(messageText))

	promoteCommands := []string{
		// Group Management Commands
		".listgroups",
		".enablegroup",
		".enablemulti",
		".disablegroup",
		".groupstatus",
		".testgroup",
		// Template Commands
		".listtemplates",
		".alltemplates",
		".previewtemplate",
		// Admin Commands
		".addtemplate",
		".edittemplate",
		".deletetemplate",
		".templatestats",
		".promotestats",
		".activegroups",
		".fetchproducts",
		".productstats",
		".deleteall",
		".deletemulti",
		".help",
	}

	for _, cmd := range promoteCommands {
		if strings.HasPrefix(lowerText, cmd) {
			return true
		}
	}

	return false
}

// HandlePromoteCommands menangani semua command auto promote
func (h *PromoteCommandHandler) HandlePromoteCommands(evt *events.Message, messageText string) string {
	lowerText := strings.ToLower(strings.TrimSpace(messageText))
	args := strings.Fields(lowerText)

	if len(args) == 0 {
		return ""
	}

	command := args[0]

	switch command {
	case ".listtemplates":
		return h.HandleListTemplatesCommand(evt)

	case ".alltemplates":
		return h.HandleAllTemplatesCommand(evt)

	case ".previewtemplate":
		return h.HandlePreviewTemplateCommand(evt, args)

	case ".help":
		return h.HandleHelpCommand(evt)

	default:
		return ""
	}
}
