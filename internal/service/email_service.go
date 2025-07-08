package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/pkg/config"
	"fmt"
	"net/smtp"
	"strings"
)

type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
}

func NewEmailService() *EmailService {
	return &EmailService{
		smtpHost:     "smtp.gmail.com",
		smtpPort:     "587",
		smtpUsername: config.GetConfig().Email.Username,
		// smtpUsername: "kachan333.333@gmail.com",
		smtpPassword: config.GetConfig().Email.Password,
		// smtpPassword: "-",
		fromEmail: config.GetConfig().Email.From,
		// fromEmail: "kachan333.333@gmail.com",
	}
}

func (s *EmailService) SendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	msg := fmt.Sprintf("From: %s\r\n", s.fromEmail)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	msg += "\r\n"
	msg += body

	err := smtp.SendMail(s.smtpHost+":"+s.smtpPort, auth, s.fromEmail, []string{to}, []byte(msg))
	return err
}

func (s *EmailService) getTranslatedText(translatedText entity.TranslatedText, language string) string {
	switch language {
	case "en":
		return translatedText.EN
	case "ru":
		return translatedText.RU
	case "es":
		return translatedText.ES
	default:
		return translatedText.EN
	}
}

func (s *EmailService) SendRegistrationConfirmation(registration entity.EventRegistration, mailTemplate entity.Mail) error {
	htmlContent := s.getTranslatedText(mailTemplate.HTMLText, registration.Language)
	if htmlContent == "" {
		htmlContent = s.getTranslatedText(mailTemplate.HTMLText, "en")
	}

	eventTitle := s.getTranslatedText(registration.Event.Title, registration.Language)
	eventLocation := s.getTranslatedText(registration.Event.Location, registration.Language)

	htmlContent = strings.ReplaceAll(htmlContent, "{{full_name}}", registration.FullName)
	htmlContent = strings.ReplaceAll(htmlContent, "{{event_title}}", eventTitle)
	htmlContent = strings.ReplaceAll(htmlContent, "{{event_start_date}}", registration.Event.StartDate.Format("2006-01-02 15:04"))
	htmlContent = strings.ReplaceAll(htmlContent, "{{event_end_date}}", registration.Event.EndDate.Format("2006-01-02 15:04"))
	htmlContent = strings.ReplaceAll(htmlContent, "{{event_location}}", eventLocation)

	subject := fmt.Sprintf("Registration Confirmation - %s", eventTitle)

	return s.SendEmail(registration.Email, subject, htmlContent)
}

func (s *EmailService) SendAdminNotification(registration entity.EventRegistration) error {
	adminEmail := config.GetConfig().Email.Admin
	// adminEmail := "kachan333.333@gmail.com"

	eventTitle := s.getTranslatedText(registration.Event.Title, "en")
	subject := fmt.Sprintf("New Event Registration - %s", eventTitle)
	body := fmt.Sprintf(`
		<h2>New Event Registration</h2>
		<p><strong>Name:</strong> %s</p>
		<p><strong>Email:</strong> %s</p>
		<p><strong>Language:</strong> %s</p>
		<p><strong>Event:</strong> %s</p>
		<p><strong>Event Date:</strong> %s</p>
		<p><strong>Registration Time:</strong> %s</p>
	`, registration.FullName, registration.Email, registration.Language,
		eventTitle, registration.Event.StartDate.Format("2006-01-02 15:04"),
		registration.CreatedAt.Format("2006-01-02 15:04"))

	return s.SendEmail(adminEmail, subject, body)
}

func (s *EmailService) SendAdminNotificationAuthor(registration entity.AuthorRegistration) error {
	adminEmail := config.GetConfig().Email.Admin
	// adminEmail := "kachan333.333@gmail.com"

	subject := fmt.Sprintf("New Author Registration - %s", registration.FullName)
	body := fmt.Sprintf(`
		<h2>New Author Registration</h2>
		<p><strong>Name:</strong> %s</p>
		<p><strong>Email:</strong> %s</p>
		<p><strong>Language:</strong> %s</p>
		<p><strong>Phone Number:</strong> %s</p>
		<p><strong>Portfolio:</strong> %s</p>
		<p><strong>Registration Time:</strong> %s</p>
	`, registration.FullName, registration.Email, registration.Language, registration.PhoneNumber, registration.Portfolio, registration.CreatedAt.Format("2006-01-02 15:04"))

	return s.SendEmail(adminEmail, subject, body)
}
