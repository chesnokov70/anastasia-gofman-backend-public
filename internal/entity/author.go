package entity

import (
	"time"
)

// @name SocialLink
type SocialLink struct {
	Instagram string `json:"instagram" example:"vovka_insta"`
	Telegram  string `json:"telegram" example:"vovka_tele"`
	Vkontakte string `json:"vkontakte" example:"vovka_vk"`
	Facebook  string `json:"facebook" example:"vovka_fb"`
	Twitter   string `json:"twitter" example:"vovka_tweet"`
	Youtube   string `json:"youtube" example:"vovka_channel"`
	Linkedin  string `json:"linkedin" example:"vovka_linkedin"`
	Whatsapp  string `json:"whatsapp" example:"+5553535"`
	Pinterest string `json:"pinterest" example:"vovka_pins"`
	Behance   string `json:"behance" example:"vovka_art"`
}

// @name ContactInfo
type ContactInfo struct {
	Email string     `json:"email" example:"vovka@sosijopa.com"`
	Phone string     `json:"phone" example:"+5553535"`
	Links SocialLink `json:"links" gorm:"embedded"`
}

// @name Author
type Author struct {
	ID   uint           `json:"id" gorm:"primary_key" example:"1"`
	Name TranslatedText `json:"name" gorm:"type:jsonb"`
	Bio  TranslatedText `json:"bio" gorm:"type:jsonb"`

	// NEW
	Biography             TranslatedText `json:"biography" gorm:"type:jsonb"`
	EducationalBackground TranslatedText `json:"educational_background" gorm:"type:jsonb"`
	Exhibitions           TranslatedText `json:"exhibitions" gorm:"type:jsonb"`
	ContactInfo           TranslatedText `json:"contact_info" gorm:"type:jsonb"`

	// NEW

	Contact   ContactInfo `json:"contact" gorm:"embedded"`
	CreatedAt time.Time   `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt time.Time   `json:"updated_at" example:"2021-01-01T00:00:00Z"`
	Position  int         `json:"position" example:"1"`
	IsActive  bool        `json:"is_active" example:"true"`
	// PhotoID   int            `json:"photo_id"`
	// Photo     Photo          `json:"photo" gorm:"foreignKey:PhotoID"`

	// PHOTOS
	MainPhotoID *uint `json:"main_photo_id" example:"1"`
	MainPhoto   Photo `json:"-" gorm:"foreignKey:MainPhotoID;references:ID"`

	PreviewPhotoID *uint `json:"preview_photo_id" example:"2"`
	PreviewPhoto   Photo `json:"-" gorm:"foreignKey:PreviewPhotoID;references:ID"`

	Photos []Photo `json:"-" gorm:"polymorphic:Owner;polymorphicValue:authors"`
}
