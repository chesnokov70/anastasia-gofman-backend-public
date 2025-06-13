package dto

import (
	"anastasia_gofman_backend/internal/entity"
	"fmt"
	"strings"
)

// CreateAuthorDTO создает мальчишку - json в теле запроса
// @name CreateAuthor
type CreateAuthorDTO struct {
	Name                  TranslatedTextDTO `json:"name" binding:"required"`
	Bio                   TranslatedTextDTO `json:"bio"`
	Biography             TranslatedTextDTO `json:"biography"`
	EducationalBackground TranslatedTextDTO `json:"educational_background"`
	Exhibitions           TranslatedTextDTO `json:"exhibitions"`
	ContactInfo           TranslatedTextDTO `json:"contact_info"`
	Contact               struct {
		Email string `json:"email" binding:"email" example:"vovka@sosijopa.com"`
		Phone string `json:"phone" example:"+5553535"`
		Links struct {
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
		} `json:"links"`
	} `json:"contact"`
	IsActive bool `json:"is_active" example:"true"`
}

// UpdateAuthorDTO обновляет мальчишку - json в теле запроса
// @name UpdateAuthor
type UpdateAuthorDTO struct {
	Name                  TranslatedTextDTO `json:"name"`
	Bio                   TranslatedTextDTO `json:"bio"`
	Biography             TranslatedTextDTO `json:"biography"`
	EducationalBackground TranslatedTextDTO `json:"educational_background"`
	Exhibitions           TranslatedTextDTO `json:"exhibitions"`
	ContactInfo           TranslatedTextDTO `json:"contact_info"`
	Contact               struct {
		Email string `json:"email" binding:"email" example:"vovka@sosijopa.com"`
		Phone string `json:"phone" example:"+5553535"`
		Links struct {
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
		} `json:"links"`
	} `json:"contact"`
	IsActive bool `json:"is_active" example:"false"`
}

// CreateAuthorWithPhotosDTO создает мальчишку с фотками - multipart/form-data в поле main_photo, preview_photo, photos(массив)
// @name CreateAuthorWithPhotosDTO
type CreateAuthorWithPhotosDTO struct {
	Name                  TranslatedTextDTO `json:"name" binding:"required"`
	Bio                   TranslatedTextDTO `json:"bio"`
	Biography             TranslatedTextDTO `json:"biography"`
	EducationalBackground TranslatedTextDTO `json:"educational_background"`
	Exhibitions           TranslatedTextDTO `json:"exhibitions"`
	ContactInfo           TranslatedTextDTO `json:"contact_info"`
	Contact               struct {
		Email string `json:"email" binding:"email" example:"vovka@sosijopa.com"`
		Phone string `json:"phone" example:"+5553535"`
		Links struct {
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
		} `json:"links"`
	} `json:"contact"`
	IsActive bool `json:"is_active" example:"true"`
}

// AuthorResponseDTO ответ на запросы об авторе
// @name AuthorResponseDTO
type AuthorResponseDTO struct {
	ID                    uint                  `json:"id"`
	Name                  entity.TranslatedText `json:"name"`
	Bio                   entity.TranslatedText `json:"bio"`
	Biography             entity.TranslatedText `json:"biography"`
	EducationalBackground entity.TranslatedText `json:"educational_background"`
	Exhibitions           entity.TranslatedText `json:"exhibitions"`
	ContactInfo           entity.TranslatedText `json:"contact_info"`
	Contact               entity.ContactInfo    `json:"contact"`
	CreatedAt             string                `json:"created_at"`
	UpdatedAt             string                `json:"updated_at"`
	Position              int                   `json:"position"`
	IsActive              bool                  `json:"is_active"`
	MainPhotoPath         string                `json:"main_photo_path,omitempty"`
	PreviewPhotoPath      string                `json:"preview_photo_path,omitempty"`
	Photos                []PhotoResponseDTO    `json:"photos"`
}

func ToAuthorResponseDTO(author entity.Author) AuthorResponseDTO {
	dto := AuthorResponseDTO{
		ID:                    author.ID,
		Name:                  author.Name,
		Bio:                   author.Bio,
		Biography:             author.Biography,
		EducationalBackground: author.EducationalBackground,
		Exhibitions:           author.Exhibitions,
		ContactInfo:           author.ContactInfo,
		Contact:               author.Contact,
		Position:              author.Position,
		IsActive:              author.IsActive,
		CreatedAt:             author.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:             author.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Photos:                []PhotoResponseDTO{},
	}

	if author.MainPhotoID != nil && author.MainPhoto.Path != "" {
		path := strings.TrimPrefix(author.MainPhoto.Path, "/")
		dto.MainPhotoPath = fmt.Sprintf("%s/%s", BaseURL, path)
	}
	if author.PreviewPhotoID != nil && author.PreviewPhoto.Path != "" {
		path := strings.TrimPrefix(author.PreviewPhoto.Path, "/")
		dto.PreviewPhotoPath = fmt.Sprintf("%s/%s", BaseURL, path)
	}

	for _, photo := range author.Photos {
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

func ToAuthorResponseDTOs(authors []entity.Author) []AuthorResponseDTO {
	authorDTOs := make([]AuthorResponseDTO, len(authors))
	for i, author := range authors {
		authorDTOs[i] = ToAuthorResponseDTO(author)
	}
	return authorDTOs
}

type AuthorResponseWithArtsDTO struct {
	AuthorResponseDTO
	Arts []ArtResponseDTO `json:"arts"`
}

func ToAuthorResponseWithArtsDTO(author entity.Author, arts []entity.Art) AuthorResponseWithArtsDTO {
	return AuthorResponseWithArtsDTO{
		AuthorResponseDTO: ToAuthorResponseDTO(author),
		Arts:              ToArtResponseDTOs(arts),
	}
}

func ToAuthorResponseWithAllArtsDTOs(authors []entity.Author, arts map[uint][]entity.Art) []AuthorResponseWithArtsDTO {
	authorDTOs := make([]AuthorResponseWithArtsDTO, len(authors))
	for i, author := range authors {
		authorDTOs[i] = ToAuthorResponseWithArtsDTO(author, arts[author.ID])
	}
	return authorDTOs
}
