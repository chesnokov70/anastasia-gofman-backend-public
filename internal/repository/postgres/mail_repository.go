package postgres

import (
	"anastasia_gofman_backend/internal/entity"

	"gorm.io/gorm"
)

type MailRepository struct {
	db *gorm.DB
}

func NewMailRepository(db *gorm.DB) *MailRepository {
	return &MailRepository{db: db}
}

func (r *MailRepository) GetMailByAction(action string) (entity.Mail, error) {
	var mail entity.Mail
	err := r.db.Where("action = ?", action).First(&mail).Error
	return mail, err
}

func (r *MailRepository) GetAllMails() ([]entity.Mail, error) {
	var mails []entity.Mail
	err := r.db.Find(&mails).Error
	return mails, err
}

func (r *MailRepository) CreateMail(mail entity.Mail) (entity.Mail, error) {
	err := r.db.Create(&mail).Error
	return mail, err
}

func (r *MailRepository) UpdateMail(mail entity.Mail) (entity.Mail, error) {
	err := r.db.Save(&mail).Error
	return mail, err
}

func (r *MailRepository) DeleteMail(id int) error {
	return r.db.Delete(&entity.Mail{}, id).Error
}