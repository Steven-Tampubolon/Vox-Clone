package configs

import (
	"log"
	"vox-clone/internal/domain"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitializeDB membuka koneksi ke MySQL dan otomatis melakukan migrasi skema tabel
func InitializeDB(cfg *Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(cfg.BuildDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal menyambungkan ke database MySQL: %v", err)
	}

	// Proses AutoMigrate dikapsulasi di sini agar tidak mengotori main.go
	log.Println("Menjalankan migrasi database otomatis...")
	if err := db.AutoMigrate(&domain.Voice{}); err != nil {
		log.Fatalf("Gagal melakukan AutoMigrate database: %v", err)
	}

	return db
}
