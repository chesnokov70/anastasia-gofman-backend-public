// swag:name TranslatedText
package dto

import (
	"anastasia_gofman_backend/internal/entity"
	"fmt"
	"strings"
)

// @name CreatePress
// CreatePressDTO defines the structure for creating a new press piece.
type CreatePressDTO struct {
	Title       TranslatedTextDTO `json:"title"`
	Description TranslatedTextDTO `json:"description"`
	FullText    TranslatedTextDTO `json:"full_text"`
	Link        string            `json:"link" example:"https://example.com"`
	Position    int               `json:"position" example:"1"`
}

// @name CreateArticle
// CreateArticleDTO defines the structure for creating a new article piece.
type CreateArticleDTO struct {
	Title       TranslatedTextDTO `json:"title"`
	Description TranslatedTextDTO `json:"description"`
	FullText    TranslatedTextDTO `json:"full_text"`
	Link        string            `json:"link" example:"https://example.com"`
	Position    int               `json:"position" example:"1"`
}

// @name UpdatePress
type UpdatePressDTO struct {
	Title       TranslatedTextDTO `json:"title"`
	Description TranslatedTextDTO `json:"description"`
	FullText    TranslatedTextDTO `json:"full_text"`
	Link        string            `json:"link" example:"https://example.com"`
	Position    int               `json:"position" example:"1"`
}

// @name UpdateArticle
type UpdateArticleDTO struct {
	Title       TranslatedTextDTO `json:"title"`
	Description TranslatedTextDTO `json:"description"`
	FullText    TranslatedTextDTO `json:"full_text"`
	Link        string            `json:"link" example:"https://example.com"`
	Position    int               `json:"position" example:"1"`
}

func (dto *CreatePressDTO) ToEntity(id *uint) entity.Press {
	press := entity.Press{
		Title:       dto.Title.ToEntity(),
		Description: dto.Description.ToEntity(),
		FullText:    dto.FullText.ToEntity(),
		Link:        dto.Link,
		Position:    dto.Position,
	}
	if id != nil {
		press.ID = *id
	}
	return press
}

func (dto *CreateArticleDTO) ToEntity(id *uint) entity.Article {
	article := entity.Article{
		Title:       dto.Title.ToEntity(),
		Description: dto.Description.ToEntity(),
		FullText:    dto.FullText.ToEntity(),
		Link:        dto.Link,
		Position:    dto.Position,
	}
	if id != nil {
		article.ID = *id
	}
	return article
}

func (dto *UpdatePressDTO) ToEntity(id *uint) entity.Press {
	press := entity.Press{
		Title:       dto.Title.ToEntity(),
		Description: dto.Description.ToEntity(),
		FullText:    dto.FullText.ToEntity(),
		Link:        dto.Link,
		Position:    dto.Position,
	}
	if id != nil {
		press.ID = *id
	}
	return press
}

func (dto *UpdateArticleDTO) ToEntity(id *uint) entity.Article {
	article := entity.Article{
		Title:       dto.Title.ToEntity(),
		Description: dto.Description.ToEntity(),
		FullText:    dto.FullText.ToEntity(),
		Link:        dto.Link,
		Position:    dto.Position,
	}
	if id != nil {
		article.ID = *id
	}
	return article
}

type PressResponseDTO struct {
	ID               uint                  `json:"id"`
	Title            entity.TranslatedText `json:"title"`
	Description      entity.TranslatedText `json:"description"`
	FullText         entity.TranslatedText `json:"full_text"`
	Link             string                `json:"link" example:"https://example.com"`
	Position         int                   `json:"position" example:"1"`
	CreatedAt        string                `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt        string                `json:"updated_at" example:"2021-01-01T00:00:00Z"`
	MainPhotoPath    string                `json:"main_photo_path,omitempty"`
	PreviewPhotoPath string                `json:"preview_photo_path,omitempty"`

	Photos []PhotoResponseDTO `json:"photos"`
}

type ArticleResponseDTO struct {
	ID               uint                  `json:"id"`
	Title            entity.TranslatedText `json:"title"`
	Description      entity.TranslatedText `json:"description"`
	FullText         entity.TranslatedText `json:"full_text"`
	Link             string                `json:"link" example:"https://example.com"`
	Position         int                   `json:"position" example:"1"`
	CreatedAt        string                `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt        string                `json:"updated_at" example:"2021-01-01T00:00:00Z"`
	MainPhotoPath    string                `json:"main_photo_path,omitempty"`
	PreviewPhotoPath string                `json:"preview_photo_path,omitempty"`

	Photos []PhotoResponseDTO `json:"photos"`
}

func ToPressResponseDTO(press entity.Press, base_url string) PressResponseDTO {
	dto := PressResponseDTO{
		ID:               press.ID,
		Title:            press.Title,
		Description:      press.Description,
		FullText:         press.FullText,
		Link:             press.Link,
		Position:         press.Position,
		CreatedAt:        press.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        press.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		MainPhotoPath:    "",
		PreviewPhotoPath: "",
		Photos:           []PhotoResponseDTO{},
	}

	if press.MainPhotoID != nil && press.MainPhoto.Path != "" {
		path := strings.TrimPrefix(press.MainPhoto.Path, "/")
		dto.MainPhotoPath = fmt.Sprintf("%s/%s", base_url, path)
	}
	if press.PreviewPhotoID != nil && press.PreviewPhoto.Path != "" {
		path := strings.TrimPrefix(press.PreviewPhoto.Path, "/")
		dto.PreviewPhotoPath = fmt.Sprintf("%s/%s", base_url, path)
	}

	for _, photo := range press.Photos {
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

func ToArticleResponseDTO(article entity.Article, base_url string) ArticleResponseDTO {
	dto := ArticleResponseDTO{
		ID:               article.ID,
		Title:            article.Title,
		Description:      article.Description,
		FullText:         article.FullText,
		Link:             article.Link,
		Position:         article.Position,
		CreatedAt:        article.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        article.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		MainPhotoPath:    "",
		PreviewPhotoPath: "",
		Photos:           []PhotoResponseDTO{},
	}

	if article.MainPhotoID != nil && article.MainPhoto.Path != "" {
		path := strings.TrimPrefix(article.MainPhoto.Path, "/")
		dto.MainPhotoPath = fmt.Sprintf("%s/%s", base_url, path)
	}
	if article.PreviewPhotoID != nil && article.PreviewPhoto.Path != "" {
		path := strings.TrimPrefix(article.PreviewPhoto.Path, "/")
		dto.PreviewPhotoPath = fmt.Sprintf("%s/%s", base_url, path)
	}

	for _, photo := range article.Photos {
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

func ToPressResponseDTOs(presses []entity.Press, base_url string) []PressResponseDTO {
	pressDTOs := make([]PressResponseDTO, len(presses))
	for i, press := range presses {
		pressDTOs[i] = ToPressResponseDTO(press, base_url)
	}
	return pressDTOs
}

func ToArticleResponseDTOs(articles []entity.Article, base_url string) []ArticleResponseDTO {
	articleDTOs := make([]ArticleResponseDTO, len(articles))
	for i, article := range articles {
		articleDTOs[i] = ToArticleResponseDTO(article, base_url)
	}
	return articleDTOs
}
