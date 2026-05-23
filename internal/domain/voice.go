package domain

import (
	"context"
	"io"
	"time"
)

// Voice melambangkan entitas suara yang berhasil dikloning di sistem kita.
// Kita menggunakan tag standar untuk pemetaan DB nanti tanpa mengikat domain ke library DB tertentu.
type Voice struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	VoiceID   string    `json:"voice_id"`
	CreatedAt time.Time `json:"created_at"`
}

// VoiceRepository adalah kontrak untuk segala operasi baca/tulis ke Database.
type VoiceRepository interface {
	Save(ctx context.Context, voice *Voice) error
	FindAll(ctx context.Context) ([]Voice, error)
}

// VoiceAIService adalah kontrak untuk berkomunikasi dengan Mesin AI (XTTS/Cloud)
type VoiceAIService interface {
	CloneVoice(ctx context.Context, name string, fileReader io.Reader, fileName string) (string, error)
	GenerateTTS(ctx context.Context, voiceID, text string) (io.ReadCloser, error)
}

// VoiceUsecase sekarang menggunakan interface baru yang sudah diganti namanya
type VoiceUsecase interface {
	RegisterNewVoice(ctx context.Context, name string, fileReader io.Reader, fileName string) (*Voice, error)
	TextToSpeech(ctx context.Context, voiceID, text string) (io.ReadCloser, error)
	GetAllVoices(ctx context.Context) ([]Voice, error)
}
