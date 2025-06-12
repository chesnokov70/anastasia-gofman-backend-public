package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/repository"
)

type authorService struct {
	authorRepository repository.AuthorRepository
}

func (s *authorService) GetAllAuthors() ([]entity.Author, error) {
	return s.authorRepository.GetAllAuthors()
}

func (s *authorService) GetAuthorByID(id uint) (entity.Author, error) {
	return s.authorRepository.GetAuthorByID(id)
}

func (s *authorService) CreateAuthor(author entity.Author) (entity.Author, error) {
	return s.authorRepository.CreateAuthor(author)
}
func (s *authorService) UpdateAuthor(author entity.Author) (entity.Author, error) {
	return s.authorRepository.UpdateAuthor(author)
}
func (s *authorService) DeleteAuthor(id uint) error {
	return s.authorRepository.DeleteAuthor(id)
}
func (s *authorService) PartialUpdateAuthor(id uint, kwargs map[string]interface{}) (entity.Author, error) {
	return s.authorRepository.PartialUpdateAuthor(id, kwargs)
}
func (s *authorService) FullUpdateAuthor(author entity.Author) (entity.Author, error) {
	return s.authorRepository.FullUpdateAuthor(author)
}

func NewAuthorService(authorRepository repository.AuthorRepository) AuthorService {
	return &authorService{authorRepository: authorRepository}
}
