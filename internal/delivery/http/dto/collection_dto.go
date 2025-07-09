package dto

import (
	"anastasia_gofman_backend/internal/entity"
)

type CreateCollectionDTO struct {
	Name        entity.TranslatedText `json:"name" binding:"required"`
	Description entity.TranslatedText `json:"description"`
}

type UpdateCollectionDTO struct {
	Name           entity.TranslatedText `json:"name"`
	Description    entity.TranslatedText `json:"description"`
	ArtsIds        []uint                `json:"arts_ids"`
	RemoveNotInIds bool                  `json:"remove_not_in_ids"`
}

type CollectionResponseDTO struct {
	ID          uint                  `json:"id"`
	Name        entity.TranslatedText `json:"name"`
	Description entity.TranslatedText `json:"description"`
	Arts        []ArtResponseDTO      `json:"arts,omitempty"`
	CreatedAt   string                `json:"created_at"`
	UpdatedAt   string                `json:"updated_at"`
}

func (dto *CreateCollectionDTO) ToEntity() entity.ArtCollection {
	return entity.ArtCollection{
		Name:        dto.Name,
		Description: dto.Description,
	}
}

func (dto *UpdateCollectionDTO) ToEntity(id uint) entity.ArtCollection {
	return entity.ArtCollection{
		ID:          id,
		Name:        dto.Name,
		Description: dto.Description,
	}
}

func (dto *UpdateCollectionDTO) ToMap(id uint) map[string]interface{} {
	return map[string]interface{}{
		"id":          id,
		"name":        dto.Name,
		"description": dto.Description,
	}
}

func ToCollectionResponseDTO(collection entity.ArtCollection, base_url string) CollectionResponseDTO {
	dto := CollectionResponseDTO{
		ID:          collection.ID,
		Name:        collection.Name,
		Description: collection.Description,
		CreatedAt:   collection.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   collection.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Arts:        []ArtResponseDTO{},
	}

	for _, art := range collection.Arts {
		dto.Arts = append(dto.Arts, ToArtResponseDTO(art, base_url))
	}

	return dto
}

func ToCollectionResponseDTOs(collections []entity.ArtCollection, base_url string) []CollectionResponseDTO {
	collectionDTOs := make([]CollectionResponseDTO, len(collections))
	for i, collection := range collections {
		collectionDTOs[i] = ToCollectionResponseDTO(collection, base_url)
	}
	return collectionDTOs
}
