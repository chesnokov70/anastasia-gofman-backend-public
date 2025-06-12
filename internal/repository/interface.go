// internal/repository/interface.go
package repository

import "anastasia_gofman_backend/internal/entity"

type AuthorRepository interface {
	GetAllAuthors() ([]entity.Author, error)
	GetAuthorByID(id uint) (entity.Author, error)
	CreateAuthor(author entity.Author) (entity.Author, error)
	UpdateAuthor(author entity.Author) (entity.Author, error)
	DeleteAuthor(id uint) error
	PartialUpdateAuthor(id uint, kwargs map[string]interface{}) (entity.Author, error)
	FullUpdateAuthor(author entity.Author) (entity.Author, error)
}

type ArtRepository interface {
	GetAllArts() ([]entity.Art, error)
	GetArtByID(id uint) (entity.Art, error)
	GetCountOfArts() (int, error)
	CreateArt(art entity.Art) (entity.Art, error)
	UpdateArt(art entity.Art) (entity.Art, error)
	DeleteArt(id uint) error
	PartialUpdateArt(id uint, kwargs map[string]interface{}) (entity.Art, error)
	FullUpdateArt(art entity.Art) (entity.Art, error)
	AddMainOrPreviewPhotoToArt(photo entity.Photo) (entity.Art, error)
	AddAuthorToArt(id uint, author_id uint) (entity.Art, error)
	UpdateArtsPosition(positions []int) error
	RemoveMainAndPreviewPhotoFromArt(artID uint) error

	// AddPhotoToArt(photo entity.Photo) (entity.Art, error)
	// GetCountOfPhotos(artID uint) int
	// PatchArtPhotos(id uint, photos []entity.Photo) (entity.Art, error)
	// AddPhotosToArt(id uint, photos []*multipart.FileHeader) ([]entity.Art, error)
	// PatchArtPhotos(id uint, photos []*multipart.FileHeader) ([]entity.Art, error)
	// AddAuthorToArt(id uint, author_id uint) (entity.Art, error)
	// GetMainPhoto(id uint) (entity.Art, error)
}

type EventRepository interface {
	GetAllEvents() ([]entity.Event, error)
	GetEventByID(id uint) (entity.Event, error)
	GetCountOfEvents() (int, error)

	CreateEvent(event entity.Event) (entity.Event, error)
	AddMainOrPreviewPhotoToEvent(eventID uint, photo entity.Photo) (entity.Event, error)

	DeleteEvent(id uint) error
	DeleteMainOrPreviewPhotoFromEvent(eventID uint, main_or_preview string) error

	FullUpdateEvent(event entity.Event) (entity.Event, error)
	UpdateEvent(event entity.Event) (entity.Event, error)
	UpdateEventsPosition(positions []int) error
	PartialUpdateEvent(id uint, kwargs map[string]interface{}) (entity.Event, error)

	// GetMainPhoto(eventID uint) (entity.Photo, error)
	// GetCountOfPhotos(eventID uint) (int, error)
	// GetPhoto(photoID uint) (entity.Photo, error)
	// GetPhotos(eventID uint) ([]entity.Photo, error)
	// DeletePhoto(photoID uint) error
	// DeleteAllPhotos(eventID uint) error
	// SavePhoto(photo entity.Photo) (entity.Photo, error)
	// PatchEventPhotos(eventID uint, photos []entity.Photo) (entity.Event, error)

	// AddPhotoToEvent(eventID uint, photo entity.Photo) (entity.Event, error)
	// AddPhotosToEvent(eventID uint, photos []entity.Photo) (entity.Event, error)
}

type PhotoRepository interface {
	// GET
	GetAllPhotosByOwnerID(ownerID uint, ownerType string) ([]entity.Photo, error)
	GetMainPhotoByOwnerID(ownerID uint, ownerType string) (entity.Photo, error)
	GetPreviewPhotoByOwnerID(ownerID uint, ownerType string) (entity.Photo, error)
	GetMainOrPreviewPhotoByOwnerID(ownerID uint, ownerType string, isMain bool) (entity.Photo, error)
	GetCountOfPhotos(ownerID uint, ownerType string) (int, error)
	GetAllNoSpecialPhotosByOwnerID(ownerID uint, ownerType string) ([]entity.Photo, error)

	// Delete
	DeletePhoto(photoID uint) error
	DeleteAllPhotos(ownerID uint, ownerType string) error
	DeleteMainPhoto(ownerID uint, ownerType string) error
	DeletePreviewPhoto(ownerID uint, ownerType string) error
	DeleteAllNoSpecialPhotos(ownerID uint, ownerType string) error

	// Create
	CreatePhoto(photo entity.Photo) (entity.Photo, error)
}
