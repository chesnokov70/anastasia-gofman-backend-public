package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/repository"
	"anastasia_gofman_backend/pkg/config"
	"encoding/base64"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type eventService struct {
	eventRepository repository.EventRepository
	photoRepository repository.PhotoRepository
}

func NewEventService(eventRepository repository.EventRepository, photoRepository repository.PhotoRepository) EventService {
	return &eventService{eventRepository: eventRepository, photoRepository: photoRepository}
}
func (s *eventService) GetAllEvents(offset int, limit int) ([]entity.Event, int64, int64, error) {
	events, total, err := s.eventRepository.GetAllEvents(offset, limit)
	if err != nil {
		return nil, 0, 0, err
	}

	var total_pages int64
	if total == 0 {
		total_pages = 0
	} else {
		pageSize := limit
		if pageSize > 0 {
			total_pages = (int64(total) + int64(pageSize) - 1) / int64(pageSize)
		} else {
			total_pages = 1
		}
	}
	return events, total_pages, int64(total), nil
}

func (s *eventService) GetEventByID(id uint) (entity.Event, error) {
	return s.eventRepository.GetEventByID(id)
}

func (s *eventService) CreateEvent(event entity.Event) (entity.Event, error) {
	count, err := s.eventRepository.GetCountOfEvents()
	if err != nil {
		count = 0
	}
	event.Position = count + 1
	return s.eventRepository.CreateEvent(event)
}

func (s *eventService) PartialUpdateEvent(id uint, kwargs map[string]interface{}) (entity.Event, error) {
	if kwargs["main_photo"] != nil {
		main_photo, err := create_photo_from_http_photo(id, "event", kwargs["main_photo"].(*multipart.FileHeader), true, false)
		if err != nil {
			return entity.Event{}, err
		}
		s.eventRepository.RemoveSpecificPhotoFromEvent(id, true)
		s.photoRepository.CreatePhoto(main_photo)
		s.eventRepository.AddMainOrPreviewPhotoToEvent(id, main_photo)
		delete(kwargs, "main_photo")
	}
	if kwargs["preview_photo"] != nil {
		preview_photo, err := create_photo_from_http_photo(id, "event", kwargs["preview_photo"].(*multipart.FileHeader), false, true)
		if err != nil {
			return entity.Event{}, err
		}
		s.eventRepository.RemoveSpecificPhotoFromEvent(id, false)
		s.photoRepository.CreatePhoto(preview_photo)
		s.eventRepository.AddMainOrPreviewPhotoToEvent(id, preview_photo)
		delete(kwargs, "preview_photo")
	}
	if kwargs["photos"] != nil {
		photos := kwargs["photos"].([]*multipart.FileHeader)
		s.photoRepository.DeleteAllNoSpecialPhotos(id, "event")
		for _, photo := range photos {
			photo, err := create_photo_from_http_photo(id, "event", photo, false, false)
			if err != nil {
				return entity.Event{}, err
			}
			s.photoRepository.CreatePhoto(photo)
		}
		delete(kwargs, "photos")
	}
	return s.eventRepository.PartialUpdateEvent(id, kwargs)
}

func (s *eventService) FullUpdateEvent(event entity.Event) (entity.Event, error) {
	return s.eventRepository.FullUpdateEvent(event)
}

func (s *eventService) UpdateEvent(event entity.Event) (entity.Event, error) {
	return s.eventRepository.UpdateEvent(event)
}

func (s *eventService) AddMainOrPreviewPhotoToEvent(eventID uint, fileHeader *multipart.FileHeader, is_main bool, is_preview bool) (entity.Event, error) {
	var oldPhoto *entity.Photo
	if is_main {
		if photo, err := s.photoRepository.GetMainOrPreviewPhotoByOwnerID(eventID, "event", true); err == nil {
			oldPhoto = &photo
		}
	} else if is_preview {
		if photo, err := s.photoRepository.GetMainOrPreviewPhotoByOwnerID(eventID, "event", false); err == nil {
			oldPhoto = &photo
		}
	}

	photo, err := create_photo_from_http_photo(eventID, "event", fileHeader, is_main, is_preview)
	if err != nil {
		return entity.Event{}, err
	}
	photo, err = s.photoRepository.CreatePhoto(photo)
	if err != nil {
		return entity.Event{}, err
	}

	event, err := s.eventRepository.AddMainOrPreviewPhotoToEvent(eventID, photo)
	if err != nil {
		return entity.Event{}, err
	}

	if oldPhoto != nil {
		filePath := oldPhoto.Path
		if strings.HasPrefix(filePath, "/") {
			filePath = filePath[1:]
		}

		if err := os.Remove(filePath); err != nil {
			fmt.Printf("ERROR DELETING OLD PHOTO FILE %s: %v\n", filePath, err)
		}

		if err := s.photoRepository.DeletePhoto(oldPhoto.ID); err != nil {
			fmt.Printf("ERROR DELETING OLD PHOTO FROM DB (ID %d): %v\n", oldPhoto.ID, err)
		}
	}

	return event, nil
}

func (s *eventService) AddPhotosToEvent(id uint, photos []*multipart.FileHeader) (entity.Event, error) {
	for _, photo := range photos {
		photo, err := create_photo_from_http_photo(id, "event", photo, false, false)
		if err != nil {
			return entity.Event{}, err
		}
		s.photoRepository.CreatePhoto(photo)
	}
	return s.GetEventByID(id)
}

func (s *eventService) AddPhotosToEventReplaceOld(id uint, photos []*multipart.FileHeader) (entity.Event, error) {

	s.DeleteAllNoSpecialPhotos(id)
	for _, photo := range photos {
		photo, err := create_photo_from_http_photo(id, "event", photo, false, false)
		if err != nil {
			return entity.Event{}, err
		}
		s.photoRepository.CreatePhoto(photo)
	}
	return s.GetEventByID(id)
}

func (s *eventService) UpdateEventsPosition(positions []int) error {
	return s.eventRepository.UpdateEventsPosition(positions)
}

func (s *eventService) UpdateMainPhotoToEvent(id uint, fileHeader *multipart.FileHeader) (entity.Event, error) {
	var oldPhoto *entity.Photo
	if photo, err := s.photoRepository.GetMainOrPreviewPhotoByOwnerID(id, "event", true); err == nil {
		oldPhoto = &photo
	}

	photo, err := create_photo_from_http_photo(id, "event", fileHeader, true, false)
	if err != nil {
		return entity.Event{}, err
	}
	photo, err = s.photoRepository.CreatePhoto(photo)
	if err != nil {
		return entity.Event{}, err
	}

	event, err := s.eventRepository.AddMainOrPreviewPhotoToEvent(id, photo)
	if err != nil {
		return entity.Event{}, err
	}

	if oldPhoto != nil {

		filePath := oldPhoto.Path
		if strings.HasPrefix(filePath, "/") {
			filePath = filePath[1:]
		}

		if err := os.Remove(filePath); err != nil {
			fmt.Printf("ERROR DELETING OLD PHOTO FILE %s: %v\n", filePath, err)
		}

		if err := s.photoRepository.DeletePhoto(oldPhoto.ID); err != nil {
			fmt.Printf("ERROR DELETING OLD PHOTO FROM DB (ID %d): %v\n", oldPhoto.ID, err)
		}
	}

	return event, nil
}

func (s *eventService) GetMainPhoto(id uint) (entity.Photo, error) {
	return s.photoRepository.GetMainOrPreviewPhotoByOwnerID(id, "event", true)
}

func (s *eventService) DeleteAllPhotos(id uint) error {
	photos, err := s.photoRepository.GetAllPhotosByOwnerID(id, "event")
	if err != nil {
		return err
	}
	for _, photo := range photos {
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

func (s *eventService) DeleteAllNoSpecialPhotos(id uint) error {
	photos, err := s.photoRepository.GetAllNoSpecialPhotosByOwnerID(id, "event")
	if err != nil {
		return err
	}
	for _, photo := range photos {
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

func (s *eventService) DeleteMainOrPreviewPhoto(id uint, is_preview bool) error {
	photo, err := s.photoRepository.GetMainOrPreviewPhotoByOwnerID(id, "event", !is_preview)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		fmt.Printf("ERROR DELETING PHOTO %v\n", err)
		return err
	}

	if err := s.eventRepository.RemoveSpecificPhotoFromEvent(id, !is_preview); err != nil {
		return fmt.Errorf("failed to remove photo reference from event: %w", err)
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

func (s *eventService) DeleteEvent(id uint) error {
	err := s.DeleteAllPhotos(id)
	if err != nil {
		return err
	}
	return s.eventRepository.DeleteEvent(id)
}

func (s *eventService) PatchEventPhotosFromStrings(eventID uint, photoStrings []string) (entity.Event, error) {
	currentEvent, err := s.GetEventByID(eventID)
	if err != nil {
		return entity.Event{}, err
	}

	currentEventPhotosByPath := make(map[string]entity.Photo)
	for _, p := range currentEvent.Photos {
		if !p.IsMain && !p.IsPreview {
			currentEventPhotosByPath[p.Path] = p
		}
	}

	requestedPaths := make(map[string]bool)

	for i, photoString := range photoStrings {
		position := i
		if strings.HasPrefix(photoString, "http") {
			idx := strings.Index(photoString, "/uploads/")
			if idx == -1 {
				return entity.Event{}, fmt.Errorf("invalid photo URL format: %s", photoString)
			}
			relativePath := photoString[idx:]
			relativePath = path.Clean(relativePath)

			requestedPaths[relativePath] = true

			photo, err := s.photoRepository.GetPhotoByPath(relativePath)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return entity.Event{}, fmt.Errorf("photo with path %s not found in database", relativePath)
				}
				return entity.Event{}, err // other db error
			}

			if photo.OwnerID != eventID || photo.OwnerType != "event" {
				err = s.photoRepository.UpdatePhotoOwnerAndPosition(photo.ID, eventID, "event", position)
				if err != nil {
					return entity.Event{}, fmt.Errorf("failed to move photo %d to event %d: %w", photo.ID, eventID, err)
				}
			} else if photo.Position != position {
				err = s.photoRepository.UpdatePhotoPosition(photo.ID, position)
				if err != nil {
					return entity.Event{}, fmt.Errorf("failed to update position for photo %d: %w", photo.ID, err)
				}
			}

		} else if strings.HasPrefix(photoString, "data:") {
			newPhoto, err := s.createPhotoFromBase64(eventID, photoString, position)
			if err != nil {
				return entity.Event{}, err
			}
			requestedPaths[newPhoto.Path] = true
		} else {
			return entity.Event{}, fmt.Errorf("unsupported photo format: %s", photoString)
		}
	}

	for path, photo := range currentEventPhotosByPath {
		if !requestedPaths[path] {
			if err := s.deletePhotoFile(photo.Path); err != nil {
				fmt.Printf("Warning: couldn't delete photo file %s: %v\n", photo.Path, err)
			}
			if err := s.photoRepository.DeletePhoto(photo.ID); err != nil {
				fmt.Printf("Warning: couldn't delete photo from DB %d: %v\n", photo.ID, err)
			}
		}
	}

	return s.GetEventByID(eventID)
}

func (s *eventService) deletePhotoFile(photoPath string) error {
	filePath := photoPath
	if strings.HasPrefix(filePath, "/") {
		filePath = filePath[1:]
	}
	return os.Remove(filePath)
}

func (s *eventService) updatePhotoPosition(photoID uint, newPosition int) error {
	return s.photoRepository.UpdatePhotoPosition(photoID, newPosition)
}

func (s *eventService) createPhotoFromBase64(eventID uint, base64Data string, position int) (entity.Photo, error) {
	parts := strings.Split(base64Data, ",")
	if len(parts) != 2 {
		return entity.Photo{}, fmt.Errorf("invalid base64 format")
	}

	mimeType := strings.Split(parts[0], ";")[0]
	mimeType = strings.TrimPrefix(mimeType, "data:")

	var ext string
	switch mimeType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		ext = ".jpg"
	}

	imageData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return entity.Photo{}, fmt.Errorf("failed to decode base64: %w", err)
	}

	subdir := "events_photos"
	filename := generateFilenameWithUUID2(eventID, "event", fmt.Sprintf("photo_%d%s", position, ext))
	fullPath := config.GetUploadFilePath(subdir, filename)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return entity.Photo{}, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(fullPath, imageData, 0644); err != nil {
		return entity.Photo{}, fmt.Errorf("failed to write file: %w", err)
	}

	relativePath := fmt.Sprintf("/uploads/%s/%s", subdir, filename)
	photo := entity.Photo{
		Path:      relativePath,
		OwnerID:   eventID,
		OwnerType: "event",
		IsMain:    false,
		IsPreview: false,
		Position:  position,
	}

	return s.photoRepository.CreatePhoto(photo)
}

func generateFilenameWithUUID2(ownerID uint, ownerType string, originalName string) string {
	uuidStr := uuid.New().String()[:8]
	switch ownerType {
	case "arts":
		return fmt.Sprintf("art_%d_%s%s", ownerID, uuidStr, filepath.Ext(originalName))
	default:
		return fmt.Sprintf("%s_%d_%s%s", ownerType, ownerID, uuidStr, filepath.Ext(originalName))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// func (s *eventService) AddMainOrPreviewPhotoToEvent(eventID uint, photo entity.Photo) (entity.Event, error) {
// 	photo.OwnerType = "event"
// 	photo, err := s.photoRepository.CreatePhoto(photo)
// 	if err != nil {
// 		return entity.Event{}, err
// 	}
// 	return s.eventRepository.AddMainOrPreviewPhotoToEvent(eventID, photo)
// }
