package postgres

import (
	"anastasia_gofman_backend/internal/entity"
	"errors"

	"gorm.io/gorm"
)

type PressAndArticleRepository struct {
	db *gorm.DB
}

func NewPressAndArticleRepository(db *gorm.DB) *PressAndArticleRepository {
	return &PressAndArticleRepository{db: db}
}

func (r *PressAndArticleRepository) GetAllPressAndArticles(offset int, limit int, with_pagination bool, article_or_press string, sorting string) ([]entity.Press, []entity.Article, int64, error) {
	var press []entity.Press
	var articles []entity.Article
	var count int64
	var query *gorm.DB
	switch article_or_press {
	case "press":
		query = r.db.Model(&entity.Press{}).Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos.position ASC")
		})
	case "article":
		query = r.db.Model(&entity.Article{}).Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos.position ASC")
		})
	default:
		return nil, nil, 0, errors.New("invalid article_or_press")
	}

	if sorting == "NEW" {
		query = query.Order("created_at DESC")
	} else if sorting == "CLOSEST" {
		query = query.Order("event_at ASC")
	} else if sorting == "FARTHEST" {
		query = query.Order("event_at DESC")
	} else {
		query = query.Order("created_at DESC")
	}

	err := query.Count(&count).Error
	if err != nil {
		return nil, nil, 0, err
	}

	if with_pagination {
		query = query.Offset(offset).Limit(limit)
	}

	switch article_or_press {
	case "press":
		err = query.Find(&press).Error
		return press, nil, count, err
	case "article":
		err = query.Find(&articles).Error
		return nil, articles, count, err
	default:
		return nil, nil, 0, errors.New("invalid article_or_press")
	}
}

func (r *PressAndArticleRepository) GetPressOrArticleByID(id uint, article_or_press string) (*entity.Press, *entity.Article, error) {
	var press entity.Press
	var article entity.Article
	var err error
	switch article_or_press {
	case "press":
		err = r.db.Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos.position ASC")
		}).First(&press, id).Error
		return &press, nil, err
	case "article":
		err = r.db.Preload("MainPhoto").Preload("PreviewPhoto").Preload("Photos", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos.position ASC")
		}).First(&article, id).Error
		return nil, &article, err
	default:
		return nil, nil, errors.New("invalid article_or_press")
	}
}

func (r *PressAndArticleRepository) CreatePressOrArticle(press_or_article string, press entity.Press, article entity.Article) (*entity.Press, *entity.Article, error) {
	var err error
	var createdID uint
	switch press_or_article {
	case "press":
		err = r.db.Create(&press).Error
		if err != nil {
			return nil, nil, err
		}
		createdID = press.ID
	case "article":
		err = r.db.Create(&article).Error
		if err != nil {
			return nil, nil, err
		}
		createdID = article.ID
	default:
		return nil, nil, errors.New("invalid article_or_press")
	}

	return r.GetPressOrArticleByID(createdID, press_or_article)
}

func (r *PressAndArticleRepository) GetCountOfPressOrArticle(article_or_press string) (int, error) {
	var count int64
	var query *gorm.DB
	switch article_or_press {
	case "press":
		query = r.db.Model(&entity.Press{})
	case "article":
		query = r.db.Model(&entity.Article{})
	}
	err := query.Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), err
}

func (r *PressAndArticleRepository) UpdatePressOrArticle(press_or_article string, press entity.Press, article entity.Article) (*entity.Press, *entity.Article, error) {
	var err error
	switch press_or_article {
	case "press":
		err = r.db.Save(&press).Error
		return &press, nil, err
	case "article":
		err = r.db.Save(&article).Error
		return nil, &article, err
	default:
		return nil, nil, errors.New("invalid article_or_press")
	}
}

func (r *PressAndArticleRepository) DeletePressOrArticle(press_or_article string, id uint) error {
	switch press_or_article {
	case "press":
		return r.db.Delete(&entity.Press{}, id).Error
	case "article":
		return r.db.Delete(&entity.Article{}, id).Error
	default:
		return errors.New("invalid article_or_press")
	}
}

func (r *PressAndArticleRepository) PartialUpdatePressOrArticle(press_or_article string, id uint, kwargs map[string]interface{}) (*entity.Press, *entity.Article, error) {
	var press entity.Press
	var article entity.Article
	var err error
	switch press_or_article {
	case "press":
		err = r.db.Model(&press).Where("id = ?", id).Updates(kwargs).Error
		// return &press, nil, err
	case "article":
		err = r.db.Model(&article).Where("id = ?", id).Updates(kwargs).Error
		// return nil, &article, err
	default:
		return nil, nil, errors.New("invalid article_or_press")
	}
	if err != nil {
		return nil, nil, err
	}
	return r.GetPressOrArticleByID(id, press_or_article)
}

func (r *PressAndArticleRepository) FullUpdatePressOrArticle(press_or_article string, press *entity.Press, article *entity.Article) (*entity.Press, *entity.Article, error) {
	var err error
	var id uint
	switch press_or_article {
	case "press":
		err = r.db.Model(press).Where("id = ?", press.ID).
			Select("title", "description", "full_text", "link", "position", "event_at").
			Updates(press).Error
		id = press.ID
	case "article":
		err = r.db.Model(article).Where("id = ?", article.ID).
			Select("title", "description", "full_text", "link", "position", "event_at").
			Updates(article).Error
		id = article.ID
	default:
		return nil, nil, errors.New("invalid article_or_press")
	}
	if err != nil {
		return nil, nil, err
	}
	return r.GetPressOrArticleByID(id, press_or_article)
}

func (r *PressAndArticleRepository) AddMainOrPreviewPhotoToPressOrArticle(photo entity.Photo, press_or_article string) (*entity.Press, *entity.Article, error) {
	var press entity.Press
	var article entity.Article
	which_photo := "main_photo_id"
	if photo.IsPreview {
		which_photo = "preview_photo_id"
	}
	switch press_or_article {
	case "press":
		err := r.db.Model(&entity.Press{}).Where("id = ?", photo.OwnerID).
			Update(which_photo, photo.ID).
			First(&press).Error
		if err != nil {
			return nil, nil, err
		}
	case "article":
		err := r.db.Model(&entity.Article{}).Where("id = ?", photo.OwnerID).
			Update(which_photo, photo.ID).
			First(&article).Error
		if err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, errors.New("invalid article_or_press")
	}
	return r.GetPressOrArticleByID(photo.OwnerID, press_or_article)
}

func (r *PressAndArticleRepository) GetCountOfPhotos(press_or_article string, id uint) (int, error) {
	var count int64
	var query *gorm.DB
	switch press_or_article {
	case "press":
		query = r.db.Model(&entity.Press{}).Where("id = ?", id)
	case "article":
		query = r.db.Model(&entity.Article{}).Where("id = ?", id)
	}
	return int(count), query.Error
}

func (r *PressAndArticleRepository) RemoveMainAndPreviewPhotoFromPressOrArticle(press_or_article string, id uint) error {
	switch press_or_article {
	case "press":
		return r.db.Model(&entity.Press{}).Where("id = ?", id).Updates(map[string]interface{}{"main_photo_id": nil, "preview_photo_id": nil}).Error
	case "article":
		return r.db.Model(&entity.Article{}).Where("id = ?", id).Updates(map[string]interface{}{"main_photo_id": nil, "preview_photo_id": nil}).Error
	default:
		return errors.New("invalid article_or_press")
	}
}
