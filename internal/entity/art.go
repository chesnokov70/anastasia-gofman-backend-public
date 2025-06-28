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

	Name         TranslatedText `json:"name" gorm:"type:jsonb"`
	Title        TranslatedText `json:"title" gorm:"type:jsonb"`
	Description  TranslatedText `json:"description" gorm:"type:jsonb"`
	Medium       TranslatedText `json:"medium" gorm:"type:jsonb"`
	Technique    TranslatedText `json:"technique" gorm:"type:jsonb"`
	DimensionX   int            `json:"dimension_x" example:"100"`
	DimensionY   int            `json:"dimension_y" example:"100"`
	DimensionStr string         `json:"dimension_str" example:"100x100*222"`
	Year         int            `json:"year" example:"2021"`
	Frame        TranslatedText `json:"frame" gorm:"type:jsonb"`
	Price        int            `json:"price" example:"100000"`

	// HUETA - wtf:
	Size      string `json:"size" example:"small"`
	Direction string `json:"direction" example:"EXCLUSIVE"`
	Style     string `json:"style" example:"abstract"`

	Type        string `json:"type" example:"archive"`
	ArchiveType string `json:"archive_type" example:"repeat"`

	CollectionID *uint         `json:"collection_id" example:"1"`
	Collection   ArtCollection `json:"-" gorm:"foreignKey:CollectionID;references:ID"`

	// END HUETA

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

type ArtFilter struct {
	PriceFrom   *int
	PriceTo     *int
	Size        *string
	Direction   *string
	Style       *string
	Author      *string
	Type        *string
	ArchiveType *string
}

type ArtCollection struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        TranslatedText `json:"name" gorm:"type:jsonb"`
	Description TranslatedText `json:"description" gorm:"type:jsonb"`
	Arts        []Art          `json:"arts,omitempty" gorm:"foreignKey:CollectionID"`
	CreatedAt   time.Time      `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt   time.Time      `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}

type Article struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       TranslatedText `json:"title" gorm:"type:jsonb"`
	Description TranslatedText `json:"description" gorm:"type:jsonb"`
	FullText    TranslatedText `json:"full_text" gorm:"type:jsonb"`
	Link        string         `json:"link" example:"https://example.com"`

	MainPhotoID *uint `json:"main_photo_id" example:"1"`
	MainPhoto   Photo `json:"-" gorm:"foreignKey:MainPhotoID;references:ID"`

	PreviewPhotoID *uint `json:"preview_photo_id" example:"2"`
	PreviewPhoto   Photo `json:"-" gorm:"foreignKey:PreviewPhotoID;references:ID"`

	Photos []Photo `json:"-" gorm:"polymorphic:Owner;polymorphicValue:article"`

	Position  int       `json:"position" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}

type Press struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       TranslatedText `json:"title" gorm:"type:jsonb"`
	Description TranslatedText `json:"description" gorm:"type:jsonb"`
	FullText    TranslatedText `json:"full_text" gorm:"type:jsonb"`
	Link        string         `json:"link" example:"https://example.com"`
	MainPhotoID *uint          `json:"main_photo_id" example:"1"`
	MainPhoto   Photo          `json:"-" gorm:"foreignKey:MainPhotoID;references:ID"`

	PreviewPhotoID *uint `json:"preview_photo_id" example:"2"`
	PreviewPhoto   Photo `json:"-" gorm:"foreignKey:PreviewPhotoID;references:ID"`

	Photos []Photo `json:"-" gorm:"polymorphic:Owner;polymorphicValue:press"`

	// PhotosIDS []uint    `json:"photos_ids"`
	Position  int       `json:"position" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}
