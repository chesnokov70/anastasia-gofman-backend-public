package dto

import (
	"anastasia_gofman_backend/internal/entity"
	"time"
)

type EventRegistrationRequestDTO struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	FullName string `json:"full_name" binding:"required" example:"John Doe"`
	Language string `json:"language" binding:"required" example:"en"`
}

type EventRegistrationResponseDTO struct {
	ID        int       `json:"id" example:"1"`
	Email     string    `json:"email" example:"user@example.com"`
	FullName  string    `json:"full_name" example:"John Doe"`
	Language  string    `json:"language" example:"en"`
	EventID   int       `json:"event_id" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
}

type MailDTO struct {
	ID                       int                   `json:"id" example:"1"`
	Action                   string                `json:"action" example:"event_registration"`
	HTMLText                 entity.TranslatedText `json:"html_text"`
	TimeOfSendingAfterAction *int                  `json:"time_of_sending_after_action" example:"0"`
	CreatedAt                time.Time             `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt                time.Time             `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}

type CreateMailDTO struct {
	Action                   string                `json:"action" example:"event_registration"`
	HTMLText                 entity.TranslatedText `json:"html_text" binding:"required"`
	TimeOfSendingAfterAction *int                  `json:"time_of_sending_after_action" example:"0"`
}

func ToEventRegistrationResponseDTO(registration entity.EventRegistration) EventRegistrationResponseDTO {
	return EventRegistrationResponseDTO{
		ID:        registration.ID,
		Email:     registration.Email,
		FullName:  registration.FullName,
		Language:  registration.Language,
		EventID:   registration.EventID,
		CreatedAt: registration.CreatedAt,
	}
}

func ToMailDTO(mail entity.Mail) MailDTO {
	return MailDTO{
		ID:                       mail.ID,
		Action:                   mail.Action,
		HTMLText:                 mail.HTMLText,
		TimeOfSendingAfterAction: mail.TimeOfSendingAfterAction,
		CreatedAt:                mail.CreatedAt,
		UpdatedAt:                mail.UpdatedAt,
	}
}

func (dto CreateMailDTO) ToEntity() entity.Mail {
	return entity.Mail{
		Action:                   dto.Action,
		HTMLText:                 dto.HTMLText,
		TimeOfSendingAfterAction: dto.TimeOfSendingAfterAction,
	}
}

type EmailSubscriptionRequestDTO struct {
	Email string `json:"email" binding:"required" example:"user@example.com"`
}

type EmailSubscriptionResponseDTO struct {
	ID        int       `json:"id" example:"1"`
	Email     string    `json:"email" example:"user@example.com"`
	Status    string    `json:"status" example:"active"`
	CreatedAt time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
}

type UpdateEmailSubscriptionStatusDTO struct {
	ID     *int    `json:"id"`
	Email  *string `json:"email"`
	Status string  `json:"status" binding:"required" example:"inactive"`
}

func ToEmailSubscriptionResponseDTO(subscription entity.EmailSubscription) EmailSubscriptionResponseDTO {
	return EmailSubscriptionResponseDTO{
		ID:        subscription.ID,
		Email:     subscription.Email,
		Status:    subscription.Status,
		CreatedAt: subscription.CreatedAt,
	}
}

type ArtRequestRequestDTO struct {
	Email       string `json:"email" binding:"required" example:"user@example.com"`
	FullName    string `json:"full_name" binding:"required" example:"John Doe"`
	Language    string `json:"language" binding:"required" example:"en"`
	PhoneNumber string `json:"phone_number" binding:"required" example:"+79995667379"`
	Request     string `json:"request" binding:"required" example:"I want to request a piece of art"`
}

type ArtRequestResponseDTO struct {
	ID          int            `json:"id" example:"1"`
	Email       string         `json:"email" example:"user@example.com"`
	FullName    string         `json:"full_name" example:"John Doe"`
	Language    string         `json:"language" example:"en"`
	PhoneNumber string         `json:"phone_number" example:"+79995667379"`
	Request     string         `json:"request" example:"I want to request a piece of art"`
	ArtID       int            `json:"art_id" example:"1"`
	Art         ArtResponseDTO `json:"art"`
	CreatedAt   time.Time      `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt   time.Time      `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}

func ToArtRequestResponseDTO(request entity.ArtRequest, base_url string) ArtRequestResponseDTO {
	return ArtRequestResponseDTO{
		ID:          request.ID,
		Email:       request.Email,
		FullName:    request.FullName,
		Language:    request.Language,
		PhoneNumber: request.PhoneNumber,
		Request:     request.Request,
		ArtID:       request.ArtID,
		Art:         ToArtResponseDTO(request.Art, base_url),
		CreatedAt:   request.CreatedAt,
		UpdatedAt:   request.UpdatedAt,
	}
}
