# 📋 PLANNING: Custom Bot Dashboard untuk Grup Pembelajaran

## 🎯 **TUJUAN UTAMA**
Mengembangkan sistem bot WhatsApp yang dapat:
1. **Digunakan hanya di grup tertentu** yang diizinkan admin
2. **Command dan response dinamis** melalui dashboard website
3. **Multi-format response**: text, gambar, sticker, video, audio/musik
4. **Akses terbatas** hanya admin yang bisa mengatur

---

## 🏗️ **ANALISIS STRUKTUR KODE SAAT INI**

### ✅ **Yang Sudah Ada:**
1. **Bot WhatsApp** menggunakan `whatsmeow` library
2. **Database system** dengan SQLite
3. **Admin management** dengan autentikasi nomor admin
4. **Group management** untuk kontrol grup
5. **Template system** untuk content dinamis
6. **Auto promote scheduler** yang bisa dimodifikasi
7. **Message handler** yang sudah support multi-format

### 📁 **Struktur File Penting:**
```
├── cmd/main.go                 # Entry point aplikasi
├── config/
│   ├── config.go              # Konfigurasi dasar
│   └── promote_config.go      # Konfigurasi auto promote
├── database/
│   ├── models.go              # Model database
│   ├── repository.go          # Database operations
│   └── migrations.go          # Database schema
├── handlers/
│   ├── message.go             # Handler pesan utama
│   ├── admin_commands.go      # Command admin
│   └── events.go              # Event handler WhatsApp
├── services/
│   ├── auto_promote.go        # Service auto promote
│   ├── template.go            # Template management
│   └── group_manager.go       # Group management
└── utils/
    ├── logger.go              # Logging system
    └── qrcode.go              # QR code generator
```

---

## 🚀 **RENCANA IMPLEMENTASI**

### **FASE 1: Custom Command System** 
> **Target: 3-5 hari**

#### 1.1 Database Schema Baru
```sql
-- Tabel untuk custom commands
CREATE TABLE custom_commands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    response_type ENUM('text', 'image', 'video', 'audio', 'sticker') DEFAULT 'text',
    response_content TEXT NOT NULL,
    media_url TEXT,
    is_active BOOLEAN DEFAULT true,
    created_by VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel untuk grup yang diizinkan
CREATE TABLE allowed_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_jid VARCHAR(100) UNIQUE NOT NULL,
    group_name VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    allowed_commands TEXT, -- JSON array command yang diizinkan
    created_by VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel untuk log penggunaan command
CREATE TABLE command_usage_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command VARCHAR(100),
    group_jid VARCHAR(100),
    user_jid VARCHAR(100),
    response_type VARCHAR(20),
    success BOOLEAN,
    error_message TEXT,
    used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 1.2 Service Layer Baru
- **`CustomCommandService`**: Mengelola custom commands
- **`GroupAccessService`**: Mengelola akses grup
- **`MediaHandlerService`**: Menangani berbagai format media

#### 1.3 Command Handler Extension
- Extend `MessageHandler` untuk custom commands
- Support multi-format response (text, image, video, audio, sticker)
- Validasi grup yang diizinkan

### **FASE 2: Dashboard Web Interface**
> **Target: 5-7 hari**

#### 2.1 Teknologi Stack
```
Frontend: HTML + CSS + JavaScript (Vanilla/Alpine.js)
Backend: Go Gin/Echo framework
Database: SQLite (existing)
Authentication: Session-based dengan admin verification
```

#### 2.2 Dashboard Features
1. **Admin Authentication**
   - Login dengan nomor WhatsApp admin
   - Session management
   - Security middleware

2. **Group Management Page**
   - List semua grup yang diikuti bot
   - Enable/disable grup untuk bot
   - Set command permissions per grup

3. **Custom Command Management**
   - CRUD operations untuk commands
   - Preview response
   - Upload media files
   - Test command functionality

4. **Analytics Dashboard**
   - Command usage statistics
   - Group activity metrics
   - Error logs dan monitoring

#### 2.3 API Endpoints
```
POST /api/auth/login
GET  /api/auth/logout
GET  /api/groups
POST /api/groups/{id}/toggle
PUT  /api/groups/{id}/commands

GET  /api/commands
POST /api/commands
PUT  /api/commands/{id}
DELETE /api/commands/{id}
POST /api/commands/{id}/test

GET  /api/analytics/usage
GET  /api/analytics/groups
GET  /api/logs
```

### **FASE 3: Media Handler & Response System**
> **Target: 3-4 hari**

#### 3.1 Media Support
- **Image**: JPG, PNG, GIF, WebP
- **Video**: MP4, MOV, AVI (dengan size limit)
- **Audio**: MP3, WAV, OGG, M4A
- **Sticker**: WebP format sticker
- **Text**: Rich formatting dengan markdown

#### 3.2 Media Storage
```
media/
├── images/
├── videos/
├── audios/
└── stickers/
```

#### 3.3 Response Generator
- Dynamic content replacement (variables)
- Random response selection
- Context-aware responses

### **FASE 4: Advanced Features**
> **Target: 4-5 hari**

#### 4.1 Dynamic Variables
```
{USER_NAME} - Nama pengirim
{GROUP_NAME} - Nama grup
{TIME} - Waktu saat ini
{DATE} - Tanggal saat ini
{RANDOM_NUMBER} - Angka random
{ADMIN_CONTACT} - Kontak admin
```

#### 4.2 Command Categories
- **Educational**: Materi pembelajaran
- **Informational**: Info penting
- **Interactive**: Quiz, polling
- **Entertainment**: Games, jokes
- **Administrative**: Rules, announcements

#### 4.3 Scheduling System
- Schedule custom messages
- Recurring reminders
- Event-based triggers

---

## 📂 **STRUKTUR FILE BARU**

```
├── web/                       # Dashboard web
│   ├── static/
│   │   ├── css/
│   │   ├── js/
│   │   └── media/
│   ├── templates/
│   └── handlers/
├── services/
│   ├── custom_command.go      # Custom command service
│   ├── group_access.go        # Group access control
│   ├── media_handler.go       # Media processing
│   └── dashboard_api.go       # Dashboard API
├── models/
│   ├── custom_command.go      # Custom command models
│   └── group_access.go        # Group access models
└── middleware/
    ├── auth.go                # Authentication
    └── group_filter.go        # Group filtering
```

---

## 🔧 **IMPLEMENTASI DETAIL**

### **Step 1: Extend Database Models**
```go
// models/custom_command.go
type CustomCommand struct {
    ID              int       `json:"id" db:"id"`
    Command         string    `json:"command" db:"command"`
    Description     string    `json:"description" db:"description"`
    ResponseType    string    `json:"response_type" db:"response_type"`
    ResponseContent string    `json:"response_content" db:"response_content"`
    MediaURL        *string   `json:"media_url" db:"media_url"`
    IsActive        bool      `json:"is_active" db:"is_active"`
    CreatedBy       string    `json:"created_by" db:"created_by"`
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type AllowedGroup struct {
    ID              int       `json:"id" db:"id"`
    GroupJID        string    `json:"group_jid" db:"group_jid"`
    GroupName       string    `json:"group_name" db:"group_name"`
    IsActive        bool      `json:"is_active" db:"is_active"`
    AllowedCommands string    `json:"allowed_commands" db:"allowed_commands"` // JSON
    CreatedBy       string    `json:"created_by" db:"created_by"`
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
```

### **Step 2: Custom Command Handler**
```go
// services/custom_command.go
func (s *CustomCommandService) ProcessCommand(groupJID, command string) (*Response, error) {
    // 1. Check if group is allowed
    if !s.IsGroupAllowed(groupJID) {
        return nil, errors.New("group not allowed")
    }
    
    // 2. Get command from database
    cmd, err := s.GetCommand(command)
    if err != nil {
        return nil, err
    }
    
    // 3. Process response based on type
    response := s.GenerateResponse(cmd)
    
    // 4. Log usage
    s.LogCommandUsage(groupJID, command, response.Type)
    
    return response, nil
}
```

### **Step 3: Dashboard Web Handler**
```go
// web/handlers/dashboard.go
func (h *DashboardHandler) HandleCommandCreate(c *gin.Context) {
    var req CreateCommandRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // Validate admin access
    if !h.isAdmin(c) {
        c.JSON(403, gin.H{"error": "access denied"})
        return
    }
    
    // Create command
    command, err := h.commandService.CreateCommand(req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(201, command)
}
```

---

## ⚡ **KEUNGGULAN SOLUSI**

### ✅ **Fleksibilitas Tinggi**
- Admin bisa menambah/edit command tanpa coding
- Support berbagai format media
- Group access control yang ketat

### ✅ **Mudah Digunakan**
- Dashboard web yang user-friendly
- Preview response sebelum deploy
- Real-time testing

### ✅ **Scalable & Maintainable**
- Menggunakan arsitektur existing
- Database schema yang extensible
- Logging dan monitoring built-in

### ✅ **Secure**
- Authentication per admin
- Group whitelist system
- Command usage tracking

---

## 📊 **ESTIMASI WAKTU & RESOURCES**

| Fase | Durasi | Complexity | Priority |
|------|--------|------------|----------|
| Database & Backend | 3-5 hari | Medium | High |
| Web Dashboard | 5-7 hari | High | High |
| Media Handler | 3-4 hari | Medium | Medium |
| Advanced Features | 4-5 hari | High | Low |
| **TOTAL** | **15-21 hari** | **Medium-High** | - |

---

## 🎯 **APAKAH FEASIBLE?**

### ✅ **SANGAT MEMUNGKINKAN** karena:
1. **Foundation sudah kuat** - Bot WhatsApp, database, admin system sudah ada
2. **Architecture compatible** - Bisa extend existing code tanpa breaking changes
3. **Technology stack familiar** - Go, SQLite, HTML/CSS/JS
4. **Clear requirements** - Scope yang jelas dan terukur

### 🚀 **LANGKAH SELANJUTNYA:**
1. **Approval planning** ini dari Anda
2. **Setup development environment** untuk dashboard
3. **Start dengan Fase 1** - Custom Command System
4. **Iterative development** dengan testing per fase

---

## 💡 **CATATAN TAMBAHAN**

### **Considerations:**
- **Media file size limits** untuk performance
- **Storage management** untuk file media
- **Backup strategy** untuk custom commands
- **Error handling** yang comprehensive

### **Future Enhancements:**
- **AI integration** untuk smart responses
- **Multi-language support**
- **Advanced analytics** dan reporting
- **Mobile app** untuk dashboard

---

**🤔 Apakah planning ini sesuai dengan ekspektasi Anda? Perlu ada modifikasi atau tambahan fitur?**