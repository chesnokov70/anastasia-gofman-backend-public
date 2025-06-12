// swag:name TranslatedText
package dto

import (
	"anastasia_gofman_backend/internal/entity"
	"fmt"
	"strings"
)

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
	AuthorID    *uint             `json:"author_id" example:"1"`
	Name        TranslatedTextDTO `json:"name" binding:"required"`
	Title       TranslatedTextDTO `json:"title"`
	Description TranslatedTextDTO `json:"description"`
	Medium      TranslatedTextDTO `json:"medium"`
	Technique   TranslatedTextDTO `json:"technique"`
	DimensionX  int               `json:"dimension_x" example:"100"`
	DimensionY  int               `json:"dimension_y" example:"100"`
	Year        int               `json:"year" example:"2021"`
	Frame       TranslatedTextDTO `json:"frame"`
	Price       int               `json:"price" example:"100000"`
}

// @name CreateArt
// CreateArtDTO defines the structure for creating a new art piece.
type CreateArtWithPhotosDTO struct {
	AuthorID    *uint             `json:"author_id" example:"1"`
	Name        TranslatedTextDTO `json:"name" binding:"required"`
	Title       TranslatedTextDTO `json:"title"`
	Description TranslatedTextDTO `json:"description"`
	Medium      TranslatedTextDTO `json:"medium"`
	Technique   TranslatedTextDTO `json:"technique"`
	DimensionX  int               `json:"dimension_x" example:"100"`
	DimensionY  int               `json:"dimension_y" example:"100"`
	Year        int               `json:"year" example:"2021"`
	Frame       TranslatedTextDTO `json:"frame"`
	Price       int               `json:"price" example:"100000"`
	// MainPhoto    *multipart.FileHeader   `json:"main_photo"`
	// PreviewPhoto *multipart.FileHeader   `json:"preview_photo"`
	// Photos       []*multipart.FileHeader `json:"photos"`
}

// UpdateArtDTO defines the structure for updating an existing art piece.
// @name UpdateArt
type UpdateArtDTO struct {
	AuthorID    *uint             `json:"author_id" example:"1"`
	Name        TranslatedTextDTO `json:"name"`
	Title       TranslatedTextDTO `json:"title"`
	Description TranslatedTextDTO `json:"description"`
	Medium      TranslatedTextDTO `json:"medium"`
	Technique   TranslatedTextDTO `json:"technique"`
	DimensionX  int               `json:"dimension_x" example:"100"`
	DimensionY  int               `json:"dimension_y" example:"100"`
	Year        int               `json:"year" example:"2021"`
	Frame       TranslatedTextDTO `json:"frame"`
	Price       int               `json:"price" example:"100000"`
}

func (dto *CreateArtDTO) ToEntity(id *uint) entity.Art {
	art := entity.Art{
		// ID:          id,
		AuthorID:    dto.AuthorID,
		Name:        dto.Name.ToEntity(),
		Title:       dto.Title.ToEntity(),
		Description: dto.Description.ToEntity(),
		Medium:      dto.Medium.ToEntity(),
		Technique:   dto.Technique.ToEntity(),
		DimensionX:  dto.DimensionX,
		DimensionY:  dto.DimensionY,
		Year:        dto.Year,
		Frame:       dto.Frame.ToEntity(),
		Price:       dto.Price,
	}
	if id != nil {
		art.ID = *id
	}
	return art
}

func (dto *CreateArtWithPhotosDTO) ToEntity(id *uint) entity.Art {
	art := entity.Art{
		// ID:          id,
		AuthorID:    dto.AuthorID,
		Name:        dto.Name.ToEntity(),
		Title:       dto.Title.ToEntity(),
		Description: dto.Description.ToEntity(),
		Medium:      dto.Medium.ToEntity(),
		Technique:   dto.Technique.ToEntity(),
		DimensionX:  dto.DimensionX,
		DimensionY:  dto.DimensionY,
		Year:        dto.Year,
		Frame:       dto.Frame.ToEntity(),
		Price:       dto.Price,
	}
	if id != nil {
		art.ID = *id
	}
	return art
}

func (dto *UpdateArtDTO) ToEntity(id *uint) entity.Art {
	art := entity.Art{
		// ID:          id,
		AuthorID:    dto.AuthorID,
		Name:        dto.Name.ToEntity(),
		Title:       dto.Title.ToEntity(),
		Description: dto.Description.ToEntity(),
		Medium:      dto.Medium.ToEntity(),
		Technique:   dto.Technique.ToEntity(),
		DimensionX:  dto.DimensionX,
		DimensionY:  dto.DimensionY,
		Year:        dto.Year,
		Frame:       dto.Frame.ToEntity(),
		Price:       dto.Price,
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
	ID               uint                  `json:"id"`
	AuthorID         *uint                 `json:"author_id,omitempty"`
	Name             entity.TranslatedText `json:"name"`
	Title            entity.TranslatedText `json:"title"`
	Description      entity.TranslatedText `json:"description"`
	Medium           entity.TranslatedText `json:"medium"`
	Technique        entity.TranslatedText `json:"technique"`
	DimensionX       int                   `json:"dimension_x"`
	DimensionY       int                   `json:"dimension_y"`
	Year             int                   `json:"year"`
	Frame            entity.TranslatedText `json:"frame"`
	Price            int                   `json:"price"`
	MainPhotoPath    string                `json:"main_photo_path,omitempty"`
	PreviewPhotoPath string                `json:"preview_photo_path,omitempty"`
	Position         int                   `json:"position"`
	CreatedAt        string                `json:"created_at"`
	UpdatedAt        string                `json:"updated_at"`
	Photos           []PhotoResponseDTO    `json:"photos"`
}

// TODO: to config
const BaseURL = "http://localhost:8010"

func ToArtResponseDTO(art entity.Art) ArtResponseDTO {
	dto := ArtResponseDTO{
		ID:          art.ID,
		AuthorID:    art.AuthorID,
		Name:        art.Name,
		Title:       art.Title,
		Description: art.Description,
		Medium:      art.Medium,
		Technique:   art.Technique,
		DimensionX:  art.DimensionX,
		DimensionY:  art.DimensionY,
		Year:        art.Year,
		Frame:       art.Frame,
		Price:       art.Price,
		Position:    art.Position,
		CreatedAt:   art.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   art.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Photos:      []PhotoResponseDTO{}, // Initialize with empty slice
	}

	if art.MainPhotoID != nil && art.MainPhoto.Path != "" {
		path := strings.TrimPrefix(art.MainPhoto.Path, "/")
		dto.MainPhotoPath = fmt.Sprintf("%s/%s", BaseURL, path)
	}
	if art.PreviewPhotoID != nil && art.PreviewPhoto.Path != "" {
		path := strings.TrimPrefix(art.PreviewPhoto.Path, "/")
		dto.PreviewPhotoPath = fmt.Sprintf("%s/%s", BaseURL, path)
	}

	for _, photo := range art.Photos {
		if !photo.IsMain && !photo.IsPreview {
			path := strings.TrimPrefix(photo.Path, "/")
			fullPath := fmt.Sprintf("%s/%s", BaseURL, path)
			dto.Photos = append(dto.Photos, PhotoResponseDTO{
				ID:       photo.ID,
				Path:     fullPath,
				Position: photo.Position,
			})
		}
	}

	return dto
}

func ToArtResponseDTOs(arts []entity.Art) []ArtResponseDTO {
	artDTOs := make([]ArtResponseDTO, len(arts))
	for i, art := range arts {
		artDTOs[i] = ToArtResponseDTO(art)
	}
	return artDTOs
}
