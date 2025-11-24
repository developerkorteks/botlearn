# 🎯 BOT PEMBELAJARAN INJEC - REAL IMPLEMENTATION

## 📋 **CONTOH REAL YANG ANDA INGINKAN:**

### **1. Command Response Text**
```
Command: .listbugs
Response: 📝 TEXT MESSAGE
```
**Contoh Output:**
```
🐛 LIST BUG SERVER VPN (PEMBELAJARAN)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🇮🇩 SERVER INDONESIA:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔴 TELKOMSEL:
• Bug 1: 104.16.0.1:443
• Bug 2: 162.159.192.1:443
• Bug 3: cf.shopee.co.id:443

🔵 XL AXIATA:
• Bug 1: quiz.vidio.com:443
• Bug 2: cache.netflix.com:443
• Bug 3: *.googlevideo.com:443

🟢 INDOSAT:
• Bug 1: m.facebook.com:443
• Bug 2: api.whatsapp.com:443
• Bug 3: edge-chat.instagram.com:443

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️ UNTUK PEMBELAJARAN SAJA
🚫 GUNAKAN DENGAN BIJAK
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Last Update: 15 Jan 2024
```

### **2. Command Response Video**
```
Command: .websocketbolakbalik
Response: 🎥 VIDEO TUTORIAL
```
**Bot akan kirim:**
- Video file: `tutorial_websocket.mp4`
- Caption: "📚 Tutorial WebSocket Bolak-Balik untuk Pembelajaran"

### **3. Command Response File/APK**
```
Command: .bahaninject
Response: 📁 FILE APK/APLIKASI
```
**Bot akan kirim:**
- File: `injector_tool_v2.apk` (atau file zip tools)
- Caption: "🛠️ Tools Injector untuk Pembelajaran - Gunakan dengan Bijak!"

### **4. Auto Response kata tertentu (Candaan)**
```
User ketik: "cape"
Bot response: 🎵 VOICE NOTE "yah cape ya bang" + 😴 STICKER

User ketik: "gabut"
Bot response: 😂 STICKER lucu + 🎵 MUSIK "gabut nih"

User ketik: "semangat"
Bot response: 💪 STICKER semangat + 🎵 LAGU motivasi
```

---

## 🛠️ **IMPLEMENTASI TEKNIS**

### **Database Schema:**
```sql
-- Tabel command custom
CREATE TABLE learning_commands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command VARCHAR(100) UNIQUE NOT NULL,        -- '.listbugs'
    title VARCHAR(255),                          -- 'List Bug VPN'
    response_type VARCHAR(20) NOT NULL,          -- 'text', 'video', 'file', 'audio', 'sticker'
    text_content TEXT,                           -- Isi text untuk .listbugs
    file_path TEXT,                              -- Path file video/apk
    caption TEXT,                                -- Caption untuk media
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel auto response kata kunci
CREATE TABLE auto_responses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    keyword VARCHAR(100) NOT NULL,               -- 'cape', 'gabut', 'semangat'
    response_type VARCHAR(20) NOT NULL,          -- 'sticker', 'audio', 'mixed'
    sticker_path TEXT,                           -- Path sticker
    audio_path TEXT,                             -- Path voice note
    text_response TEXT,                          -- Text tambahan
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel grup yang diizinkan
CREATE TABLE allowed_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_jid VARCHAR(100) UNIQUE NOT NULL,
    group_name VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### **Message Handler Logic:**
```go
// handlers/learning_message.go
func (h *LearningMessageHandler) HandleMessage(evt *events.Message) {
    // 1. Cek apakah grup diizinkan
    if !h.isGroupAllowed(evt.Info.Chat.String()) {
        return // Bot DIAM jika grup tidak diizinkan
    }
    
    messageText := h.getMessageText(evt.Message)
    
    // 2. Cek command (.command)
    if strings.HasPrefix(messageText, ".") {
        h.handleCommand(evt, messageText)
        return
    }
    
    // 3. Cek kata kunci untuk auto response candaan
    h.handleAutoResponse(evt, messageText)
}

func (h *LearningMessageHandler) handleCommand(evt *events.Message, command string) {
    // Get command dari database
    cmd, err := h.repository.GetLearningCommand(command)
    if err != nil {
        return // Command tidak ditemukan
    }
    
    switch cmd.ResponseType {
    case "text":
        h.sendTextMessage(evt.Info.Chat, cmd.TextContent)
        
    case "video":
        h.sendVideoFile(evt.Info.Chat, cmd.FilePath, cmd.Caption)
        
    case "file":
        h.sendDocumentFile(evt.Info.Chat, cmd.FilePath, cmd.Caption)
        
    case "audio":
        h.sendAudioFile(evt.Info.Chat, cmd.FilePath)
        
    case "sticker":
        h.sendStickerFile(evt.Info.Chat, cmd.FilePath)
    }
}

func (h *LearningMessageHandler) handleAutoResponse(evt *events.Message, text string) {
    lowerText := strings.ToLower(text)
    
    // Cek kata kunci
    responses, err := h.repository.GetAutoResponsesByKeyword(lowerText)
    if err != nil || len(responses) == 0 {
        return
    }
    
    for _, response := range responses {
        // Kirim sticker jika ada
        if response.StickerPath != "" {
            h.sendStickerFile(evt.Info.Chat, response.StickerPath)
        }
        
        // Kirim audio jika ada
        if response.AudioPath != "" {
            h.sendAudioFile(evt.Info.Chat, response.AudioPath)
        }
        
        // Kirim text jika ada
        if response.TextResponse != "" {
            h.sendTextMessage(evt.Info.Chat, response.TextResponse)
        }
    }
}
```

### **Media Sender Implementation:**
```go
// Send Video Tutorial
func (h *LearningMessageHandler) sendVideoFile(chatJID types.JID, videoPath, caption string) {
    videoData, err := os.ReadFile(videoPath)
    if err != nil {
        h.logger.Errorf("Failed to read video: %v", err)
        return
    }
    
    uploaded, err := h.client.Upload(context.Background(), videoData, whatsmeow.MediaVideo)
    if err != nil {
        h.logger.Errorf("Failed to upload video: %v", err)
        return
    }
    
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
    h.logger.Infof("Video sent: %s", videoPath)
}

// Send APK/File
func (h *LearningMessageHandler) sendDocumentFile(chatJID types.JID, filePath, caption string) {
    fileData, err := os.ReadFile(filePath)
    if err != nil {
        h.logger.Errorf("Failed to read file: %v", err)
        return
    }
    
    fileName := filepath.Base(filePath)
    uploaded, err := h.client.Upload(context.Background(), fileData, whatsmeow.MediaDocument)
    if err != nil {
        h.logger.Errorf("Failed to upload file: %v", err)
        return
    }
    
    msg := &waProto.Message{
        DocumentMessage: &waProto.DocumentMessage{
            Caption:       &caption,
            FileName:      &fileName,
            Url:           &uploaded.URL,
            DirectPath:    &uploaded.DirectPath,
            MediaKey:      uploaded.MediaKey,
            FileEncSha256: uploaded.FileEncSHA256,
            FileSha256:    uploaded.FileSHA256,
            FileLength:    &uploaded.FileLength,
        },
    }
    
    h.client.SendMessage(context.Background(), chatJID, msg)
    h.logger.Infof("File sent: %s", filePath)
}

// Send Voice Note/Music
func (h *LearningMessageHandler) sendAudioFile(chatJID types.JID, audioPath string) {
    audioData, err := os.ReadFile(audioPath)
    if err != nil {
        h.logger.Errorf("Failed to read audio: %v", err)
        return
    }
    
    uploaded, err := h.client.Upload(context.Background(), audioData, whatsmeow.MediaAudio)
    if err != nil {
        h.logger.Errorf("Failed to upload audio: %v", err)
        return
    }
    
    msg := &waProto.Message{
        AudioMessage: &waProto.AudioMessage{
            Url:           &uploaded.URL,
            DirectPath:    &uploaded.DirectPath,
            MediaKey:      uploaded.MediaKey,
            FileEncSha256: uploaded.FileEncSHA256,
            FileSha256:    uploaded.FileSHA256,
            FileLength:    &uploaded.FileLength,
            Ptt:           proto.Bool(true), // Voice note
        },
    }
    
    h.client.SendMessage(context.Background(), chatJID, msg)
    h.logger.Infof("Audio sent: %s", audioPath)
}
```

---

## 🌐 **DASHBOARD WEB INTERFACE**

### **1. Command Management Page:**
```html
<div class="command-manager">
    <h2>🛠️ Kelola Command Pembelajaran</h2>
    
    <!-- Form Tambah Command -->
    <form class="add-command-form">
        <div class="form-group">
            <label>Command:</label>
            <input type="text" placeholder=".listbugs" name="command" required>
        </div>
        
        <div class="form-group">
            <label>Response Type:</label>
            <select name="response_type" onchange="toggleInputs(this.value)">
                <option value="text">📝 Text Message</option>
                <option value="video">🎥 Video File</option>
                <option value="file">📁 Document/APK</option>
                <option value="audio">🎵 Audio/Voice</option>
                <option value="sticker">😄 Sticker</option>
            </select>
        </div>
        
        <!-- Text Content (untuk .listbugs) -->
        <div id="text-input" class="form-group">
            <label>Text Content:</label>
            <textarea name="text_content" rows="10" 
                placeholder="🐛 LIST BUG SERVER VPN..."></textarea>
        </div>
        
        <!-- File Upload (untuk video/apk/audio) -->
        <div id="file-input" class="form-group" style="display:none">
            <label>Upload File:</label>
            <input type="file" name="media_file" 
                accept="video/*,audio/*,.apk,.zip,.pdf">
            <small>Max 50MB</small>
        </div>
        
        <div class="form-group">
            <label>Caption:</label>
            <input type="text" name="caption" 
                placeholder="📚 Tutorial WebSocket untuk Pembelajaran">
        </div>
        
        <button type="submit">💾 Simpan Command</button>
    </form>
    
    <!-- List Command yang Ada -->
    <div class="commands-list">
        <div class="command-item">
            <span class="command">.listbugs</span>
            <span class="type">📝 Text</span>
            <span class="status">✅ Active</span>
            <button class="btn-edit">✏️ Edit</button>
            <button class="btn-delete">🗑️ Delete</button>
        </div>
        
        <div class="command-item">
            <span class="command">.websocketbolakbalik</span>
            <span class="type">🎥 Video</span>
            <span class="status">✅ Active</span>
            <button class="btn-edit">✏️ Edit</button>
            <button class="btn-delete">🗑️ Delete</button>
        </div>
    </div>
</div>
```

### **2. Auto Response Management:**
```html
<div class="auto-response-manager">
    <h2>🤖 Auto Response Candaan</h2>
    
    <form class="add-auto-response-form">
        <div class="form-group">
            <label>Kata Kunci:</label>
            <input type="text" placeholder="cape" name="keyword" required>
        </div>
        
        <div class="form-group">
            <label>Response Type:</label>
            <select name="response_type">
                <option value="sticker">😄 Sticker</option>
                <option value="audio">🎵 Voice Note</option>
                <option value="mixed">🎭 Sticker + Audio</option>
            </select>
        </div>
        
        <div class="form-group">
            <label>Upload Sticker (.webp):</label>
            <input type="file" name="sticker_file" accept=".webp">
        </div>
        
        <div class="form-group">
            <label>Upload Voice Note:</label>
            <input type="file" name="audio_file" accept="audio/*">
        </div>
        
        <button type="submit">💾 Simpan Auto Response</button>
    </form>
</div>
```

---

## 📁 **FILE STRUCTURE**

```
media/
├── commands/
│   ├── videos/
│   │   └── websocket_tutorial.mp4
│   ├── files/
│   │   ├── injector_tool_v2.apk
│   │   └── bug_list_tools.zip
│   └── audios/
│       └── tutorial_audio.mp3
├── auto_responses/
│   ├── stickers/
│   │   ├── cape_sticker.webp
│   │   ├── gabut_sticker.webp
│   │   └── semangat_sticker.webp
│   └── voices/
│       ├── yah_cape_ya_bang.mp3
│       ├── gabut_nih.mp3
│       └── semangat_terus.mp3
└── temp/
    └── uploads/
```

---

## 🎯 **REAL USAGE EXAMPLE**

### **Setup via Dashboard:**
1. Admin login ke `http://localhost:8080/dashboard`
2. Tambah command `.listbugs` → type "text" → isi content bug list
3. Tambah command `.websocketbolakbalik` → type "video" → upload video tutorial
4. Tambah command `.bahaninject` → type "file" → upload APK tools
5. Tambah auto response "cape" → upload sticker + voice note

### **Usage di Grup:**
```
User: .listbugs
Bot: [Kirim text list bug VPN lengkap]

User: .websocketbolakbalik
Bot: [Kirim video tutorial websocket]

User: .bahaninject  
Bot: [Kirim file APK injector tools]

User: wah cape banget nih
Bot: [Kirim sticker cape + voice "yah cape ya bang"]

User di grup lain (tidak diaktifkan): .listbugs
Bot: [DIAM TOTAL - tidak ada response]
```

---

## ✅ **KESIMPULAN**

**SANGAT BISA DILAKUKAN!** 🚀

- ✅ Command custom (.listbugs, .websocketbolakbalik, .bahaninject)
- ✅ Multi-format response (text, video, file, audio, sticker)  
- ✅ Auto response kata kunci untuk candaan
- ✅ Dashboard web untuk manage semua
- ✅ Grup whitelist system
- ✅ File upload & management

**Estimasi: 10-14 hari**

**Apakah implementasi ini sesuai dengan yang Anda maksud?** 🤔