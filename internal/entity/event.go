package entity

import (
	"time"

	"gorm.io/gorm"
)

// swag:name Event
type Event struct {
	ID          int            `json:"id" gorm:"primaryKey" example:"1"`
	Title       TranslatedText `json:"title" gorm:"type:jsonb"`
	Description TranslatedText `json:"description" gorm:"type:jsonb"`
	StartDate   time.Time      `json:"start_date" example:"2021-01-01T00:00:00Z"`
	EndDate     time.Time      `json:"end_date" example:"2021-01-01T00:00:00Z"`
	Location    TranslatedText `json:"location" gorm:"type:jsonb"`
	Language    TranslatedText `json:"language" gorm:"type:jsonb"`
	Organizer   TranslatedText `json:"organizer" gorm:"type:jsonb"`
	Format      TranslatedText `json:"format" gorm:"type:jsonb"`
	Venue       TranslatedText `json:"venue" gorm:"type:jsonb"`

	MainPhotoID *int  `json:"main_photo_id" example:"1"`
	MainPhoto   Photo `json:"-" gorm:"foreignKey:MainPhotoID;references:ID"`

	PreviewPhotoID *int  `json:"preview_photo_id" example:"2"`
	PreviewPhoto   Photo `json:"-" gorm:"foreignKey:PreviewPhotoID;references:ID"`

	Photos []Photo `json:"-" gorm:"polymorphic:Owner;polymorphicValue:event"`

	Position   int       `json:"position" example:"1"`
	CreatedAt  time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt  time.Time `json:"updated_at" example:"2021-01-01T00:00:00Z"`
	IsFinished bool      `json:"is_finished" example:"false"`
}

func (e *Event) BeforeDelete(tx *gorm.DB) error {
	return tx.Where("owner_id = ? AND owner_type = ?", e.ID, "event").Delete(&Photo{}).Error
}
