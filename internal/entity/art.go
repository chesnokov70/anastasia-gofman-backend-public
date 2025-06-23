package entity

import (
	"time"

	"gorm.io/gorm"
)

// @name Art
// type Art struct {
// 	ID       int `json:"id" gorm:"primaryKey"`
// 	AuthorID int `json:"author_id" gorm:"foreignKey:AuthorID"`
// 	// Name        string    `json:"name"`
// 	Name        TranslatedText `json:"name" gorm:"type:jsonb"`
// 	Title       TranslatedText `json:"title" gorm:"type:jsonb"`
// 	Description TranslatedText `json:"description" gorm:"type:jsonb"`
// 	Medium      TranslatedText `json:"medium" gorm:"type:jsonb"`
// 	Technique   TranslatedText `json:"technique" gorm:"type:jsonb"`
// 	DimensionX  int            `json:"dimension_x"`
// 	DimensionY  int            `json:"dimension_y"`
// 	Year        int            `json:"year"`
// 	// Frame       string         `json:"frame"`
// 	Frame TranslatedText `json:"frame" gorm:"type:jsonb"`
// 	Price int            `json:"price"`

// 	MainPhotoID int   `json:"main_photo_id"`
// 	MainPhoto   Photo `json:"-" gorm:"foreignKey:MainPhotoID;references:ID"`

// 	PreviewPhotoID int   `json:"preview_photo_id"`
// 	PreviewPhoto   Photo `json:"-" gorm:"foreignKey:PreviewPhotoID;references:ID"`

// 	Photos    []Photo   `json:"-" gorm:"foreignKey:ID;references:ArtID"`
// 	Position  int       `json:"position"`
// 	CreatedAt time.Time `json:"created_at"`
// 	UpdatedAt time.Time `json:"updated_at"`
// }

// @name Art

type Art struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	AuthorID *uint  `json:"author_id"`
	Author   Author `json:"-" gorm:"foreignKey:AuthorID;references:ID"`

	Name        TranslatedText `json:"name" gorm:"type:jsonb"`
	Title       TranslatedText `json:"title" gorm:"type:jsonb"`
	Description TranslatedText `json:"description" gorm:"type:jsonb"`
	Medium      TranslatedText `json:"medium" gorm:"type:jsonb"`
	Technique   TranslatedText `json:"technique" gorm:"type:jsonb"`
	DimensionX  int            `json:"dimension_x" example:"100"`
	DimensionY  int            `json:"dimension_y" example:"100"`
	Year        int            `json:"year" example:"2021"`
	Frame       TranslatedText `json:"frame" gorm:"type:jsonb"`
	Price       int            `json:"price" example:"100000"`

	MainPhotoID *uint `json:"main_photo_id" example:"1"`
	MainPhoto   Photo `json:"-" gorm:"foreignKey:MainPhotoID;references:ID"`

	PreviewPhotoID *uint `json:"preview_photo_id" example:"2"`
	PreviewPhoto   Photo `json:"-" gorm:"foreignKey:PreviewPhotoID;references:ID"`

	Photos []Photo `json:"-" gorm:"polymorphic:Owner;polymorphicValue:arts"`

	// PhotosIDS []uint    `json:"photos_ids"`
	Position  int       `json:"position" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2021-01-01T00:00:00Z"`

	NameForStripe        string `json:"name_for_stripe" example:"some name in english"`
	DescriptionForStripe string `json:"description_for_stripe" example:"some description in english"`
	StripeProductID      string `json:"stripe_product_id" example:"prod_1234567890"`
	PaymentLink          string `json:"payment_link" example:"https://buy.stripe.com/test_cN29C0000000000000"`
}

func (a *Art) BeforeDelete(tx *gorm.DB) error {
	return tx.Where("owner_id = ? AND owner_type = ?", a.ID, "arts").Delete(&Photo{}).Error
}
