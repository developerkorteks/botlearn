# 📊 Fitur Cek Kuota XL/AXIS

## 📝 Overview

Fitur **Cek Kuota** memungkinkan bot untuk mengecek informasi kuota dan paket dari nomor XL/AXIS menggunakan API eksternal. Fitur ini hanya bisa digunakan di grup yang sudah di-whitelist.

---

## 🎯 Cara Penggunaan

### Command Format:
```
.cekkuota <nomor>
```

### Contoh Penggunaan:
```
.cekkuota 087817739901
.cekkuota 6287817739901
```

### Format Nomor yang Didukung:
- ✅ Format `08xxx` → `087817739901`
- ✅ Format `628xxx` → `6287817739901`
- ✅ Format dengan spasi → `0878 1773 9901`
- ✅ Format dengan tanda hubung → `0878-1773-9901`
- ✅ Format dengan `+` → `+6287817739901`

### Catatan Penting:
⚠️ **API hanya support nomor XL/AXIS**. Nomor operator lain (Telkomsel, Indosat, dll) akan mengembalikan error.

---

## 🔒 Aturan Akses

### ✅ Bisa Digunakan:
- **Grup yang sudah di-whitelist** (`.addgroup`)
- Semua member di grup yang diizinkan

### ❌ Tidak Bisa Digunakan:
- Grup yang belum di-whitelist
- Chat personal/private
- Nomor bot sendiri (IsFromMe)

---

## 📊 Output/Response

Bot akan mengirimkan informasi lengkap dengan format yang rapi:

### 1️⃣ Informasi Kartu
- Nomor HP (MSISDN)
- Provider (XL/AXIS)
- Tipe Jaringan (2G/3G/4G/5G)
- Lama Berlangganan (Tenure)
- Masa Aktif
- Grace Period

### 2️⃣ Status VoLTE
- Device support
- Area support
- Simcard support

### 3️⃣ Paket Aktif
Untuk setiap paket yang aktif, ditampilkan:
- Nama paket
- Tanggal expired
- Detail kuota per jenis:
  - Nama kuota
  - Total kuota
  - Sisa kuota
  - Persentase terpakai

### 4️⃣ Emoji & Formatting
Output menggunakan emoji untuk membedakan jenis kuota:
- 📶 → Kuota data umum
- 📞 → Kuota nelpon/call
- 💬 → Kuota SMS
- 🚗 → Kuota aplikasi (Gojek, Grab, Waze)

---

## 🔧 Implementasi Teknis

### File yang Dibuat/Dimodifikasi:

#### 1. `services/quota_checker.go` (NEW)
Service baru untuk handle cek kuota:
```go
type QuotaChecker struct {
    client *http.Client
}

func NewQuotaChecker() *QuotaChecker
func (qc *QuotaChecker) CheckQuota(phoneNumber string) (string, error)
func (qc *QuotaChecker) NormalizePhoneNumber(number string) string
func (qc *QuotaChecker) FormatQuotaResponse(resp *QuotaResponse) string
```

**Features:**
- HTTP client dengan timeout 30 detik
- Normalisasi format nomor otomatis
- Header mimic browser untuk bypass restrictions
- Error handling yang comprehensive
- Output formatting dengan emoji dan box drawing

#### 2. `handlers/learning_message.go` (MODIFIED)
Penambahan handler untuk command `.cekkuota`:

**Perubahan:**
- Tambah field `quotaChecker *services.QuotaChecker` di struct
- Initialize QuotaChecker di constructor
- Intercept command `.cekkuota` sebelum IsFromMe check
- Validasi grup whitelist
- Handler `handleCekKuotaCommand()` untuk proses request

**Flow:**
```
User input: .cekkuota 087817739901
    ↓
Check: Apakah di grup yang diizinkan?
    ↓ YES
Send: "⏳ Sedang mengecek kuota..."
    ↓
Call: quotaChecker.CheckQuota()
    ↓
API: https://bendith.my.id/end.php
    ↓
Parse & Format response
    ↓
Send: Formatted quota info
```

---

## 🌐 API Integration

### Endpoint:
```
GET https://bendith.my.id/end.php?check=package&number={number}&version=2
```

### Request Headers:
```
Accept: */*
Accept-Language: en-US,en;q=0.9
Referer: https://bendith.my.id/
User-Agent: Mozilla/5.0 (X11; Linux x86_64) ...
Sec-Fetch-Dest: empty
Sec-Fetch-Mode: cors
Sec-Fetch-Site: same-origin
```

### Response:
HTTP Status: **201 Created** (bukan 200 OK!)

```json
{
  "success": true,
  "code": "000",
  "message": "",
  "data": {
    "subs_info": {
      "msisdn": "6287817739901",
      "operator": "XL",
      "net_type": "4G",
      "exp_date": "05-12-2025",
      ...
    },
    "package_info": {
      "packages": [
        {
          "name": "Bundling 45GB Setahun",
          "expiry": "22-12-2025",
          "quotas": [...]
        }
      ]
    }
  }
}
```

### Error Response:
```json
{
  "success": false,
  "code": "999",
  "message": "Silakan masukkan nomor XL/AXIS",
  "data": null
}
```

---

## 🧪 Testing

Fitur telah ditest dengan berbagai skenario:

### ✅ Test Cases Passed:
1. Format `08xxx` → ✅ Success
2. Format `628xxx` → ✅ Success  
3. Format dengan spasi → ✅ Success
4. Format dengan dash → ✅ Success
5. Nomor non-XL → ✅ Error handling correct
6. Invalid format → ✅ Error handling correct
7. Normalisasi nomor → ✅ All formats converted correctly

---

## 💡 Usage Example

### Scenario: Member grup mau cek kuota XL

```
User: .cekkuota 087817739901
Bot: ⏳ Sedang mengecek kuota...
     📱 Nomor: 087817739901
     Mohon tunggu sebentar...

Bot: ╔═══════════════════════════╗
     ║   📊 INFO KUOTA & PAKET   ║
     ╚═══════════════════════════╝
     
     📱 *Info Kartu:*
     ├─ Nomor: 6287817739901
     ├─ Provider: XL
     ├─ Jaringan: 4G
     ├─ Lama Berlangganan: 2 Tahun 0 Bulan
     ├─ Masa Aktif: 05-12-2025
     └─ Grace Period: 04-01-2026
     
     🔊 *VoLTE Status:*
     ├─ Device: ❌
     ├─ Area: ✅
     └─ Simcard: ✅
     
     📦 *Paket Aktif:*
     ━━━━━━━━━━━━━━━━━━━━━━━━━━━
     
     🎁 *Paket 1: Bundling 45GB Setahun*
     └─ Berlaku hingga: 22-12-2025
     
     📊 *Detail Kuota:*
     ├─ 📶 *Kuota Bonus WhatsApp*
        ├─ Total: 2.8GB
        ├─ Sisa: 2.8GB
        └─ Terpakai: 100.0%
     
     └─ 📶 *Kuota Utama*
        ├─ Total: 1GB
        ├─ Sisa: 1GB
        └─ Terpakai: 100.0%
     
     ━━━━━━━━━━━━━━━━━━━━━━━━━━━
     ✅ *Data berhasil diambil*
     🕐 Waktu: 24-11-2025 16:06:18
```

---

## 🔐 Security & Rate Limiting

### Rate Limiting:
- Command cooldown: **3 detik per user** (shared dengan command lain)
- Mencegah spam dan abuse API

### Security:
- ✅ Whitelist grup enforcement
- ✅ Input sanitization (normalisasi nomor)
- ✅ HTTP timeout (30 detik)
- ✅ Error handling untuk API failures
- ✅ Tidak store nomor atau data sensitif

---

## 🚀 Deployment

Fitur sudah terintegrasi dengan sistem bot, tidak perlu konfigurasi tambahan.

### Build:
```bash
go build -o bot cmd/main.go
```

### Run:
```bash
./bot
```

Bot akan otomatis support command `.cekkuota` di grup yang sudah di-whitelist.

---

## 🐛 Error Handling

### Error Messages:

#### 1. Format Nomor Salah
```
❌ Format salah!

📝 Contoh penggunaan:
• .cekkuota 081234567890
• .cekkuota 6281234567890
```

#### 2. Nomor Bukan XL/AXIS
```
❌ Gagal cek kuota!

📱 Nomor: 081234567890
⚠️ Error: API error: 999 - Silakan masukkan nomor XL/AXIS

💡 Tips:
• Pastikan format nomor benar (08xxx atau 628xxx)
• Coba lagi beberapa saat
```

#### 3. API Timeout/Error
```
❌ Gagal cek kuota!

📱 Nomor: 087817739901
⚠️ Error: Gagal mengakses API: context deadline exceeded

💡 Tips:
• Pastikan format nomor benar (08xxx atau 628xxx)
• Coba lagi beberapa saat
```

---

## 📌 Future Improvements

Potensi enhancement untuk versi berikutnya:

1. **Support Multi-Provider**
   - Tambah API untuk Telkomsel, Indosat, Tri, dll
   - Auto-detect provider dari prefix nomor

2. **Caching**
   - Cache hasil cek kuota untuk 5-10 menit
   - Reduce load ke API eksternal

3. **Statistics**
   - Track usage per user/group
   - Popular paket/provider analytics

4. **Notification**
   - Alert saat kuota hampir habis
   - Reminder sebelum masa aktif expired

5. **Dashboard Integration**
   - View cek kuota history di web dashboard
   - Export reports

---

## 📞 Support

Jika ada pertanyaan atau issues:
- Check logs untuk error details
- Dashboard: `http://localhost:1462`
- Command help: `.help`

---

**Created:** 2025-11-24  
**Version:** 1.0.0  
**Status:** ✅ Production Ready
