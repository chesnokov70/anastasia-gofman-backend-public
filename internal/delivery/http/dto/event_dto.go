package dto

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"anastasia_gofman_backend/internal/entity"
)

// FlexibleTime поддерживает несколько форматов дат
// @name FlexibleTime
type FlexibleTime struct {
	time.Time
}

// @name FlexibleTimeSwaggerDoc
// swagger:model FlexibleTime
type FlexibleTimeSwaggerDoc struct {
	Time string `json:"time" example:"2024-01-01T00:00:00Z"`
}

// UnmarshalJSON реализует кастомную логику парсинга для нескольких форматов дат
func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	// Убираем кавычки
	str := strings.Trim(string(data), `"`)

	// Поддерживаемые форматы
	formats := []string{
		"2006-01-02T15:04:05Z07:00", // RFC3339
		"2006-01-02T15:04:05Z",      // RFC3339 UTC
		"2006-01-02T15:04:05",       // ISO без таймзоны
		"2006-01-02 15:04:05",       // SQL datetime
		"2006-01-02",                // Только дата
		"02.01.2006",                // dd.mm.yyyy
		"02.01.2006 15:04:05",       // dd.mm.yyyy hh:mm:ss
		"02.01.2006 15:04",          // dd.mm.yyyy hh:mm
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			ft.Time = t
			return nil
		}
	}

	return fmt.Errorf("unable to parse time: %s", str)
}

// MarshalJSON сериализует время в RFC3339 формате
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ft.Time.Format(time.RFC3339))
}

// ToTime возвращает time.Time
func (ft FlexibleTime) ToTime() time.Time {
	return ft.Time
}

// @name CreateEvent
// CreateEventDTO defines the structure for creating a new event.
type CreateEventDTO struct {
	Title       TranslatedTextDTO `json:"title" binding:"required"`
	Description TranslatedTextDTO `json:"description"`
	Location    TranslatedTextDTO `json:"location"`
	StartDate   FlexibleTime      `json:"start_date" swaggertype:"string" example:"2024-01-01T00:00:00Z"`
	EndDate     FlexibleTime      `json:"end_date" swaggertype:"string" example:"2024-01-02T00:00:00Z"`
}

// @name CreateEventWithPhotos
// CreateEventWithPhotosDTO defines the structure for creating a new event with photos.
type CreateEventWithPhotosDTO struct {
	Title       TranslatedTextDTO `json:"title" binding:"required"`
	Description TranslatedTextDTO `json:"description"`
	Location    TranslatedTextDTO `json:"location"`
	StartDate   FlexibleTime      `json:"start_date" example:"2024-01-01T00:00:00Z или 01.01.2024"`
	EndDate     FlexibleTime      `json:"end_date" example:"2024-01-02T00:00:00Z или 02.01.2024"`
}

func (dto *CreateEventDTO) ToEntity(id *int) entity.Event {
	event := entity.Event{
		Title:       dto.Title.ToEntity(),
		Description: dto.Description.ToEntity(),
		Location:    dto.Location.ToEntity(),
		StartDate:   dto.StartDate.ToTime(),
		EndDate:     dto.EndDate.ToTime(),
	}
	if id != nil {
		event.ID = *id
	}
	return event
}

func (dto *CreateEventWithPhotosDTO) ToEntity(id *int) entity.Event {
	event := entity.Event{
		Title:       dto.Title.ToEntity(),
		Description: dto.Description.ToEntity(),
		Location:    dto.Location.ToEntity(),
		StartDate:   dto.StartDate.ToTime(),
		EndDate:     dto.EndDate.ToTime(),
	}
	if id != nil {
		event.ID = *id
	}
	return event
}

// @name UpdateEvent
// UpdateEventDTO defines the structure for updating an existing event.
type UpdateEventDTO struct {
	Title       TranslatedTextDTO `json:"title"`
	Description TranslatedTextDTO `json:"description"`
	Location    TranslatedTextDTO `json:"location"`
	StartDate   FlexibleTime      `json:"start_date" example:"2024-01-01T00:00:00Z или 01.01.2024"`
	EndDate     FlexibleTime      `json:"end_date" example:"2024-01-02T00:00:00Z или 02.01.2024"`
}

func (dto *UpdateEventDTO) ToEntity(id *int) entity.Event {
	event := entity.Event{
		Title:       dto.Title.ToEntity(),
		Description: dto.Description.ToEntity(),
		Location:    dto.Location.ToEntity(),
		StartDate:   dto.StartDate.ToTime(),
		EndDate:     dto.EndDate.ToTime(),
	}
	if id != nil {
		event.ID = *id
	}
	return event
}

// func (dto *UpdateArtDTO) ToEntity() entity.Art {
// 	art := entity.Art{
// 		// ID:          id,
// 		AuthorID:    dto.AuthorID,
// 		Name:        dto.Name.ToEntity(),
// 		Title:       dto.Title.ToEntity(),
// 		Description: dto.Description.ToEntity(),
// 		Medium:      dto.Medium.ToEntity(),
// 		Technique:   dto.Technique.ToEntity(),
// 		DimensionX:  dto.DimensionX,
// 		DimensionY:  dto.DimensionY,
// 		Year:        dto.Year,
// 		Frame:       dto.Frame.ToEntity(),
// 		Price:       dto.Price,
// 	}
// 	if id != nil {
// 		art.ID = *id
// 	}
// 	return art
// }

// @name EventResponse
type EventResponseDTO struct {
	ID               int                   `json:"id"`
	Title            entity.TranslatedText `json:"title"`
	Description      entity.TranslatedText `json:"description"`
	Location         entity.TranslatedText `json:"location"`
	StartDate        string                `json:"start_date"`
	EndDate          string                `json:"end_date"`
	MainPhotoPath    string                `json:"main_photo_path,omitempty"`
	PreviewPhotoPath string                `json:"preview_photo_path,omitempty"`
	Position         int                   `json:"position"`
	CreatedAt        string                `json:"created_at"`
	UpdatedAt        string                `json:"updated_at"`
	Photos           []PhotoResponseDTO    `json:"photos"`
}

func ToEventResponseDTO(event entity.Event) EventResponseDTO {
	dto := EventResponseDTO{
		ID:               event.ID,
		Title:            event.Title,
		Description:      event.Description,
		Location:         event.Location,
		StartDate:        event.StartDate.Format("2006-01-02T15:04:05Z07:00"),
		EndDate:          event.EndDate.Format("2006-01-02T15:04:05Z07:00"),
		Position:         event.Position,
		CreatedAt:        event.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        event.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		MainPhotoPath:    "",
		PreviewPhotoPath: "",
		Photos:           []PhotoResponseDTO{}, // Initialize with empty slice
	}

	if event.MainPhotoID != nil && event.MainPhoto.Path != "" {
		path := strings.TrimPrefix(event.MainPhoto.Path, "/")
		dto.MainPhotoPath = fmt.Sprintf("%s/%s", BaseURL, path)
	}
	if event.PreviewPhotoID != nil && event.PreviewPhoto.Path != "" {
		path := strings.TrimPrefix(event.PreviewPhoto.Path, "/")
		dto.PreviewPhotoPath = fmt.Sprintf("%s/%s", BaseURL, path)
	}

	for _, photo := range event.Photos {
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

func ToEventResponseDTOs(events []entity.Event) []EventResponseDTO {
	eventDTOs := make([]EventResponseDTO, len(events))
	for i, event := range events {
		eventDTOs[i] = ToEventResponseDTO(event)
	}
	return eventDTOs
}
