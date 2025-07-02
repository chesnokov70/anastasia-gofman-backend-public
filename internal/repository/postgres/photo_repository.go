package postgres

import (
	"anastasia_gofman_backend/internal/entity"

	"gorm.io/gorm"
)

type PhotoRepository struct {
	db *gorm.DB
}

func NewPhotoRepository(db *gorm.DB) *PhotoRepository {
	return &PhotoRepository{db: db}
}

// GETTTTTTTTTTTTTTTTT
func (r *PhotoRepository) GetAllPhotosByOwnerID(ownerID uint, ownerType string) ([]entity.Photo, error) {
	var photos []entity.Photo
	err := r.db.Where("owner_id = ? AND owner_type = ?", ownerID, ownerType).Find(&photos).Error
	return photos, err
}

func (r *PhotoRepository) GetMainPhotoByOwnerID(ownerID uint, ownerType string) (entity.Photo, error) {
	var photo entity.Photo
	err := r.db.Where("owner_id = ? AND owner_type = ? AND is_main = ?", ownerID, ownerType, true).First(&photo).Error
	return photo, err
}

func (r *PhotoRepository) GetAllNoSpecialPhotosByOwnerID(ownerID uint, ownerType string) ([]entity.Photo, error) {
	var photos []entity.Photo
	err := r.db.Where("owner_id = ? AND owner_type = ? AND is_main = ? AND is_preview = ?", ownerID, ownerType, false, false).Find(&photos).Error
	return photos, err
}

func (r *PhotoRepository) GetMainOrPreviewPhotoByOwnerID(ownerID uint, ownerType string, isMain bool) (entity.Photo, error) {
	var photo entity.Photo
	var err error = nil
	if isMain {
		err = r.db.Where("owner_id = ? AND owner_type = ? AND is_main = ?", ownerID, ownerType, true).First(&photo).Error
	} else {
		err = r.db.Where("owner_id = ? AND owner_type = ? AND is_preview = ?", ownerID, ownerType, true).First(&photo).Error
	}
	return photo, err
}

func (r *PhotoRepository) GetPreviewPhotoByOwnerID(ownerID uint, ownerType string) (entity.Photo, error) {
	var photo entity.Photo
	err := r.db.Where("owner_id = ? AND owner_type = ? AND is_preview = ?", ownerID, ownerType, true).First(&photo).Error
	return photo, err
}

func (r *PhotoRepository) GetCountOfPhotos(ownerID uint, ownerType string) (int, error) {
	var count int64
	err := r.db.Model(&entity.Photo{}).Where("owner_id = ? AND owner_type = ?", ownerID, ownerType).Count(&count).Error
	return int(count), err
}

// Delete
func (r *PhotoRepository) DeletePhoto(photoID uint) error {
	return r.db.Delete(&entity.Photo{}, photoID).Error
}

func (r *PhotoRepository) DeleteAllPhotos(ownerID uint, ownerType string) error {
	return r.db.Where("owner_id = ? AND owner_type = ?", ownerID, ownerType).Delete(&entity.Photo{}).Error
}

func (r *PhotoRepository) DeleteMainPhoto(ownerID uint, ownerType string) error {
	return r.db.Where("owner_id = ? AND owner_type = ? AND is_main = ?", ownerID, ownerType, true).Delete(&entity.Photo{}).Error
}

func (r *PhotoRepository) DeletePreviewPhoto(ownerID uint, ownerType string) error {
	return r.db.Where("owner_id = ? AND owner_type = ? AND is_preview = ?", ownerID, ownerType, true).Delete(&entity.Photo{}).Error
}

func (r *PhotoRepository) DeleteAllNoSpecialPhotos(ownerID uint, ownerType string) error {
	return r.db.Where("owner_id = ? AND owner_type = ? AND is_main = ? AND is_preview = ?", ownerID, ownerType, false, false).Delete(&entity.Photo{}).Error
}

// Create
func (r *PhotoRepository) CreatePhoto(photo entity.Photo) (entity.Photo, error) {
	if photo.Position == 0 {
		photo.Position, _ = r.GetCountOfPhotos(photo.OwnerID, photo.OwnerType)
		photo.Position += 1
	}
	return photo, r.db.Create(&photo).Error
}

// Update
func (r *PhotoRepository) UpdatePhotoPosition(photoID uint, newPosition int) error {
	return r.db.Model(&entity.Photo{}).Where("id = ?", photoID).Update("position", newPosition).Error
}
