package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nabilulilalbab/promote/config"
	"github.com/nabilulilalbab/promote/database"
	"github.com/nabilulilalbab/promote/handlers"
	"github.com/nabilulilalbab/promote/services"
	"github.com/nabilulilalbab/promote/utils"
	"github.com/nabilulilalbab/promote/web"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// WhatsApp Bot dengan struktur yang rapi dan mudah dipelajari
// File ini adalah entry point utama aplikasi
func main() {
	// STEP 1: Load konfigurasi
	// Konfigurasi berisi semua pengaturan bot seperti database path, auto reply, dll
	cfg := config.NewConfig()

	// STEP 2: Setup logger
	// Logger untuk menampilkan informasi dengan format yang rapi
	logger := utils.NewLogger("BOT", true)
	logger.Info("Memulai WhatsApp Bot...")

	// STEP 3: Setup QR code generator
	// QR code generator untuk menampilkan QR code visual di terminal
	qrGen := utils.NewQRCodeGenerator(cfg.QRCodePath)

	// STEP 3.5: Ensure data dir writable
	if err := os.MkdirAll("data", 0o775); err != nil {
		fmt.Printf("⚠️ Warning: gagal membuat folder data: %v\n", err)
	}
	if f, err := os.CreateTemp("data", ".permcheck_*"); err == nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
	} else {
		fmt.Printf("⚠️ Warning: folder data mungkin tidak writable: %v\n", err)
	}

	// STEP 4: Setup Enhanced WhatsApp Manager dengan anti-spam protection
	logger.Info("Menginisialisasi Enhanced WhatsApp Manager...")
	waManager, err := utils.NewWAManager(cfg.DatabasePath, logger)
	if err != nil {
		logger.Errorf("Gagal membuat WhatsApp Manager: %v", err)
		os.Exit(1)
	}

	// STEP 5: Buat WhatsApp client dengan proteksi anti-spam
	err = waManager.CreateClient()
	if err != nil {
		logger.Errorf("Gagal membuat WhatsApp client: %v", err)
		os.Exit(1)
	}

	// Gunakan client dari manager
	client := waManager.Client

	// STEP 7: Setup Learning System
	logger.Info("Initializing Learning System...")

	// Setup database untuk learning
	learningDB, learningRepo, err := database.InitializeLearningDatabase("data/learning.db")
	if err != nil {
		logger.Errorf("Failed to initialize learning database: %v", err)
		os.Exit(1)
	}
	defer learningDB.Close()

	// Setup XRay converter service (create first)
	xrayConverterService := services.NewXRayConverterService(learningRepo, logger)

	// Setup learning service (with XRay service)
	learningService := services.NewLearningService(client, learningRepo, logger, xrayConverterService)

	// Insert default XRay converters
	logger.Info("Setting up default XRay converters...")
	if err := database.InsertDefaultConverters(learningRepo); err != nil {
		logger.Errorf("Failed to insert default converters: %v", err)
	} else {
		logger.Success("Default XRay converters setup complete!")
	}

	// Setup learning message handler
	// Hapus promoteCfg.AdminNumbers karena sudah tidak ada bedanya pengguna promote atau tidak
	learningMessageHandler := handlers.NewLearningMessageHandler(client, learningService, xrayConverterService, logger, []string{})

	// Setup dashboard server dengan WhatsApp pairing support
	dashboardServer := web.NewDashboardServer(learningRepo, logger)
	dashboardServer.SetWhatsAppClient(client)
	dashboardServer.SetWAManager(waManager)
	dashboardServer.SetQRGenerator(qrGen)

	// STEP 9: Setup handlers untuk menangani pesan dan event
	// Gunakan learning message handler sebagai handler utama
	// Event handler menangani semua event WhatsApp (koneksi, pesan, dll)
	eventHandler := handlers.NewEventHandler(client, learningMessageHandler)

	// Tambahkan handler Hot Swap WA Session tanpa kill app
	dashboardServer.SetOnReloadSessionDB(func() error {
		logger.Info("Executing Hot-Swap WA Session re-initialization...")

		// STEP 1: Buat WhatsApp Manager baru
		// waManager lama sudah di-disconnect di dashboard_server.go
		newWaManager, err := utils.NewWAManager(cfg.DatabasePath, logger)
		if err != nil {
			return fmt.Errorf("gagal membuat WhatsApp Manager baru: %w", err)
		}

		// STEP 2: Buat WhatsApp client baru
		err = newWaManager.CreateClient()
		if err != nil {
			return fmt.Errorf("gagal membuat WhatsApp client baru: %w", err)
		}

		// STEP 3: Update reference pada semua services yang jalan secara dinamis
		waManager = newWaManager
		client = newWaManager.Client

		dashboardServer.SetWhatsAppClient(client)
		dashboardServer.SetWAManager(waManager)
		learningService.SetWhatsAppClient(client)
		learningMessageHandler.SetWhatsAppClient(client)

		// Event handler butuh di-setup lagi ke client yang baru
		// handler logic-nya (messageHandler dsb) tetap utuh
		eventHandler.SetWhatsAppClient(client)
		client.AddEventHandler(eventHandler.HandleEvent)

		logger.Success("Hot-Swap WA Session completed.")
		return nil
	})

	// STEP 10: Daftarkan event handler ke client
	client.AddEventHandler(eventHandler.HandleEvent)

	// STEP 11: Start Dashboard Server
	logger.Info("Starting Dashboard Server...")
	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil || port == 0 {
		port = 42981 // Default port
	}

	go func() {
		if err := dashboardServer.StartServer(port); err != nil {
			logger.Errorf("Dashboard server error: %v", err)
			logger.Error("Dashboard gagal start — bot dihentikan agar tidak jalan buta (silent failure).")
			os.Exit(1)
		}
	}()
	logger.Successf("Dashboard server started on http://localhost:%d", port)

	// STEP 12: Connect ke WhatsApp dengan Enhanced Protection
	if !waManager.IsLoggedIn() {
		// Belum login, pairing dimulai dari dashboard (hindari bentrok QR)
		autoQR := os.Getenv("AUTO_QR_ON_START") == "true" || os.Getenv("AUTO_QR_ON_START") == "1"
		if autoQR {
			logger.Warning("Belum login, AUTO_QR_ON_START=true → memulai QR pairing otomatis...")
			if err = waManager.ConnectWithQRSafely(qrGen); err != nil {
				logger.Errorf("Gagal connect dengan QR: %v", err)
				os.Exit(1)
			}
		} else {
			logger.Warning("Belum login. Mulai pairing dari Dashboard > WhatsApp Pairing.")
		}
	} else {
		// Sudah login sebelumnya, langsung connect dengan proteksi
		logger.Info("Sudah login sebelumnya, connecting dengan proteksi anti-spam...")
		if err = waManager.ConnectSafely(); err != nil {
			logger.Errorf("Gagal connect: %v", err)
			logger.Info("💡 Tips: Jika error rate limit, tunggu beberapa menit")
			os.Exit(1)
		}
	}

	// STEP 13: (Dihapus: Start Auto Promote Scheduler)

	// STEP 14: Bot siap digunakan
	logger.Success("Bot berhasil terhubung ke WhatsApp!")
	logger.Info("Bot siap menerima pesan...")

	logger.Info("Tekan Ctrl+C untuk menghentikan bot")

	// STEP 15: Tampilkan informasi learning system
	logger.Success("🚀 Learning Bot System is READY!")
	logger.Infof("Dashboard: http://localhost:%d", port)
	logger.Info("Admin commands: .addgroup, .removegroup, .listgroups, .stats, .logs")
	logger.Info("Learning commands: .help, .info, .listbugs (and more via dashboard)")

	// STEP 16: Tampilkan informasi XRay converter
	logger.Success("🔄 XRay Converter System is READY!")
	logger.Info("Converter commands: .convertbizz, .convertinsta, .convertnetflix, .convertgopay, .convertgrpc")
	logger.Info("Usage: .convertbizz vmess://xxx | .convertinsta trojan://xxx")

	// STEP 16: Wait for interrupt signal (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// STEP 17: Graceful shutdown
	logger.Info("Menghentikan bot...")

	// Disconnect menggunakan enhanced manager
	waManager.Disconnect()
	logger.Success("Bot berhasil dihentikan. Sampai jumpa!")
}

// connectWithQR function removed - now using enhanced WAManager.ConnectWithQRSafely()
// which provides better anti-spam protection and rate limiting
