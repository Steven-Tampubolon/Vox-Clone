package main

import (
	"log"
	"vox-clone/configs"
	"vox-clone/internal/delivery/http"
	"vox-clone/internal/repository"
	"vox-clone/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Ambil Konfigurasi Sistem dari .env
	cfg := configs.LoadConfig()
	if cfg.ElevenLabsAPIKey == "" {
		log.Fatalf("Error: ELEVENLABS_API_KEY wajib diisi di dalam file .env!")
	}

	// 2. Inisialisasi Database & Jalankan AutoMigrate
	db := configs.InitializeDB(cfg)

	// 3. Inisialisasi HTTP Router Engine & Terapkan CORS Middleware
	r := gin.Default()
	r.Use(http.CORSMiddleware())

	// 4. Dependency Injection (Perakitan Komponen Arsitektur)
	voiceRepo := repository.NewVoiceDBRepository(db)
	elevenService := repository.NewElevenLabsService(cfg.ElevenLabsAPIKey)
	voiceUsecase := usecase.NewVoiceUsecase(voiceRepo, elevenService)

	// Daftarkan Handler HTTP
	http.NewVoiceHandler(r, voiceUsecase)

	// 5. Hidupkan Mesin Aplikasi
	log.Printf("[VoxClone] Backend sukses berjalan di port :%s\n", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Server gagal menyala: %v", err)
	}
}
