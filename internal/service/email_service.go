package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/pkg/config"
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"mime/multipart"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
)

type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
}

type EmailAttachment struct {
	Filename string
	Content  []byte
	MimeType string
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
	return s.SendEmailWithAttachments(to, subject, body, nil)
}

func (s *EmailService) SendEmailWithAttachments(to, subject, body string, attachments []EmailAttachment) error {
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Создаем multipart сообщение
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Headers
	boundary := writer.Boundary()
	buf.WriteString(fmt.Sprintf("From: %s\r\n", s.fromEmail))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
	buf.WriteString("\r\n")

	// HTML body part
	bodyHeader := make(map[string][]string)
	bodyHeader["Content-Type"] = []string{"text/html; charset=UTF-8"}
	bodyPart, err := writer.CreatePart(bodyHeader)
	if err != nil {
		return fmt.Errorf("failed to create body part: %w", err)
	}
	bodyPart.Write([]byte(body))

	// Attachments
	for _, attachment := range attachments {
		attachHeader := make(map[string][]string)
		attachHeader["Content-Type"] = []string{fmt.Sprintf("%s; name=\"%s\"", attachment.MimeType, attachment.Filename)}
		attachHeader["Content-Disposition"] = []string{fmt.Sprintf("attachment; filename=\"%s\"", attachment.Filename)}
		attachHeader["Content-Transfer-Encoding"] = []string{"base64"}

		attachPart, err := writer.CreatePart(attachHeader)
		if err != nil {
			log.Printf("Failed to create attachment part for %s: %v", attachment.Filename, err)
			continue
		}

		// Кодируем в base64
		encoded := base64.StdEncoding.EncodeToString(attachment.Content)
		// Разбиваем на строки по 76 символов (стандарт RFC)
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			attachPart.Write([]byte(encoded[i:end] + "\r\n"))
		}
	}

	writer.Close()

	err = smtp.SendMail(s.smtpHost+":"+s.smtpPort, auth, s.fromEmail, []string{to}, buf.Bytes())
	return err
}

func (s *EmailService) SendToAllAdmins(subject, body string) error {
	return s.SendToAllAdminsWithAttachments(subject, body, nil)
}

func (s *EmailService) SendToAllAdminsWithAttachments(subject, body string, attachments []EmailAttachment) error {
	adminEmails := config.GetConfig().Email.Admin

	if len(adminEmails) == 0 {
		log.Printf("No admin emails configured, skipping notification")
		return nil
	}

	var errors []error

	for _, adminEmail := range adminEmails {
		if adminEmail == "" {
			continue
		}

		err := s.SendEmailWithAttachments(adminEmail, subject, body, attachments)
		if err != nil {
			log.Printf("Failed to send email to admin %s: %v", adminEmail, err)
			errors = append(errors, fmt.Errorf("failed to send to %s: %w", adminEmail, err))
		} else {
			log.Printf("Successfully sent email to admin: %s", adminEmail)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send to %d admin(s): %v", len(errors), errors)
	}

	return nil
}

func CreateAttachmentFromFile(filePath string) (EmailAttachment, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return EmailAttachment{}, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	filename := filepath.Base(filePath)
	mimeType := getMimeTypeFromExtension(filepath.Ext(filePath))

	return EmailAttachment{
		Filename: filename,
		Content:  content,
		MimeType: mimeType,
	}, nil
}

// Определяем MIME тип по расширению файла
func getMimeTypeFromExtension(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
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

	return s.SendToAllAdmins(subject, body)
}

func (s *EmailService) SendAdminNotificationAuthor(registration entity.AuthorRegistration) error {
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

	return s.SendToAllAdmins(subject, body)
}
