package usecase

import (
	"context"
	"io"
	"vox-clone/internal/domain"
)

type voiceUsecase struct {
	repo    domain.VoiceRepository
	voiceAI domain.VoiceAIService // Nama variabel diubah agar lebih umum
}

func NewVoiceUsecase(r domain.VoiceRepository, ai domain.VoiceAIService) domain.VoiceUsecase {
	return &voiceUsecase{
		repo:    r,
		voiceAI: ai,
	}
}

func (u *voiceUsecase) RegisterNewVoice(ctx context.Context, name string, fileReader io.Reader, fileName string) (*domain.Voice, error) {
	// Memanggil service AI yang baru
	voiceID, err := u.voiceAI.CloneVoice(ctx, name, fileReader, fileName)
	if err != nil {
		return nil, err
	}

	voice := &domain.Voice{
		Name:    name,
		VoiceID: voiceID,
	}

	if err := u.repo.Save(ctx, voice); err != nil {
		return nil, err
	}

	return voice, nil
}

func (u *voiceUsecase) TextToSpeech(ctx context.Context, voiceID, text string) (io.ReadCloser, error) {
	return u.voiceAI.GenerateTTS(ctx, voiceID, text)
}

func (u *voiceUsecase) GetAllVoices(ctx context.Context) ([]domain.Voice, error) {
	return u.repo.FindAll(ctx)
}
