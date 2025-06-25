package service

import (
	"anastasia_gofman_backend/internal/entity"
	"mime/multipart"
)

type AuthorService interface {
	GetAllAuthors(with_arts bool, page int, size int, with_pagination bool) ([]entity.Author, map[uint][]entity.Art, int64, int64, error)
	GetAuthorsBySpecialization(specializations []string, with_arts bool, page int, size int, with_pagination bool) ([]entity.Author, map[uint][]entity.Art, int64, int64, error)
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
}

type ArtService interface {
	GetAllArts(page int, size int, with_pagination bool, sorting string, filtering *entity.ArtFilter, without_collection bool) ([]entity.Art, int64, int64, error)
	GetArtByID(id uint) (entity.Art, error)
	CreateArt(art entity.Art, with_stripe bool) (entity.Art, error)
	UpdateArt(art entity.Art) (entity.Art, error)
	DeleteArt(id uint) error
	PartialUpdateArt(id uint, kwargs map[string]interface{}, with_stripe bool) (entity.Art, error)
	FullUpdateArt(art entity.Art, with_stripe bool) (entity.Art, error)
	AddMainOrPreviewPhotoToArt(artID uint, fileHeader *multipart.FileHeader, is_main bool, is_preview bool) (entity.Art, error)
	AddPhotosToArt(id uint, photos []*multipart.FileHeader) (entity.Art, error)
	PatchArtPhotos(id uint, photos []*multipart.FileHeader) (entity.Art, error)
	AddAuthorToArt(id uint, author_id uint) (entity.Art, error)
	GetMainPhoto(id uint) (entity.Photo, error)
	UpdateArtsPosition(positions []int) error
	UpdatePhotosInStripe(id uint) error
	DeleteProductInStripe(id uint) error
	GetMinAndMaxPrice() (int, int, error)
	DeleteArtsByCollectionID(id uint) error
	DeleteArtsByCollectionIDSync(id uint) error
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
}
