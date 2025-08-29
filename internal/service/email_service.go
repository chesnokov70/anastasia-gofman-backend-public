package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/pkg/config"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
}

type EmailAttachment struct {
	Filename string
	Content  []byte
	MimeType string
}

func NewEmailService() *EmailService {
	cfg := config.GetConfig()
	svc := &EmailService{
		smtpHost:     firstNonEmpty(cfg.Email.SmtpHost, "smtp.titan.email"),
		smtpPort:     firstNonEmpty(cfg.Email.SmtpPort, "587"),
		smtpUsername: strings.TrimSpace(cfg.Email.Username),
		smtpPassword: strings.TrimSpace(cfg.Email.Password),
		fromEmail:    strings.TrimSpace(cfg.Email.From),
		fromName:     cfg.Email.FromName,
	}
	log.Printf("EmailService init: host=%s port=%s from=%s", svc.smtpHost, svc.smtpPort, svc.fromEmail)
	return svc
}

func (s *EmailService) SendEmail(to, subject, body string) error {
	return s.SendEmailWithAttachments(to, subject, body, nil)
}

func (s *EmailService) SendEmailWithAttachments(to, subject, body string, attachments []EmailAttachment) error {
	fromAddr := mail.Address{Name: s.fromName, Address: s.fromEmail}
	toAddr := mail.Address{Address: to}

	log.Printf("EmailService: send start -> to=%s subject=%q via %s:%s", to, subject, s.smtpHost, s.smtpPort)
	msg, err := s.buildMessage(fromAddr, []mail.Address{toAddr}, subject, body, attachments)
	if err != nil {
		log.Printf("EmailService: buildMessage error: %v", err)
		return err
	}
	log.Printf("EmailService: MIME built (%d bytes)", len(msg))

	if s.smtpPort == "465" {
		log.Printf("EmailService: using implicit TLS (465)")
		return s.sendSMTP465(toAddr.Address, msg)
	}

	addr := net.JoinHostPort(s.smtpHost, s.smtpPort)
	log.Printf("EmailService: dialing %s", addr)
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: s.smtpHost, MinVersion: tls.VersionTLS12}
		if err = c.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
		log.Printf("EmailService: STARTTLS negotiated")
	}

	if ok, _ := c.Extension("AUTH"); ok {
		auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
		log.Printf("EmailService: AUTH success")
	}

	if err = c.Mail(fromAddr.Address); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}
	log.Printf("EmailService: MAIL FROM accepted")
	if err = c.Rcpt(toAddr.Address); err != nil {
		return fmt.Errorf("RCPT TO: %w", err)
	}
	log.Printf("EmailService: RCPT TO accepted")

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("write body: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("data close: %w", err)
	}
	log.Printf("EmailService: DATA sent")

	if err := c.Quit(); err != nil {
		return fmt.Errorf("quit: %w", err)
	}
	log.Printf("EmailService: send complete")
	return nil
}

func (s *EmailService) buildMessage(from mail.Address, to []mail.Address, subject, htmlBody string, atts []EmailAttachment) ([]byte, error) {
	var buf bytes.Buffer

	date := time.Now().Format(time.RFC1123Z)
	msgID := fmt.Sprintf("<%d.%s@%s>", time.Now().UnixNano(), "app", domainPart(from.Address))
	toList := make([]string, 0, len(to))
	for _, a := range to {
		toList = append(toList, a.String())
	}

	encodedSubject := mime.QEncoding.Encode("utf-8", subject)

	mixed := multipart.NewWriter(&buf)

	writeHeader(&buf, map[string]string{
		"From":         from.String(),
		"To":           strings.Join(toList, ", "),
		"Subject":      encodedSubject,
		"Date":         date,
		"Message-ID":   msgID,
		"MIME-Version": "1.0",
		"Content-Type": fmt.Sprintf("multipart/mixed; boundary=%s", mixed.Boundary()),
	})
	buf.WriteString("\r\n")

	altBoundary := "alt-" + mixed.Boundary()
	altHdr := textprotoMIMEHeader(map[string]string{
		"Content-Type": fmt.Sprintf("multipart/alternative; boundary=%s", altBoundary),
	})
	altPart, _ := mixed.CreatePart(altHdr)
	alt := multipart.NewWriter(altPart)
	_ = alt.SetBoundary(altBoundary)

	htmlHdr := textprotoMIMEHeader(map[string]string{
		"Content-Type":              "text/html; charset=UTF-8",
		"Content-Transfer-Encoding": "quoted-printable",
	})
	htmlPart, _ := alt.CreatePart(htmlHdr)
	qp := quotedprintable.NewWriter(htmlPart)
	_, _ = qp.Write([]byte(htmlBody))
	_ = qp.Close()
	_ = alt.Close()

	for _, a := range atts {
		if a.Filename == "" || len(a.Content) == 0 {
			continue
		}
		dispName := mime.QEncoding.Encode("utf-8", filepath.Base(a.Filename))
		attHdr := textprotoMIMEHeader(map[string]string{
			"Content-Type":              fmt.Sprintf("%s; name=\"%s\"", a.MimeType, dispName),
			"Content-Disposition":       fmt.Sprintf("attachment; filename=\"%s\"", dispName),
			"Content-Transfer-Encoding": "base64",
		})
		p, _ := mixed.CreatePart(attHdr)
		enc := base64.StdEncoding.EncodeToString(a.Content)
		for i := 0; i < len(enc); i += 76 {
			end := i + 76
			if end > len(enc) {
				end = len(enc)
			}
			_, _ = p.Write([]byte(enc[i:end] + "\r\n"))
		}
	}
	_ = mixed.Close()

	return buf.Bytes(), nil
}

func writeHeader(buf *bytes.Buffer, kv map[string]string) {
	for k, v := range kv {
		buf.WriteString(k + ": " + v + "\r\n")
	}
}

func textprotoMIMEHeader(kv map[string]string) textproto.MIMEHeader {
	h := make(textproto.MIMEHeader, len(kv))
	for k, v := range kv {
		h[k] = []string{v}
	}
	return h
}

func domainPart(addr string) string {
	if i := strings.LastIndex(addr, "@"); i != -1 {
		return addr[i+1:]
	}
	return "localhost"
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func (s *EmailService) sendSMTP465(to string, msg []byte) error {
	host := s.smtpHost
	port := "465"
	addr := net.JoinHostPort(host, port)

	log.Printf("EmailService: TLS dialing %s", addr)
	tlsConn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}

	log.Printf("EmailService: creating SMTP client over TLS")
	c, err := smtp.NewClient(tlsConn, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Quit()

	if ok, _ := c.Extension("AUTH"); ok {
		auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, host)
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
		log.Printf("EmailService: AUTH success (implicit TLS)")
	}

	if err := c.Mail(s.fromEmail); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}
	log.Printf("EmailService: MAIL FROM accepted (implicit TLS)")
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO: %w", err)
	}
	log.Printf("EmailService: RCPT TO accepted (implicit TLS)")

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("data close: %w", err)
	}
	log.Printf("EmailService: DATA sent (implicit TLS)")
	return nil
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

func (s *EmailService) SendAdminNotificationArtRequest(request entity.ArtRequest) error {
	subject := fmt.Sprintf("New Art Request from %s", request.FullName)
	body := fmt.Sprintf(`
		<h2>New Art Request</h2>
		<p><strong>Name:</strong> %s</p>
		<p><strong>Email:</strong> %s</p>
		<p><strong>Language:</strong> %s</p>
		<p><strong>Phone Number:</strong> %s</p>
		<p><strong>Request:</strong> %s</p>
		<p><strong>Art:</strong> %s</p>
		<p><strong>Request Time:</strong> %s</p>
	`, request.FullName, request.Email, request.Language, request.PhoneNumber, request.Request, request.Art.Name.EN, request.CreatedAt.Format("2006-01-02 15:04"))

	return s.SendToAllAdmins(subject, body)
}
