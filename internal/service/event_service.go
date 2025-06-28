package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/repository"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"strings"

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
		s.eventRepository.DeleteMainOrPreviewPhotoFromEvent(id, "main")
		s.photoRepository.CreatePhoto(main_photo)
		s.eventRepository.AddMainOrPreviewPhotoToEvent(id, main_photo)
		delete(kwargs, "main_photo")
	}
	if kwargs["preview_photo"] != nil {
		preview_photo, err := create_photo_from_http_photo(id, "event", kwargs["preview_photo"].(*multipart.FileHeader), false, true)
		if err != nil {
			return entity.Event{}, err
		}
		s.eventRepository.DeleteMainOrPreviewPhotoFromEvent(id, "preview")
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

func (s *eventService) DeleteMainOrPreviewPhoto(id uint, type_of_photo string) error {
	var is_main bool = true
	if type_of_photo == "preview" {
		is_main = false
	}
	photo, err := s.photoRepository.GetMainOrPreviewPhotoByOwnerID(id, "event", is_main)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		fmt.Printf("ERROR DELETING PHOTO %v\n", err)
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

func (s *eventService) DeleteEvent(id uint) error {
	err := s.DeleteAllPhotos(id)
	if err != nil {
		return err
	}
	return s.eventRepository.DeleteEvent(id)
}

// func (s *eventService) AddMainOrPreviewPhotoToEvent(eventID uint, photo entity.Photo) (entity.Event, error) {
// 	photo.OwnerType = "event"
// 	photo, err := s.photoRepository.CreatePhoto(photo)
// 	if err != nil {
// 		return entity.Event{}, err
// 	}
// 	return s.eventRepository.AddMainOrPreviewPhotoToEvent(eventID, photo)
// }
