package service

import (
	"anastasia_gofman_backend/internal/entity"
	"mime/multipart"
)

type AuthorService interface {
	GetAllAuthors() ([]entity.Author, error)
	GetAuthorByID(id uint) (entity.Author, error)
	CreateAuthor(author entity.Author) (entity.Author, error)
	UpdateAuthor(author entity.Author) (entity.Author, error)
	DeleteAuthor(id uint) error
	PartialUpdateAuthor(id uint, kwargs map[string]interface{}) (entity.Author, error)
	FullUpdateAuthor(author entity.Author) (entity.Author, error)
}

type ArtService interface {
	GetAllArts() ([]entity.Art, error)
	GetArtByID(id uint) (entity.Art, error)
	CreateArt(art entity.Art) (entity.Art, error)
	UpdateArt(art entity.Art) (entity.Art, error)
	DeleteArt(id uint) error
	PartialUpdateArt(id uint, kwargs map[string]interface{}) (entity.Art, error)
	FullUpdateArt(art entity.Art) (entity.Art, error)
	AddMainOrPreviewPhotoToArt(artID uint, fileHeader *multipart.FileHeader, is_main bool, is_preview bool) (entity.Art, error)
	AddPhotosToArt(id uint, photos []*multipart.FileHeader) (entity.Art, error)
	PatchArtPhotos(id uint, photos []*multipart.FileHeader) (entity.Art, error)
	AddAuthorToArt(id uint, author_id uint) (entity.Art, error)
	GetMainPhoto(id uint) (entity.Photo, error)
	// PatchArtPhotos(id uint, photos []*multipart.FileHeader) ([]entity.Art, error)
	UpdateArtsPosition(positions []int) error
}

type EventService interface {
	GetAllEvents() ([]entity.Event, error)
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
