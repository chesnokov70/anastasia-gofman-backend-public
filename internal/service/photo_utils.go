package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/repository"
	"anastasia_gofman_backend/pkg/config"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PhotoOwnerGetter interface {
	GetByID(id uint) (interface{}, error)
}

type PhotoPatchConfig struct {
	OwnerType      string
	SubdirName     string
	FilenamePrefix string
}

var PhotoConfigs = map[string]PhotoPatchConfig{
	"event": {
		OwnerType:      "event",
		SubdirName:     "events_photos",
		FilenamePrefix: "event",
	},
	"arts": {
		OwnerType:      "arts",
		SubdirName:     "arts_photos",
		FilenamePrefix: "art",
	},
	"press": {
		OwnerType:      "press",
		SubdirName:     "press_photos",
		FilenamePrefix: "press",
	},
	"article": {
		OwnerType:      "article",
		SubdirName:     "article_photos",
		FilenamePrefix: "article",
	},
	"authors": {
		OwnerType:      "authors",
		SubdirName:     "authors_photos",
		FilenamePrefix: "author",
	},
}

func PatchPhotosFromStrings(
	ownerID uint,
	ownerType string,
	photoStrings []string,
	photoRepo repository.PhotoRepository,
	getCurrentPhotos func() []entity.Photo,
) error {
	cfg, exists := PhotoConfigs[ownerType]
	if !exists {
		return fmt.Errorf("unsupported owner type: %s", ownerType)
	}

	currentPhotosByPath := make(map[string]entity.Photo)
	for _, p := range getCurrentPhotos() {
		if !p.IsMain && !p.IsPreview {
			currentPhotosByPath[p.Path] = p
		}
	}

	requestedPaths := make(map[string]bool)

	for path, photo := range currentPhotosByPath {
		isRequested := false
		for _, photoString := range photoStrings {
			if strings.HasPrefix(photoString, "http") {
				relativePath, err := extractRelativePathFromURL(photoString)
				if err == nil && relativePath == path {
					isRequested = true
					break
				}
			}
		}

		if !isRequested {
			if err := DeletePhotoFile(photo.Path); err != nil {
				fmt.Printf("Warning: couldn't delete photo file %s: %v\n", photo.Path, err)
			}
			if err := photoRepo.DeletePhoto(photo.ID); err != nil {
				fmt.Printf("Warning: couldn't delete photo from DB %d: %v\n", photo.ID, err)
			}
		}
	}

	for i, photoString := range photoStrings {
		position := i + 1
		if strings.HasPrefix(photoString, "http") {
			relativePath, err := extractRelativePathFromURL(photoString)
			if err != nil {
				return err
			}

			requestedPaths[relativePath] = true

			photo, err := photoRepo.GetPhotoByPath(relativePath)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("photo with path %s not found in database", relativePath)
				}
				return err
			}

			if photo.OwnerID != ownerID || photo.OwnerType != cfg.OwnerType {
				err = photoRepo.UpdatePhotoOwnerAndPosition(photo.ID, ownerID, cfg.OwnerType, position)
				if err != nil {
					return fmt.Errorf("failed to move photo %d to %s %d: %w", photo.ID, cfg.OwnerType, ownerID, err)
				}
			} else if photo.Position != position {
				err = photoRepo.UpdatePhotoPosition(photo.ID, position)
				if err != nil {
					return fmt.Errorf("failed to update position for photo %d: %w", photo.ID, err)
				}
			}

		} else if strings.HasPrefix(photoString, "data:") {
			newPhoto, err := CreatePhotoFromBase64(ownerID, cfg, photoString, position, photoRepo)
			if err != nil {
				return err
			}
			requestedPaths[newPhoto.Path] = true
		} else {
			return fmt.Errorf("unsupported photo format: %s", photoString)
		}
	}

	return nil
}

func extractRelativePathFromURL(photoURL string) (string, error) {
	idx := strings.Index(photoURL, "/uploads/")
	if idx == -1 {
		return "", fmt.Errorf("invalid photo URL format: %s", photoURL)
	}
	relativePath := photoURL[idx:]
	return path.Clean(relativePath), nil
}

func DeletePhotoFile(photoPath string) error {
	filePath := photoPath
	if strings.HasPrefix(filePath, "/uploads/") {
		filePath = strings.TrimPrefix(filePath, "/uploads/")
		parts := strings.Split(filePath, "/")
		if len(parts) >= 2 {
			filePath = config.GetUploadFilePath(parts[0], strings.Join(parts[1:], "/"))
		}
	} else if strings.HasPrefix(filePath, "/") {
		filePath = filePath[1:]
	}

	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func CreatePhotoFromBase64(ownerID uint, cfg PhotoPatchConfig, base64Data string, position int, photoRepo repository.PhotoRepository) (entity.Photo, error) {
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

	filename := GenerateFilenameWithUUID(ownerID, cfg.FilenamePrefix, fmt.Sprintf("photo_%d%s", position, ext))
	fullPath := config.GetUploadFilePath(cfg.SubdirName, filename)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return entity.Photo{}, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(fullPath, imageData, 0644); err != nil {
		return entity.Photo{}, fmt.Errorf("failed to write file: %w", err)
	}

	relativePath := fmt.Sprintf("/uploads/%s/%s", cfg.SubdirName, filename)
	photo := entity.Photo{
		Path:      relativePath,
		OwnerID:   ownerID,
		OwnerType: cfg.OwnerType,
		IsMain:    false,
		IsPreview: false,
		Position:  position,
	}

	return photoRepo.CreatePhoto(photo)
}

func GenerateFilenameWithUUID(ownerID uint, prefix string, originalName string) string {
	uuidStr := uuid.New().String()[:8]
	return fmt.Sprintf("%s_%d_%s%s", prefix, ownerID, uuidStr, filepath.Ext(originalName))
}
