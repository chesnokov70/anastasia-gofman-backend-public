package entity

import (
	"time"
)

type EventRegistration struct {
	ID        int       `json:"id" gorm:"primaryKey" example:"1"`
	Email     string    `json:"email" gorm:"not null" example:"user@example.com"`
	FullName  string    `json:"full_name" gorm:"not null" example:"John Doe"`
	Language  string    `json:"language" gorm:"not null" example:"en"`
	EventID   int       `json:"event_id" gorm:"not null" example:"1"`
	Event     Event     `json:"event" gorm:"foreignKey:EventID;references:ID"`
	CreatedAt time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}

type AuthorRegistration struct {
	ID          int       `json:"id" gorm:"primaryKey" example:"1"`
	Email       string    `json:"email" example:"user@example.com"`
	FullName    string    `json:"full_name" example:"John Doe"`
	Language    string    `json:"language" example:"en"`
	PhoneNumber string    `json:"phone_number" example:"+79991234567"`
	Portfolio   string    `json:"portfolio" example:"https://example.com"`
	CreatedAt   time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}

type EmailSubscription struct {
	ID        int       `json:"id" gorm:"primaryKey" example:"1"`
	Email     string    `json:"email" gorm:"not null" example:"user@example.com"`
	Status    string    `json:"status" gorm:"not null" example:"active"`
	CreatedAt time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}

type ArtRequest struct {
	ID          int       `json:"id" gorm:"primaryKey" example:"1"`
	Email       string    `json:"email" gorm:"not null" example:"user@example.com"`
	FullName    string    `json:"full_name" gorm:"not null" example:"John Doe"`
	Language    string    `json:"language" gorm:"not null" example:"en"`
	PhoneNumber string    `json:"phone_number" gorm:"not null" example:"+79995667379"`
	Request     string    `json:"request" gorm:"not null" example:"I want to request a piece of art"`
	ArtID       int       `json:"art_id" gorm:"not null" example:"1"`
	Art         Art       `json:"art" gorm:"foreignKey:ArtID;references:ID"`
	CreatedAt   time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}
