package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/repository"
	"anastasia_gofman_backend/pkg/config"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"strings"

	"gorm.io/gorm"
)

type pressAndArticleService struct {
	pressAndArticleRepository repository.PressAndArticleRepository
	photoRepository           repository.PhotoRepository
}

func NewPressAndArticleService(pressAndArticleRepository repository.PressAndArticleRepository, photoRepository repository.PhotoRepository) PressAndArticleService {
	return &pressAndArticleService{
		pressAndArticleRepository: pressAndArticleRepository,
		photoRepository:           photoRepository,
	}
}

func (s *pressAndArticleService) GetAllPressAndArticles(page int, size int, with_pagination bool, article_or_press string) ([]entity.Press, []entity.Article, int64, int64, error) {
	offset, limit := 0, 0
	if page > 0 && size > 0 {
		offset = (page - 1) * size
		limit = size
	}
	press, articles, total, err := s.pressAndArticleRepository.GetAllPressAndArticles(offset, limit, with_pagination, article_or_press)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	var total_pages int64
	if total == 0 {
		total_pages = 0
	} else {
		total_pages = (int64(total) + int64(size) - 1) / int64(size)
	}
	return press, articles, total_pages, int64(total), nil
}

func (s *pressAndArticleService) GetPressOrArticleByID(id uint, article_or_press string) (*entity.Press, *entity.Article, error) {
	return s.pressAndArticleRepository.GetPressOrArticleByID(id, article_or_press)
}

func (s *pressAndArticleService) CreatePressOrArticle(press_or_article string, press entity.Press, article entity.Article) (*entity.Press, *entity.Article, error) {
	count, err := s.pressAndArticleRepository.GetCountOfPressOrArticle(press_or_article)
	if err != nil {
		count = 0
	}
	press.Position = count + 1
	article.Position = count + 1
	return s.pressAndArticleRepository.CreatePressOrArticle(press_or_article, press, article)
}

func (s *pressAndArticleService) UpdatePressOrArticle(press_or_article string, press entity.Press, article entity.Article) (*entity.Press, *entity.Article, error) {
	return s.pressAndArticleRepository.UpdatePressOrArticle(press_or_article, press, article)
}

func (s *pressAndArticleService) DeletePressOrArticle(press_or_article string, id uint) error {
	// if err := s.pressAndArticleRepository.RemoveMainAndPreviewPhotoFromArt(id); err != nil {
	// 	return err
	// }

	err1 := s.pressAndArticleRepository.DeletePressOrArticle(press_or_article, id)

	err2 := s.DeleteAllPhotos(id, press_or_article)
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

func (s *pressAndArticleService) PartialUpdatePressOrArticle(press_or_article string, id uint, kwargs map[string]interface{}) (*entity.Press, *entity.Article, error) {

	_, _, err := s.pressAndArticleRepository.GetPressOrArticleByID(id, press_or_article)
	if err != nil {
		return nil, nil, err
	}

	// Обработка main_photo
	if kwargs["main_photo"] != nil {
		mainPhotoHeader, ok := kwargs["main_photo"].(*multipart.FileHeader)
		if !ok {
			return nil, nil, errors.New("invalid main_photo format")
		}

		pos, _ := s.photoRepository.GetCountOfPhotos(id, press_or_article)
		main_photo, err := create_photo_from_http_photo(id, press_or_article, mainPhotoHeader, true, false, pos)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create main photo: %w", err)
		}

		if _, err := s.photoRepository.CreatePhoto(main_photo); err != nil {
			return nil, nil, fmt.Errorf("failed to save main photo: %w", err)
		}

		if _, _, err := s.pressAndArticleRepository.AddMainOrPreviewPhotoToPressOrArticle(main_photo, press_or_article); err != nil {
			return nil, nil, fmt.Errorf("failed to link main photo: %w", err)
		}
		delete(kwargs, "main_photo")
	}

	// Обработка preview_photo
	if kwargs["preview_photo"] != nil {
		previewPhotoHeader, ok := kwargs["preview_photo"].(*multipart.FileHeader)
		if !ok {
			return nil, nil, errors.New("invalid preview_photo format")
		}

		pos, _ := s.photoRepository.GetCountOfPhotos(id, press_or_article)
		preview_photo, err := create_photo_from_http_photo(id, press_or_article, previewPhotoHeader, false, true, pos)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create preview photo: %w", err)
		}

		if _, err := s.photoRepository.CreatePhoto(preview_photo); err != nil {
			return nil, nil, fmt.Errorf("failed to save preview photo: %w", err)
		}

		if _, _, err := s.pressAndArticleRepository.AddMainOrPreviewPhotoToPressOrArticle(preview_photo, press_or_article); err != nil {
			return nil, nil, fmt.Errorf("failed to link preview photo: %w", err)
		}
		delete(kwargs, "preview_photo")
	}

	// Обработка photos
	if kwargs["photos"] != nil {
		photos, ok := kwargs["photos"].([]*multipart.FileHeader)
		if !ok {
			return nil, nil, errors.New("invalid photos format")
		}

		if err := s.photoRepository.DeleteAllNoSpecialPhotos(id, press_or_article); err != nil {
			return nil, nil, fmt.Errorf("failed to delete old photos: %w", err)
		}

		pos, _ := s.photoRepository.GetCountOfPhotos(id, press_or_article)
		for i, photo := range photos {
			photoEntity, err := create_photo_from_http_photo(id, press_or_article, photo, false, false, pos+1+i)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create photo %d: %w", i, err)
			}
			if _, err := s.photoRepository.CreatePhoto(photoEntity); err != nil {
				return nil, nil, fmt.Errorf("failed to save photo %d: %w", i, err)
			}
		}
		delete(kwargs, "photos")
	}

	return s.pressAndArticleRepository.PartialUpdatePressOrArticle(press_or_article, id, kwargs)
}

func (s *pressAndArticleService) FullUpdatePressOrArticle(press_or_article string, press entity.Press, article entity.Article) (*entity.Press, *entity.Article, error) {
	return s.pressAndArticleRepository.FullUpdatePressOrArticle(press_or_article, &press, &article)
}

func (s *pressAndArticleService) AddMainOrPreviewPhotoToPressOrArticle(press_or_article string, id uint, fileHeader *multipart.FileHeader, is_main bool, is_preview bool) (*entity.Press, *entity.Article, error) {
	var oldPhoto *entity.Photo
	if is_main {
		if photo, err := s.photoRepository.GetMainOrPreviewPhotoByOwnerID(id, press_or_article, true); err == nil {
			oldPhoto = &photo
		}
	} else if is_preview {
		if photo, err := s.photoRepository.GetMainOrPreviewPhotoByOwnerID(id, press_or_article, false); err == nil {
			oldPhoto = &photo
		}
	}

	pos, _ := s.photoRepository.GetCountOfPhotos(id, press_or_article)
	main_photo, err := create_photo_from_http_photo(id, press_or_article, fileHeader, is_main, is_preview, pos)
	if err != nil {
		return nil, nil, err
	}
	photo, err := s.photoRepository.CreatePhoto(main_photo)
	if err != nil {
		return nil, nil, err
	}

	press, article, err := s.pressAndArticleRepository.AddMainOrPreviewPhotoToPressOrArticle(photo, press_or_article)
	if err != nil {
		return nil, nil, err
	}

	if oldPhoto != nil {

		filePath := oldPhoto.Path
		if strings.HasPrefix(filePath, "/uploads/") {
			filePath = strings.TrimPrefix(filePath, "/uploads/")
			filePath = config.GetUploadFilePath(strings.Split(filePath, "/")[0], strings.Join(strings.Split(filePath, "/")[1:], "/"))
		} else if strings.HasPrefix(filePath, "/") {
			filePath = filePath[1:]
		}

		if err := os.Remove(filePath); err != nil {
			fmt.Printf("ERROR DELETING OLD PHOTO FILE %s: %v\n", filePath, err)
		}

		if err := s.photoRepository.DeletePhoto(oldPhoto.ID); err != nil {
			fmt.Printf("ERROR DELETING OLD PHOTO FROM DB (ID %d): %v\n", oldPhoto.ID, err)
		}
	}

	return press, article, nil
}

func (s *pressAndArticleService) AddPhotosToPressOrArticle(id uint, press_or_article string, photos []*multipart.FileHeader) (*entity.Press, *entity.Article, error) {
	current_count_of_photos, err := s.photoRepository.GetCountOfPhotos(id, press_or_article)
	if err != nil {
		return nil, nil, err
	}
	for i, photo := range photos {
		photo, err := create_photo_from_http_photo(id, press_or_article, photo, false, false, i+current_count_of_photos)
		if err != nil {
			return nil, nil, err
		}
		_, err = s.photoRepository.CreatePhoto(photo)
		if err != nil {
			return nil, nil, err
		}
	}
	return s.pressAndArticleRepository.GetPressOrArticleByID(id, press_or_article)
}

func (s *pressAndArticleService) PatchPressOrArticlePhotos(id uint, press_or_article string, photos []*multipart.FileHeader) (*entity.Press, *entity.Article, error) {

	s.DeleteAllNoSpecialPhotos(id, press_or_article)
	current_count_of_photos, err := s.photoRepository.GetCountOfPhotos(id, press_or_article)
	if err != nil {
		return nil, nil, err
	}
	for i, photo := range photos {
		photo, err := create_photo_from_http_photo(id, press_or_article, photo, false, false, i+current_count_of_photos+1)
		if err != nil {
			return nil, nil, err
		}
		s.photoRepository.CreatePhoto(photo)
	}
	return s.pressAndArticleRepository.GetPressOrArticleByID(id, press_or_article)
}

func (s *pressAndArticleService) GetMainPhoto(id uint, press_or_article string) (entity.Photo, error) {
	press, article, err := s.pressAndArticleRepository.GetPressOrArticleByID(id, press_or_article)
	if err != nil {
		return entity.Photo{}, err
	}
	if press_or_article == "press" {
		return press.MainPhoto, nil
	} else {
		return article.MainPhoto, nil
	}
}

func (s *pressAndArticleService) DeleteAllPhotos(id uint, press_or_article string) error {
	photos, err := s.photoRepository.GetAllPhotosByOwnerID(id, press_or_article)
	if err != nil {
		return err
	}
	for _, photo := range photos {
		filePath := photo.Path
		if strings.HasPrefix(filePath, "/uploads/") {
			filePath = strings.TrimPrefix(filePath, "/uploads/")
			filePath = config.GetUploadFilePath(strings.Split(filePath, "/")[0], strings.Join(strings.Split(filePath, "/")[1:], "/"))
		} else if strings.HasPrefix(filePath, "/") {

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

func (s *pressAndArticleService) DeleteAllNoSpecialPhotos(id uint, press_or_article string) error {
	photos, err := s.photoRepository.GetAllNoSpecialPhotosByOwnerID(id, press_or_article)
	if err != nil {
		return err
	}
	for _, photo := range photos {
		filePath := photo.Path
		if strings.HasPrefix(filePath, "/uploads/") {

			filePath = strings.TrimPrefix(filePath, "/uploads/")

			filePath = config.GetUploadFilePath(strings.Split(filePath, "/")[0], strings.Join(strings.Split(filePath, "/")[1:], "/"))
		} else if strings.HasPrefix(filePath, "/") {
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

func (s *pressAndArticleService) DeleteMainOrPreviewPhoto(id uint, press_or_article string, type_of_photo string) error {
	var is_main bool = true
	if type_of_photo == "preview" {
		is_main = false
	}
	photo, err := s.photoRepository.GetMainOrPreviewPhotoByOwnerID(id, press_or_article, is_main)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		fmt.Println("ERROR DELETING PHOTO", err)
		return err
	}

	filePath := photo.Path
	if strings.HasPrefix(filePath, "/uploads/") {
		filePath = strings.TrimPrefix(filePath, "/uploads/")
		filePath = config.GetUploadFilePath(strings.Split(filePath, "/")[0], strings.Join(strings.Split(filePath, "/")[1:], "/"))
	} else if strings.HasPrefix(filePath, "/") {
		filePath = filePath[1:]
	}

	err = os.Remove(filePath)
	if err != nil {
		fmt.Printf("ERROR DELETING PHOTO %s: %v\n", filePath, err)
	}

	return s.photoRepository.DeletePhoto(photo.ID)
}
