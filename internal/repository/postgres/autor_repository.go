package postgres

import (
	"anastasia_gofman_backend/internal/entity"
	"fmt"

	"gorm.io/gorm"
)

type AuthorRepository struct {
	db *gorm.DB
}

func (r *AuthorRepository) GetAllAuthors(offset int, limit int) ([]entity.Author, int64, error) {
	var authors []entity.Author
	var count int64
	query := r.db.Model(&entity.Author{}).Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
		return db.Order("photos.position ASC")
	}).Order("position ASC")

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}

	err = query.Find(&authors).Error
	return authors, count, err
}

func (r *AuthorRepository) GetAuthorByID(id uint) (entity.Author, error) {
	var author entity.Author
	err := r.db.Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
		return db.Order("photos.position ASC")
	}).First(&author, id).Error
	return author, err
}

func (r *AuthorRepository) GetCountOfAuthors() (int, error) {
	var count int64
	err := r.db.Model(&entity.Author{}).Count(&count).Error
	return int(count), err
}

func (r *AuthorRepository) CreateAuthor(author entity.Author) (entity.Author, error) {
	count, _ := r.GetCountOfAuthors()
	author.Position = count + 1
	err := r.db.Create(&author).Error
	return author, err
}

func (r *AuthorRepository) PartialUpdateAuthor(id uint, kwargs map[string]interface{}) (entity.Author, error) {
	if contact, ok := kwargs["contact"]; ok {
		if contactMap, ok := contact.(map[string]interface{}); ok {
			if email, ok := contactMap["email"]; ok {
				kwargs["email"] = email
			}
			if phone, ok := contactMap["phone"]; ok {
				kwargs["phone"] = phone
			}
			if links, ok := contactMap["links"]; ok {
				if linksMap, ok := links.(map[string]interface{}); ok {
					for key, value := range linksMap {
						kwargs[key] = value
					}
				}
			}
		}
		delete(kwargs, "contact")
	}

	err := r.db.Model(&entity.Author{}).Where("id = ?", id).Updates(kwargs).Error
	if err != nil {
		return entity.Author{}, err
	}
	return r.GetAuthorByID(id)
}

func (r *AuthorRepository) FullUpdateAuthor(author entity.Author) (entity.Author, error) {
	fmt.Printf("%+v\n", author)
	if author.Position == 0 {
		pos, _ := r.GetCountOfAuthors()
		author.Position = pos
	}

	var existingAuthor entity.Author
	if err := r.db.Preload("Photos").Preload("MainPhoto").Preload("PreviewPhoto").First(&existingAuthor, author.ID).Error; err != nil {
		return author, err
	}
	author.CreatedAt = existingAuthor.CreatedAt
	author.Photos = existingAuthor.Photos
	author.MainPhoto = existingAuthor.MainPhoto
	author.PreviewPhoto = existingAuthor.PreviewPhoto
	author.MainPhotoID = existingAuthor.MainPhotoID
	author.PreviewPhotoID = existingAuthor.PreviewPhotoID
	err := r.db.Save(&author).Error
	return author, err
}

func (r *AuthorRepository) UpdateAuthor(author entity.Author) (entity.Author, error) {
	err := r.db.Save(&author).Error
	return author, err
}

func (r *AuthorRepository) DeleteAuthor(id uint) error {
	return r.db.Delete(&entity.Author{}, id).Error
}

func (r *AuthorRepository) AddMainOrPreviewPhotoToAuthor(photo entity.Photo) (entity.Author, error) {
	var author entity.Author
	which_photo := "main_photo_id"
	if photo.IsPreview {
		which_photo = "preview_photo_id"
	}
	err := r.db.Model(&entity.Author{}).Where("id = ?", photo.OwnerID).
		Update(which_photo, photo.ID).
		First(&author).Error
	if err != nil {
		return entity.Author{}, err
	}
	return r.GetAuthorByID(author.ID)
}

func (r *AuthorRepository) UpdateAuthorsPosition(positions []int) error {
	for i, position := range positions {
		err := r.db.Model(&entity.Author{}).Where("id = ?", i+1).Update("position", position).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *AuthorRepository) RemoveMainOrPreviewPhotoFromAuthor(authorID uint, isMain bool) error {
	field := "preview_photo_id"
	if isMain {
		field = "main_photo_id"
	}
	return r.db.Model(&entity.Author{}).Where("id = ?", authorID).Update(field, nil).Error
}

func NewAuthorRepository(db *gorm.DB) *AuthorRepository {
	return &AuthorRepository{db: db}
}
