package entity

import (
	"time"
)

type Mail struct {
	ID                       int            `json:"id" gorm:"primaryKey" example:"1"`
	Action                   string         `json:"action" gorm:"not null" example:"event_registration"`
	HTMLText                 TranslatedText `json:"html_text" gorm:"type:jsonb"`
	TimeOfSendingAfterAction *int           `json:"time_of_sending_after_action" example:"0"`
	CreatedAt                time.Time      `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt                time.Time      `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}
