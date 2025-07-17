package postgres

import (
	"anastasia_gofman_backend/internal/entity"
	"encoding/json"

	"gorm.io/gorm"
)

type SettingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) GetSettings() (entity.Settings, error) {
	var settings entity.Settings
	if err := r.db.First(&settings, 1).Error; err != nil {
		return entity.Settings{}, err
	}
	return settings, nil
}

func (r *SettingsRepository) UpdateSettings(settings entity.Settings) (entity.Settings, error) {
	if err := r.db.Save(&settings).Error; err != nil {
		return entity.Settings{}, err
	}
	return settings, nil
}

func (r *SettingsRepository) CreateDefaultSettings() {
	var count int64
	r.db.Model(&entity.Settings{}).Count(&count)
	if count == 0 {
		emptyJSON, _ := json.Marshal(map[string]interface{}{})
		defaultSettings := entity.Settings{
			ID:   1,
			Data: emptyJSON,
		}
		r.db.Create(&defaultSettings)
	}
}
