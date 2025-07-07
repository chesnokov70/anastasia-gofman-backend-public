// internal/repository/interface.go
package repository

import (
	"anastasia_gofman_backend/internal/entity"
)

type AuthorRepository interface {
	GetAllAuthors(offset int, limit int, with_pagination bool, full bool) ([]entity.Author, int64, error)
	GetAuthorsBySpecialization(specializations []string, offset int, limit int, with_pagination bool, full bool) ([]entity.Author, int64, error)
	GetAuthorByID(id uint) (entity.Author, error)
	GetCountOfAuthors() (int, error)
	CreateAuthor(author entity.Author) (entity.Author, error)
	UpdateAuthor(author entity.Author) (entity.Author, error)
	DeleteAuthor(id uint) error
	PartialUpdateAuthor(id uint, kwargs map[string]interface{}) (entity.Author, error)
	FullUpdateAuthor(author entity.Author) (entity.Author, error)
	AddMainOrPreviewPhotoToAuthor(photo entity.Photo) (entity.Author, error)
	UpdateAuthorsPosition(positions []int) error
	RemoveMainOrPreviewPhotoFromAuthor(authorID uint, isMain bool) error
	CreateDefaultAuthor()
}

type ArtRepository interface {
	GetAllArts(offset int, limit int, with_pagination bool, sorting string, filtering *entity.ArtFilter, without_collection bool, with_type_discrimination bool) ([]entity.Art, int64, error)
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
	RemoveSpecificPhotoFromArt(artID uint, is_main bool) error
	GetArtsByAuthorID(authorID uint) ([]entity.Art, error)
	SplitArtsByAuthors(authors []entity.Author) (map[uint][]entity.Art, error)
	GetMinAndMaxPrice() (int, int, error)

	GetArtsByCollectionID(collectionID uint) ([]entity.Art, error)
	RemoveCollectionFromArts(collectionID uint) error
	DeleteArtsByCollectionID(collectionID uint) error

	// AddPhotoToArt(photo entity.Photo) (entity.Art, error)
	// GetCountOfPhotos(artID uint) int
	// PatchArtPhotos(id uint, photos []entity.Photo) (entity.Art, error)
	// AddPhotosToArt(id uint, photos []*multipart.FileHeader) ([]entity.Art, error)
	// PatchArtPhotos(id uint, photos []*multipart.FileHeader) ([]entity.Art, error)
	// AddAuthorToArt(id uint, author_id uint) (entity.Art, error)
	// GetMainPhoto(id uint) (entity.Art, error)
}

type EventRepository interface {
	GetAllEvents(offset int, limit int) ([]entity.Event, int64, error)
	GetEventByID(id uint) (entity.Event, error)
	GetCountOfEvents() (int, error)

	CreateEvent(event entity.Event) (entity.Event, error)
	AddMainOrPreviewPhotoToEvent(eventID uint, photo entity.Photo) (entity.Event, error)

	DeleteEvent(id uint) error
	RemoveSpecificPhotoFromEvent(eventID uint, is_main bool) error

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
	CreatePhoto(photo entity.Photo) (entity.Photo, error)
	DeletePhoto(id uint) error
	GetMainOrPreviewPhotoByOwnerID(ownerID uint, ownerType string, isMain bool) (entity.Photo, error)
	GetAllPhotosByOwnerID(ownerID uint, ownerType string) ([]entity.Photo, error)
	GetCountOfPhotos(ownerID uint, ownerType string) (int, error)
	GetAllNoSpecialPhotosByOwnerID(ownerID uint, ownerType string) ([]entity.Photo, error)
	UpdatePhotoPosition(photoID uint, newPosition int) error
	GetPhotoByPath(path string) (entity.Photo, error)
	UpdatePhotoOwnerAndPosition(id uint, ownerID uint, ownerType string, position int) error

	GetMainPhotoByOwnerID(ownerID uint, ownerType string) (entity.Photo, error)
	GetPreviewPhotoByOwnerID(ownerID uint, ownerType string) (entity.Photo, error)

	DeleteAllPhotos(ownerID uint, ownerType string) error
	DeleteMainPhoto(ownerID uint, ownerType string) error
	DeletePreviewPhoto(ownerID uint, ownerType string) error
	DeleteAllNoSpecialPhotos(ownerID uint, ownerType string) error
}

type ArtCollectionRepository interface {
	GetAllCollections(sorting string, withArts bool) ([]entity.ArtCollection, error)
	GetCollectionByID(id uint, with_arts bool) (entity.ArtCollection, error)
	CreateCollection(collection entity.ArtCollection) (entity.ArtCollection, error)
	UpdateCollection(collection entity.ArtCollection) (entity.ArtCollection, error)
	DeleteCollection(id uint, delete_action string) error
	PartialUpdateCollection(id uint, kwargs map[string]interface{}) (entity.ArtCollection, error)
	FullUpdateCollection(collection entity.ArtCollection) (entity.ArtCollection, error)
	GetArtsByCollectionID(collectionID uint) ([]entity.Art, error)
	AddArtsToCollection(id uint, arts []uint) (entity.ArtCollection, error)
}

type PressAndArticleRepository interface {
	GetAllPressAndArticles(offset int, limit int, with_pagination bool, article_or_press string, sorting string) ([]entity.Press, []entity.Article, int64, error)
	GetPressOrArticleByID(id uint, article_or_press string) (*entity.Press, *entity.Article, error)
	CreatePressOrArticle(press_or_article string, press entity.Press, article entity.Article) (*entity.Press, *entity.Article, error)
	GetCountOfPressOrArticle(article_or_press string) (int, error)
	UpdatePressOrArticle(press_or_article string, press entity.Press, article entity.Article) (*entity.Press, *entity.Article, error)
	DeletePressOrArticle(press_or_article string, id uint) error
	PartialUpdatePressOrArticle(press_or_article string, id uint, kwargs map[string]interface{}) (*entity.Press, *entity.Article, error)
	FullUpdatePressOrArticle(press_or_article string, press *entity.Press, article *entity.Article) (*entity.Press, *entity.Article, error)
	AddMainOrPreviewPhotoToPressOrArticle(photo entity.Photo, press_or_article string) (*entity.Press, *entity.Article, error)
	RemoveMainAndPreviewPhotoFromPressOrArticle(press_or_article string, id uint) error
	RemoveSpecificPhotoFromPressOrArticle(press_or_article string, id uint, is_main bool) error
}
