package postgres

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/repository"
	"log"

	"gorm.io/gorm"
)

type ArtCollectionRepository struct {
	db            *gorm.DB
	artRepository repository.ArtRepository
}

func NewArtCollectionRepository(db *gorm.DB, artRepository repository.ArtRepository) repository.ArtCollectionRepository {
	return &ArtCollectionRepository{db: db, artRepository: artRepository}
}

func (r *ArtCollectionRepository) GetAllCollections(sorting string, withArts bool) ([]entity.ArtCollection, error) {
	var collections []entity.ArtCollection
	query := r.db.Model(&entity.ArtCollection{})

	if withArts {
		query = query.Preload("Arts", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Author").Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
				return db.Order("photos.position ASC")
			}).Order("position ASC")
		})
	}

	switch sorting {
	case "NEW":
		query = query.Order("created_at DESC")
	case "OLD":
		query = query.Order("created_at ASC")
	case "BIG":
		query = query.Order("(SELECT COUNT(*) FROM arts WHERE arts.collection_id = art_collections.id) DESC")
	case "SMALL":
		query = query.Order("(SELECT COUNT(*) FROM arts WHERE arts.collection_id = art_collections.id) ASC")
	case "AVALIBLE":
		query = query.Order("(SELECT COUNT(*) FROM arts WHERE arts.collection_id = art_collections.id) DESC")
	default:
		query = query.Order("created_at DESC")
	}

	err := query.Find(&collections).Error
	return collections, err
}

func (r *ArtCollectionRepository) GetCollectionByID(id uint, with_arts bool) (entity.ArtCollection, error) {
	var collection entity.ArtCollection
	err := r.db.First(&collection, id).Error
	if with_arts {
		err = r.db.Preload("Arts", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Author").Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
				return db.Order("photos.position ASC")
			}).Order("position ASC")
		}).First(&collection, id).Error
	}
	log.Println(collection)
	return collection, err
}

func (r *ArtCollectionRepository) CreateCollection(collection entity.ArtCollection) (entity.ArtCollection, error) {
	err := r.db.Create(&collection).Error
	return collection, err
}

func (r *ArtCollectionRepository) UpdateCollection(collection entity.ArtCollection) (entity.ArtCollection, error) {
	err := r.db.Save(&collection).Error
	return collection, err
}

func (r *ArtCollectionRepository) DeleteCollection(id uint, deleteAction string) error {

	return r.db.Transaction(func(tx *gorm.DB) error {

		if deleteAction == "SAVE_ART" {
			if err := r.artRepository.RemoveCollectionFromArts(id); err != nil {
				return err
			}
		}

		return tx.Delete(&entity.ArtCollection{}, id).Error
	})
}

func (r *ArtCollectionRepository) GetArtsByCollectionID(collectionID uint) ([]entity.Art, error) {
	var arts []entity.Art
	err := r.db.Preload("Author").Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
		return db.Order("photos.position ASC")
	}).Where("collection_id = ?", collectionID).Find(&arts).Error
	return arts, err
}

func (r *ArtCollectionRepository) PartialUpdateCollection(id uint, kwargs map[string]interface{}) (entity.ArtCollection, error) {
	var collection entity.ArtCollection
	err := r.db.First(&collection, id).Error
	if err != nil {
		return entity.ArtCollection{}, err
	}

	for key, value := range kwargs {
		if key == "name" {
			collection.Name = value.(entity.TranslatedText)
		}
		if key == "description" {
			collection.Description = value.(entity.TranslatedText)
		}
	}

	err = r.db.Save(&collection).Error
	return collection, err
}

func (r *ArtCollectionRepository) FullUpdateCollection(collection entity.ArtCollection) (entity.ArtCollection, error) {
	err := r.db.Save(&collection).Error
	return collection, err
}

func (r *ArtCollectionRepository) AddArtsToCollection(id uint, arts []uint) (entity.ArtCollection, error) {
	for _, art := range arts {
		err := r.db.Model(&entity.Art{}).Where("id = ?", art).Update("collection_id", id).Error
		if err != nil {
			return entity.ArtCollection{}, err
		}
	}
	return r.GetCollectionByID(id, true)
}
