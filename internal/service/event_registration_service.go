package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/repository/postgres"
	"fmt"
)

type EventRegistrationService struct {
	registrationRepo *postgres.EventRegistrationRepository
	eventRepo        *postgres.EventRepository
	mailService      *MailService
	emailService     *EmailService
}

func NewEventRegistrationService(
	registrationRepo *postgres.EventRegistrationRepository,
	eventRepo *postgres.EventRepository,
	mailService *MailService,
	emailService *EmailService,
) *EventRegistrationService {
	return &EventRegistrationService{
		registrationRepo: registrationRepo,
		eventRepo:        eventRepo,
		mailService:      mailService,
		emailService:     emailService,
	}
}

func (s *EventRegistrationService) RegisterForEvent(email, fullName, language string, eventID int) (entity.EventRegistration, error) {
	alreadyRegistered, err := s.registrationRepo.CheckIfUserAlreadyRegistered(email, eventID)
	if err != nil {
		return entity.EventRegistration{}, err
	}
	if alreadyRegistered {
		return entity.EventRegistration{}, fmt.Errorf("user already registered for this event")
	}

	event, err := s.eventRepo.GetEventByID(uint(eventID))
	if err != nil {
		return entity.EventRegistration{}, fmt.Errorf("event not found")
	}

	registration := entity.EventRegistration{
		Email:    email,
		FullName: fullName,
		Language: language,
		EventID:  eventID,
	}

	createdRegistration, err := s.registrationRepo.CreateRegistration(registration)
	if err != nil {
		return entity.EventRegistration{}, err
	}

	createdRegistration.Event = event

	// mailTemplate, err := s.mailService.GetMailByAction("event_registration")
	// if err == nil {
	// 	go func() {
	// 		err = s.emailService.SendRegistrationConfirmation(createdRegistration, mailTemplate)
	// 		if err != nil {
	// 			fmt.Printf("Warning: Failed to send confirmation email: %v\n", err)
	// 		}
	// 	}()
	// }

	go func() {
		err = s.emailService.SendAdminNotification(createdRegistration)
		if err != nil {
			fmt.Printf("Warning: Failed to send admin notification: %v\n", err)
		}
	}()

	return createdRegistration, nil
}

func (s *EventRegistrationService) GetRegistrationsByEventID(eventID int) ([]entity.EventRegistration, error) {
	return s.registrationRepo.GetRegistrationsByEventID(eventID)
}

func (s *EventRegistrationService) GetAllRegistrations() ([]entity.EventRegistration, error) {
	return s.registrationRepo.GetAllRegistrations()
}

func (s *EventRegistrationService) RegisterForAuthor(email, fullName, language, phoneNumber, portfolio string) (entity.AuthorRegistration, error) {
	registration := entity.AuthorRegistration{
		Email:       email,
		FullName:    fullName,
		Language:    language,
		PhoneNumber: phoneNumber,
		Portfolio:   portfolio,
	}

	createdRegistration, err := s.registrationRepo.CreateAuthorRegistration(registration)
	if err != nil {
		return entity.AuthorRegistration{}, err
	}

	go func() {
		err = s.emailService.SendAdminNotificationAuthor(createdRegistration)
		if err != nil {
			fmt.Printf("Warning: Failed to send admin notification: %v\n", err)
		}
	}()

	return createdRegistration, nil
}

func (s *EventRegistrationService) GetAllAuthorRegistrations() ([]entity.AuthorRegistration, error) {
	return s.registrationRepo.GetAllAuthorRegistrations()
}

func (s *EventRegistrationService) GetAuthorRegistrationByID(id int) (entity.AuthorRegistration, error) {
	return s.registrationRepo.GetAuthorRegistrationByID(id)
}

func (s *EventRegistrationService) SubscribeEmail(email string) (entity.EmailSubscription, error) {
	isSubscribed, _ := s.registrationRepo.CheckIfEmailSubscribed(email)
	if isSubscribed {
		return entity.EmailSubscription{}, fmt.Errorf("email already subscribed")
	}

	subscription := entity.EmailSubscription{
		Email:  email,
		Status: "active",
	}
	return s.registrationRepo.CreateEmailSubscription(subscription)
}

func (s *EventRegistrationService) GetAllEmailSubscriptions(page, size int, withPagination bool, status string, createdAtSort string) ([]entity.EmailSubscription, int64, int64, error) {
	offset, limit := 0, 0
	if page > 0 && size > 0 {
		offset = (page - 1) * size
		limit = size
	}
	subscriptions, total, err := s.registrationRepo.GetAllEmailSubscriptions(offset, limit, withPagination, status, createdAtSort)
	if err != nil {
		return nil, 0, 0, err
	}
	var totalPages int64
	if total == 0 {
		totalPages = 0
	} else {
		totalPages = (total + int64(size) - 1) / int64(size)
	}
	return subscriptions, totalPages, total, nil
}

func (s *EventRegistrationService) UpdateEmailSubscriptionStatus(id int, status string) (entity.EmailSubscription, error) {
	subscription, err := s.registrationRepo.GetEmailSubscriptionByID(id)
	if err != nil {
		return entity.EmailSubscription{}, err
	}
	subscription.Status = status
	return s.registrationRepo.UpdateEmailSubscription(subscription)
}

func (s *EventRegistrationService) DeleteEmailSubscription(id int) error {
	return s.registrationRepo.DeleteEmailSubscription(id)
}
