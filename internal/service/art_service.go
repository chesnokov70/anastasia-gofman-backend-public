package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/repository"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

type artService struct {
	artRepository   repository.ArtRepository
	photoRepository repository.PhotoRepository
}

func NewArtService(artRepository repository.ArtRepository, photoRepository repository.PhotoRepository) ArtService {
	return &artService{artRepository: artRepository, photoRepository: photoRepository}
}

func (s *artService) GetAllArts() ([]entity.Art, error) {

	return s.artRepository.GetAllArts()
}

func (s *artService) GetArtByID(id uint) (entity.Art, error) {
	return s.artRepository.GetArtByID(id)
}

func (s *artService) CreateArt(art entity.Art) (entity.Art, error) {
	count, err := s.artRepository.GetCountOfArts()
	if err != nil {
		count = 0
	}
	art.Position = count + 1
	return s.artRepository.CreateArt(art)
}

func (s *artService) UpdateArt(art entity.Art) (entity.Art, error) {
	return s.artRepository.UpdateArt(art)
}

func (s *artService) DeleteArt(id uint) error {
	// if err := s.artRepository.RemoveMainAndPreviewPhotoFromArt(id); err != nil {
	// 	return err
	// }
	err1 := s.artRepository.DeleteArt(id)

	err2 := s.DeleteAllPhotos(id)
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

func (s *artService) PartialUpdateArt(id uint, kwargs map[string]interface{}) (entity.Art, error) {
	if kwargs["main_photo"] != nil {
		pos, _ := s.photoRepository.GetCountOfPhotos(id, "arts")
		main_photo, err := create_photo_from_http_photo(id, "arts", kwargs["main_photo"].(*multipart.FileHeader), true, false, pos)
		if err != nil {
			return entity.Art{}, err
		}
		s.photoRepository.CreatePhoto(main_photo)
		s.artRepository.AddMainOrPreviewPhotoToArt(main_photo)
		delete(kwargs, "main_photo")
	}
	if kwargs["preview_photo"] != nil {
		pos, _ := s.photoRepository.GetCountOfPhotos(id, "arts")
		preview_photo, err := create_photo_from_http_photo(id, "arts", kwargs["preview_photo"].(*multipart.FileHeader), false, true, pos)
		if err != nil {
			return entity.Art{}, err
		}
		s.photoRepository.CreatePhoto(preview_photo)
		s.artRepository.AddMainOrPreviewPhotoToArt(preview_photo)
		delete(kwargs, "preview_photo")
	}
	if kwargs["photos"] != nil {
		photos := kwargs["photos"].([]*multipart.FileHeader)
		s.photoRepository.DeleteAllNoSpecialPhotos(id, "arts")
		pos, _ := s.photoRepository.GetCountOfPhotos(id, "arts")
		for i, photo := range photos {
			photo, err := create_photo_from_http_photo(id, "arts", photo, false, false, pos+1+i)
			if err != nil {
				return entity.Art{}, err
			}
			s.photoRepository.CreatePhoto(photo)
		}
		delete(kwargs, "photos")
	}
	return s.artRepository.PartialUpdateArt(id, kwargs)
}

func (s *artService) FullUpdateArt(art entity.Art) (entity.Art, error) {
	return s.artRepository.FullUpdateArt(art)
}

func (s *artService) AddMainOrPreviewPhotoToArt(artID uint, fileHeader *multipart.FileHeader, is_main bool, is_preview bool) (entity.Art, error) {
	if is_main {
		if err := s.DeleteMainOrPreviewPhoto(artID, "main"); err != nil {
			return entity.Art{}, err
		}
		s.photoRepository.DeleteMainPhoto(artID, "arts")
	} else if is_preview {
		if err := s.DeleteMainOrPreviewPhoto(artID, "preview"); err != nil {
			return entity.Art{}, err
		}
		s.photoRepository.DeletePreviewPhoto(artID, "arts")
	}
	// s.photoRepository.DeletePhoto(artID, is_main, is_preview)
	pos, _ := s.photoRepository.GetCountOfPhotos(artID, "arts")
	main_photo, err := create_photo_from_http_photo(artID, "arts", fileHeader, is_main, is_preview, pos)
	if err != nil {
		return entity.Art{}, err
	}
	main_photo, err = s.photoRepository.CreatePhoto(main_photo)
	if err != nil {
		return entity.Art{}, err
	}
	return s.artRepository.AddMainOrPreviewPhotoToArt(main_photo)
}

func create_photo_from_http_photo(OwnerID uint, OwnerType string, photo *multipart.FileHeader, is_main bool, is_preview bool, position_of_photo int) (entity.Photo, error) {

	file, err := photo.Open()
	if err != nil {
		return entity.Photo{}, err
	}
	defer file.Close()
	var filename string
	if OwnerType == "arts" {
		filename = fmt.Sprintf("uploads/arts_photos/art_%d_photo_%d%s", OwnerID, position_of_photo, filepath.Ext(photo.Filename))
	} else if OwnerType == "event" {
		filename = fmt.Sprintf("uploads/events_photos/event_%d_photo_%d%s", OwnerID, position_of_photo, filepath.Ext(photo.Filename))
	} else {
		return entity.Photo{}, errors.New("invalid type of photo")
	}

	out, err := os.Create(filename)
	if err != nil {
		return entity.Photo{}, err
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		return entity.Photo{}, err
	}
	var res_photo entity.Photo
	if OwnerType == "arts" {
		res_photo = entity.Photo{
			Path:      "/" + filename,
			OwnerID:   OwnerID,
			OwnerType: OwnerType,
			IsMain:    is_main,
			IsPreview: is_preview,
		}
	} else if OwnerType == "event" {
		res_photo = entity.Photo{
			Path:      "/" + filename,
			OwnerID:   OwnerID,
			OwnerType: OwnerType,
			IsMain:    is_main,
			IsPreview: is_preview,
		}
	}
	return res_photo, nil
}

func (s *artService) AddPhotosToArt(id uint, photos []*multipart.FileHeader) (entity.Art, error) {
	current_count_of_photos, err := s.photoRepository.GetCountOfPhotos(id, "arts")
	if err != nil {
		return entity.Art{}, err
	}
	for i, photo := range photos {
		photo, err := create_photo_from_http_photo(id, "arts", photo, false, false, i+current_count_of_photos)
		if err != nil {
			return entity.Art{}, err
		}
		s.photoRepository.CreatePhoto(photo)
	}
	return s.artRepository.GetArtByID(id)
}

func (s *artService) PatchArtPhotos(id uint, photos []*multipart.FileHeader) (entity.Art, error) {

	s.DeleteAllNoSpecialPhotos(id)
	current_count_of_photos, err := s.photoRepository.GetCountOfPhotos(id, "arts")
	if err != nil {
		return entity.Art{}, err
	}
	for i, photo := range photos {
		photo, err := create_photo_from_http_photo(id, "arts", photo, false, false, i+current_count_of_photos+1)
		if err != nil {
			return entity.Art{}, err
		}
		s.photoRepository.CreatePhoto(photo)
	}
	return s.artRepository.GetArtByID(id)
}

func (s *artService) AddAuthorToArt(id uint, author_id uint) (entity.Art, error) {
	art, err := s.artRepository.AddAuthorToArt(id, author_id)
	if err != nil {
		return entity.Art{}, err
	}
	return art, nil
}

func (s *artService) GetMainPhoto(id uint) (entity.Photo, error) {
	art, err := s.artRepository.GetArtByID(id)
	if err != nil {
		return entity.Photo{}, err
	}
	return art.MainPhoto, nil
}

func (s *artService) UpdateArtsPosition(positions []int) error {
	return s.artRepository.UpdateArtsPosition(positions)
}

func (s *artService) DeleteAllPhotos(id uint) error {
	photos, err := s.photoRepository.GetAllPhotosByOwnerID(id, "arts")
	if err != nil {
		return err
	}
	for _, photo := range photos {
		// Убираем ведущий слеш, чтобы получить относительный путь
		filePath := photo.Path
		if strings.HasPrefix(filePath, "/") {
			filePath = filePath[1:]
		}

		err := os.Remove(filePath)
		if err != nil {
			fmt.Printf("ERROR DELETING PHOTO %s: %v\n", filePath, err)
		} else {
			fmt.Printf("PHOTO DELETED: %s\n", filePath)
		}

		s.photoRepository.DeletePhoto(photo.ID)
	}
	return nil
}

func (s *artService) DeleteAllNoSpecialPhotos(id uint) error {
	photos, err := s.photoRepository.GetAllNoSpecialPhotosByOwnerID(id, "arts")
	if err != nil {
		return err
	}
	for _, photo := range photos {
		// Убираем ведущий слеш, чтобы получить относительный путь
		filePath := photo.Path
		if strings.HasPrefix(filePath, "/") {
			filePath = filePath[1:]
		}

		err := os.Remove(filePath)
		if err != nil {
			fmt.Printf("ERROR DELETING PHOTO %s: %v\n", filePath, err)
		}

		s.photoRepository.DeletePhoto(photo.ID)
	}
	return nil
}

func (s *artService) DeleteMainOrPreviewPhoto(id uint, type_of_photo string) error {
	var is_main bool = true
	if type_of_photo == "preview" {
		is_main = false
	}
	photo, err := s.photoRepository.GetMainOrPreviewPhotoByOwnerID(id, "arts", is_main)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		fmt.Println("ERROR DELETING PHOTO", err)
		return err
	}

	filePath := photo.Path
	if strings.HasPrefix(filePath, "/") {
		filePath = filePath[1:]
	}

	err = os.Remove(filePath)
	if err != nil {
		fmt.Printf("ERROR DELETING PHOTO %s: %v\n", filePath, err)
	}

	return s.photoRepository.DeletePhoto(photo.ID)
}

// func (s *artService) CreateArtWithPhotos(art entity.Art) (entity.Art, error) {
