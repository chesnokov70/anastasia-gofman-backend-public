package postgres

import (
	"anastasia_gofman_backend/internal/entity"
	"strings"

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

func (r *EventRegistrationRepository) CreateAuthorRegistration(registration entity.AuthorRegistration) (entity.AuthorRegistration, error) {
	err := r.db.Create(&registration).Error
	return registration, err
}

func (r *EventRegistrationRepository) GetAllAuthorRegistrations() ([]entity.AuthorRegistration, error) {
	var registrations []entity.AuthorRegistration
	err := r.db.Find(&registrations).Error
	return registrations, err
}

func (r *EventRegistrationRepository) GetAuthorRegistrationByID(id int) (entity.AuthorRegistration, error) {
	var registration entity.AuthorRegistration
	err := r.db.First(&registration, id).Error
	return registration, err
}

func (r *EventRegistrationRepository) CreateEmailSubscription(subscription entity.EmailSubscription) (entity.EmailSubscription, error) {
	err := r.db.Create(&subscription).Error
	return subscription, err
}

func (r *EventRegistrationRepository) GetAllEmailSubscriptions(offset int, limit int, withPagination bool, status string, createdAtSort string) ([]entity.EmailSubscription, int64, error) {
	var subscriptions []entity.EmailSubscription
	var count int64
	query := r.db.Model(&entity.EmailSubscription{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	orderBy := "created_at DESC"
	if strings.ToUpper(createdAtSort) == "ASC" {
		orderBy = "created_at ASC"
	}
	query = query.Order(orderBy)

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	if withPagination {
		query = query.Offset(offset).Limit(limit)
	}

	err = query.Find(&subscriptions).Error
	return subscriptions, count, err
}

func (r *EventRegistrationRepository) UpdateEmailSubscription(subscription entity.EmailSubscription) (entity.EmailSubscription, error) {
	err := r.db.Save(&subscription).Error
	return subscription, err
}

func (r *EventRegistrationRepository) GetEmailSubscriptionByID(id int) (entity.EmailSubscription, error) {
	var subscription entity.EmailSubscription
	err := r.db.First(&subscription, id).Error
	return subscription, err
}

func (r *EventRegistrationRepository) GetEmailSubscriptionByEmail(email string) (entity.EmailSubscription, error) {
	var subscription entity.EmailSubscription
	err := r.db.Where("email = ?", email).First(&subscription).Error
	return subscription, err
}

func (r *EventRegistrationRepository) DeleteEmailSubscription(id int) error {
	return r.db.Delete(&entity.EmailSubscription{}, id).Error
}

func (r *EventRegistrationRepository) CheckIfEmailSubscribed(email string) (bool, error) {
	var count int64
	err := r.db.Model(&entity.EmailSubscription{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
