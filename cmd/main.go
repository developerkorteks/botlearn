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
	promoteCfg := config.NewPromoteConfig()
	
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
		_ = f.Close(); _ = os.Remove(f.Name())
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
	learningMessageHandler := handlers.NewLearningMessageHandler(client, learningService, xrayConverterService, logger, promoteCfg.AdminNumbers)
	
	// Setup dashboard server dengan WhatsApp pairing support
	dashboardServer := web.NewDashboardServer(learningRepo, logger, promoteCfg.AdminNumbers)
	dashboardServer.SetWhatsAppClient(client)
	dashboardServer.SetWAManager(waManager)
	dashboardServer.SetQRGenerator(qrGen)
	
	logger.Success("Learning System initialized!")
	
	// STEP 8: Setup Auto Promote System (jika diaktifkan)
	var autoPromoteService *services.AutoPromoteService
	
	if promoteCfg.EnableAutoPromote {
		logger.Info("Initializing Auto Promote System...")
		
		// Setup database untuk auto promote
		promoteDB, promoteRepo, err := database.InitializeDatabase(promoteCfg.PromoteDatabasePath)
		if err != nil {
			logger.Errorf("Failed to initialize promote database: %v", err)
			os.Exit(1)
		}
		defer promoteDB.Close()
		
		// Setup services (template service jika diperlukan)
		// templateService := services.NewTemplateService(promoteRepo, logger)
		autoPromoteService = services.NewAutoPromoteService(client, promoteRepo, logger)
		// Set interval dari konfigurasi
		autoPromoteService.SetInterval(promoteCfg.AutoPromoteInterval)
		// Services untuk auto promote (jika diperlukan nanti)
		// apiProductService := services.NewAPIProductService(templateService, logger)
		// groupManagerService := services.NewGroupManagerService(client, promoteRepo, logger)
		
		// Setup command handlers (if needed for specific use cases)
		// promoteCommandHandler := handlers.NewPromoteCommandHandler(autoPromoteService, templateService, logger)
		// adminCommandHandler := handlers.NewAdminCommandHandler(autoPromoteService, templateService, apiProductService, groupManagerService, logger, promoteCfg.AdminNumbers)
		
		logger.Success("Auto Promote System initialized!")
	}
	
	// STEP 9: Setup handlers untuk menangani pesan dan event
	// Gunakan learning message handler sebagai handler utama
	// Event handler menangani semua event WhatsApp (koneksi, pesan, dll)
	eventHandler := handlers.NewEventHandler(client, learningMessageHandler)
	
	// STEP 10: Daftarkan event handler ke client
	client.AddEventHandler(eventHandler.HandleEvent)
	
	// STEP 11: Start Dashboard Server
	logger.Info("Starting Dashboard Server...")
	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil || port == 0 {
		port = 1462 // Default port
	}

	go func() {
		if err := dashboardServer.StartServer(port); err != nil {
			logger.Errorf("Dashboard server error: %v", err)
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
	
	// STEP 13: Start Auto Promote Scheduler (jika diaktifkan)
	if autoPromoteService != nil {
		logger.Info("Starting Auto Promote Scheduler...")
		autoPromoteService.StartScheduler()
		
		// Log konfigurasi auto promote
		logger.Infof("Auto Promote Config: %d admin(s), %d hour interval", 
			len(promoteCfg.AdminNumbers), promoteCfg.AutoPromoteInterval)
	}
	
	// STEP 14: Bot siap digunakan
	logger.Success("Bot berhasil terhubung ke WhatsApp!")
	logger.Info("Bot siap menerima pesan...")
	
	if promoteCfg.EnableAutoPromote {
		logger.Success("🚀 Auto Promote System is READY!")
		logger.Info("Commands: .aca, .disableaca, .promotehelp")
	}
	
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
	
	// Stop auto promote scheduler jika berjalan
	if autoPromoteService != nil {
		logger.Info("Stopping Auto Promote Scheduler...")
		autoPromoteService.StopScheduler()
	}
	
	// Disconnect menggunakan enhanced manager
	waManager.Disconnect()
	logger.Success("Bot berhasil dihentikan. Sampai jumpa!")
}

// connectWithQR function removed - now using enhanced WAManager.ConnectWithQRSafely()
// which provides better anti-spam protection and rate limiting