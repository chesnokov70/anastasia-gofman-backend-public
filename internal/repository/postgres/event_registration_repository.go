package postgres

import (
	"anastasia_gofman_backend/internal/entity"

	"gorm.io/gorm"
)

type EventRegistrationRepository struct {
	db *gorm.DB
}

func NewEventRegistrationRepository(db *gorm.DB) *EventRegistrationRepository {
	return &EventRegistrationRepository{db: db}
}

func (r *EventRegistrationRepository) CreateRegistration(registration entity.EventRegistration) (entity.EventRegistration, error) {
	err := r.db.Create(&registration).Error
	return registration, err
}

func (r *EventRegistrationRepository) GetRegistrationsByEventID(eventID int) ([]entity.EventRegistration, error) {
	var registrations []entity.EventRegistration
	err := r.db.Where("event_id = ?", eventID).Preload("Event").Find(&registrations).Error
	return registrations, err
}

func (r *EventRegistrationRepository) GetRegistrationByID(id int) (entity.EventRegistration, error) {
	var registration entity.EventRegistration
	err := r.db.Preload("Event").First(&registration, id).Error
	return registration, err
}

func (r *EventRegistrationRepository) CheckIfUserAlreadyRegistered(email string, eventID int) (bool, error) {
	var count int64
	err := r.db.Model(&entity.EventRegistration{}).Where("email = ? AND event_id = ?", email, eventID).Count(&count).Error
	return count > 0, err
}

func (r *EventRegistrationRepository) GetAllRegistrations() ([]entity.EventRegistration, error) {
	var registrations []entity.EventRegistration
	err := r.db.Preload("Event").Find(&registrations).Error
	return registrations, err
}