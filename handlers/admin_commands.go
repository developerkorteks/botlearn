// Package handlers - Admin command handlers untuk mengelola template dan sistem
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

// AdminCommandHandler menangani command admin untuk auto promote
type AdminCommandHandler struct {
	autoPromoteService  *services.AutoPromoteService
	templateService     *services.TemplateService
	apiProductService   *services.APIProductService
	groupManagerService *services.GroupManagerService
	logger              *utils.Logger
	adminNumbers        []string // Daftar nomor admin yang bisa menggunakan command admin
	client              *whatsmeow.Client
	typingSimulator     *utils.TypingSimulator
}

// NewAdminCommandHandler membuat handler baru
func NewAdminCommandHandler(
	autoPromoteService *services.AutoPromoteService,
	templateService *services.TemplateService,
	apiProductService *services.APIProductService,
	groupManagerService *services.GroupManagerService,
	logger *utils.Logger,
	adminNumbers []string,
	client *whatsmeow.Client,
) *AdminCommandHandler {
	return &AdminCommandHandler{
		autoPromoteService:  autoPromoteService,
		templateService:     templateService,
		apiProductService:   apiProductService,
		groupManagerService: groupManagerService,
		logger:              logger,
		adminNumbers:        adminNumbers,
		client:              client,
		typingSimulator:     utils.NewTypingSimulator(client, logger),
	}
}

// sendMessageWithTyping mengirim pesan dengan simulasi typing yang natural
func (h *AdminCommandHandler) sendMessageWithTyping(chatJID types.JID, message string) error {
	// Simulasi typing berdasarkan kompleksitas pesan
	h.typingSimulator.SmartTypingDelay(chatJID, message)
	
	// Kirim pesan setelah simulasi typing selesai
	_, err := h.client.SendMessage(context.Background(), chatJID, &waProto.Message{
		Conversation: &message,
	})
	
	if err != nil {
		h.logger.Errorf("Gagal kirim pesan admin ke %s: %v", chatJID, err)
		return err
	}
	
	h.logger.Debugf("✅ Pesan admin terkirim dengan typing delay ke %s", chatJID)
	return nil
}

// isAdmin mengecek apakah user adalah admin dengan validasi ketat
func (h *AdminCommandHandler) isAdmin(userNumber string) bool {
	// Validasi input
	if userNumber == "" {
		h.logger.Warning("Empty user number provided for admin check")
		return false
	}

	// Log attempt untuk security monitoring
	h.logger.Debugf("Admin check for user: %s", userNumber)

	// Cek apakah user ada dalam daftar admin
	for _, admin := range h.adminNumbers {
		if admin == userNumber {
			h.logger.Infof("Admin access granted for: %s", userNumber)
			return true
		}
	}

	// Log unauthorized attempt
	h.logger.Warningf("Unauthorized admin attempt from: %s", userNumber)
	return false
}

// IsUserAdmin adalah method public untuk mengecek admin dari luar
func (h *AdminCommandHandler) IsUserAdmin(userNumber string) bool {
	return h.isAdmin(userNumber)
}

// HandleAddTemplateCommand menangani command .addtemplate
func (h *AdminCommandHandler) HandleAddTemplateCommand(evt *events.Message, args []string) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return `❌ *AKSES DITOLAK*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TIDAK ADA IZIN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Command ini hanya bisa digunakan oleh admin

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Hanya admin yang memiliki akses
• Hubungi admin untuk bantuan
• Gunakan /help untuk command umum

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔒 *Akses terbatas untuk keamanan sistem*`
	}

	// Format: .addtemplate "Judul" "Kategori" "Konten"
	if len(args) < 4 {
		return `❌ *FORMAT SALAH*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *CARA PENGGUNAAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *FORMAT COMMAND*
*.addtemplate* "Judul" "Kategori" "Konten"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 *CONTOH PENGGUNAAN*
*.addtemplate* "Flash Sale Hari Ini" "flashsale" "🔥 FLASH SALE! Diskon 50% hanya hari ini! Order: 08123456789"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *TIPS PENTING*
• Gunakan tanda kutip untuk teks spasi
• Kategori: produk, diskon, testimoni, flashsale
• Konten bisa pakai emoji dan formatting WhatsApp`
	}

	// Parse arguments (simplified parsing)
	fullText := strings.Join(args[1:], " ")
	parts := h.parseQuotedArgs(fullText)

	if len(parts) < 3 {
		return `❌ *FORMAT SALAH*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *PARAMETER TIDAK VALID*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Format salah. Gunakan: .addtemplate "Judul" "Kategori" "Konten"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *CONTOH PENGGUNAAN*
• .addtemplate "Flash Sale" "diskon" "🔥 FLASH SALE! Diskon 50%!"
• Gunakan tanda kutip untuk teks dengan spasi
• Kategori: produk, diskon, testimoni, flashsale

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *Coba lagi dengan format yang benar*`
	}

	title := parts[0]
	category := parts[1]
	content := parts[2]

	// Buat template
	template, err := h.templateService.CreateTemplate(title, content, category)
	if err != nil {
		h.logger.Errorf("Failed to create template: %v", err)
		return fmt.Sprintf(`❌ *GAGAL MEMBUAT TEMPLATE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal membuat template: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Template dengan judul sama sudah ada
• Konten template terlalu panjang
• Kategori tidak valid
• Masalah koneksi database

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi admin*`, err.Error())
	}

	return fmt.Sprintf(`✅ *TEMPLATE BERHASIL DIBUAT!*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *BERHASIL TERSIMPAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 *DETAIL TEMPLATE*
🆔 *ID:* %d
🏷️ *Judul:* %s
📂 *Kategori:* %s
✅ *Status:* Aktif

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *KONTEN TEMPLATE*
%s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *COMMANDS SELANJUTNYA*
• *.previewtemplate %d* - Preview template
• *.edittemplate %d* - Edit template
• *.listtemplates* - Lihat semua template

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎉 *Template siap digunakan!*`,
		template.ID, template.Title, template.Category, template.Content, template.ID, template.ID)
}

// HandleEditTemplateCommand menangani command .edittemplate
func (h *AdminCommandHandler) HandleEditTemplateCommand(evt *events.Message, args []string) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return `❌ *AKSES DITOLAK*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TIDAK ADA IZIN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Command ini hanya bisa digunakan oleh admin

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Hanya admin yang memiliki akses
• Hubungi admin untuk bantuan
• Gunakan /help untuk command umum

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔒 *Akses terbatas untuk keamanan sistem*`
	}

	// Format: .edittemplate [ID] "Judul" "Kategori" "Konten"
	if len(args) < 5 {
		return `❌ *FORMAT SALAH*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *CARA PENGGUNAAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *FORMAT COMMAND*
*.edittemplate* [ID] "Judul" "Kategori" "Konten"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 *CONTOH PENGGUNAAN*
*.edittemplate* 1 "Promo Terbaru" "diskon" "🎉 Promo spesial! Diskon 30%"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *TIPS PENTING*
• Gunakan .listtemplates untuk melihat ID template
• Gunakan tanda kutip untuk teks dengan spasi
• ID harus berupa angka yang valid`
	}

	// Parse ID
	templateID, err := strconv.Atoi(args[1])
	if err != nil {
		return `❌ *ID TIDAK VALID*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *FORMAT ID SALAH*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 ID template harus berupa angka

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *CONTOH YANG BENAR*
• .edittemplate 1 "Judul" "Kategori" "Konten"
• .deletetemplate 5
• .previewtemplate 3

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *Gunakan .listtemplates untuk melihat ID*`
	}

	// Parse arguments
	fullText := strings.Join(args[2:], " ")
	parts := h.parseQuotedArgs(fullText)

	if len(parts) < 3 {
		return `❌ *FORMAT SALAH*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *PARAMETER TIDAK VALID*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Format salah. Gunakan: .edittemplate [ID] "Judul" "Kategori" "Konten"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *CONTOH PENGGUNAAN*
• .edittemplate 1 "Promo Terbaru" "diskon" "🎉 Promo spesial! Diskon 30%"
• Gunakan tanda kutip untuk teks dengan spasi
• ID harus berupa angka

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *Coba lagi dengan format yang benar*`
	}

	title := parts[0]
	category := parts[1]
	content := parts[2]

	// Update template
	err = h.templateService.UpdateTemplate(templateID, title, content, category, true)
	if err != nil {
		h.logger.Errorf("Failed to update template %d: %v", templateID, err)
		return fmt.Sprintf(`❌ *GAGAL MENGUPDATE TEMPLATE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mengupdate template: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Template dengan ID tersebut tidak ditemukan
• Judul template sudah digunakan
• Konten template terlalu panjang
• Masalah koneksi database

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi admin*`, err.Error())
	}

	return fmt.Sprintf(`✅ *TEMPLATE BERHASIL DIUPDATE!*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *BERHASIL DIPERBARUI*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 *DETAIL TEMPLATE*
🆔 *ID:* %d
🏷️ *Judul:* %s
📂 *Kategori:* %s
✅ *Status:* Aktif

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *KONTEN BARU*
%s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *COMMANDS SELANJUTNYA*
• *.previewtemplate %d* - Preview template
• *.listtemplates* - Lihat semua template
• *.deletetemplate %d* - Hapus template

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎉 *Perubahan langsung berlaku untuk auto promote!*`,
		templateID, title, category, content, templateID, templateID)
}

// HandleDeleteTemplateCommand menangani command .deletetemplate
func (h *AdminCommandHandler) HandleDeleteTemplateCommand(evt *events.Message, args []string) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return `❌ *AKSES DITOLAK*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TIDAK ADA IZIN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Command ini hanya bisa digunakan oleh admin

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Hanya admin yang memiliki akses
• Hubungi admin untuk bantuan
• Gunakan /help untuk command umum

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔒 *Akses terbatas untuk keamanan sistem*`
	}

	if len(args) < 2 {
		return `❌ *FORMAT SALAH*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *CARA PENGGUNAAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *FORMAT COMMAND*
*.deletetemplate* [ID]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 *CONTOH PENGGUNAAN*
*.deletetemplate* 5

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *TIPS PENTING*
• Gunakan .listtemplates untuk melihat ID template
• ID harus berupa angka yang valid
• Template yang dihapus tidak bisa dikembalikan`
	}

	// Parse ID
	templateID, err := strconv.Atoi(args[1])
	if err != nil {
		return `❌ *ID TIDAK VALID*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *FORMAT ID SALAH*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 ID template harus berupa angka

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *CONTOH YANG BENAR*
• .edittemplate 1 "Judul" "Kategori" "Konten"
• .deletetemplate 5
• .previewtemplate 3

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *Gunakan .listtemplates untuk melihat ID*`
	}

	// Ambil info template sebelum dihapus
	template, err := h.templateService.GetTemplateByID(templateID)
	if err != nil {
		return fmt.Sprintf(`❌ *GAGAL MENDAPATKAN TEMPLATE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mendapatkan template: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Template dengan ID tersebut tidak ada
• Masalah koneksi database
• ID template tidak valid

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Gunakan .listtemplates untuk melihat ID yang valid*`, err.Error())
	}

	if template == nil {
		return fmt.Sprintf(`❌ *TEMPLATE TIDAK DITEMUKAN*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *ID TIDAK VALID*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Template dengan ID %d tidak ditemukan

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Template sudah dihapus sebelumnya
• ID salah atau tidak ada
• Template tidak pernah dibuat

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *COMMANDS SELANJUTNYA*
• .listtemplates - Lihat template yang ada
• .alltemplates - Lihat semua template
• .addtemplate - Buat template baru

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔍 *Periksa ID template yang valid*`, templateID)
	}

	// Hapus template
	err = h.templateService.DeleteTemplate(templateID)
	if err != nil {
		h.logger.Errorf("Failed to delete template %d: %v", templateID, err)
		return fmt.Sprintf(`❌ *GAGAL MENGHAPUS TEMPLATE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal menghapus template: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Template sedang digunakan oleh sistem
• Masalah koneksi database
• Template sudah dihapus sebelumnya

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi admin*`, err.Error())
	}

	return fmt.Sprintf(`🗑️ *TEMPLATE BERHASIL DIHAPUS!*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *BERHASIL DIHAPUS*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 *DETAIL TEMPLATE YANG DIHAPUS*
🆔 *ID:* %d
🏷️ *Judul:* %s
📂 *Kategori:* %s
🗑️ *Status:* Dihapus

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️ *PERINGATAN*
• Template telah dihapus permanen
• Tidak bisa dikembalikan lagi
• Auto promote akan menggunakan template lain

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎮 *COMMANDS SELANJUTNYA*
• *.listtemplates* - Lihat template tersisa
• *.addtemplate* - Tambah template baru
• *.templatestats* - Statistik template

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ *Template berhasil dihapus!*`,
		templateID, template.Title, template.Category)
}

// HandleTemplateStatsCommand menangani command .templatestats
func (h *AdminCommandHandler) HandleTemplateStatsCommand(evt *events.Message) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return `❌ *AKSES DITOLAK*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TIDAK ADA IZIN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Command ini hanya bisa digunakan oleh admin

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Hanya admin yang memiliki akses
• Hubungi admin untuk bantuan
• Gunakan /help untuk command umum

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔒 *Akses terbatas untuk keamanan sistem*`
	}

	stats, err := h.templateService.GetTemplateStats()
	if err != nil {
		h.logger.Errorf("Failed to get template stats: %v", err)
		return `❌ *GAGAL MENDAPATKAN STATISTIK*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mendapatkan statistik template

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Masalah koneksi database
• Service template tidak tersedia
• Error internal sistem

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi dalam beberapa saat*`
	}

	var result strings.Builder
	result.WriteString("📊 *STATISTIK TEMPLATE*\n\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("           *RINGKASAN TEMPLATE*\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	result.WriteString(fmt.Sprintf("📝 *Total Template:* %d\n", stats["total"]))
	result.WriteString(fmt.Sprintf("✅ *Template Aktif:* %d\n", stats["active"]))
	result.WriteString(fmt.Sprintf("❌ *Template Tidak Aktif:* %d\n", stats["inactive"]))

	result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("           *DETAIL PER KATEGORI*\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	categories := stats["categories"].(map[string]int)
	if len(categories) > 0 {
		for category, count := range categories {
			result.WriteString(fmt.Sprintf("• *%s:* %d template\n", strings.Title(category), count))
		}
	} else {
		result.WriteString("Tidak ada kategori yang ditemukan.\n")
	}

	result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("           *COMMANDS TERKAIT*\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString("• *.listtemplates* - Lihat template aktif\n")
	result.WriteString("• *.alltemplates* - Lihat semua template\n")
	result.WriteString("• *.addtemplate* - Tambah template baru")

	return result.String()
}

// HandlePromoteStatsCommand menangani command .promotestats
func (h *AdminCommandHandler) HandlePromoteStatsCommand(evt *events.Message) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return `❌ *AKSES DITOLAK*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TIDAK ADA IZIN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Command ini hanya bisa digunakan oleh admin

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Hanya admin yang memiliki akses
• Hubungi admin untuk bantuan
• Gunakan /help untuk command umum

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔒 *Akses terbatas untuk keamanan sistem*`
	}

	// Ambil jumlah grup aktif
	activeCount, err := h.autoPromoteService.GetActiveGroupsCount()
	if err != nil {
		h.logger.Errorf("Failed to get active groups count: %v", err)
		return `❌ *GAGAL MENDAPATKAN STATISTIK*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mendapatkan statistik auto promote

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Service auto promote tidak tersedia
• Masalah koneksi database
• Error internal sistem

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi dalam beberapa saat*`
	}

	return fmt.Sprintf(`📊 *STATISTIK AUTO PROMOTE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
          *OVERVIEW SISTEM*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *STATUS UTAMA*
🔥 *Grup Aktif:* %d grup
⏰ *Interval:* Sesuai konfigurasi
🤖 *Scheduler:* Berjalan
📊 *Mode:* Auto Promote

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📈 *BREAKDOWN GRUP*
📋 *Total Terdaftar:* %d grup
✅ *Aktif:* %d grup
❌ *Tidak Aktif:* %d grup
📊 *Tingkat Aktivasi:* %.1f%%

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎮 *COMMANDS MONITORING*
• *.activegroups* - Detail grup aktif
• *.listgroups* - Semua grup
• *.groupstatus [ID]* - Status spesifik

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *Statistik diperbarui real-time*`,
		activeCount, activeCount, activeCount, 0,
		func() float64 {
			if activeCount > 0 {
				return float64(activeCount) / float64(activeCount) * 100
			}
			return 0.0
		}())
}

// HandleActiveGroupsCommand menangani command .activegroups
func (h *AdminCommandHandler) HandleActiveGroupsCommand(evt *events.Message) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return `❌ *AKSES DITOLAK*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TIDAK ADA IZIN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Command ini hanya bisa digunakan oleh admin

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Hanya admin yang memiliki akses
• Hubungi admin untuk bantuan
• Gunakan /help untuk command umum

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔒 *Akses terbatas untuk keamanan sistem*`
	}

	// Ambil daftar grup aktif dari service
	activeGroups, err := h.autoPromoteService.GetActiveGroups()
	if err != nil {
		h.logger.Errorf("Failed to get active groups: %v", err)
		return `❌ *GAGAL MENDAPATKAN DAFTAR GRUP*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mendapatkan daftar grup aktif

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Service auto promote tidak tersedia
• Masalah koneksi database
• Tidak ada grup yang terdaftar

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau gunakan .listgroups*`
	}

	if len(activeGroups) == 0 {
		return `👥 *GRUP AKTIF AUTO PROMOTE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
            *TIDAK ADA GRUP AKTIF*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

❌ Tidak ada grup yang menggunakan auto promote

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Gunakan *.listgroups* untuk melihat semua grup
• Gunakan *.enablegroup [ID]* untuk mengaktifkan
• Auto promote akan muncul di sini setelah aktif

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *Mulai dengan .listgroups untuk setup*`
	}

	var result strings.Builder
	result.WriteString("👥 *GRUP AKTIF AUTO PROMOTE*\n\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString(fmt.Sprintf("        *TOTAL: %d GRUP AKTIF*\n", len(activeGroups)))
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, group := range activeGroups {
		if i >= 20 { // Batasi tampilan maksimal 20 grup
			result.WriteString(fmt.Sprintf("... dan %d grup lainnya\n", len(activeGroups)-20))
			break
		}

		// Format group JID untuk tampilan
		groupDisplay := h.formatGroupJID(group.GroupJID)

		result.WriteString(fmt.Sprintf("*%d.* 👥 %s\n", i+1, groupDisplay))
		result.WriteString(fmt.Sprintf("   *Status:* ✅ Aktif\n"))

		if group.StartedAt != nil {
			result.WriteString(fmt.Sprintf("   *Dimulai:* %s\n", group.StartedAt.Format("02 Jan 2006, 15:04")))
		}

		if group.LastPromoteAt != nil {
			result.WriteString(fmt.Sprintf("   *Terakhir Promote:* %s\n", group.LastPromoteAt.Format("02 Jan 2006, 15:04")))
		} else {
			result.WriteString("   *Terakhir Promote:* Belum pernah\n")
		}

		if i < len(activeGroups)-1 && i < 19 {
			result.WriteString("\n")
		}
	}

	result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("           *COMMANDS TERKAIT*\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString("• *.promotestats* - Statistik umum\n")
	result.WriteString("• *.disablegroup [ID]* - Nonaktifkan grup")

	return result.String()
}

// formatGroupJID memformat group JID untuk tampilan yang lebih readable
func (h *AdminCommandHandler) formatGroupJID(groupJID string) string {
	// Ambil hanya bagian ID grup (sebelum @g.us)
	if strings.Contains(groupJID, "@g.us") {
		parts := strings.Split(groupJID, "@")
		if len(parts) > 0 {
			return fmt.Sprintf("Grup-%s", parts[0][len(parts[0])-8:]) // 8 digit terakhir
		}
	}
	return groupJID
}

// HandleFetchProductsCommand menangani command .fetchproducts
func (h *AdminCommandHandler) HandleFetchProductsCommand(evt *events.Message) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return `❌ *AKSES DITOLAK*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TIDAK ADA IZIN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Command ini hanya bisa digunakan oleh admin

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Hanya admin yang memiliki akses
• Hubungi admin untuk bantuan
• Gunakan /help untuk command umum

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔒 *Akses terbatas untuk keamanan sistem*`
	}

	if h.apiProductService == nil {
		return `❌ *SERVICE TIDAK TERSEDIA*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		         *KESALAHAN SISTEM*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Service untuk produk API tidak dikonfigurasi

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Konfigurasi API belum diatur
• Service tidak diinisialisasi saat start-up
• Terjadi error internal

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Hubungi developer untuk perbaikan*`
	}

	h.logger.Info("Admin requesting product fetch from API...")

	result, err := h.apiProductService.FetchProductsAndCreateTemplates()
	if err != nil {
		h.logger.Errorf("Failed to fetch products: %v", err)
		return fmt.Sprintf(`❌ *GAGAL MENGAMBIL PRODUK*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		         *KESALAHAN API*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mengambil produk dari API: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Koneksi ke server API gagal
• URL API salah atau tidak valid
• API Key tidak valid atau kadaluwarsa
• Server API sedang down

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Periksa koneksi dan konfigurasi API*`, err.Error())
	}

	return result
}

// HandleProductStatsCommand menangani command .productstats
func (h *AdminCommandHandler) HandleProductStatsCommand(evt *events.Message) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return `❌ *AKSES DITOLAK*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TIDAK ADA IZIN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Command ini hanya bisa digunakan oleh admin

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Hanya admin yang memiliki akses
• Hubungi admin untuk bantuan
• Gunakan /help untuk command umum

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔒 *Akses terbatas untuk keamanan sistem*`
	}

	if h.apiProductService == nil {
		return `❌ *SERVICE TIDAK TERSEDIA*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		         *KESALAHAN SISTEM*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Service untuk produk API tidak dikonfigurasi

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Konfigurasi API belum diatur
• Service tidak diinisialisasi saat start-up
• Terjadi error internal

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Hubungi developer untuk perbaikan*`
	}

	result, err := h.apiProductService.GetProductStats()
	if err != nil {
		h.logger.Errorf("Failed to get product stats: %v", err)
		return fmt.Sprintf(`❌ *GAGAL MENDAPATKAN STATISTIK*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		         *KESALAHAN API*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mendapatkan statistik produk: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Koneksi ke server API gagal
• Database tidak dapat diakses
• Terjadi error internal

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi developer*`, err.Error())
	}

	return result
}

// HandleDeleteAllTemplatesCommand menangani command .deleteall
func (h *AdminCommandHandler) HandleDeleteAllTemplatesCommand(evt *events.Message) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return `❌ *AKSES DITOLAK*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TIDAK ADA IZIN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Command ini hanya bisa digunakan oleh admin

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Hanya admin yang memiliki akses
• Hubungi admin untuk bantuan
• Gunakan /help untuk command umum

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔒 *Akses terbatas untuk keamanan sistem*`
	}

	// Ambil semua template
	templates, err := h.templateService.GetAllTemplates()
	if err != nil {
		return fmt.Sprintf(`❌ *GAGAL MENDAPATKAN TEMPLATE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *KESALAHAN DATABASE*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mendapatkan daftar template: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Koneksi database terputus
• Tabel template tidak ditemukan
• Terjadi error internal

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi developer*`, err.Error())
	}

	if len(templates) == 0 {
		return `ℹ️ *TIDAK ADA TEMPLATE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *DATABASE KOSONG*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Tidak ada template yang perlu dihapus.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Database template sudah bersih.
• Gunakan *.fetchproducts* untuk mengisi ulang.
• Gunakan *.addtemplate* untuk menambah manual.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎉 *Semua bersih!*`
	}

	// Hapus semua template
	deletedCount := 0
	var errors []string

	for _, template := range templates {
		err := h.templateService.DeleteTemplate(template.ID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("ID %d: %v", template.ID, err))
		} else {
			deletedCount++
		}
	}

	var result strings.Builder
	result.WriteString("🗑️ *HASIL HAPUS SEMUA TEMPLATE*\n\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("           *RINGKASAN OPERASI*\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString(fmt.Sprintf("✅ *Berhasil Dihapus:* %d template\n", deletedCount))

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("❌ *Gagal Dihapus:* %d template\n", len(errors)))
	}

	result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("           *PERINGATAN PENTING*\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString("• Tindakan ini *tidak dapat* dibatalkan.\n")
	result.WriteString("• Semua template telah dihapus permanen.\n")
	result.WriteString("• Auto promote mungkin berhenti jika kehabisan template.\n")

	result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("           *LANGKAH SELANJUTNYA*\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString("• Gunakan *.fetchproducts* untuk isi ulang dari API.\n")
	result.WriteString("• Gunakan *.addtemplate* untuk menambah manual.")

	return result.String()
}

// HandleDeleteMultipleTemplatesCommand menangani command .deletemulti [ID1,ID2,ID3]
func (h *AdminCommandHandler) HandleDeleteMultipleTemplatesCommand(evt *events.Message, args []string) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return `❌ *AKSES DITOLAK*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TIDAK ADA IZIN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Command ini hanya bisa digunakan oleh admin

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Hanya admin yang memiliki akses
• Hubungi admin untuk bantuan
• Gunakan /help untuk command umum

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔒 *Akses terbatas untuk keamanan sistem*`
	}

	if len(args) < 2 {
		return `❌ *FORMAT SALAH*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *CARA PENGGUNAAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *FORMAT COMMAND*
*.deletemulti* [ID1,ID2,ID3]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 *CONTOH PENGGUNAAN*
*.deletemulti* 1,5,8,12

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *TIPS PENTING*
• Pisahkan ID dengan koma tanpa spasi
• Gunakan .alltemplates untuk melihat ID
• Maksimal 20 ID sekaligus
• Template yang dihapus tidak bisa dikembalikan`
	}

	// Parse ID dari argument
	idsStr := strings.Join(args[1:], "")
	idStrings := strings.Split(idsStr, ",")

	if len(idStrings) > 20 {
		return `❌ *TERLALU BANYAK ID*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *BATAS MAKSIMAL*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Maksimal 20 template sekaligus

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *SOLUSI*
• Bagi menjadi beberapa command
• Contoh: .deletemulti 1,2,3,4,5
• Lalu: .deletemulti 6,7,8,9,10

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi dengan ID yang lebih sedikit*`
	}

	var ids []int
	for _, idStr := range idStrings {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			return fmt.Sprintf(`❌ *ID TIDAK VALID*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *FORMAT ID SALAH*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 ID tidak valid: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *TIPS PERBAIKAN*
• Semua ID harus berupa angka
• Pisahkan dengan koma tanpa spasi
• Contoh yang benar: 1,5,8,12

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi dengan format yang benar*`, idStr)
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return `❌ *TIDAK ADA ID VALID*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TIDAK ADA DATA*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Tidak ada ID yang valid ditemukan

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Format ID salah (harus angka)
• ID kosong atau hanya koma
• Spasi berlebihan dalam input

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Contoh yang benar: .deletemulti 1,5,8*`
	}

	// Hapus template berdasarkan ID
	deletedCount := 0
	var errors []string
	var deletedTitles []string

	for _, id := range ids {
		// Ambil info template sebelum dihapus
		template, err := h.templateService.GetTemplateByID(id)
		if err != nil {
			errors = append(errors, fmt.Sprintf("ID %d: tidak ditemukan", id))
			continue
		}

		err = h.templateService.DeleteTemplate(id)
		if err != nil {
			errors = append(errors, fmt.Sprintf("ID %d: %v", id, err))
		} else {
			deletedCount++
			deletedTitles = append(deletedTitles, fmt.Sprintf("ID %d: %s", id, template.Title))
		}
	}

	var result strings.Builder
	result.WriteString("🗑️ *HASIL HAPUS MULTIPLE TEMPLATE*\n\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("           *RINGKASAN OPERASI*\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString(fmt.Sprintf("✅ *Berhasil Dihapus:* %d template\n", deletedCount))

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("❌ *Gagal Dihapus:* %d template\n", len(errors)))
	}

	if len(deletedTitles) > 0 {
		result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		result.WriteString("           *DETAIL YANG DIHAPUS*\n")
		result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
		for i, title := range deletedTitles {
			if i < 10 { // Batasi tampilan
				result.WriteString(fmt.Sprintf("• %s\n", title))
			} else {
				result.WriteString(fmt.Sprintf("... dan %d lainnya.", len(deletedTitles)-10))
				break
			}
		}
	}

	if len(errors) > 0 {
		result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		result.WriteString("           *DETAIL KEGAGALAN*\n")
		result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
		for i, e := range errors {
			if i < 5 { // Batasi tampilan
				result.WriteString(fmt.Sprintf("• %s\n", e))
			} else {
				result.WriteString(fmt.Sprintf("... dan %d lainnya.", len(errors)-5))
				break
			}
		}
	}

	result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString("💡 Gunakan *.listtemplates* untuk melihat sisa template.")

	return result.String()
}

// === GROUP MANAGEMENT COMMANDS ===

// HandleListGroupsCommand menangani command .listgroups
func (h *AdminCommandHandler) HandleListGroupsCommand(evt *events.Message) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return "" // Tidak ada response untuk non-admin
	}

	if h.groupManagerService == nil {
		return `❌ *SERVICE TIDAK TERSEDIA*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		         *KESALAHAN SISTEM*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Service untuk manajemen grup tidak dikonfigurasi

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Service tidak diinisialisasi saat start-up
• Terjadi error internal

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Hubungi developer untuk perbaikan*`
	}

	h.logger.Info("Admin requesting list of joined groups...")

	groups, err := h.groupManagerService.GetAllJoinedGroups()
	if err != nil {
		h.logger.Errorf("Failed to get joined groups: %v", err)
		return fmt.Sprintf(`❌ *GAGAL MENDAPATKAN GRUP*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		         *KESALAHAN DATABASE*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mendapatkan daftar grup: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Koneksi database terputus
• Tabel grup tidak ditemukan
• Terjadi error internal

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi developer*`, err.Error())
	}

	if len(groups) == 0 {
		return `👥 *DAFTAR GRUP YANG DIIKUTI*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
            *TIDAK ADA GRUP*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

❌ Bot belum bergabung dengan grup manapun

*LANGKAH SELANJUTNYA:*
1️⃣ Tambahkan bot ke grup
2️⃣ Ketik *.listgroups* lagi
3️⃣ Gunakan *.enablegroup [ID]*`
	}

	var result strings.Builder
	result.WriteString("👥 *DAFTAR GRUP YANG DIIKUTI*\n\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString(fmt.Sprintf("        *TOTAL: %d GRUP*\n", len(groups)))
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, group := range groups {
		statusIcon := "🔴"
		statusText := "*TIDAK AKTIF*"
		if group.IsActive {
			statusIcon = "🟢"
			statusText = "*AKTIF*"
		}

		result.WriteString(fmt.Sprintf("%s *ID: %d* - %s\n", statusIcon, group.ID, group.Name))
		result.WriteString(fmt.Sprintf("👥 Member: *%d orang*\n", group.MemberCount))
		result.WriteString(fmt.Sprintf("🤖 Status: %s\n", statusText))

		if group.Description != "" && len(group.Description) > 0 {
			desc := group.Description
			if len(desc) > 50 {
				desc = desc[:50] + "..."
			}
			result.WriteString(fmt.Sprintf("📝 %s\n", desc))
		}

		if i < len(groups)-1 {
			result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
		}
	}

	result.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("            *COMMANDS*\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString("• *.enablegroup [ID]*\n")
	result.WriteString("  _Aktifkan auto promote_\n\n")
	result.WriteString("• *.disablegroup [ID]*\n")
	result.WriteString("  _Nonaktifkan auto promote_\n\n")
	result.WriteString("• *.groupstatus [ID]*\n")
	result.WriteString("  _Status detail grup_\n\n")
	result.WriteString("• *.testgroup [ID]*\n")
	result.WriteString("  _Test kirim promosi_\n\n")
	result.WriteString("💡 *Contoh:* .enablegroup 3 atau .testgroup 5")

	return result.String()
}

// Helper function untuk max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// HandleEnableGroupCommand menangani command .enablegroup [ID]
func (h *AdminCommandHandler) HandleEnableGroupCommand(evt *events.Message, args []string) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return "" // Tidak ada response untuk non-admin
	}

	if len(args) < 2 {
		return `❌ *FORMAT SALAH*

📝 **Format:** .enablegroup [ID]
📋 **Contoh:** .enablegroup 3

💡 Gunakan .listgroups untuk melihat ID grup`
	}

	if h.groupManagerService == nil {
		return `❌ *SERVICE TIDAK TERSEDIA*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *KESALAHAN SISTEM*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Service untuk manajemen grup tidak dikonfigurasi

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Service tidak diinisialisasi saat start-up
• Terjadi error internal

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Hubungi developer untuk perbaikan*`
	}

	// Parse ID grup
	groupID, err := strconv.Atoi(args[1])
	if err != nil {
		return `❌ *ID TIDAK VALID*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *FORMAT ID SALAH*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 ID grup harus berupa angka.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *CONTOH PENGGUNAAN*
• .enablegroup 3
• .disablegroup 5

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *Gunakan .listgroups untuk melihat ID*`
	}

	// Aktifkan auto promote
	err = h.groupManagerService.EnableAutoPromoteForGroup(groupID)
	if err != nil {
		h.logger.Errorf("Failed to enable auto promote for group %d: %v", groupID, err)
		return fmt.Sprintf(`❌ *GAGAL MENGAKTIFKAN PROMOTE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mengaktifkan auto promote: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Grup dengan ID tersebut tidak ditemukan
• Masalah koneksi database
• Grup sudah aktif

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi admin*`, err.Error())
	}

	// Ambil info grup untuk response
	groupInfo, err := h.groupManagerService.GetGroupByID(groupID)
	if err != nil {
		return `✅ *AUTO PROMOTE BERHASIL DIAKTIFKAN!*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *STATUS TELAH DIUBAH*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 Auto promote berhasil diaktifkan.
⚠️ Namun, info grup tidak dapat diambil saat ini.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *INFORMASI*
• Gunakan *.groupstatus* untuk cek detail.
• Auto promote sudah berjalan.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚀 *Sistem siap bekerja!*`
	}

	return fmt.Sprintf(`✅ *AUTO PROMOTE DIAKTIFKAN!*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *BERHASIL AKTIF*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *DETAIL GRUP*
👥 *Nama:* %s
🆔 *ID:* %d
👤 *Member:* %d orang
⏰ *Mulai:* Sekarang
🤖 *Status:* AKTIF

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 *INFORMASI SISTEM*
• Bot akan kirim promosi otomatis
• Template dipilih secara random
• Interval sesuai konfigurasi
• Monitoring real-time tersedia

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎮 *COMMANDS SELANJUTNYA*

• *.groupstatus %d*
  _Monitor status grup_

• *.testgroup %d*
  _Test kirim promosi_

• *.disablegroup %d*
  _Nonaktifkan jika perlu_

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚀 *Auto promote siap bekerja!*`,
		groupInfo.Name, groupInfo.ID, groupInfo.MemberCount, groupID, groupID, groupID)
}

// HandleEnableMultipleGroupsCommand menangani command .enablemulti [ID1,ID2,...]
func (h *AdminCommandHandler) HandleEnableMultipleGroupsCommand(evt *events.Message, args []string) string {
	if !h.isAdmin(evt.Info.Sender.User) {
		return `❌ *AKSES DITOLAK*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *TIDAK ADA IZIN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🚫 Command ini hanya bisa digunakan oleh admin.`
	}

	if len(args) < 2 {
		return `❌ *FORMAT SALAH*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *CARA PENGGUNAAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📝 *Format:* .enablemulti [ID1,ID2,ID3]
📋 *Contoh:* .enablemulti 1,5,8
💡 Gunakan .listgroups untuk melihat ID grup.`
	}

	if h.groupManagerService == nil {
		return `❌ *SERVICE TIDAK TERSEDIA*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *KESALAHAN SISTEM*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🚫 Service untuk manajemen grup tidak dikonfigurasi.`
	}

	idsStr := strings.Join(args[1:], "")
	idStrings := strings.Split(idsStr, ",")

	var successCount, failCount int
	var successDetails, failDetails []string

	for _, idStr := range idStrings {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			failCount++
			failDetails = append(failDetails, fmt.Sprintf("ID '%s': bukan angka", idStr))
			continue
		}

		err = h.groupManagerService.EnableAutoPromoteForGroup(id)
		if err != nil {
			failCount++
			failDetails = append(failDetails, fmt.Sprintf("ID %d: %v", id, err))
		} else {
			successCount++
			successDetails = append(successDetails, fmt.Sprintf("ID %d", id))
		}
	}

	var result strings.Builder
	result.WriteString("🚀 *HASIL AKTIVASI MULTIPLE GRUP*\n\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("           *RINGKASAN OPERASI*\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString(fmt.Sprintf("✅ *Berhasil Diaktifkan:* %d grup\n", successCount))
	result.WriteString(fmt.Sprintf("❌ *Gagal Diaktifkan:* %d grup\n", failCount))

	if len(successDetails) > 0 {
		result.WriteString("\n*Grup yang berhasil diaktifkan:*\n")
		result.WriteString(strings.Join(successDetails, ", "))
	}

	if len(failDetails) > 0 {
		result.WriteString("\n\n*Detail Kegagalan:*\n")
		for i, detail := range failDetails {
			if i < 5 { // Batasi 5 error
				result.WriteString(fmt.Sprintf("• %s\n", detail))
			} else {
				result.WriteString(fmt.Sprintf("... dan %d error lainnya.", len(failDetails)-5))
				break
			}
		}
	}

	result.WriteString("\n\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("💡 Gunakan *.activegroups* untuk melihat semua grup yang aktif.")

	return result.String()
}

// HandleDisableGroupCommand menangani command .disablegroup [ID]
func (h *AdminCommandHandler) HandleDisableGroupCommand(evt *events.Message, args []string) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return "" // Tidak ada response untuk non-admin
	}

	if len(args) < 2 {
		return `❌ *FORMAT SALAH*

📝 **Format:** .disablegroup [ID]
📋 **Contoh:** .disablegroup 3

💡 Gunakan .listgroups untuk melihat ID grup`
	}

	if h.groupManagerService == nil {
		return `❌ *SERVICE TIDAK TERSEDIA*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *KESALAHAN SISTEM*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Service untuk manajemen grup tidak dikonfigurasi

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Service tidak diinisialisasi saat start-up
• Terjadi error internal

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Hubungi developer untuk perbaikan*`
	}

	// Parse ID grup
	groupID, err := strconv.Atoi(args[1])
	if err != nil {
		return `❌ *ID TIDAK VALID*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *FORMAT ID SALAH*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 ID grup harus berupa angka.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *CONTOH PENGGUNAAN*
• .disablegroup 3
• .groupstatus 5

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *Gunakan .listgroups untuk melihat ID*`
	}

	// Ambil info grup sebelum dinonaktifkan
	groupInfo, err := h.groupManagerService.GetGroupByID(groupID)
	if err != nil {
		return fmt.Sprintf(`❌ *GRUP TIDAK DITEMUKAN*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *ID TIDAK VALID*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Grup dengan ID %d tidak ditemukan di database.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• ID grup salah atau tidak ada.
• Bot belum join grup tersebut.
• Grup sudah dihapus.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *Gunakan .listgroups untuk melihat ID yang valid*`, groupID)
	}

	// Nonaktifkan auto promote
	err = h.groupManagerService.DisableAutoPromoteForGroup(groupID)
	if err != nil {
		h.logger.Errorf("Failed to disable auto promote for group %d: %v", groupID, err)
		return fmt.Sprintf(`❌ *GAGAL MENONAKTIFKAN PROMOTE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal menonaktifkan auto promote: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Grup dengan ID tersebut tidak ditemukan
• Masalah koneksi database
• Grup sudah tidak aktif

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi admin*`, err.Error())
	}

	return fmt.Sprintf(`🛑 *AUTO PROMOTE DINONAKTIFKAN!*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *BERHASIL DINONAKTIFKAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

👥 *Grup:* %s
🆔 *ID:* %d
⏰ *Dihentikan:* Sekarang

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *INFORMASI PENTING*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

• Auto promote telah dihentikan untuk grup ini.
• Bot tidak akan mengirim promosi lagi ke grup ini.
• Gunakan *.enablegroup %d* untuk mengaktifkan kembali.
• Data grup dan statistik tetap tersimpan di sistem.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ *Perubahan berhasil disimpan!*`,
		groupInfo.Name, groupInfo.ID, groupID)
}

// HandleGroupStatusCommand menangani command .groupstatus [ID]
func (h *AdminCommandHandler) HandleGroupStatusCommand(evt *events.Message, args []string) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return "" // Tidak ada response untuk non-admin
	}

	if len(args) < 2 {
		return `❌ *FORMAT SALAH*

📝 **Format:** .groupstatus [ID]
📋 **Contoh:** .groupstatus 3

💡 Gunakan .listgroups untuk melihat ID grup`
	}

	if h.groupManagerService == nil {
		return `❌ *SERVICE TIDAK TERSEDIA*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		         *KESALAHAN SISTEM*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Service untuk manajemen grup tidak dikonfigurasi

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Service tidak diinisialisasi saat start-up
• Terjadi error internal

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Hubungi developer untuk perbaikan*`
	}

	// Parse ID grup
	groupID, err := strconv.Atoi(args[1])
	if err != nil {
		return `❌ *ID TIDAK VALID*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		         *FORMAT ID SALAH*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 ID grup harus berupa angka.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *CONTOH PENGGUNAAN*
• .groupstatus 3
• .listgroups

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *Gunakan .listgroups untuk melihat ID*`
	}

	// Ambil status grup
	groupInfo, dbGroup, err := h.groupManagerService.GetGroupStatus(groupID)
	if err != nil {
		h.logger.Errorf("Failed to get group status for %d: %v", groupID, err)
		return fmt.Sprintf(`❌ *GAGAL MENDAPATKAN STATUS*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		         *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mendapatkan status grup: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Grup dengan ID tersebut tidak ditemukan
• Masalah koneksi database
• Error internal sistem

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi admin*`, err.Error())
	}

	// Format status
	status := "❌ Tidak Aktif"
	if dbGroup != nil && dbGroup.IsActive {
		status = "✅ Aktif"
	}

	var startedInfo string
	if dbGroup != nil && dbGroup.StartedAt != nil {
		startedInfo = dbGroup.StartedAt.Format("2006-01-02 15:04")
	} else {
		startedInfo = "Belum pernah"
	}

	var lastPromoteInfo string
	if dbGroup != nil && dbGroup.LastPromoteAt != nil {
		lastPromoteInfo = dbGroup.LastPromoteAt.Format("2006-01-02 15:04")
	} else {
		lastPromoteInfo = "Belum pernah"
	}

	// Ambil jumlah template aktif
	templates, _ := h.templateService.GetActiveTemplates()
	templateCount := len(templates)

	return fmt.Sprintf(`📊 *STATUS GRUP AUTO PROMOTE*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *DETAIL GRUP*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

👥 *Nama Grup:* %s
🆔 *ID Grup:* %d
👤 *Jumlah Member:* %d orang

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *STATUS PROMOTE*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *Status Auto Promote:* %s
📅 *Promote Dimulai:* %s
⏰ *Promosi Terakhir:* %s
📝 *Total Template Aktif:* %d template

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *INFORMASI TEKNIS*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔧 *JID Grup:*
%s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *COMMANDS TERKAIT*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

• *.enablegroup %d*
	 _Aktifkan auto promote_

• *.disablegroup %d*
	 _Nonaktifkan auto promote_

• *.testgroup %d*
	 _Kirim promosi test_

• *.listgroups*
	 _Kembali ke daftar grup_`,
		groupInfo.Name, groupInfo.ID, groupInfo.MemberCount, status,
		startedInfo, lastPromoteInfo, templateCount, groupInfo.JID,
		groupID, groupID, groupID)
}

// HandleTestGroupCommand menangani command .testgroup [ID]
func (h *AdminCommandHandler) HandleTestGroupCommand(evt *events.Message, args []string) string {
	// Cek admin permission
	if !h.isAdmin(evt.Info.Sender.User) {
		return "" // Tidak ada response untuk non-admin
	}

	if len(args) < 2 {
		return `❌ *FORMAT SALAH*

📝 **Format:** .testgroup [ID]
📋 **Contoh:** .testgroup 3

💡 Gunakan .listgroups untuk melihat ID grup`
	}

	if h.groupManagerService == nil {
		return `❌ *SERVICE TIDAK TERSEDIA*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *KESALAHAN SISTEM*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Service untuk manajemen grup tidak dikonfigurasi

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Service tidak diinisialisasi saat start-up
• Terjadi error internal

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Hubungi developer untuk perbaikan*`
	}

	// Parse ID grup
	groupID, err := strconv.Atoi(args[1])
	if err != nil {
		return `❌ *ID TIDAK VALID*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *FORMAT ID SALAH*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 ID grup harus berupa angka.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *CONTOH PENGGUNAAN*
• .testgroup 3
• .listgroups

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *Gunakan .listgroups untuk melihat ID*`
	}

	// Ambil info grup
	groupInfo, err := h.groupManagerService.GetGroupByID(groupID)
	if err != nil {
		return fmt.Sprintf(`❌ *GRUP TIDAK DITEMUKAN*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *ID TIDAK VALID*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Grup dengan ID %d tidak ditemukan di database.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• ID grup salah atau tidak ada.
• Bot belum join grup tersebut.
• Grup sudah dihapus.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 *Gunakan .listgroups untuk melihat ID yang valid*`, groupID)
	}

	// Kirim test promosi
	err = h.groupManagerService.SendTestPromoteToGroup(groupID)
	if err != nil {
		h.logger.Errorf("Failed to send test promote to group %d: %v", groupID, err)
		return fmt.Sprintf(`❌ *GAGAL MENGIRIM TEST*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	          *TERJADI KESALAHAN*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 Gagal mengirim test promosi: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *KEMUNGKINAN PENYEBAB*
• Tidak ada template promosi yang aktif
• Bot tidak memiliki izin kirim pesan di grup
• Masalah koneksi WhatsApp
• Grup tidak aktif untuk promosi

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔄 *Coba lagi atau hubungi admin*`, err.Error())
	}

	return fmt.Sprintf(`🚀 *PROMOSI BERHASIL DIKIRIM!*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
          *BERHASIL TERKIRIM*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *DETAIL PENGIRIMAN*
👥 *Grup:* %s
🆔 *ID:* %d
📤 *Status:* TERKIRIM
🎲 *Template:* Random
⏰ *Waktu:* Sekarang

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 *INFORMASI*
• Promosi telah dikirim ke grup
• Template dipilih secara otomatis
• Tidak mempengaruhi jadwal rutin
• Silakan cek grup untuk melihat

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎮 *MONITORING*

• *.groupstatus %d*
  _Cek status grup_

• *.listgroups*
  _Kembali ke daftar grup_

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ *Cek grup untuk melihat hasilnya!*`,
		groupInfo.Name, groupInfo.ID, groupID)
}

// parseQuotedArgs memparse argument yang menggunakan tanda kutip
func (h *AdminCommandHandler) parseQuotedArgs(text string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false

	for i, char := range text {
		if char == '"' {
			if inQuotes {
				// End of quoted string
				args = append(args, current.String())
				current.Reset()
				inQuotes = false
			} else {
				// Start of quoted string
				inQuotes = true
			}
		} else if char == ' ' && !inQuotes {
			// Space outside quotes - separator
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(char)
		}

		// Handle end of string
		if i == len(text)-1 && current.Len() > 0 {
			args = append(args, current.String())
		}
	}

	return args
}

// HandleAdminCommands menangani semua command admin
func (h *AdminCommandHandler) HandleAdminCommands(evt *events.Message, messageText string) string {
	args := strings.Fields(messageText) // Gunakan original text untuk preserve case

	if len(args) == 0 {
		return ""
	}

	command := strings.ToLower(args[0])

	switch command {
	// Group Management Commands
	case ".listgroups":
		return h.HandleListGroupsCommand(evt)

	case ".enablegroup":
		return h.HandleEnableGroupCommand(evt, args)

	case ".enablemulti":
		return h.HandleEnableMultipleGroupsCommand(evt, args)

	case ".disablegroup":
		return h.HandleDisableGroupCommand(evt, args)

	case ".groupstatus":
		return h.HandleGroupStatusCommand(evt, args)

	case ".testgroup":
		return h.HandleTestGroupCommand(evt, args)

	// Template Management Commands
	case ".addtemplate":
		return h.HandleAddTemplateCommand(evt, args)

	case ".edittemplate":
		return h.HandleEditTemplateCommand(evt, args)

	case ".deletetemplate":
		return h.HandleDeleteTemplateCommand(evt, args)

	case ".templatestats":
		return h.HandleTemplateStatsCommand(evt)

	case ".promotestats":
		return h.HandlePromoteStatsCommand(evt)

	case ".activegroups":
		return h.HandleActiveGroupsCommand(evt)

	case ".fetchproducts":
		return h.HandleFetchProductsCommand(evt)

	case ".productstats":
		return h.HandleProductStatsCommand(evt)

	case ".deleteall":
		return h.HandleDeleteAllTemplatesCommand(evt)

	case ".deletemulti":
		return h.HandleDeleteMultipleTemplatesCommand(evt, args)

	default:
		return ""
	}
}
