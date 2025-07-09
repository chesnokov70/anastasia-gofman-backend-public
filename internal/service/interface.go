package service

import (
	"anastasia_gofman_backend/internal/delivery/http/dto"
	"anastasia_gofman_backend/internal/entity"
	"mime/multipart"
)

type AuthorService interface {
	GetAllAuthors(with_arts bool, page int, size int, with_pagination bool, full bool) ([]entity.Author, map[uint][]entity.Art, int64, int64, error)
	GetAuthorsBySpecialization(specializations []string, with_arts bool, page int, size int, with_pagination bool, full bool) ([]entity.Author, map[uint][]entity.Art, int64, int64, error)
	GetAuthorByID(id uint, with_arts bool) (entity.Author, map[uint][]entity.Art, error)
	CreateAuthor(author entity.Author) (entity.Author, error)
	UpdateAuthor(author entity.Author) (entity.Author, error)
	DeleteAuthor(id uint) error
	PartialUpdateAuthor(id uint, kwargs map[string]interface{}) (entity.Author, error)
	FullUpdateAuthor(author entity.Author) (entity.Author, error)
	AddMainOrPreviewPhotoToAuthor(authorID uint, fileHeader *multipart.FileHeader, is_main bool, is_preview bool) (entity.Author, error)
	AddPhotosToAuthor(id uint, photos []*multipart.FileHeader) (entity.Author, error)
	PatchAuthorPhotos(id uint, photos []*multipart.FileHeader) (entity.Author, error)
	GetMainPhoto(id uint) (entity.Photo, error)
	UpdateAuthorsPosition(positions []int) error
	GetAuthorWithArts(id uint) (entity.Author, []entity.Art, error)
	DeleteMainOrPreviewPhoto(id uint, is_main bool, is_preview bool) error
	PatchAuthorPhotosFromStrings(authorID uint, photoStrings []string) (entity.Author, error)
}

type ArtService interface {
	GetAllArts(page int, size int, with_pagination bool, sorting string, filtering *entity.ArtFilter, without_collection bool, with_type_discrimination bool) ([]entity.Art, int64, int64, error)
	GetArtByID(id uint) (entity.Art, error)
	GetArtByStripeProductID(stripeProductID string) (entity.Art, error)
	CreateArt(art entity.Art, with_stripe bool) (entity.Art, error)
	UpdateArt(art entity.Art) (entity.Art, error)
	DeleteArt(id uint) error
	PartialUpdateArt(id uint, kwargs map[string]interface{}, with_stripe bool) (entity.Art, error)
	FullUpdateArt(art entity.Art, with_stripe bool) (entity.Art, error)
	AddMainOrPreviewPhotoToArt(artID uint, fileHeader *multipart.FileHeader, is_main bool, is_preview bool) (entity.Art, error)
	AddPhotosToArt(id uint, photos []*multipart.FileHeader) (entity.Art, error)
	PatchArtPhotos(id uint, photos []*multipart.FileHeader) (entity.Art, error)
	AddAuthorToArt(id uint, author_id uint) (entity.Art, error)
	DeleteMainOrPreviewPhoto(id uint, is_preview bool) error
	GetMainPhoto(id uint) (entity.Photo, error)
	UpdateArtsPosition(positions []int) error
	UpdatePhotosInStripe(id uint) error
	DeleteProductInStripe(id uint) error
	GetMinAndMaxPrice() (int, int, error)
	DeleteArtsByCollectionID(id uint) error
	DeleteArtsByCollectionIDSync(id uint) error
	PatchArtPhotosFromStrings(artID uint, photoStrings []string) (entity.Art, error)
	ChangeArtTypeAfterBuy(stripeProductID string) (entity.Art, error)
}

type EventService interface {
	GetAllEvents(offset int, limit int) ([]entity.Event, int64, int64, error)
	GetEventByID(id uint) (entity.Event, error)
	GetMainPhoto(id uint) (entity.Photo, error)

	CreateEvent(event entity.Event) (entity.Event, error)
	FullUpdateEvent(event entity.Event) (entity.Event, error)
	UpdateEvent(event entity.Event) (entity.Event, error)
	UpdateEventsPosition(positions []int) error
	UpdateMainPhotoToEvent(id uint, fileHeader *multipart.FileHeader) (entity.Event, error)
	PartialUpdateEvent(id uint, kwargs map[string]interface{}) (entity.Event, error)

	AddMainOrPreviewPhotoToEvent(eventID uint, fileHeader *multipart.FileHeader, is_main bool, is_preview bool) (entity.Event, error)
	AddPhotosToEvent(id uint, photos []*multipart.FileHeader) (entity.Event, error)
	AddPhotosToEventReplaceOld(id uint, photos []*multipart.FileHeader) (entity.Event, error)
	PatchEventPhotosFromStrings(eventID uint, photoStrings []string) (entity.Event, error)

	DeleteMainOrPreviewPhoto(id uint, is_preview bool) error

	DeleteEvent(id uint) error
}

type ArtCollectionService interface {
	GetAllCollections(sorting string, with_arts bool) ([]entity.ArtCollection, error)
	GetCollectionByID(id uint, with_arts bool) (entity.ArtCollection, error)
	CreateCollection(collection entity.ArtCollection) (entity.ArtCollection, error)
	UpdateCollection(collection entity.ArtCollection) (entity.ArtCollection, error)
	DeleteCollection(id uint, delete_action string, artService ArtService) error
	PartialUpdateCollection(id uint, kwargs map[string]interface{}) (entity.ArtCollection, error)
	FullUpdateCollection(collection entity.ArtCollection) (entity.ArtCollection, error)
	AddArtsToCollection(id uint, arts []uint, remove_not_in_ids bool) (entity.ArtCollection, error)
}

type PressAndArticleService interface {
	GetAllPressAndArticles(page int, size int, with_pagination bool, article_or_press string, sorting string) ([]entity.Press, []entity.Article, int64, int64, error)
	GetPressOrArticleByID(id uint, article_or_press string) (*entity.Press, *entity.Article, error)
	CreatePressOrArticle(press_or_article string, press entity.Press, article entity.Article) (*entity.Press, *entity.Article, error)

	UpdatePressOrArticle(press_or_article string, press entity.Press, article entity.Article) (*entity.Press, *entity.Article, error)
	DeletePressOrArticle(press_or_article string, id uint) error
	PartialUpdatePressOrArticle(press_or_article string, id uint, kwargs map[string]interface{}) (*entity.Press, *entity.Article, error)
	FullUpdatePressOrArticle(press_or_article string, press entity.Press, article entity.Article) (*entity.Press, *entity.Article, error)

	DeleteMainOrPreviewPhoto(id uint, press_or_article string, is_preview bool) error
	AddMainOrPreviewPhotoToPressOrArticle(press_or_article string, id uint, fileHeader *multipart.FileHeader, is_main bool, is_preview bool) (*entity.Press, *entity.Article, error)
	AddPhotosToPressOrArticle(id uint, press_or_article string, photos []*multipart.FileHeader) (*entity.Press, *entity.Article, error)
	PatchPressOrArticlePhotos(id uint, press_or_article string, photos []*multipart.FileHeader) (*entity.Press, *entity.Article, error)
	GetMainPhoto(id uint, press_or_article string) (entity.Photo, error)
	DeleteAllPhotos(id uint, press_or_article string) error
	DeleteAllNoSpecialPhotos(id uint, press_or_article string) error
	PatchPhotosFromStrings(press_or_article string, objectID uint, photoStrings []string) (*entity.Press, *entity.Article, error)
}

type TranslationService interface {
	TranslateText(text string, languages []string) (map[string]string, error)
	AutoCompleteTranslation(text map[string]string, supportedLanguages []string, maxRetries int) (map[string]string, error)
	AutoCompleteTranslatedTextDTO(textDTO dto.TranslatedTextDTO, maxRetries int) (dto.TranslatedTextDTO, error)
	AutoCompleteEventTranslations(eventDTO *dto.CreateEventDTO, maxRetries int) error
	AutoCompleteEventWithPhotosTranslations(eventDTO *dto.CreateEventWithPhotosDTO, maxRetries int) error
}
