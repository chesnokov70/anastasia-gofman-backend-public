package postgres

import (
	"anastasia_gofman_backend/internal/entity"

	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) GetAllEvents(offset int, limit int, only_id bool) ([]entity.Event, int64, error) {
	var events []entity.Event
	var count int64
	var query *gorm.DB
	if only_id {
		query = r.db.Model(&entity.Event{}).Select("id").Order("created_at DESC")
	} else {
		query = r.db.Model(&entity.Event{}).Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos.position ASC")
		}).Order("created_at DESC")
	}

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}

	err = query.Find(&events).Error
	return events, count, err
}

func (r *EventRepository) GetEventByID(id uint) (entity.Event, error) {
	var event entity.Event
	err := r.db.Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
		return db.Order("photos.position ASC")
	}).First(&event, id).Error
	return event, err
}

func (r *EventRepository) CreateEvent(event entity.Event) (entity.Event, error) {
	err := r.db.Create(&event).Error
	return event, err
}

func (r *EventRepository) GetCountOfEvents() (int, error) {
	var count int64
	err := r.db.Model(&entity.Event{}).Count(&count).Error
	return int(count), err
}

func (r *EventRepository) PartialUpdateEvent(id uint, kwargs map[string]interface{}) (entity.Event, error) {
	var event entity.Event
	err := r.db.Model(&event).Where("id = ?", id).Updates(kwargs).Error
	if err != nil {
		return entity.Event{}, err
	}
	return r.GetEventByID(id)
}

func (r *EventRepository) FullUpdateEvent(event entity.Event) (entity.Event, error) {
	err := r.db.Model(&event).Where("id = ?", event.ID).
		Select("title", "description", "start_date", "end_date", "location", "position", "is_finished").
		Updates(event).Error
	if err != nil {
		return entity.Event{}, err
	}
	return r.GetEventByID(uint(event.ID))
}

func (r *EventRepository) UpdateEvent(event entity.Event) (entity.Event, error) {
	err := r.db.Save(&event).Error
	return event, err
}

func (r *EventRepository) DeleteEvent(id uint) error {
	return r.db.Delete(&entity.Event{}, id).Error
}

func (r *EventRepository) UpdateEventsPosition(positions []int) error {
	for i, position := range positions {
		err := r.db.Model(&entity.Event{}).Where("id = ?", i+1).Update("position", position).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *EventRepository) AddMainOrPreviewPhotoToEvent(eventID uint, photo entity.Photo) (entity.Event, error) {
	// return entity.Event{}, nil
	var event entity.Event
	err := r.db.First(&event, eventID).Error
	if err != nil {
		return entity.Event{}, err
	}
	which_photo := "main_photo_id"
	if photo.IsPreview {
		which_photo = "preview_photo_id"
	}
	err = r.db.Model(&event).Update(which_photo, photo.ID).Error
	if err != nil {
		return entity.Event{}, err
	}
	return event, nil
}

func (r *EventRepository) RemoveSpecificPhotoFromEvent(eventID uint, is_main bool) error {
	var event entity.Event
	err := r.db.First(&event, eventID).Error
	if err != nil {
		return err
	}
	var update_data = make(map[string]interface{})
	if is_main {
		update_data["main_photo_id"] = nil
	} else {
		update_data["preview_photo_id"] = nil
	}
	err = r.db.Model(&event).Updates(update_data).Error
	if err != nil {
		return err
	}
	// err = r.db.Model(&event).Update(which_photo, nil).Error
	// if err != nil {
	// 	return err
	// }
	return nil
}

// func (r *EventRepository) AddPhotoToEvent(eventID uint, photo entity.Photo) (entity.Event, error) {
// 	err := r.db.Create(&photo).Error
// 	if err != nil {
// 		return entity.Event{}, err
// 	}
// 	var event entity.Event
// 	err = r.db.First(&event, eventID).Error
// 	if err != nil {
// 		return entity.Event{}, err
// 	}
// 	err = r.db.Model(&event).Update("photos", append(event.Photos, photo)).Error
// 	return event, err
// }

// func (r *EventRepository) AddPhotosToEvent(eventID uint, photos []entity.Photo) (entity.Event, error) {

// 	var event entity.Event
// 	err := r.db.First(&event, eventID).Error
// 	if err != nil {
// 		return entity.Event{}, err
// 	}
// 	for _, photo := range photos {
// 		err = r.db.Model(&event).Update("photos", append(event.Photos, photo)).Error
// 		if err != nil {
// 			return entity.Event{}, err
// 		}
// 	}
// 	err = r.db.Save(&event).Error
// 	if err != nil {
// 		return entity.Event{}, err
// 	}
// 	return event, nil
// }

// func (r *EventRepository) PatchEventPhotos(eventID uint, photos []entity.Photo) (entity.Event, error) {
// 	for _, photo := range photos {
// 		err := r.db.Create(&photo).Error
// 		if err != nil {
// 			return entity.Event{}, err
// 		}
// 	}
// 	var event entity.Event
// 	err := r.db.Where("id = ?", eventID).Preload("Photos").First(&event).Error
// 	if err != nil {
// 		return entity.Event{}, err
// 	}
// 	err = r.db.Model(&event).Association("Photos").Replace(photos)
// 	if err != nil {
// 		return entity.Event{}, err
// 	}
// 	return event, nil
// }

// func (r *EventRepository) DeleteMainPhotoFromEvent(eventID uint) error {
// 	return r.db.Transaction(func(tx *gorm.DB) error {
// 		var event entity.Event
// 		// First, retrieve the event to get the MainPhotoID
// 		if err := tx.First(&event, eventID).Error; err != nil {
// 			// Handle case where event is not found or other DB error
// 			return err
// 		}

// 		// Store the MainPhotoID before nullifying it
// 		photoIDToDelete := event.MainPhotoID

// 		// If MainPhotoID is 0 (or whatever your convention for "no photo" is),
// 		// there's nothing to delete from the photos table.
// 		// We still proceed to ensure MainPhotoID is null in the event.
// 		if photoIDToDelete == 0 {
// 			// Ensure MainPhotoID is nil (or null in DB)
// 			if err := tx.Model(&entity.Event{}).Where("id = ?", eventID).Update("main_photo_id", nil).Error; err != nil {
// 				return err
// 			}
// 			return nil // No photo to delete, but ensure the field is nulled
// 		}

// 		// Update the event's main_photo_id to nil
// 		if err := tx.Model(&entity.Event{}).Where("id = ?", eventID).Update("main_photo_id", nil).Error; err != nil {
// 			return err
// 		}

// 		// Delete the actual photo record from the photos table
// 		// Assuming Photo entity has an "ID" field of type int.
// 		if err := tx.Delete(&entity.Photo{}, photoIDToDelete).Error; err != nil {
// 			return err
// 		}

// 		return nil // Transaction success
// 	})
// }

// func (r *EventRepository) GetMainPhoto(eventID uint) (entity.Photo, error) {
// 	var photo entity.Photo
// 	err := r.db.Where("event_id = ?", eventID).First(&photo).Error
// 	return photo, err
// }

// func (r *EventRepository) SavePhoto(photo entity.Photo) (entity.Photo, error) {
// 	err := r.db.Save(&photo).Error
// 	return photo, err
// }

// func (r *EventRepository) DeletePhoto(photoID uint) error {
// 	return r.db.Delete(&entity.Photo{}, photoID).Error
// }

// func (r *EventRepository) GetPhoto(photoID uint) (entity.Photo, error) {
// 	var photo entity.Photo
// 	err := r.db.First(&photo, photoID).Error
// 	return photo, err
// }

// func (r *EventRepository) GetCountOfPhotos(eventID uint) (int, error) {
// 	var count int64
// 	err := r.db.Model(&entity.Photo{}).Where("event_id = ?", eventID).Count(&count).Error
// 	return int(count), err
// }

// func (r *EventRepository) GetPhotos(eventID uint) ([]entity.Photo, error) {
// 	var photos []entity.Photo
// 	err := r.db.Where("event_id = ?", eventID).Find(&photos).Error
// 	return photos, err
// }

// func (r *EventRepository) DeleteAllPhotos(eventID uint) error {
// 	return r.db.Where("event_id = ?", eventID).Delete(&entity.Photo{}).Error
// }
