package postgres

import (
	"anastasia_gofman_backend/internal/entity"
	"strings"

	"gorm.io/gorm"
)

type ArtRepository struct {
	db *gorm.DB
}

func NewArtRepository(db *gorm.DB) *ArtRepository {
	return &ArtRepository{db: db}
}

func (r *ArtRepository) GetAllArts(offset int, limit int, with_pagination bool, sorting string, filtering *entity.ArtFilter, without_collection bool, with_type_discrimination bool) ([]entity.Art, int64, error) {
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
		if filtering.Type != nil {
			if *filtering.Type == "" || *filtering.Type == "_common" {
				query = query.Where("type IS NULL OR type = ''")
			} else {
				query = query.Where("type = ?", *filtering.Type)
			}
		}
		if filtering.ArchiveType != nil {
			query = query.Where("archive_type = ?", *filtering.ArchiveType)
		}

		if filtering.Search != nil {
			searchConditions := []string{}
			searchArgs := []interface{}{}

			if filtering.Search.Name != nil {
				if *filtering.Search.Name != "" {
					searchConditions = append(searchConditions, "name->>'en' ILIKE ?")
					searchArgs = append(searchArgs, "%"+*filtering.Search.Name+"%")
					searchConditions = append(searchConditions, "name->>'ru' ILIKE ?")
					searchArgs = append(searchArgs, "%"+*filtering.Search.Name+"%")
					searchConditions = append(searchConditions, "name->>'es' ILIKE ?")
					searchArgs = append(searchArgs, "%"+*filtering.Search.Name+"%")
				}
			}

			if filtering.Search.Title != nil {
				if filtering.Search.Title.EN != "" {
					searchConditions = append(searchConditions, "title->>'en' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Title.EN+"%")
				}
				if filtering.Search.Title.RU != "" {
					searchConditions = append(searchConditions, "title->>'ru' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Title.RU+"%")
				}
				if filtering.Search.Title.ES != "" {
					searchConditions = append(searchConditions, "title->>'es' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Title.ES+"%")
				}
			}

			if filtering.Search.Description != nil {
				if filtering.Search.Description.EN != "" {
					searchConditions = append(searchConditions, "description->>'en' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Description.EN+"%")
				}
				if filtering.Search.Description.RU != "" {
					searchConditions = append(searchConditions, "description->>'ru' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Description.RU+"%")
				}
				if filtering.Search.Description.ES != "" {
					searchConditions = append(searchConditions, "description->>'es' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Description.ES+"%")
				}
			}

			if filtering.Search.Medium != nil {
				if filtering.Search.Medium.EN != "" {
					searchConditions = append(searchConditions, "medium->>'en' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Medium.EN+"%")
				}
				if filtering.Search.Medium.RU != "" {
					searchConditions = append(searchConditions, "medium->>'ru' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Medium.RU+"%")
				}
				if filtering.Search.Medium.ES != "" {
					searchConditions = append(searchConditions, "medium->>'es' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Medium.ES+"%")
				}
			}

			if filtering.Search.Technique != nil {
				if filtering.Search.Technique.EN != "" {
					searchConditions = append(searchConditions, "technique->>'en' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Technique.EN+"%")
				}
				if filtering.Search.Technique.RU != "" {
					searchConditions = append(searchConditions, "technique->>'ru' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Technique.RU+"%")
				}
				if filtering.Search.Technique.ES != "" {
					searchConditions = append(searchConditions, "technique->>'es' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Technique.ES+"%")
				}
			}

			if filtering.Search.Frame != nil {
				if filtering.Search.Frame.EN != "" {
					searchConditions = append(searchConditions, "frame->>'en' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Frame.EN+"%")
				}
				if filtering.Search.Frame.RU != "" {
					searchConditions = append(searchConditions, "frame->>'ru' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Frame.RU+"%")
				}
				if filtering.Search.Frame.ES != "" {
					searchConditions = append(searchConditions, "frame->>'es' ILIKE ?")
					searchArgs = append(searchArgs, "%"+filtering.Search.Frame.ES+"%")
				}
			}

			// Поиск по обычным полям
			if filtering.Search.Year != nil && *filtering.Search.Year != "" {
				searchConditions = append(searchConditions, "CAST(year AS TEXT) ILIKE ?")
				searchArgs = append(searchArgs, "%"+*filtering.Search.Year+"%")
			}

			if filtering.Search.Style != nil && *filtering.Search.Style != "" {
				searchConditions = append(searchConditions, "style ILIKE ?")
				searchArgs = append(searchArgs, "%"+*filtering.Search.Style+"%")
			}

			if filtering.Search.Direction != nil && *filtering.Search.Direction != "" {
				searchConditions = append(searchConditions, "direction ILIKE ?")
				searchArgs = append(searchArgs, "%"+*filtering.Search.Direction+"%")
			}

			if filtering.Search.DimensionStr != nil && *filtering.Search.DimensionStr != "" {
				searchConditions = append(searchConditions, "dimension_str ILIKE ?")
				searchArgs = append(searchArgs, "%"+*filtering.Search.DimensionStr+"%")
			}
			if filtering.IsCarousel != nil {
				query = query.Where("is_carousel = ?", *filtering.IsCarousel)
			}

			if len(searchConditions) > 0 {
				searchQuery := "(" + strings.Join(searchConditions, " OR ") + ")"
				query = query.Where(searchQuery, searchArgs...)
			}
		}
	}
	if without_collection {
		query = query.Where("collection_id IS NULL")
	}

	if with_type_discrimination {
		query = query.Where("type IS NULL OR type = '' OR type = 'archive'")
	}

	orderBy := ""
	switch sorting {
	case "NEW":
		orderBy = "created_at DESC"
	case "RATED":
		orderBy = "position ASC"
	case "PRICE_HIGH":
		orderBy = "price DESC"
	case "PRICE_LOW":
		orderBy = "price ASC"
	default:
		orderBy = "position ASC"
	}

	if with_type_discrimination {
		query = query.Order("CASE WHEN type IS NULL OR type = '' THEN 0 WHEN type = 'archive' THEN 1 END, " + orderBy)
	} else {
		query = query.Order(orderBy)
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

func (r *ArtRepository) GetArtByStripeProductID(stripeProductID string) (entity.Art, error) {
	var art entity.Art
	err := r.db.Where("stripe_product_id = ?", stripeProductID).
		Preload("Photos").
		Preload("Author").
		Preload("MainPhoto").
		Preload("PreviewPhoto").
		Preload("Collection").
		First(&art).Error
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
	err := r.db.Model(&art).Where("id = ?", art.ID).
		Select("author_id", "name", "title", "description", "medium", "technique",
			"dimension_x", "dimension_y", "dimension_str", "year", "frame", "price",
			"size", "direction", "style", "type", "archive_type", "collection_id",
			"name_for_stripe", "description_for_stripe", "stripe_product_id",
			"payment_link", "position", "is_carousel").
		Updates(art).Error
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

func (r *ArtRepository) RemoveSpecificPhotoFromArt(artID uint, is_main bool) error {
	var update_data = make(map[string]interface{})
	if is_main {
		update_data["main_photo_id"] = nil
	} else {
		update_data["preview_photo_id"] = nil
	}
	return r.db.Model(&entity.Art{}).Where("id = ?", artID).Updates(update_data).Error
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

// UpdateCarousel resets all current carousel flags and positions, then applies the new order
func (r *ArtRepository) UpdateCarousel(ids []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Art{}).
			Where("is_carousel = ?", true).
			Updates(map[string]interface{}{"is_carousel": false, "carousel_position": 0}).Error; err != nil {
			return err
		}

		for i, id := range ids {
			pos := i + 1
			if err := tx.Model(&entity.Art{}).Where("id = ?", id).
				Updates(map[string]interface{}{"is_carousel": true, "carousel_position": pos}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetCarousel returns arts marked for carousel ordered by carousel_position ASC
func (r *ArtRepository) GetCarousel() ([]entity.Art, error) {
	var arts []entity.Art
	err := r.db.Where("is_carousel = ?", true).
		Preload("MainPhoto").
		Order("carousel_position ASC").
		Find(&arts).Error
	return arts, err
}
