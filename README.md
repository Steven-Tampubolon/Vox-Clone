# VoxClone — AI Voice Cloning Backend

VoxClone adalah backend REST API berbasis Go yang memungkinkan pengguna melakukan **kloning suara** dan **konversi teks ke audio** menggunakan model AI [XTTS v2](https://github.com/coqui-ai/TTS) yang berjalan di Google Colab.

---

## Tech Stack

- **Backend** — Go + Gin Framework
- **Database** — MySQL + GORM
- **AI Model** — XTTS v2 (Coqui TTS)
- **Infrastruktur AI** — Google Colab (GPU T4 Free Tier)
- **Tunnel** — Localtunnel

---

## Arsitektur Sistem

```
┌─────────────┐         ┌──────────────────────────┐
│  Dashboard  │         │       Google Colab        │
│  (Browser)  │         │                           │
└──────┬──────┘         │  Port 5002                │
       │                │  XTTS v2 Server (GPU T4)  │
       │ HTTP           │                           │
       ▼                │  Port 5003                │
┌──────────────┐  HTTP  │  Upload Server            │
│  Backend Go  │◄──────►│                           │
│  Port 8080   │        │  /tmp/xtts_speakers/      │
└──────┬───────┘        │       │ (symlink)         │
       │                │       ▼                   │
       ▼                │  Google Drive             │
┌──────────────┐        │  /XTTS_Speakers/          │
│   Database   │        └──────────────────────────┘
│   MySQL      │               ▲
└──────────────┘          Localtunnel
                          (URL publik)
```

---

## Flow Bisnis

### 1. Clone Suara
```
User upload file audio (.wav) + nama suara
        ▼
POST /api/voices/clone
        ▼
File disimpan lokal di uploads/
        ▼
File diupload ke Google Colab (port 5003)
        ▼
File tersimpan permanen di Google Drive/XTTS_Speakers/
        ▼
voice_id disimpan ke database
        ▼
Response: data voice berhasil didaftarkan
```

### 2. Generate TTS
```
User kirim teks + voice_id
        ▼
POST /api/voices/tts
        ▼
Backend Go kirim request ke XTTS v2 (port 5002)
{text, speaker_wav, language: "en"}
        ▼
XTTS generate audio menggunakan karakteristik suara
dari file di Google Drive/XTTS_Speakers/
        ▼
Audio binary dikembalikan ke browser
```

### 3. Get All Voices
```
GET /api/voices
        ▼
Ambil semua data voice dari database
        ▼
Response: list semua voice yang terdaftar
```

---

## Prerequisites

- [Go](https://golang.org/) 1.21+
- [MySQL](https://www.mysql.com/) 8.0+
- Google Account (untuk Google Colab & Google Drive)
- File `.env` (lihat bagian konfigurasi)

---

## Instalasi & Setup

### 1. Clone Repository
```bash
git clone https://github.com/Steven-Tampubolon/Vox-Clone.git
cd Vox-Vlone
```

### 2. Install Dependencies
```bash
go mod tidy
```

### 3. Setup Database
Buat database MySQL:
```sql
CREATE DATABASE voxclone_db;
```

### 4. Konfigurasi `.env`
Buat file `.env` atau gunakan `.env.example` di root project:
```env
# App
PORT=8080

# Database
DB_USER=root
DB_PASS=password
DB_HOST=localhost
DB_PORT=3306
DB_NAME=voxclone_db

# XTTS Server (diupdate setiap sesi Colab)
XTTS_SERVER_URL=https://xxx.loca.lt
XTTS_UPLOAD_URL=https://yyy.loca.lt
```

### 5. Jalankan Backend
```bash
go run cmd/api/main.go
```

---

## Setup Google Colab (Wajib)

Backend ini membutuhkan server XTTS v2 yang berjalan di [Google Colab](https://colab.research.google.com). Semua script sudah tersedia di file `vox-clone-colab.ipynb` — upload file tersebut ke Google Colab.

### Urutan Menjalankan (Setiap Sesi)

| Langkah | Cell | Keterangan |
|---------|------|------------|
| 1 | **Cell 1** | Cek GPU + Mount Drive |
| 2 | **Cell 2** | Install dependencies + Patch → lalu **Restart session** |
| 3 | **Cell 1** | Jalankan lagi setelah restart |
| 4 | **Cell 3** | Jalankan server XTTS |

### Setelah Cell 3 Selesai
1. Copy kedua URL yang muncul di output
2. Update file `.env` Go kamu:
   ```
   XTTS_SERVER_URL=https://xxx.loca.lt
   XTTS_UPLOAD_URL=https://yyy.loca.lt
   ```
3. Jalankan ulang backend Go

### Catatan Penting
- GPU T4 Colab free tier memiliki **limit harian** — hemat penggunaan
- Session Colab otomatis mati **~90 menit idle** — jalankan cell Keep Alive setelah Cell 3
- Model XTTS v2 (~1.9GB) tersimpan di **Google Drive** agar tidak download ulang
- File speaker tersimpan permanen di `Google Drive/XTTS_Speakers/`
- URL localtunnel **berubah setiap sesi** — selalu update `.env` setelah Cell 3

### Tips Hemat GPU
1. **Jangan restart session** kalau tidak perlu — gunakan Cell Debug 5 untuk kill port saja
2. **Selesaikan debug dulu** sebelum load model
3. **Tutup session** kalau sedang tidak dipakai, jangan biarkan jalan terus
4. **Pakai CPU** saat hanya testing kode Go, bukan testing kualitas suara

---

## API Documentation

### Clone Suara
```
POST /api/voices/clone
Content-Type: multipart/form-data

Form fields:
  name  (string, required) — nama untuk suara ini
  audio (file, required)   — file audio sampel (.wav, 15-30 detik)

Response 201:
{
  "message": "Kloning suara berhasil didaftarkan!",
  "data": {
    "id": 1,
    "name": "Steven",
    "voice_id": "1779573958_steven.wav",
    "created_at": "2026-05-24T..."
  }
}
```

### Generate TTS
```
POST /api/voices/tts
Content-Type: application/json

Body:
{
  "voice_id": "1779573958_steven.wav",
  "text": "Hello, this is a test."
}

Response 200:
  audio binary (audio/mpeg)
```

### Get All Voices
```
GET /api/voices

Response 200:
{
  "status": "success",
  "message": "Daftar suara berhasil diambil!",
  "data": [
    {
      "id": 1,
      "name": "Steven",
      "voice_id": "1779573958_steven.wav",
      "created_at": "2026-05-24T..."
    }
  ]
}
```

---

## Tips Sampel Audio yang Baik

Untuk hasil kloning suara yang optimal:

- ✅ Durasi **15–30 detik**
- ✅ Format **WAV, 22050Hz, Mono**
- ✅ Rekam di ruangan **sepi tanpa gema**
- ✅ Jarak mikrofon **10–15 cm** dari mulut
- ✅ Gunakan **Audacity** untuk noise reduction sebelum upload
- ❌ Hindari background noise, musik, atau suara lain

---

## Keterbatasan

- Bahasa output hanya **English (`en`)** — XTTS v2 tidak support Bahasa Indonesia
- Membutuhkan **Google Colab** yang aktif untuk generate audio
- URL localtunnel **berubah setiap sesi** Colab — `.env` harus diupdate manual
- GPU T4 Colab free tier memiliki **batas pemakaian harian**

---

## Struktur Project

```
vox-clone/
├── cmd/
│   └── api/
│       └── main.go
├── configs/
│   ├── config.go
│   └── database.go
├── internal/
│   ├── delivery/
│   │   └── http/
│   │       ├── middleware.go
│   │       └── voice_handler.go
│   ├── domain/
│   │   └── voice.go
│   ├── repository/
│   │   ├── voice_db.go
│   │   └── xtts_api.go
│   └── usecase/
│       └── voice_usecase.go
├── web/
│   ├── app.js
│   └── index.html
├── uploads/          # File audio lokal (tidak di-push ke git)
├── vox-clone-colab.ipynb  # Script Google Colab untuk menjalankan XTTS v2
├── .env              # Konfigurasi (tidak di-push ke git)
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

---

## .gitignore

```
.env
uploads/
*.wav
*.pt
*.pth
```
