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
