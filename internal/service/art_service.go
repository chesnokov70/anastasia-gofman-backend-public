package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/repository"
	"anastasia_gofman_backend/pkg/config"
	"anastasia_gofman_backend/pkg/service"
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
	stripeService   *service.StripeService
}

func NewArtService(artRepository repository.ArtRepository, photoRepository repository.PhotoRepository, stripeService *service.StripeService) ArtService {
	return &artService{
		artRepository:   artRepository,
		photoRepository: photoRepository,
		stripeService:   stripeService,
	}
}

func (s *artService) GetAllArts(page int, size int, with_pagination bool, sorting string, filtering *entity.ArtFilter) ([]entity.Art, int64, int64, error) {
	offset, limit := 0, 0
	if page > 0 && size > 0 {
		offset = (page - 1) * size
		limit = size
	}
	arts, total, err := s.artRepository.GetAllArts(offset, limit, with_pagination, sorting, filtering)
	if err != nil {
		return nil, 0, 0, err
	}
	var total_pages int64
	if total == 0 {
		total_pages = 0
	} else {
		total_pages = (int64(total) + int64(size) - 1) / int64(size)
	}
	return arts, total_pages, int64(total), nil
}

func (s *artService) GetArtByID(id uint) (entity.Art, error) {
	return s.artRepository.GetArtByID(id)
}

func (s *artService) CreateArt(art entity.Art, with_stripe bool) (entity.Art, error) {
	count, err := s.artRepository.GetCountOfArts()
	if err != nil {
		count = 0
	}
	art.Position = count + 1
	var product *service.ProductWithPriceAndLink = nil
	if with_stripe && art.Price != 0 {
		nameForStripe, descriptionForStripe := get_name_and_description_for_stripe(art)

		product, err = s.stripeService.CreateProduct(nameForStripe, descriptionForStripe, []string{}, int64(art.Price), "usd")

		if err != nil {
			return entity.Art{}, err
		}
		art.StripeProductID = product.Product.ID
		art.PaymentLink = product.Link.URL
	} else {
		art.StripeProductID = ""
		art.PaymentLink = ""
	}
	return s.artRepository.CreateArt(art)
}

func get_name_and_description_for_stripe(art entity.Art) (string, string) {
	nameForStripe := art.NameForStripe
	descriptionForStripe := art.DescriptionForStripe
	if nameForStripe == "" {
		nameForStripe = art.Name.EN
	}
	if descriptionForStripe == "" {
		descriptionForStripe = art.Description.EN
	}
	return nameForStripe, descriptionForStripe
}

func (s *artService) UpdateArt(art entity.Art) (entity.Art, error) {

	currency := "usd"
	active := true
	current_price := int64(art.Price)
	if current_price != 0 {
		nameForStripe, descriptionForStripe := get_name_and_description_for_stripe(art)
		product, err := s.stripeService.UpdateProduct(art.StripeProductID, &nameForStripe, &descriptionForStripe, &[]string{}, &current_price, &currency, &active)
		if err != nil {
			return entity.Art{}, err
		}
		art.StripeProductID = product.Product.ID
		art.PaymentLink = product.Link.URL
	}

	return s.artRepository.UpdateArt(art)
}

func (s *artService) DeleteArt(id uint) error {
	// if err := s.artRepository.RemoveMainAndPreviewPhotoFromArt(id); err != nil {
	// 	return err
	// }
	art, err := s.artRepository.GetArtByID(id)
	if err != nil {
		return err
	}
	if art.StripeProductID != "" {
		s.stripeService.DeleteProduct(art.StripeProductID)
	}
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

func (s *artService) PartialUpdateArt(id uint, kwargs map[string]interface{}, with_stripe bool) (entity.Art, error) {
	// Получаем текущее состояние artwork
	art, err := s.artRepository.GetArtByID(id)
	if err != nil {
		return entity.Art{}, err
	}

	// Обработка main_photo
	if kwargs["main_photo"] != nil {
		mainPhotoHeader, ok := kwargs["main_photo"].(*multipart.FileHeader)
		if !ok {
			return entity.Art{}, errors.New("invalid main_photo format")
		}

		pos, _ := s.photoRepository.GetCountOfPhotos(id, "arts")
		main_photo, err := create_photo_from_http_photo(id, "arts", mainPhotoHeader, true, false, pos)
		if err != nil {
			return entity.Art{}, fmt.Errorf("failed to create main photo: %w", err)
		}

		if _, err := s.photoRepository.CreatePhoto(main_photo); err != nil {
			return entity.Art{}, fmt.Errorf("failed to save main photo: %w", err)
		}

		if _, err := s.artRepository.AddMainOrPreviewPhotoToArt(main_photo); err != nil {
			return entity.Art{}, fmt.Errorf("failed to link main photo: %w", err)
		}
		delete(kwargs, "main_photo")
	}

	// Обработка preview_photo
	if kwargs["preview_photo"] != nil {
		previewPhotoHeader, ok := kwargs["preview_photo"].(*multipart.FileHeader)
		if !ok {
			return entity.Art{}, errors.New("invalid preview_photo format")
		}

		pos, _ := s.photoRepository.GetCountOfPhotos(id, "arts")
		preview_photo, err := create_photo_from_http_photo(id, "arts", previewPhotoHeader, false, true, pos)
		if err != nil {
			return entity.Art{}, fmt.Errorf("failed to create preview photo: %w", err)
		}

		if _, err := s.photoRepository.CreatePhoto(preview_photo); err != nil {
			return entity.Art{}, fmt.Errorf("failed to save preview photo: %w", err)
		}

		if _, err := s.artRepository.AddMainOrPreviewPhotoToArt(preview_photo); err != nil {
			return entity.Art{}, fmt.Errorf("failed to link preview photo: %w", err)
		}
		delete(kwargs, "preview_photo")
	}

	// Обработка photos
	if kwargs["photos"] != nil {
		photos, ok := kwargs["photos"].([]*multipart.FileHeader)
		if !ok {
			return entity.Art{}, errors.New("invalid photos format")
		}

		if err := s.photoRepository.DeleteAllNoSpecialPhotos(id, "arts"); err != nil {
			return entity.Art{}, fmt.Errorf("failed to delete old photos: %w", err)
		}

		pos, _ := s.photoRepository.GetCountOfPhotos(id, "arts")
		for i, photo := range photos {
			photoEntity, err := create_photo_from_http_photo(id, "arts", photo, false, false, pos+1+i)
			if err != nil {
				return entity.Art{}, fmt.Errorf("failed to create photo %d: %w", i, err)
			}
			if _, err := s.photoRepository.CreatePhoto(photoEntity); err != nil {
				return entity.Art{}, fmt.Errorf("failed to save photo %d: %w", i, err)
			}
		}
		delete(kwargs, "photos")
	}

	// Безопасное извлечение данных для Stripe
	var nameForStripe, descriptionForStripe *string
	var price *int64

	if val, exists := kwargs["name_for_stripe"]; exists && val != nil {
		if strPtr, ok := val.(*string); ok {
			nameForStripe = strPtr
		} else if str, ok := val.(string); ok {
			nameForStripe = &str
		} else {
			return entity.Art{}, errors.New("invalid name_for_stripe format")
		}
	}

	if val, exists := kwargs["description_for_stripe"]; exists && val != nil {
		if strPtr, ok := val.(*string); ok {
			descriptionForStripe = strPtr
		} else if str, ok := val.(string); ok {
			descriptionForStripe = &str
		} else {
			return entity.Art{}, errors.New("invalid description_for_stripe format")
		}
	}

	if val, exists := kwargs["price"]; exists && val != nil {
		if intPtr, ok := val.(*int64); ok {
			price = intPtr
		} else if intVal, ok := val.(int64); ok {
			price = &intVal
		} else if intVal, ok := val.(int); ok {
			int64Val := int64(intVal)
			price = &int64Val
		} else {
			return entity.Art{}, errors.New("invalid price format")
		}
	}

	// Обновление в Stripe только если есть изменения в соответствующих полях
	if with_stripe && (nameForStripe != nil || descriptionForStripe != nil || price != nil) {

		product, err := partial_update_stripe_product(s, art, nameForStripe, descriptionForStripe, price)
		if err != nil {
			return entity.Art{}, fmt.Errorf("failed to update stripe product: %w", err)
		}

		// Обновляем локальные данные для сохранения в БД
		if product != nil {
			if nameForStripe != nil {
				kwargs["name_for_stripe"] = *nameForStripe
			}
			if descriptionForStripe != nil {
				kwargs["description_for_stripe"] = *descriptionForStripe
			}
			if price != nil {
				kwargs["price"] = int(*price)
			}
			kwargs["stripe_product_id"] = product.Product.ID
			kwargs["payment_link"] = product.Link.URL
		}
	}

	// Финальное обновление в репозитории
	return s.artRepository.PartialUpdateArt(id, kwargs)
}

func partial_update_stripe_product(s *artService, art entity.Art, nameForStripe *string, descriptionForStripe *string, price *int64) (*service.ProductWithPriceAndLink, error) {
	currency := "usd"
	active := true
	product, err := s.stripeService.UpdateProduct(art.StripeProductID, nameForStripe, descriptionForStripe, &[]string{}, price, &currency, &active)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (s *artService) FullUpdateArt(art entity.Art, with_stripe bool) (entity.Art, error) {
	nameForStripe, descriptionForStripe := get_name_and_description_for_stripe(art)
	price := int64(art.Price)
	currency := "usd"
	active := true
	if with_stripe {
		product, err := s.stripeService.UpdateProduct(art.StripeProductID, &nameForStripe, &descriptionForStripe, &[]string{}, &price, &currency, &active)
		if err != nil {
			return entity.Art{}, err
		}
		if product != nil {
			art.StripeProductID = product.Product.ID
			art.PaymentLink = product.Link.URL
		}
	}
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

	var subdir string
	var filename string
	if OwnerType == "arts" {
		subdir = "arts_photos"
		filename = fmt.Sprintf("art_%d_photo_%d%s", OwnerID, position_of_photo, filepath.Ext(photo.Filename))
	} else if OwnerType == "event" {
		subdir = "events_photos"
		filename = fmt.Sprintf("event_%d_photo_%d%s", OwnerID, position_of_photo, filepath.Ext(photo.Filename))
	} else {
		return entity.Photo{}, errors.New("invalid type of photo")
	}

	// Получаем полный путь для сохранения файла
	fullPath := config.GetUploadFilePath(subdir, filename)

	// Создаем директорию если её нет
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

	// Для базы данных сохраняем относительный путь
	relativePath := fmt.Sprintf("/uploads/%s/%s", subdir, filename)

	var res_photo entity.Photo
	if OwnerType == "arts" {
		res_photo = entity.Photo{
			Path:      relativePath,
			OwnerID:   OwnerID,
			OwnerType: OwnerType,
			IsMain:    is_main,
			IsPreview: is_preview,
		}
	} else if OwnerType == "event" {
		res_photo = entity.Photo{
			Path:      relativePath,
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
		// Преобразуем путь из БД в реальный путь файла
		filePath := photo.Path
		if strings.HasPrefix(filePath, "/uploads/") {
			// Убираем /uploads/ из начала
			filePath = strings.TrimPrefix(filePath, "/uploads/")
			// Получаем полный путь
			filePath = config.GetUploadFilePath(strings.Split(filePath, "/")[0], strings.Join(strings.Split(filePath, "/")[1:], "/"))
		} else if strings.HasPrefix(filePath, "/") {
			// Старый формат пути
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

// func (s *artService) CreateArtWithPhotos(art entity.Art) (entity.Art, error) {

func (s *artService) UpdatePhotosInStripe(id uint) error {
	BaseURL := config.GetBaseURL()
	art, err := s.artRepository.GetArtByID(id)
	if err != nil {
		return err
	}
	list_of_photos := []string{}
	if art.MainPhotoID != nil && art.MainPhoto.Path != "" {
		path := strings.TrimPrefix(art.MainPhoto.Path, "/")
		list_of_photos = append(list_of_photos, fmt.Sprintf("%s/%s", BaseURL, path))
	}
	if art.PreviewPhotoID != nil && art.PreviewPhoto.Path != "" {
		path := strings.TrimPrefix(art.PreviewPhoto.Path, "/")
		list_of_photos = append(list_of_photos, fmt.Sprintf("%s/%s", BaseURL, path))
	}

	for _, photo := range art.Photos {
		if !photo.IsMain && !photo.IsPreview {
			path := strings.TrimPrefix(photo.Path, "/")
			fullPath := fmt.Sprintf("%s/%s", BaseURL, path)
			list_of_photos = append(list_of_photos, fullPath)
		}
	}
	current_price := int64(art.Price)
	currency := "usd"
	active := true
	_, err = s.stripeService.UpdateProduct(art.StripeProductID, &art.Name.EN, &art.Description.EN, &list_of_photos, &current_price, &currency, &active)
	if err != nil {
		return err
	}
	return nil
}

func (s *artService) DeleteProductInStripe(id uint) error {
	art, err := s.artRepository.GetArtByID(id)
	if err != nil {
		return err
	}
	return s.stripeService.DeleteProduct(art.StripeProductID)
}

func (s *artService) GetMinAndMaxPrice() (int, int, error) {
	min_price, max_price, err := s.artRepository.GetMinAndMaxPrice()
	if err != nil {
		return 0, 0, err
	}
	return min_price, max_price, nil
}
