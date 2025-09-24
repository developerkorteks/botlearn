# 📚 PLANNING: Bot Pembelajaran/Injec untuk Grup Tertentu

## 🎯 **KEBUTUHAN YANG DIMINTA:**

1. ✅ **Bot untuk grup tertentu** - khusus kebutuhan injec/pembelajaran/belajar/informasi
2. ✅ **Command dan response dinamis** - admin bisa atur via dashboard website
3. ✅ **Hanya grup yang di-set admin** - kalau tidak di-set, bot tidak bisa digunakan
4. ✅ **Multi-format response** - text, pesan, sticker, video, suara/musik
5. ✅ **Dashboard web only** - semua pengaturan via website

---

## 🏗️ **ANALISIS STRUKTUR EXISTING**

### ✅ **SUDAH ADA & BISA DIPAKAI:**
```
├── WhatsApp Bot (whatsmeow) ✅
├── Admin Authentication ✅  
├── Group Management ✅
├── Database SQLite ✅
├── Message Handler ✅
├── Media Support ✅
└── Template System ✅ (bisa dimodifikasi)
```

### 🔄 **YANG PERLU DIMODIFIKASI:**
1. **Group Access Control** - bot hanya respon di grup yang diizinkan
2. **Custom Command System** - replace auto-promote dengan learning commands
3. **Dashboard Web** - interface untuk admin kelola semua
4. **Media Response Handler** - support text, sticker, video, audio

---

## 🚀 **IMPLEMENTASI PLAN**

### **FASE 1: Group Access Control System**
> **Target: 2-3 hari**

#### Database Schema Baru:
```sql
-- Grup yang diizinkan untuk bot pembelajaran
CREATE TABLE learning_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_jid VARCHAR(100) UNIQUE NOT NULL,
    group_name VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    description TEXT,
    created_by VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Command pembelajaran custom
CREATE TABLE learning_commands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command VARCHAR(100) NOT NULL,
    title VARCHAR(255),
    description TEXT,
    response_type ENUM('text', 'image', 'video', 'audio', 'sticker') DEFAULT 'text',
    response_content TEXT NOT NULL,
    media_file_path TEXT,
    category VARCHAR(100), -- 'injec', 'pembelajaran', 'informasi', dll
    is_active BOOLEAN DEFAULT true,
    usage_count INTEGER DEFAULT 0,
    created_by VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### Group Filter Logic:
```go
// services/learning_access.go
func (s *LearningAccessService) IsGroupAllowed(groupJID string) bool {
    group, err := s.repository.GetLearningGroup(groupJID)
    if err != nil || group == nil {
        return false // Grup tidak terdaftar = tidak bisa pakai bot
    }
    return group.IsActive
}

func (s *MessageHandler) HandleMessage(evt *events.Message) {
    // Cek apakah ini grup
    isGroup := evt.Info.Chat.Server == types.GroupServer
    
    if isGroup {
        // HANYA RESPON DI GRUP YANG DIIZINKAN
        if !s.learningService.IsGroupAllowed(evt.Info.Chat.String()) {
            // Bot DIAM TOTAL - tidak ada response apapun
            return
        }
    }
    
    // Lanjut proses command jika grup diizinkan
    s.processLearningCommand(evt)
}
```

### **FASE 2: Custom Learning Command System**
> **Target: 3-4 hari**

#### Command Handler:
```go
// handlers/learning_commands.go
func (h *LearningCommandHandler) ProcessCommand(evt *events.Message, command string) {
    // Get command dari database
    cmd, err := h.repository.GetLearningCommand(command)
    if err != nil {
        return // Command tidak ditemukan, bot diam
    }
    
    // Generate response berdasarkan tipe
    switch cmd.ResponseType {
    case "text":
        h.sendTextResponse(evt.Info.Chat, cmd.ResponseContent)
    case "image":
        h.sendImageResponse(evt.Info.Chat, cmd.ResponseContent, cmd.MediaFilePath)
    case "video":
        h.sendVideoResponse(evt.Info.Chat, cmd.ResponseContent, cmd.MediaFilePath)
    case "audio":
        h.sendAudioResponse(evt.Info.Chat, cmd.ResponseContent, cmd.MediaFilePath)
    case "sticker":
        h.sendStickerResponse(evt.Info.Chat, cmd.MediaFilePath)
    }
    
    // Update usage count
    h.repository.IncrementCommandUsage(cmd.ID)
}
```

#### Media Response Implementation:
```go
// Text Response
func (h *LearningCommandHandler) sendTextResponse(chatJID types.JID, content string) {
    msg := &waProto.Message{
        Conversation: &content,
    }
    h.client.SendMessage(context.Background(), chatJID, msg)
}

// Image Response  
func (h *LearningCommandHandler) sendImageResponse(chatJID types.JID, caption, imagePath string) {
    imageData, _ := os.ReadFile(imagePath)
    uploaded, _ := h.client.Upload(context.Background(), imageData, whatsmeow.MediaImage)
    
    msg := &waProto.Message{
        ImageMessage: &waProto.ImageMessage{
            Caption:       &caption,
            Url:           &uploaded.URL,
            DirectPath:    &uploaded.DirectPath,
            MediaKey:      uploaded.MediaKey,
            FileEncSha256: uploaded.FileEncSHA256,
            FileSha256:    uploaded.FileSHA256,
            FileLength:    &uploaded.FileLength,
        },
    }
    h.client.SendMessage(context.Background(), chatJID, msg)
}

// Video Response
func (h *LearningCommandHandler) sendVideoResponse(chatJID types.JID, caption, videoPath string) {
    videoData, _ := os.ReadFile(videoPath)
    uploaded, _ := h.client.Upload(context.Background(), videoData, whatsmeow.MediaVideo)
    
    msg := &waProto.Message{
        VideoMessage: &waProto.VideoMessage{
            Caption:       &caption,
            Url:           &uploaded.URL,
            DirectPath:    &uploaded.DirectPath,
            MediaKey:      uploaded.MediaKey,
            FileEncSha256: uploaded.FileEncSHA256,
            FileSha256:    uploaded.FileSHA256,
            FileLength:    &uploaded.FileLength,
        },
    }
    h.client.SendMessage(context.Background(), chatJID, msg)
}

// Audio Response
func (h *LearningCommandHandler) sendAudioResponse(chatJID types.JID, caption, audioPath string) {
    audioData, _ := os.ReadFile(audioPath)
    uploaded, _ := h.client.Upload(context.Background(), audioData, whatsmeow.MediaAudio)
    
    msg := &waProto.Message{
        AudioMessage: &waProto.AudioMessage{
            Url:           &uploaded.URL,
            DirectPath:    &uploaded.DirectPath,
            MediaKey:      uploaded.MediaKey,
            FileEncSha256: uploaded.FileEncSHA256,
            FileSha256:    uploaded.FileSHA256,
            FileLength:    &uploaded.FileLength,
        },
    }
    h.client.SendMessage(context.Background(), chatJID, msg)
}

// Sticker Response
func (h *LearningCommandHandler) sendStickerResponse(chatJID types.JID, stickerPath string) {
    stickerData, _ := os.ReadFile(stickerPath)
    uploaded, _ := h.client.Upload(context.Background(), stickerData, whatsmeow.MediaImage)
    
    msg := &waProto.Message{
        StickerMessage: &waProto.StickerMessage{
            Url:           &uploaded.URL,
            DirectPath:    &uploaded.DirectPath,
            MediaKey:      uploaded.MediaKey,
            FileEncSha256: uploaded.FileEncSHA256,
            FileSha256:    uploaded.FileSHA256,
            FileLength:    &uploaded.FileLength,
        },
    }
    h.client.SendMessage(context.Background(), chatJID, msg)
}
```

### **FASE 3: Dashboard Web Interface**
> **Target: 4-5 hari**

#### Tech Stack:
```
Frontend: HTML + CSS + JavaScript (Alpine.js)
Backend: Go Gin Framework
Upload: Multipart file upload
Storage: local filesystem + database
```

#### Dashboard Pages:

1. **Admin Login** (`/login`)
   - Autentikasi nomor admin
   - Session management

2. **Group Management** (`/groups`)
   ```html
   <div class="groups-page">
     <h2>📱 Kelola Grup Pembelajaran</h2>
     
     <!-- List semua grup yang diikuti bot -->
     <div class="groups-list">
       <div class="group-item">
         <span>👥 Grup Belajar Coding</span>
         <button class="btn-enable">✅ Aktifkan</button>
       </div>
       <div class="group-item disabled">
         <span>👥 Grup Random Chat</span>
         <button class="btn-disable">❌ Nonaktif</button>
       </div>
     </div>
   </div>
   ```

3. **Command Management** (`/commands`)
   ```html
   <div class="commands-page">
     <h2>💬 Kelola Command Pembelajaran</h2>
     
     <!-- Form tambah command baru -->
     <form class="add-command-form">
       <input type="text" placeholder="Command (misal: /html)" required>
       <input type="text" placeholder="Judul pembelajaran">
       <select name="category">
         <option value="injec">💉 Injec</option>
         <option value="pembelajaran">📚 Pembelajaran</option>
         <option value="informasi">ℹ️ Informasi</option>
       </select>
       <select name="response_type">
         <option value="text">📝 Text</option>
         <option value="image">🖼️ Gambar</option>
         <option value="video">🎥 Video</option>
         <option value="audio">🎵 Audio</option>
         <option value="sticker">😄 Sticker</option>
       </select>
       
       <!-- Text content -->
       <textarea placeholder="Isi response..."></textarea>
       
       <!-- File upload untuk media -->
       <input type="file" accept="image/*,video/*,audio/*">
       
       <button type="submit">💾 Simpan Command</button>
     </form>
     
     <!-- List command yang sudah ada -->
     <div class="commands-list">
       <div class="command-item">
         <span>/html</span>
         <span>📚 HTML Dasar</span>
         <span>📝 Text</span>
         <button class="btn-edit">✏️</button>
         <button class="btn-delete">🗑️</button>
       </div>
     </div>
   </div>
   ```

4. **Preview & Test** (`/test`)
   - Test command di grup tertentu
   - Preview response sebelum deploy

#### API Endpoints:
```go
// Web server routes
r.POST("/api/login", dashboardHandler.Login)
r.GET("/api/groups", dashboardHandler.GetGroups)
r.POST("/api/groups/:id/toggle", dashboardHandler.ToggleGroup)

r.GET("/api/commands", dashboardHandler.GetCommands)
r.POST("/api/commands", dashboardHandler.CreateCommand)
r.PUT("/api/commands/:id", dashboardHandler.UpdateCommand)
r.DELETE("/api/commands/:id", dashboardHandler.DeleteCommand)
r.POST("/api/commands/test", dashboardHandler.TestCommand)

r.POST("/api/upload", dashboardHandler.UploadMedia)
```

### **FASE 4: Integration & Testing**
> **Target: 2-3 hari**

#### File Structure Baru:
```
├── web/                       # Dashboard web
│   ├── static/
│   │   ├── css/dashboard.css
│   │   ├── js/dashboard.js
│   │   └── uploads/           # Media files
│   │       ├── images/
│   │       ├── videos/
│   │       ├── audios/
│   │       └── stickers/
│   ├── templates/
│   │   ├── login.html
│   │   ├── groups.html
│   │   ├── commands.html
│   │   └── layout.html
│   └── handlers/
│       └── dashboard.go
├── services/
│   ├── learning_access.go     # Group access control
│   ├── learning_command.go    # Custom command processing
│   └── media_handler.go       # Media file handling
└── handlers/
    └── learning_message.go    # Learning message handler
```

---

## 🎯 **CONTOH PENGGUNAAN**

### **Scenario 1: Admin Setup**
1. Admin login ke dashboard `http://localhost:8080/dashboard`
2. Pilih grup "Belajar Programming" → Aktifkan
3. Tambah command `/html` dengan response text tentang HTML dasar
4. Tambah command `/video-css` dengan upload video tutorial CSS
5. Bot siap digunakan di grup tersebut

### **Scenario 2: User di Grup**
```
User: /html
Bot: 📚 HTML DASAR

HTML (HyperText Markup Language) adalah...
[penjelasan lengkap]

Contoh:
<html>
  <head><title>Hello</title></head>
  <body><h1>Hello World!</h1></body>
</html>

User: /video-css  
Bot: [Kirim video tutorial CSS]

User di grup lain (tidak diizinkan): /html
Bot: [DIAM - tidak ada response]
```

---

## ✅ **KESIMPULAN**

### **SANGAT BISA DILAKUKAN!** 🚀

**Alasan:**
1. ✅ **Foundation kuat** - Bot WhatsApp sudah jalan
2. ✅ **Database ready** - SQLite bisa extend dengan mudah  
3. ✅ **Admin system ada** - tinggal integrate ke dashboard
4. ✅ **Media support** - whatsmeow support semua format
5. ✅ **Group management** - sudah ada struktur dasarnya

**Estimasi Total: 11-15 hari**

**Yang Dihasilkan:**
- ✅ Bot hanya aktif di grup yang di-set admin
- ✅ Dashboard web untuk kelola command dan response
- ✅ Support text, gambar, video, audio, sticker
- ✅ Admin bisa tambah/edit/hapus command secara dinamis
- ✅ Bot diam total di grup yang tidak diizinkan

---

## 🚀 **SIAP EKSEKUSI?**

**Langkah Selanjutnya:**
1. **Konfirmasi planning** ini sesuai kebutuhan?
2. **Mulai Fase 1** - Group Access Control
3. **Development iteratif** dengan testing per fase
4. **Deploy dashboard** untuk admin testing

**Apakah planning ini sudah sesuai dengan yang Anda inginkan?** 🤔