package postgres

import (
	"anastasia_gofman_backend/internal/entity"
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"
)

type AuthorRepository struct {
	db *gorm.DB
}

func (r *AuthorRepository) GetAllAuthors(offset int, limit int, with_pagination bool) ([]entity.Author, int64, error) {
	var authors []entity.Author
	var count int64
	query := r.db.Model(&entity.Author{}).Where("id != ?", 3333).Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
		return db.Order("photos.position ASC")
	}).Order("position ASC")

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	if with_pagination && limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}

	err = query.Find(&authors).Error
	return authors, count, err
}

func (r *AuthorRepository) GetAuthorsBySpecialization(specializations []string, offset int, limit int, with_pagination bool) ([]entity.Author, int64, error) {
	var authors []entity.Author
	var count int64

	query := r.db.Model(&entity.Author{}).Where("id != ?", 3333).Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
		return db.Order("photos.position ASC")
	}).Order("position ASC")

	if len(specializations) > 0 {
		for _, spec := range specializations {
			query = query.Where("specialization::jsonb ? ?", spec)
		}
	}

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	if with_pagination && limit > 0 {
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
	err := r.db.Model(&entity.Author{}).Where("id != ?", 3333).Count(&count).Error
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
	if id == 3333 {
		return errors.New("cannot delete default author")
	}
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
		err := r.db.Model(&entity.Author{}).Where("id = ? AND id != ?", i+1, 3333).Update("position", position).Error
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

func (r *AuthorRepository) CreateDefaultAuthor() {
	var count int64
	r.db.Model(&entity.Author{}).Where("id = ?", 3333).Count(&count)

	if count == 0 {
		defaultAuthor := entity.Author{
			ID: 3333,
			Name: entity.TranslatedText{
				EN: "Anastasia Gofman",
				RU: "Анастасия Гофман",
				ES: "Anastasia Gofman",
			},
			Bio: entity.TranslatedText{
				EN: "Artist and Gallery Owner",
				RU: "Художник и владелец галереи",
				ES: "Artista y propietario de galería",
			},
			Biography: entity.TranslatedText{
				EN: "Anastasia Gofman is a talented artist and the owner of this gallery.",
				RU: "Анастасия Гофман - талантливый художник и владелец этой галереи.",
				ES: "Anastasia Gofman es una artista talentosa y propietaria de esta galería.",
			},
			Contact: entity.ContactInfo{
				Email: "anastasia@gallery.com",
			},
			Position: 3333,
			IsActive: true,
		}

		if err := r.db.Create(&defaultAuthor).Error; err != nil {
			log.Printf("Warning: Failed to create default author: %v", err)
		} else {
			log.Printf("Default author 'Anastasia Gofman' created with ID: 3333")
		}
	}
}
