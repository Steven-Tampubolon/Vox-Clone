package usecase

import (
	"context"
	"io"
	"vox-clone/internal/domain"
)

type voiceUsecase struct {
	repo       domain.VoiceRepository
	elevenLabs domain.ElevenLabsService
}

// NewVoiceUsecase adalah constructor untuk menginisialisasi business logic layer
func NewVoiceUsecase(r domain.VoiceRepository, e domain.ElevenLabsService) domain.VoiceUsecase {
	return &voiceUsecase{
		repo:       r,
		elevenLabs: e,
	}
}

func (u *voiceUsecase) RegisterNewVoice(ctx context.Context, name string, fileReader io.Reader, fileName string) (*domain.Voice, error) {
	// 1. Kirim data rekaman/file audio ke Cloud ElevenLabs untuk di-clone
	voiceID, err := u.elevenLabs.CloneVoice(ctx, name, fileReader, fileName)
	if err != nil {
		return nil, err
	}

	// 2. Petakan hasil ke dalam entitas domain Voice
	voice := &domain.Voice{
		Name:    name,
		VoiceID: voiceID,
	}

	// 3. Simpan data entitas tersebut ke dalam database via GORM Repository
	if err := u.repo.Save(ctx, voice); err != nil {
		return nil, err
	}

	return voice, nil
}

func (u *voiceUsecase) TextToSpeech(ctx context.Context, voiceID, text string) (io.ReadCloser, error) {
	// Panggil layanan ElevenLabs untuk melakukan konversi Teks menjadi Aliran Audio (.mp3)
	return u.elevenLabs.GenerateTTS(ctx, voiceID, text)
}

func (u *voiceUsecase) GetAllVoices(ctx context.Context) ([]domain.Voice, error) {
	// Ambil semua daftar suara yang sudah pernah dikloning dari database
	return u.repo.FindAll(ctx)
}
