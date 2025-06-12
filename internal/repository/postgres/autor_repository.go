package postgres

import (
	"anastasia_gofman_backend/internal/entity"
	"fmt"

	"gorm.io/gorm"
)

type AuthorRepository struct {
	db *gorm.DB
}

func (r *AuthorRepository) GetAllAuthors() ([]entity.Author, error) {
	var authors []entity.Author
	err := r.db.Find(&authors).Error
	return authors, err
}

func (r *AuthorRepository) GetAuthorByID(id uint) (entity.Author, error) {
	var author entity.Author
	err := r.db.First(&author, id).Error
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
	var author entity.Author
	err := r.db.Model(&author).Where("id = ?", id).Updates(kwargs).Error
	return author, err
}

func (r *AuthorRepository) FullUpdateAuthor(author entity.Author) (entity.Author, error) {
	fmt.Printf("%+v\n", author)
	if author.Position == 0 {
		pos, _ := r.GetCountOfAuthors()
		author.Position = pos
	}

	var existingAuthor entity.Author
	if err := r.db.First(&existingAuthor, author.ID).Error; err != nil {
		return author, err
	}
	author.CreatedAt = existingAuthor.CreatedAt
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

func NewAuthorRepository(db *gorm.DB) *AuthorRepository {
	return &AuthorRepository{db: db}
}
