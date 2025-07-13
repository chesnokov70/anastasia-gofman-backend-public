// swag:name TranslatedText
package dto

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/pkg/config"
	"fmt"
	"strings"
)

type ArtFilteringDTO struct {
	PriceFrom   *int          `json:"price_from"`
	PriceTo     *int          `json:"price_to"`
	Size        *string       `json:"size"`
	Direction   *string       `json:"direction"`
	Style       *string       `json:"style"`
	Author      *string       `json:"author"`
	Type        *string       `json:"type"`
	ArchiveType *string       `json:"archive_type"`
	Search      *ArtSearchDTO `json:"search"`
}

type ArtSearchDTO struct {
	// Name         *TranslatedTextDTO `json:"name"`
	Name         *string            `json:"name"`
	Title        *TranslatedTextDTO `json:"title"`
	Description  *TranslatedTextDTO `json:"description"`
	Medium       *TranslatedTextDTO `json:"medium"`
	Technique    *TranslatedTextDTO `json:"technique"`
	Year         *string            `json:"year"`
	Frame        *TranslatedTextDTO `json:"frame"`
	Style        *string            `json:"style"`
	Direction    *string            `json:"direction"`
	DimensionStr *string            `json:"dimension_str"`
}

func (dto *ArtFilteringDTO) ToEntity() *entity.ArtFilter {
	if dto == nil {
		return nil
	}

	var search *entity.ArtSearch
	if dto.Search != nil {
		search = &entity.ArtSearch{}
		if dto.Search.Name != nil {
			search.Name = dto.Search.Name
		}
		if dto.Search.Title != nil {
			titleEntity := dto.Search.Title.ToEntity()
			search.Title = &titleEntity
		}
		if dto.Search.Description != nil {
			descEntity := dto.Search.Description.ToEntity()
			search.Description = &descEntity
		}
		if dto.Search.Medium != nil {
			mediumEntity := dto.Search.Medium.ToEntity()
			search.Medium = &mediumEntity
		}
		if dto.Search.Technique != nil {
			techniqueEntity := dto.Search.Technique.ToEntity()
			search.Technique = &techniqueEntity
		}
		if dto.Search.Frame != nil {
			frameEntity := dto.Search.Frame.ToEntity()
			search.Frame = &frameEntity
		}
		search.Year = dto.Search.Year
		search.Style = dto.Search.Style
		search.Direction = dto.Search.Direction
		search.DimensionStr = dto.Search.DimensionStr
	}

	return &entity.ArtFilter{
		PriceFrom:   dto.PriceFrom,
		PriceTo:     dto.PriceTo,
		Size:        dto.Size,
		Direction:   dto.Direction,
		Style:       dto.Style,
		Author:      dto.Author,
		Type:        dto.Type,
		ArchiveType: dto.ArchiveType,
		Search:      search,
	}
}

// TranslatedTextDTO defines a structure for localized text.
// @name TranslatedTextDTO
type TranslatedTextDTO struct {
	EN string `json:"en" example:"Tung, tung, tung, tung, tung, tung, tung, tung, tung, Sahur."`
	RU string `json:"ru" example:"Тун, тун, тун, Сахур."`
	ES string `json:"es" example:"Golybini Shpionini"`
}

func (t TranslatedTextDTO) ToEntity() entity.TranslatedText {
	return entity.TranslatedText{
		EN: t.EN,
		RU: t.RU,
		ES: t.ES,
	}
}

// @name CreateArt
// CreateArtDTO defines the structure for creating a new art piece.
type CreateArtDTO struct {
	AuthorID             *uint             `json:"author_id" example:"1"`
	Name                 TranslatedTextDTO `json:"name" binding:"required"`
	Title                TranslatedTextDTO `json:"title"`
	Description          TranslatedTextDTO `json:"description"`
	Medium               TranslatedTextDTO `json:"medium"`
	Technique            TranslatedTextDTO `json:"technique"`
	Direction            string            `json:"direction" example:"EXCLUSIVE"`
	Style                string            `json:"style" example:"abstract"`
	DimensionX           int               `json:"dimension_x" example:"100"`
	DimensionY           int               `json:"dimension_y" example:"100"`
	DimensionStr         string            `json:"dimension_str" example:"100x100*222"`
	Year                 int               `json:"year" example:"2021"`
	Frame                TranslatedTextDTO `json:"frame"`
	Price                int               `json:"price" example:"100000"`
	Size                 string            `json:"size" example:"small"`
	NameForStripe        string            `json:"name_for_stripe" example:"some name in english"`
	DescriptionForStripe string            `json:"description_for_stripe" example:"some description in english"`
	CollectionID         *uint             `json:"collection_id" example:"1"`
	Type                 string            `json:"type" example:"archive"`
	ArchiveType          string            `json:"archive_type" example:"repeat"`
}

// @name CreateArt
// CreateArtDTO defines the structure for creating a new art piece.
type CreateArtWithPhotosDTO struct {
	AuthorID             *uint             `json:"author_id" example:"1"`
	Name                 TranslatedTextDTO `json:"name" binding:"required"`
	Title                TranslatedTextDTO `json:"title"`
	Description          TranslatedTextDTO `json:"description"`
	Medium               TranslatedTextDTO `json:"medium"`
	Style                string            `json:"style" example:"abstract"`
	Technique            TranslatedTextDTO `json:"technique"`
	Direction            string            `json:"direction" example:"EXCLUSIVE"`
	DimensionX           int               `json:"dimension_x" example:"100"`
	DimensionY           int               `json:"dimension_y" example:"100"`
	DimensionStr         string            `json:"dimension_str" example:"100x100*222"`
	Year                 int               `json:"year" example:"2021"`
	Frame                TranslatedTextDTO `json:"frame"`
	Price                int               `json:"price" example:"100000"`
	Size                 string            `json:"size" example:"small"`
	NameForStripe        string            `json:"name_for_stripe" example:"some name in english"`
	DescriptionForStripe string            `json:"description_for_stripe" example:"some description in english"`
	CollectionID         *uint             `json:"collection_id" example:"1"`
	// MainPhoto    *multipart.FileHeader   `json:"main_photo"`
	// PreviewPhoto *multipart.FileHeader   `json:"preview_photo"`
	// Photos       []*multipart.FileHeader `json:"photos"`
	Type        string `json:"type" example:"archive"`
	ArchiveType string `json:"archive_type" example:"repeat"`
}

// UpdateArtDTO defines the structure for updating an existing art piece.
// @name UpdateArt
type UpdateArtDTO struct {
	AuthorID             *uint             `json:"author_id" example:"1"`
	Name                 TranslatedTextDTO `json:"name"`
	Title                TranslatedTextDTO `json:"title"`
	Description          TranslatedTextDTO `json:"description"`
	Medium               TranslatedTextDTO `json:"medium"`
	Technique            TranslatedTextDTO `json:"technique"`
	Style                string            `json:"style" example:"abstract"`
	Direction            string            `json:"direction" example:"EXCLUSIVE"`
	DimensionX           int               `json:"dimension_x" example:"100"`
	DimensionY           int               `json:"dimension_y" example:"100"`
	DimensionStr         string            `json:"dimension_str" example:"100x100*222"`
	Year                 int               `json:"year" example:"2021"`
	Size                 string            `json:"size" example:"small"`
	Frame                TranslatedTextDTO `json:"frame"`
	Price                int               `json:"price" example:"100000"`
	NameForStripe        string            `json:"name_for_stripe" example:"some name in english"`
	DescriptionForStripe string            `json:"description_for_stripe" example:"some description in english"`
	CollectionID         *uint             `json:"collection_id" example:"1"`
	Type                 string            `json:"type" example:"archive"`
	ArchiveType          string            `json:"archive_type" example:"repeat"`
}

func (dto *CreateArtDTO) ToEntity(id *uint) entity.Art {
	art := entity.Art{
		// ID:          id,
		AuthorID:             dto.AuthorID,
		Name:                 dto.Name.ToEntity(),
		Title:                dto.Title.ToEntity(),
		Description:          dto.Description.ToEntity(),
		Medium:               dto.Medium.ToEntity(),
		Technique:            dto.Technique.ToEntity(),
		DimensionX:           dto.DimensionX,
		DimensionY:           dto.DimensionY,
		Style:                dto.Style,
		Direction:            dto.Direction,
		DimensionStr:         dto.DimensionStr,
		Year:                 dto.Year,
		Size:                 dto.Size,
		Frame:                dto.Frame.ToEntity(),
		Price:                dto.Price,
		NameForStripe:        dto.NameForStripe,
		DescriptionForStripe: dto.DescriptionForStripe,
		CollectionID:         dto.CollectionID,
		Type:                 dto.Type,
		ArchiveType:          dto.ArchiveType,
	}
	if id != nil {
		art.ID = *id
	}
	return art
}

func (dto *CreateArtWithPhotosDTO) ToEntity(id *uint) entity.Art {
	art := entity.Art{
		// ID:          id,
		AuthorID:             dto.AuthorID,
		Name:                 dto.Name.ToEntity(),
		Title:                dto.Title.ToEntity(),
		Description:          dto.Description.ToEntity(),
		Medium:               dto.Medium.ToEntity(),
		Technique:            dto.Technique.ToEntity(),
		DimensionX:           dto.DimensionX,
		DimensionY:           dto.DimensionY,
		Style:                dto.Style,
		Direction:            dto.Direction,
		DimensionStr:         dto.DimensionStr,
		Year:                 dto.Year,
		Size:                 dto.Size,
		Frame:                dto.Frame.ToEntity(),
		Price:                dto.Price,
		NameForStripe:        dto.NameForStripe,
		DescriptionForStripe: dto.DescriptionForStripe,
		CollectionID:         dto.CollectionID,
		Type:                 dto.Type,
		ArchiveType:          dto.ArchiveType,
	}
	if id != nil {
		art.ID = *id
	}
	return art
}

func (dto *UpdateArtDTO) ToEntity(id *uint) entity.Art {
	art := entity.Art{
		// ID:          id,
		AuthorID:             dto.AuthorID,
		Name:                 dto.Name.ToEntity(),
		Title:                dto.Title.ToEntity(),
		Description:          dto.Description.ToEntity(),
		Medium:               dto.Medium.ToEntity(),
		Technique:            dto.Technique.ToEntity(),
		DimensionX:           dto.DimensionX,
		DimensionY:           dto.DimensionY,
		Style:                dto.Style,
		Direction:            dto.Direction,
		DimensionStr:         dto.DimensionStr,
		Year:                 dto.Year,
		Size:                 dto.Size,
		Frame:                dto.Frame.ToEntity(),
		Price:                dto.Price,
		NameForStripe:        dto.NameForStripe,
		DescriptionForStripe: dto.DescriptionForStripe,
		CollectionID:         dto.CollectionID,
		Type:                 dto.Type,
		ArchiveType:          dto.ArchiveType,
	}
	if id != nil {
		art.ID = *id
	}
	return art
}

type PhotoResponseDTO struct {
	ID       uint   `json:"id"`
	Path     string `json:"path"`
	Position int    `json:"position"`
}

type ArtResponseDTO struct {
	ID                   uint                  `json:"id"`
	AuthorID             *uint                 `json:"author_id,omitempty"`
	AuthorName           entity.TranslatedText `json:"author_name"`
	Name                 entity.TranslatedText `json:"name"`
	Style                string                `json:"style"`
	Title                entity.TranslatedText `json:"title"`
	Description          entity.TranslatedText `json:"description"`
	Medium               entity.TranslatedText `json:"medium"`
	Technique            entity.TranslatedText `json:"technique"`
	DimensionX           int                   `json:"dimension_x"`
	DimensionY           int                   `json:"dimension_y"`
	DimensionStr         string                `json:"dimension_str"`
	Direction            string                `json:"direction"`
	Year                 int                   `json:"year"`
	Frame                entity.TranslatedText `json:"frame"`
	Price                int                   `json:"price"`
	MainPhotoPath        string                `json:"main_photo_path,omitempty"`
	PreviewPhotoPath     string                `json:"preview_photo_path,omitempty"`
	Position             int                   `json:"position"`
	CreatedAt            string                `json:"created_at"`
	UpdatedAt            string                `json:"updated_at"`
	Size                 string                `json:"size"`
	Photos               []PhotoResponseDTO    `json:"photos"`
	StripeProductID      string                `json:"stripe_product_id"`
	PaymentLink          string                `json:"payment_link"`
	NameForStripe        string                `json:"name_for_stripe"`
	DescriptionForStripe string                `json:"description_for_stripe"`
	CollectionID         *uint                 `json:"collection_id,omitempty"`
	Type                 string                `json:"type"`
	ArchiveType          string                `json:"archive_type"`
}

// TODO: to config
// const BaseURL = "http://localhost:8010"
var BaseURL = config.GetBaseURL()

func ToArtResponseDTO(art entity.Art, base_url string) ArtResponseDTO {
	dto := ArtResponseDTO{
		ID:                   art.ID,
		AuthorID:             art.AuthorID,
		AuthorName:           art.Author.Name,
		Name:                 art.Name,
		Style:                art.Style,
		Title:                art.Title,
		Description:          art.Description,
		Medium:               art.Medium,
		Technique:            art.Technique,
		DimensionX:           art.DimensionX,
		DimensionY:           art.DimensionY,
		DimensionStr:         art.DimensionStr,
		Direction:            art.Direction,
		Year:                 art.Year,
		Frame:                art.Frame,
		Price:                art.Price,
		Size:                 art.Size,
		Position:             art.Position,
		CreatedAt:            art.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:            art.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Photos:               []PhotoResponseDTO{}, // Initialize with empty slice
		StripeProductID:      art.StripeProductID,
		PaymentLink:          art.PaymentLink,
		NameForStripe:        art.NameForStripe,
		DescriptionForStripe: art.DescriptionForStripe,
		CollectionID:         art.CollectionID,
		Type:                 art.Type,
		ArchiveType:          art.ArchiveType,
	}

	if art.MainPhotoID != nil && art.MainPhoto.Path != "" {
		path := strings.TrimPrefix(art.MainPhoto.Path, "/")
		dto.MainPhotoPath = fmt.Sprintf("%s/%s", base_url, path)
	}
	if art.PreviewPhotoID != nil && art.PreviewPhoto.Path != "" {
		path := strings.TrimPrefix(art.PreviewPhoto.Path, "/")
		dto.PreviewPhotoPath = fmt.Sprintf("%s/%s", base_url, path)
	}

	for _, photo := range art.Photos {
		if !photo.IsMain && !photo.IsPreview {
			path := strings.TrimPrefix(photo.Path, "/")
			fullPath := fmt.Sprintf("%s/%s", base_url, path)
			dto.Photos = append(dto.Photos, PhotoResponseDTO{
				ID:       photo.ID,
				Path:     fullPath,
				Position: photo.Position,
			})
		}
	}

	return dto
}

func ToArtResponseDTOs(arts []entity.Art, base_url string) []ArtResponseDTO {
	artDTOs := make([]ArtResponseDTO, len(arts))
	for i, art := range arts {
		artDTOs[i] = ToArtResponseDTO(art, base_url)
	}
	return artDTOs
}
