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
