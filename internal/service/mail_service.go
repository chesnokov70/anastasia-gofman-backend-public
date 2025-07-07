package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/repository/postgres"
)

type MailService struct {
	mailRepo *postgres.MailRepository
}

func NewMailService(mailRepo *postgres.MailRepository) *MailService {
	return &MailService{
		mailRepo: mailRepo,
	}
}

func (s *MailService) GetMailByAction(action string) (entity.Mail, error) {
	return s.mailRepo.GetMailByAction(action)
}

func (s *MailService) GetAllMails() ([]entity.Mail, error) {
	return s.mailRepo.GetAllMails()
}

func (s *MailService) CreateMail(mail entity.Mail) (entity.Mail, error) {
	return s.mailRepo.CreateMail(mail)
}

func (s *MailService) UpdateMail(mail entity.Mail) (entity.Mail, error) {
	return s.mailRepo.UpdateMail(mail)
}

func (s *MailService) DeleteMail(id int) error {
	return s.mailRepo.DeleteMail(id)
}