package entity

import (
	"time"
)

// type Photo struct {
// 	ID        int       `json:"id" gorm:"primaryKey"`
// 	Path      string    `json:"path"`
// 	ArtID     int       `json:"art_id" gorm:"index foreignKey:ArtID"`
// 	Art       *Art      `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
// 	CreatedAt time.Time `json:"created_at"`
// 	UpdatedAt time.Time `json:"updated_at"`
// 	EventID   int       `json:"event_id" gorm:"index foreignKey:EventID"`
// 	Event     *Event    `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
// }

// @name Photo
type Photo struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Path      string    `json:"path"`
	OwnerID   uint      `json:"owner_id"`
	OwnerType string    `json:"owner_type"` // "arts", "event"
	Position  int       `json:"position"`
	IsMain    bool      `json:"is_main"`
	IsPreview bool      `json:"is_preview"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
