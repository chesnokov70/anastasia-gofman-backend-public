package postgres

import (
	"anastasia_gofman_backend/internal/entity"

	"gorm.io/gorm"
)

type ArtRepository struct {
	db *gorm.DB
}

func NewArtRepository(db *gorm.DB) *ArtRepository {
	return &ArtRepository{db: db}
}

func (r *ArtRepository) GetAllArts(offset int, limit int, with_pagination bool, sorting string, filtering *entity.ArtFilter, without_collection bool) ([]entity.Art, int64, error) {
	var arts []entity.Art
	var count int64
	query := r.db.Model(&entity.Art{}).Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
		return db.Order("photos.position ASC")
	})

	if filtering != nil {
		if filtering.PriceFrom != nil {
			query = query.Where("price >= ?", *filtering.PriceFrom)
		}
		if filtering.PriceTo != nil {
			query = query.Where("price <= ?", *filtering.PriceTo)
		}
		if filtering.Size != nil {
			query = query.Where("size = ?", *filtering.Size)
		}
		if filtering.Direction != nil {
			query = query.Where("direction = ?", *filtering.Direction)
		}
		if filtering.Style != nil {
			query = query.Where("style = ?", *filtering.Style)
		}
		if filtering.Author != nil {
			query = query.Joins("LEFT JOIN authors ON arts.author_id = authors.id").
				Where("authors.name->>'en' ILIKE ? OR authors.name->>'ru' ILIKE ? OR authors.name->>'es' ILIKE ?",
					"%"+*filtering.Author+"%", "%"+*filtering.Author+"%", "%"+*filtering.Author+"%")
		}
	}
	if without_collection {
		query = query.Where("collection_id IS NULL")
	}

	switch sorting {
	case "NEW":
		query = query.Order("created_at DESC")
	case "RATED":
		query = query.Order("position ASC")
	case "PRICE_HIGH":
		query = query.Order("price DESC")
	case "PRICE_LOW":
		query = query.Order("price ASC")
	default:
		query = query.Order("position ASC")
	}

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	if with_pagination {
		query = query.Offset(offset).Limit(limit)
	}

	err = query.Find(&arts).Error
	return arts, count, err
}

func (r *ArtRepository) GetArtByID(id uint) (entity.Art, error) {
	var art entity.Art
	err := r.db.Preload("Author").Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
		return db.Order("photos.position ASC")
	}).First(&art, id).Error
	return art, err
}

func (r *ArtRepository) CreateArt(art entity.Art) (entity.Art, error) {
	err := r.db.Create(&art).Error
	return art, err
}

func (r *ArtRepository) GetCountOfArts() (int, error) {
	var count int64
	err := r.db.Model(&entity.Art{}).Count(&count).Error
	return int(count), err
}

func (r *ArtRepository) UpdateArt(art entity.Art) (entity.Art, error) {
	err := r.db.Save(&art).Error
	return art, err
}

func (r *ArtRepository) DeleteArt(id uint) error {
	return r.db.Delete(&entity.Art{}, id).Error
}

func (r *ArtRepository) PartialUpdateArt(id uint, kwargs map[string]interface{}) (entity.Art, error) {
	var art entity.Art
	err := r.db.Model(&art).Where("id = ?", id).Updates(kwargs).Error
	if err != nil {
		return entity.Art{}, err
	}
	return r.GetArtByID(id)
}

func (r *ArtRepository) FullUpdateArt(art entity.Art) (entity.Art, error) {
	err := r.db.Model(&art).Where("id = ?", art.ID).Updates(art).Error
	if err != nil {
		return entity.Art{}, err
	}
	return r.GetArtByID(art.ID)
}

func (r *ArtRepository) AddMainOrPreviewPhotoToArt(photo entity.Photo) (entity.Art, error) {
	// err := r.db.Create(&photo).Error
	// if err != nil {
	// 	return entity.Art{}, err
	// }

	// Находим и обновляем art
	var art entity.Art
	which_photo := "main_photo_id"
	if photo.IsPreview {
		which_photo = "preview_photo_id"
	}
	err := r.db.Model(&entity.Art{}).Where("id = ?", photo.OwnerID).
		Update(which_photo, photo.ID).
		First(&art).Error
	if err != nil {
		return entity.Art{}, err
	}
	return r.GetArtByID(art.ID)
}

func (r *ArtRepository) GetCountOfPhotos(artID uint) int {
	var count int64
	r.db.Model(&entity.Photo{}).Where("art_id = ?", artID).Count(&count)
	return int(count)
}

func (r *ArtRepository) AddAuthorToArt(id uint, author_id uint) (entity.Art, error) {
	var art entity.Art
	err := r.db.Where("id = ?", id).Preload("Author").First(&art).Error
	if err != nil {
		return entity.Art{}, err
	}
	err = r.db.Model(&art).Association("Author").Replace(&entity.Author{ID: author_id})
	if err != nil {
		return entity.Art{}, err
	}
	return r.GetArtByID(id)
}

func (r *ArtRepository) UpdateArtsPosition(positions []int) error {
	for i, position := range positions {
		err := r.db.Model(&entity.Art{}).Where("id = ?", i+1).Update("position", position).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ArtRepository) RemoveMainAndPreviewPhotoFromArt(artID uint) error {
	return r.db.Model(&entity.Art{}).Where("id = ?", artID).Updates(map[string]interface{}{"main_photo_id": nil, "preview_photo_id": nil}).Error
}

func (r *ArtRepository) GetArtsByAuthorID(authorID uint) ([]entity.Art, error) {
	var arts []entity.Art
	err := r.db.Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
		return db.Order("photos.position ASC")
	}).Order("position ASC").Where("author_id = ?", authorID).Find(&arts).Error
	return arts, err
}

func (r *ArtRepository) SplitArtsByAuthors(authors []entity.Author) (map[uint][]entity.Art, error) {
	arts := make(map[uint][]entity.Art)
	var err error
	for _, author := range authors {
		arts[author.ID], err = r.GetArtsByAuthorID(author.ID)
		if err != nil {
			return nil, err
		}
	}
	return arts, nil
}

func (r *ArtRepository) GetMinAndMaxPrice() (int, int, error) {
	var result struct {
		MinPrice int `gorm:"column:min_price"`
		MaxPrice int `gorm:"column:max_price"`
	}

	err := r.db.Model(&entity.Art{}).Select("MIN(price) as min_price, MAX(price) as max_price").Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}

	return result.MinPrice, result.MaxPrice, nil
}

func (r *ArtRepository) GetArtsByCollectionID(collectionID uint) ([]entity.Art, error) {
	var arts []entity.Art
	err := r.db.Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
		return db.Order("photos.position ASC")
	}).Where("collection_id = ?", collectionID).Find(&arts).Error
	return arts, err
}

func (r *ArtRepository) RemoveCollectionFromArts(collectionID uint) error {
	return r.db.Model(&entity.Art{}).Where("collection_id = ?", collectionID).Update("collection_id", nil).Error
}

func (r *ArtRepository) DeleteArtsByCollectionID(collectionID uint) error {
	return r.db.Where("collection_id = ?", collectionID).Delete(&entity.Art{}).Error
}

// func (r *ArtRepository) AddPhotoToArt(photo entity.Photo) (entity.Art, error) {
// 	err := r.db.Create(&photo).Error
// 	if err != nil {
// 		return entity.Art{}, err
// 	}

// 	var art entity.Art
// 	err = r.db.Where("id = ?", photo.OwnerID).Preload("Photos").First(&art).Error
// 	if err != nil {
// 		return entity.Art{}, err
// 	}

// 	err = r.db.Model(&art).Association("Photos").Append(&photo)
// 	if err != nil {
// 		return entity.Art{}, err
// 	}

// 	return art, nil
// }

// func (r *ArtRepository) PatchArtPhotos(id uint, photos []entity.Photo) (entity.Art, error) {
// 	for _, photo := range photos {
// 		err := r.db.Create(&photo).Error
// 		if err != nil {
// 			return entity.Art{}, err
// 		}
// 	}
// 	var art entity.Art
// 	err := r.db.Where("id = ?", id).Preload("Photos").First(&art).Error
// 	if err != nil {
// 		return entity.Art{}, err
// 	}

// 	err = r.db.Model(&art).Association("Photos").Replace(photos)
// 	if err != nil {
// 		return entity.Art{}, err
// 	}
// 	return art, nil
// }
