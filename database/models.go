// Package database berisi model dan struktur data untuk fitur auto promote dan learning
package database

import (
	"time"
)

// AutoPromoteGroup menyimpan status auto promote per grup
type AutoPromoteGroup struct {
	ID            int       `json:"id" db:"id"`
	GroupJID      string    `json:"group_jid" db:"group_jid"`           // JID grup WhatsApp
	IsActive      bool      `json:"is_active" db:"is_active"`           // Status aktif/tidak
	StartedAt     *time.Time `json:"started_at" db:"started_at"`        // Waktu mulai auto promote
	LastPromoteAt *time.Time `json:"last_promote_at" db:"last_promote_at"` // Waktu terakhir kirim promosi
	CreatedAt     time.Time `json:"created_at" db:"created_at"`         // Waktu dibuat
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`         // Waktu diupdate
}

// PromoteTemplate menyimpan template promosi bisnis
type PromoteTemplate struct {
	ID        int       `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`         // Judul template (misal: "Produk Unggulan")
	Content   string    `json:"content" db:"content"`     // Isi template promosi
	Category  string    `json:"category" db:"category"`   // Kategori (produk, diskon, testimoni, dll)
	IsActive  bool      `json:"is_active" db:"is_active"` // Status aktif/tidak
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PromoteLog menyimpan log pengiriman promosi untuk tracking
type PromoteLog struct {
	ID         int       `json:"id" db:"id"`
	GroupJID   string    `json:"group_jid" db:"group_jid"`     // JID grup tujuan
	TemplateID int       `json:"template_id" db:"template_id"` // ID template yang digunakan
	Content    string    `json:"content" db:"content"`         // Isi pesan yang dikirim
	SentAt     time.Time `json:"sent_at" db:"sent_at"`         // Waktu pengiriman
	Success    bool      `json:"success" db:"success"`         // Status berhasil/gagal
	ErrorMsg   *string   `json:"error_msg" db:"error_msg"`     // Pesan error jika gagal
}

// PromoteStats menyimpan statistik promosi untuk monitoring
type PromoteStats struct {
	ID              int       `json:"id" db:"id"`
	Date            string    `json:"date" db:"date"`                         // Tanggal (YYYY-MM-DD)
	TotalGroups     int       `json:"total_groups" db:"total_groups"`         // Total grup aktif
	TotalMessages   int       `json:"total_messages" db:"total_messages"`     // Total pesan terkirim
	SuccessMessages int       `json:"success_messages" db:"success_messages"` // Pesan berhasil
	FailedMessages  int       `json:"failed_messages" db:"failed_messages"`   // Pesan gagal
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// DefaultPromoteTemplates berisi template default untuk promosi bisnis
var DefaultPromoteTemplates = []PromoteTemplate{
	{
		Title:    "Produk Unggulan",
		Category: "produk",
		Content: `🌟 *PRODUK UNGGULAN HARI INI* 🌟

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✨ *Kualitas Premium* | 💎 *Harga Terjangkau*
🚀 *Stok Terbatas* | ⚡ *Pengiriman Cepat*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🛒 *ORDER SEKARANG JUGA!*

📱 *WhatsApp:* wa.me/6208123456789
🌐 *Website:* bit.ly/produk-unggulan
💳 *Pembayaran:* Transfer/COD/E-Wallet

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⏰ *Jangan sampai kehabisan!*
🎁 *Bonus untuk 50 pembeli pertama*

#ProdukUnggulan #KualitasPremium #OrderSekarang`,
		IsActive: true,
	},
	{
		Title:    "Diskon & Promo",
		Category: "diskon",
		Content: `🎉 *PROMO SPESIAL HARI INI* 🎉

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💥 *DISKON HINGGA 50%* 💥
🎯 *Semua Produk* | ⏰ *Terbatas Waktu*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎁 *BONUS SPESIAL:*
• Gratis ongkir min. 100k
• Cashback 10% untuk member
• Voucher belanja berikutnya

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🛒 *BURUAN ORDER!*

📱 *WhatsApp:* wa.me/6208123456789
💳 *Pembayaran:* Transfer/COD/E-Wallet/QRIS
⏰ *Berakhir:* {DATE} 23:59 WIB

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔥 *Jangan sampai menyesal!*
✨ *Kesempatan emas ini terbatas!*

#PromoSpesial #Diskon50Persen #TerbatasWaktu`,
		IsActive: true,
	},
	{
		Title:    "Testimoni Customer",
		Category: "testimoni",
		Content: `⭐ *TESTIMONI CUSTOMER SETIA* ⭐

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💬 *"Produknya bagus banget, sesuai ekspektasi!"*
👤 Bu Sarah, Jakarta
⭐⭐⭐⭐⭐

💬 *"Pelayanan ramah, pengiriman cepat!"*
👤 Pak Budi, Surabaya  
⭐⭐⭐⭐⭐

💬 *"Harga murah, kualitas juara!"*
👤 Mbak Siti, Bandung
⭐⭐⭐⭐⭐

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🏆 *RATING 4.9/5* dari 1000+ customer
🎯 *99% Customer Puas* dengan pelayanan kami

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🛒 *BERGABUNGLAH DENGAN MEREKA!*

📱 *Order Sekarang:* wa.me/6208123456789
🌟 *Dapatkan Pengalaman Terbaik!*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🙏 *Terima kasih atas kepercayaan Anda!*

#TestimoniCustomer #KepuasanPelanggan #Terpercaya`,
		IsActive: true,
	},
	{
		Title:    "Flash Sale",
		Category: "flashsale",
		Content: `⚡ *FLASH SALE ALERT!* ⚡

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔥 *HANYA 2 JAM LAGI!* 🔥
💰 *Harga Super Murah* | 🏃‍♂️ *Stok Terbatas*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 *STOK TERSISA:*
🟢 Produk A: *5 pcs* tersisa
🟡 Produk B: *3 pcs* tersisa  
🟢 Produk C: *8 pcs* tersisa

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚡ *ORDER SEKARANG!*

📱 *WhatsApp:* wa.me/6208123456789
⏰ *Berakhir:* 23:59 WIB
🚀 *Checkout Cepat:* bit.ly/flashsale-now

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💨 *BURUAN! SEBELUM KEHABISAN!*
🎯 *First come, first served!*

#FlashSale #StokTerbatas #BuruanOrder`,
		IsActive: true,
	},
	{
		Title:    "Produk Baru",
		Category: "produk_baru",
		Content: `🆕 *LAUNCHING PRODUK TERBARU!* 🆕

🎊 Kami bangga memperkenalkan inovasi terbaru!
✨ Fitur canggih, desain modern
🏆 Kualitas terbaik di kelasnya

🎁 *PROMO LAUNCHING:*
• Diskon 30% untuk 100 pembeli pertama
• Gratis ongkir seluruh Indonesia
• Garansi resmi 1 tahun

📱 Pre-order: 08123456789
🚀 Jadilah yang pertama memilikinya!

#ProdukBaru #Launching #PreOrder`,
		IsActive: true,
	},
	{
		Title:    "Bundle Package",
		Category: "bundle",
		Content: `📦 *PAKET HEMAT BUNDLE!* 📦

💡 Beli 1 dapat 3? Why not!
🎯 Hemat hingga 40% dari harga normal
🎁 Bonus eksklusif untuk paket lengkap

📋 *Paket yang tersedia:*
• Paket A: 3 produk = 150k (normal 250k)
• Paket B: 5 produk = 200k (normal 350k)
• Paket C: 10 produk = 350k (normal 600k)

💰 Makin banyak makin hemat!
📱 Order: 08123456789

#BundlePackage #PaketHemat #MakinBanyakMakinHemat`,
		IsActive: true,
	},
	{
		Title:    "Free Ongkir",
		Category: "ongkir",
		Content: `🚚 *GRATIS ONGKIR SELURUH INDONESIA!* 🚚

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎉 *TANPA MINIMUM PEMBELIAN!*
📦 *Pengiriman Aman & Terpercaya*
⏰ *Estimasi 1-3 Hari Kerja*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🌟 *KEUNTUNGAN EKSKLUSIF:*

📦 *Packing Aman & Rapi*
🛡️ *Asuransi Pengiriman*
📍 *Tracking Real-Time*
🤝 *Customer Service 24/7*
🚀 *Ekspedisi Terpercaya*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🛒 *ORDER SEKARANG JUGA!*

📱 *WhatsApp:* wa.me/6208123456789
🌐 *Website:* bit.ly/free-ongkir
📱 *Cek Ongkir:* bit.ly/cek-ongkir

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *Hemat lebih banyak dengan gratis ongkir!*
✨ *Belanja makin untung!*

#GratisOngkir #PengirimanAman #OrderSekarang`,
		IsActive: true,
	},
	{
		Title:    "Cashback & Reward",
		Category: "cashback",
		Content: `💰 *PROGRAM CASHBACK & REWARD!* 💰

🎁 Belanja makin untung dengan reward points!
💎 Tukar poin dengan produk gratis
🔄 Cashback langsung ke rekening

🏆 *Benefit member:*
• Cashback 5% setiap pembelian
• Poin reward setiap transaksi
• Diskon eksklusif member
• Akses produk limited edition

👑 Daftar member sekarang!
📱 WhatsApp: 08123456789

#CashbackReward #MemberExclusive #BelanjaMakinUntung`,
		IsActive: true,
	},
	{
		Title:    "Limited Stock",
		Category: "limited",
		Content: `⚠️ *STOK TERBATAS - SEGERA HABIS!* ⚠️

🔥 Produk favorite hampir sold out!
📊 Sisa stok: 7 pcs saja
⏰ Kemungkinan habis dalam 24 jam

😱 *Jangan sampai menyesal!*
• Produk best seller #1
• Rating 5 bintang dari customer
• Sudah terjual 500+ pcs bulan ini

🏃‍♂️ BURUAN ORDER SEBELUM KEHABISAN!
📱 WhatsApp: 08123456789

#StokTerbatas #BestSeller #BuruanOrder`,
		IsActive: true,
	},
	{
		Title:    "Contact Info",
		Category: "contact",
		Content: `📞 *HUBUNGI KAMI UNTUK ORDER!* 📞

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🛒 *CARA ORDER:*

📱 *WhatsApp:* wa.me/6208123456789
📲 *Telegram:* t.me/tokoonline
📸 *Instagram:* instagram.com/toko.online
🌐 *Website:* www.tokoonline.com

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💳 *METODE PEMBAYARAN:*

🏦 *Transfer Bank:* BCA, Mandiri, BRI
💰 *E-Wallet:* OVO, DANA, GoPay, ShopeePay
📱 *QRIS:* Scan & Pay
🚚 *COD:* Area Jabodetabek

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⏰ *JAM OPERASIONAL:*

🗓️ *Senin-Sabtu:* 08:00-22:00 WIB
🗓️ *Minggu:* 10:00-20:00 WIB
🤖 *Auto Reply:* 24/7

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✨ *Siap melayani Anda dengan sepenuh hati!*

#ContactInfo #CaraOrder #JamOperasional`,
		IsActive: true,
	},
}

// ===============================
// LEARNING BOT MODELS
// ===============================

// LearningGroup menyimpan grup yang diizinkan untuk bot pembelajaran
type LearningGroup struct {
	ID          int       `json:"id" db:"id"`
	GroupJID    string    `json:"group_jid" db:"group_jid"`       // JID grup WhatsApp
	GroupName   string    `json:"group_name" db:"group_name"`     // Nama grup
	IsActive    bool      `json:"is_active" db:"is_active"`       // Status aktif/tidak
	Description string    `json:"description" db:"description"`   // Deskripsi grup
	CreatedBy   string    `json:"created_by" db:"created_by"`     // Admin yang menambahkan
	CreatedAt   time.Time `json:"created_at" db:"created_at"`     // Waktu dibuat
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`     // Waktu diupdate
}

// LearningCommand menyimpan command pembelajaran custom
type LearningCommand struct {
	ID              int       `json:"id" db:"id"`
	Command         string    `json:"command" db:"command"`                   // Command seperti ".listbugs"
	Title           string    `json:"title" db:"title"`                       // Judul command
	Description     string    `json:"description" db:"description"`           // Deskripsi command
	ResponseType    string    `json:"response_type" db:"response_type"`       // "text", "image", "video", "audio", "sticker", "file"
	TextContent     *string   `json:"text_content" db:"text_content"`         // Konten text untuk response
	MediaFilePath   *string   `json:"media_file_path" db:"media_file_path"`   // Path file media
	Caption         *string   `json:"caption" db:"caption"`                   // Caption untuk media
	Category        string    `json:"category" db:"category"`                 // "injec", "pembelajaran", "informasi", dll
	IsActive        bool      `json:"is_active" db:"is_active"`               // Status aktif/tidak
	UsageCount      int       `json:"usage_count" db:"usage_count"`           // Jumlah penggunaan
	CreatedBy       string    `json:"created_by" db:"created_by"`             // Admin yang membuat
	CreatedAt       time.Time `json:"created_at" db:"created_at"`             // Waktu dibuat
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`             // Waktu diupdate
}

// AutoResponse menyimpan auto response untuk kata kunci tertentu (candaan)
type AutoResponse struct {
	ID            int       `json:"id" db:"id"`
	Keyword       string    `json:"keyword" db:"keyword"`                   // Kata kunci seperti "cape", "gabut"
	ResponseType  string    `json:"response_type" db:"response_type"`       // "sticker", "audio", "text", "mixed"
	StickerPath   *string   `json:"sticker_path" db:"sticker_path"`         // Path file sticker
	AudioPath     *string   `json:"audio_path" db:"audio_path"`             // Path file audio/voice
	TextResponse  *string   `json:"text_response" db:"text_response"`       // Text response tambahan
	IsActive      bool      `json:"is_active" db:"is_active"`               // Status aktif/tidak
	UsageCount    int       `json:"usage_count" db:"usage_count"`           // Jumlah penggunaan
	CreatedBy     string    `json:"created_by" db:"created_by"`             // Admin yang membuat
	CreatedAt     time.Time `json:"created_at" db:"created_at"`             // Waktu dibuat
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`             // Waktu diupdate
}

// CommandUsageLog menyimpan log penggunaan command untuk monitoring
type CommandUsageLog struct {
	ID           int       `json:"id" db:"id"`
	CommandType  string    `json:"command_type" db:"command_type"`         // "learning_command" atau "auto_response"
	CommandValue string    `json:"command_value" db:"command_value"`       // Nilai command atau keyword
	GroupJID     string    `json:"group_jid" db:"group_jid"`               // JID grup
	UserJID      string    `json:"user_jid" db:"user_jid"`                 // JID user
	ResponseType string    `json:"response_type" db:"response_type"`       // Tipe response yang dikirim
	Success      bool      `json:"success" db:"success"`                   // Status berhasil/gagal
	ErrorMessage *string   `json:"error_message" db:"error_message"`       // Pesan error jika gagal
	UsedAt       time.Time `json:"used_at" db:"used_at"`                   // Waktu penggunaan
}

// DefaultLearningCommands berisi command default untuk pembelajaran
var DefaultLearningCommands = []LearningCommand{
	{
		Command:      ".help",
		Title:        "Bantuan Command",
		Description:  "Menampilkan daftar command yang tersedia",
		ResponseType: "text",
		TextContent:  stringPtr(`📚 *BANTUAN BOT PEMBELAJARAN* 📚

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
           *COMMAND TERSEDIA*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔧 *INJEC & VPN:*
• .listbugs - List bug server VPN
• .config - Config file VPN
• .tutorial - Tutorial setup VPN

📚 *PEMBELAJARAN:*
• .html - Tutorial HTML dasar
• .css - Tutorial CSS dasar
• .js - Tutorial JavaScript dasar

ℹ️ *INFORMASI:*
• .info - Info tentang bot
• .help - Bantuan ini

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎯 *Bot ini untuk pembelajaran saja*
🚫 *Gunakan dengan bijak*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`),
		Category:     "informasi",
		IsActive:     true,
		CreatedBy:    "system",
	},
	{
		Command:      ".info",
		Title:        "Informasi Bot",
		Description:  "Informasi tentang bot pembelajaran",
		ResponseType: "text",
		TextContent:  stringPtr(`ℹ️ *BOT PEMBELAJARAN & INJEC* ℹ️

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
            *TENTANG BOT*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🤖 *Nama:* Learning Injec Bot
📝 *Versi:* 1.0.0
🎯 *Tujuan:* Pembelajaran & Edukasi
💻 *Platform:* WhatsApp Bot
🛠️ *Bahasa:* Go (Golang)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎓 *FITUR UTAMA:*
• Command pembelajaran dinamis
• Tutorial VPN & Injec
• Auto response candaan
• Multi-format media support
• Group access control

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️ *DISCLAIMER:*
Bot ini dibuat untuk tujuan pembelajaran.
Segala penggunaan di luar pembelajaran
menjadi tanggung jawab user.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📞 *Support:* Hubungi admin grup`),
		Category:     "informasi",
		IsActive:     true,
		CreatedBy:    "system",
	},
}

// DefaultAutoResponses berisi auto response default untuk candaan
var DefaultAutoResponses = []AutoResponse{
	{
		Keyword:      "cape",
		ResponseType: "text",
		TextResponse: stringPtr("😴 Yah cape ya bang... istirahat dulu!"),
		IsActive:     true,
		CreatedBy:    "system",
	},
	{
		Keyword:      "gabut",
		ResponseType: "text",
		TextResponse: stringPtr("😂 Gabut nih? Coba pelajari command .help deh!"),
		IsActive:     true,
		CreatedBy:    "system",
	},
	{
		Keyword:      "semangat",
		ResponseType: "text",
		TextResponse: stringPtr("💪 SEMANGAT TERUS! Belajar itu kunci sukses!"),
		IsActive:     true,
		CreatedBy:    "system",
	},
}

// Helper function untuk string pointer
func stringPtr(s string) *string {
	return &s
}