package repository

import (
	"context"
	"vox-clone/internal/domain"

	"gorm.io/gorm"
)

type voiceDBRepository struct {
	db *gorm.DB
}

// NewVoiceDBRepository adalah constructor untuk menginisialisasi repository database
func NewVoiceDBRepository(db *gorm.DB) domain.VoiceRepository {
	return &voiceDBRepository{db: db}
}

func (r *voiceDBRepository) Save(ctx context.Context, voice *domain.Voice) error {
	// GORM otomatis mencocokkan struct domain.Voice dengan tabel 'voices' di database
	return r.db.WithContext(ctx).Create(voice).Error
}

func (r *voiceDBRepository) FindAll(ctx context.Context) ([]domain.Voice, error) {
	var voices []domain.Voice
	err := r.db.WithContext(ctx).Find(&voices).Error
	return voices, err
}
