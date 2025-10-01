package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	
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
	
	// STEP 4: Setup database untuk session WhatsApp
	// Database SQLite untuk menyimpan session agar tidak perlu login berulang
	logger.Info("Menginisialisasi database session...")
	dbLog := waLog.Noop
	container, err := sqlstore.New(context.Background(), "sqlite3", "file:"+cfg.DatabasePath+"?_foreign_keys=on", dbLog)
	if err != nil {
		logger.Errorf("Gagal membuat database: %v", err)
		os.Exit(1)
	}
	
	// STEP 5: Ambil device store dari database
	// Device store berisi informasi device WhatsApp yang tersimpan
	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		logger.Errorf("Gagal mendapatkan device store: %v", err)
		os.Exit(1)
	}
	
	// STEP 6: Buat WhatsApp client
	// Client adalah objek utama untuk berinteraksi dengan WhatsApp
	logger.Info("Membuat WhatsApp client...")
	clientLog := waLog.Stdout("Client", cfg.LogLevel, true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	
	// STEP 7: Setup Learning System
	logger.Info("Initializing Learning System...")
	
	// Setup database untuk learning
	learningDB, learningRepo, err := database.InitializeLearningDatabase("data/learning.db")
	if err != nil {
		logger.Errorf("Failed to initialize learning database: %v", err)
		os.Exit(1)
	}
	defer learningDB.Close()
	
	// Setup learning service
	learningService := services.NewLearningService(client, learningRepo, logger)
	
	// Setup XRay converter service
	xrayConverterService := services.NewXRayConverterService(learningRepo, logger)
	
	// Insert default XRay converters
	logger.Info("Setting up default XRay converters...")
	if err := database.InsertDefaultConverters(learningRepo); err != nil {
		logger.Errorf("Failed to insert default converters: %v", err)
	} else {
		logger.Success("Default XRay converters setup complete!")
	}
	
	// Setup learning message handler
	learningMessageHandler := handlers.NewLearningMessageHandler(client, learningService, xrayConverterService, logger, promoteCfg.AdminNumbers)
	
	// Setup dashboard server
	dashboardServer := web.NewDashboardServer(learningRepo, logger, promoteCfg.AdminNumbers)
	dashboardServer.SetWhatsAppClient(client)
	
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
	
	// STEP 12: Connect ke WhatsApp
	if client.Store.ID == nil {
		// Belum login, perlu scan QR code
		logger.Warning("Belum login, memerlukan QR code...")
		err = connectWithQR(client, qrGen, logger)
		if err != nil {
			logger.Errorf("Gagal connect dengan QR: %v", err)
			os.Exit(1)
		}
	} else {
		// Sudah login sebelumnya, langsung connect
		logger.Info("Sudah login sebelumnya, connecting...")
		err = client.Connect()
		if err != nil {
			logger.Errorf("Gagal connect: %v", err)
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
	
	client.Disconnect()
	logger.Success("Bot berhasil dihentikan. Sampai jumpa!")
}

// connectWithQR menangani proses koneksi dengan QR code
// Fungsi ini akan menampilkan QR code dan menunggu user untuk scan
func connectWithQR(client *whatsmeow.Client, qrGen *utils.QRCodeGenerator, logger *utils.Logger) error {
	// Dapatkan channel untuk menerima QR code dari WhatsApp
	qrChan, err := client.GetQRChannel(context.Background())
	if err != nil {
		return err
	}
	
	// Mulai proses koneksi
	err = client.Connect()
	if err != nil {
		return err
	}
	
	// Loop untuk menangani event QR code
	for evt := range qrChan {
		switch evt.Event {
		case "code":
			// QR code baru diterima, tampilkan ke user
			logger.Info("QR code diterima, menampilkan...")
			err = qrGen.GenerateAndDisplay(evt.Code)
			if err != nil {
				logger.Errorf("Gagal menampilkan QR code: %v", err)
				// Tetap lanjut, tampilkan QR code sebagai text
				logger.Infof("QR Code (text): %s", evt.Code)
			}
			
		case "success":
			// Login berhasil
			logger.Success("QR code berhasil di-scan! Login berhasil.")
			return nil
			
		case "timeout":
			// QR code timeout, akan generate yang baru
			logger.Warning("QR code timeout, generating QR code baru...")
			
		case "error":
			// Error dalam proses login
			logger.Error("Error dalam proses login QR code")
			return fmt.Errorf("QR code login error")
			
		default:
			// Event lain
			logger.Debugf("QR code event: %s", evt.Event)
		}
	}
	
	return nil
}