package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/repository"
	"anastasia_gofman_backend/pkg/config"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

type authorService struct {
	authorRepository repository.AuthorRepository
	photoRepository  repository.PhotoRepository
	artRepository    repository.ArtRepository
}

func (s *authorService) GetAllAuthors(with_arts bool, page int, size int) ([]entity.Author, map[uint][]entity.Art, int64, error) {
	offset, limit := 0, 0
	if page > 0 && size > 0 {
		offset = (page - 1) * size
		limit = size
	}
	if !with_arts {
		all_autors, total, err := s.authorRepository.GetAllAuthors(offset, limit)
		if err != nil {
			return nil, nil, 0, err
		}
		var total_pages int64
		if total == 0 {
			total_pages = 0
		} else {
			total_pages = max(int64(total/int64(size)), 1)
		}
		return all_autors, nil, total_pages, nil
	}

	authors, total, err := s.authorRepository.GetAllAuthors(offset, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	arts := make(map[uint][]entity.Art)
	arts, err = s.artRepository.SplitArtsByAuthors(authors)
	if err != nil {
		return nil, nil, 0, err
	}
	var total_pages int64
	if total == 0 {
		total_pages = 0
	} else {
		total_pages = max(int64(total/int64(size)), 1)
	}
	return authors, arts, total_pages, nil
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
	err1 := s.authorRepository.DeleteAuthor(id)
	err2 := s.DeleteAllPhotos(id)
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

func (s *authorService) PartialUpdateAuthor(id uint, kwargs map[string]interface{}) (entity.Author, error) {
	if kwargs["main_photo"] != nil {
		pos, _ := s.photoRepository.GetCountOfPhotos(id, "authors")
		main_photo, err := create_photo_from_http_photo_author(id, "authors", kwargs["main_photo"].(*multipart.FileHeader), true, false, pos)
		if err != nil {
			return entity.Author{}, err
		}
		s.photoRepository.CreatePhoto(main_photo)
		s.authorRepository.AddMainOrPreviewPhotoToAuthor(main_photo)
		delete(kwargs, "main_photo")
	}
	if kwargs["preview_photo"] != nil {
		pos, _ := s.photoRepository.GetCountOfPhotos(id, "authors")
		preview_photo, err := create_photo_from_http_photo_author(id, "authors", kwargs["preview_photo"].(*multipart.FileHeader), false, true, pos)
		if err != nil {
			return entity.Author{}, err
		}
		s.photoRepository.CreatePhoto(preview_photo)
		s.authorRepository.AddMainOrPreviewPhotoToAuthor(preview_photo)
		delete(kwargs, "preview_photo")
	}
	if kwargs["photos"] != nil {
		photos := kwargs["photos"].([]*multipart.FileHeader)
		s.photoRepository.DeleteAllNoSpecialPhotos(id, "authors")
		pos, _ := s.photoRepository.GetCountOfPhotos(id, "authors")
		for i, photo := range photos {
			photo, err := create_photo_from_http_photo_author(id, "authors", photo, false, false, pos+1+i)
			if err != nil {
				return entity.Author{}, err
			}
			s.photoRepository.CreatePhoto(photo)
		}
		delete(kwargs, "photos")
	}
	return s.authorRepository.PartialUpdateAuthor(id, kwargs)
}

func (s *authorService) FullUpdateAuthor(author entity.Author) (entity.Author, error) {
	return s.authorRepository.FullUpdateAuthor(author)
}

func (s *authorService) AddMainOrPreviewPhotoToAuthor(authorID uint, fileHeader *multipart.FileHeader, is_main bool, is_preview bool) (entity.Author, error) {
	var oldPhoto entity.Photo
	var err error

	if is_main {
		oldPhoto, _ = s.photoRepository.GetMainOrPreviewPhotoByOwnerID(authorID, "authors", true)
	} else if is_preview {
		oldPhoto, _ = s.photoRepository.GetMainOrPreviewPhotoByOwnerID(authorID, "authors", false)
	}

	pos, _ := s.photoRepository.GetCountOfPhotos(authorID, "authors")
	newPhoto, err := create_photo_from_http_photo_author(authorID, "authors", fileHeader, is_main, is_preview, pos)
	if err != nil {
		return entity.Author{}, err
	}

	createdPhoto, err := s.photoRepository.CreatePhoto(newPhoto)
	if err != nil {
		deletePhotoFile(newPhoto.Path)
		return entity.Author{}, err
	}

	author, err := s.authorRepository.AddMainOrPreviewPhotoToAuthor(createdPhoto)
	if err != nil {
		s.photoRepository.DeletePhoto(createdPhoto.ID)
		deletePhotoFile(createdPhoto.Path)
		return entity.Author{}, err
	}

	if oldPhoto.ID != 0 {
		err := s.photoRepository.DeletePhoto(oldPhoto.ID)
		if err != nil {
			fmt.Printf("Failed to delete old photo record %d: %v\n", oldPhoto.ID, err)
		}
		err = deletePhotoFile(oldPhoto.Path)
		if err != nil {
			fmt.Printf("Failed to delete old photo file %s: %v\n", oldPhoto.Path, err)
		}
	}

	return author, nil
}

func create_photo_from_http_photo_author(OwnerID uint, OwnerType string, photo *multipart.FileHeader, is_main bool, is_preview bool, position_of_photo int) (entity.Photo, error) {
	file, err := photo.Open()
	if err != nil {
		return entity.Photo{}, err
	}
	defer file.Close()

	var subdir string
	var filename string
	if OwnerType == "authors" {
		subdir = "authors_photos"
		filename = fmt.Sprintf("author_%d_photo_%d%s", OwnerID, position_of_photo, filepath.Ext(photo.Filename))
	} else {
		return entity.Photo{}, errors.New("invalid type of photo")
	}

	fullPath := config.GetUploadFilePath(subdir, filename)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return entity.Photo{}, fmt.Errorf("failed to create directory: %w", err)
	}

	out, err := os.Create(fullPath)
	if err != nil {
		return entity.Photo{}, fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return entity.Photo{}, fmt.Errorf("failed to copy file: %w", err)
	}

	relativePath := fmt.Sprintf("/uploads/%s/%s", subdir, filename)

	res_photo := entity.Photo{
		Path:      relativePath,
		OwnerID:   OwnerID,
		OwnerType: OwnerType,
		IsMain:    is_main,
		IsPreview: is_preview,
	}
	return res_photo, nil
}

func deletePhotoFile(path string) error {
	filePath := path
	if strings.HasPrefix(filePath, "/uploads/") {
		filePath = strings.TrimPrefix(filePath, "/uploads/")
		parts := strings.Split(filePath, "/")
		filePath = config.GetUploadFilePath(parts[0], strings.Join(parts[1:], "/"))
	} else if strings.HasPrefix(filePath, "/") {
		filePath = filePath[1:]
	}

	err := os.Remove(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("PHOTO FILE NOT FOUND (already deleted?): %s\n", filePath)
			return nil
		}
		fmt.Printf("ERROR DELETING PHOTO FILE %s: %v\n", filePath, err)
		return err
	}

	fmt.Printf("PHOTO FILE DELETED: %s\n", filePath)
	return nil
}

func (s *authorService) AddPhotosToAuthor(id uint, photos []*multipart.FileHeader) (entity.Author, error) {
	current_count_of_photos, err := s.photoRepository.GetCountOfPhotos(id, "authors")
	if err != nil {
		return entity.Author{}, err
	}
	for i, photo := range photos {
		photo, err := create_photo_from_http_photo_author(id, "authors", photo, false, false, i+current_count_of_photos)
		if err != nil {
			return entity.Author{}, err
		}
		s.photoRepository.CreatePhoto(photo)
	}
	return s.authorRepository.GetAuthorByID(id)
}

func (s *authorService) PatchAuthorPhotos(id uint, photos []*multipart.FileHeader) (entity.Author, error) {
	s.DeleteAllNoSpecialPhotos(id)
	current_count_of_photos, err := s.photoRepository.GetCountOfPhotos(id, "authors")
	if err != nil {
		return entity.Author{}, err
	}
	for i, photo := range photos {
		photo, err := create_photo_from_http_photo_author(id, "authors", photo, false, false, i+current_count_of_photos+1)
		if err != nil {
			return entity.Author{}, err
		}
		s.photoRepository.CreatePhoto(photo)
	}
	return s.authorRepository.GetAuthorByID(id)
}

func (s *authorService) GetMainPhoto(id uint) (entity.Photo, error) {
	author, err := s.authorRepository.GetAuthorByID(id)
	if err != nil {
		return entity.Photo{}, err
	}
	return author.MainPhoto, nil
}

func (s *authorService) UpdateAuthorsPosition(positions []int) error {
	return s.authorRepository.UpdateAuthorsPosition(positions)
}

func (s *authorService) DeleteAllPhotos(id uint) error {
	photos, err := s.photoRepository.GetAllPhotosByOwnerID(id, "authors")
	if err != nil {
		return err
	}
	for _, photo := range photos {
		deletePhotoFile(photo.Path)
		s.photoRepository.DeletePhoto(photo.ID)
	}
	return nil
}

func (s *authorService) DeleteAllNoSpecialPhotos(id uint) error {
	photos, err := s.photoRepository.GetAllNoSpecialPhotosByOwnerID(id, "authors")
	if err != nil {
		return err
	}
	for _, photo := range photos {
		deletePhotoFile(photo.Path)
		s.photoRepository.DeletePhoto(photo.ID)
	}
	return nil
}

func (s *authorService) DeleteMainOrPreviewPhoto(id uint, is_main bool, is_preview bool) error {
	photo, err := s.photoRepository.GetMainOrPreviewPhotoByOwnerID(id, "authors", is_main)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		fmt.Printf("Error finding photo for author %d to delete: %v\n", id, err)
		return err
	}

	if err := s.authorRepository.RemoveMainOrPreviewPhotoFromAuthor(id, is_main); err != nil {
		fmt.Printf("Error unlinking photo from author %d: %v\n", id, err)
		return err
	}

	if err := s.photoRepository.DeletePhoto(photo.ID); err != nil {
		fmt.Printf("Error deleting photo record %d: %v\n", photo.ID, err)
		return err
	}

	if err := deletePhotoFile(photo.Path); err != nil {
		fmt.Printf("Failed to delete photo file %s, but database record is cleaned up: %v\n", photo.Path, err)
	}

	return nil
}

func (s *authorService) GetAuthorWithArts(id uint) (entity.Author, []entity.Art, error) {
	author, err := s.authorRepository.GetAuthorByID(id)
	if err != nil {
		return entity.Author{}, nil, err
	}
	arts, err := s.artRepository.GetArtsByAuthorID(id)
	if err != nil {
		return entity.Author{}, nil, err
	}
	return author, arts, nil
}

func NewAuthorService(authorRepository repository.AuthorRepository, photoRepository repository.PhotoRepository, artRepository repository.ArtRepository) AuthorService {
	return &authorService{
		authorRepository: authorRepository,
		photoRepository:  photoRepository,
		artRepository:    artRepository,
	}
}
