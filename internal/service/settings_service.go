package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/repository"
	"encoding/json"
)

type settingsService struct {
	repo repository.SettingsRepository
}

func NewSettingsService(repo repository.SettingsRepository) SettingsService {
	return &settingsService{repo: repo}
}

func (s *settingsService) GetSettings() (entity.Settings, error) {
	return s.repo.GetSettings()
}

func (s *settingsService) UpdateSettings(data map[string]interface{}) (entity.Settings, error) {
	settings, err := s.repo.GetSettings()
	if err != nil {
		return entity.Settings{}, err
	}

	var currentData map[string]interface{}
	if err := json.Unmarshal(settings.Data, &currentData); err != nil {
		return entity.Settings{}, err
	}

	for key, value := range data {
		currentData[key] = value
	}

	jsonData, err := json.Marshal(currentData)
	if err != nil {
		return entity.Settings{}, err
	}
	settings.Data = jsonData

	return s.repo.UpdateSettings(settings)
}

func (s *settingsService) OverwriteSettings(data map[string]interface{}) (entity.Settings, error) {
	settings, err := s.repo.GetSettings()
	if err != nil {
		return entity.Settings{}, err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return entity.Settings{}, err
	}
	settings.Data = jsonData

	return s.repo.UpdateSettings(settings)
}
