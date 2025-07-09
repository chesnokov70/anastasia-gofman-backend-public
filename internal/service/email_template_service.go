package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/pkg/config"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/stripe/stripe-go/v82"
)

type EmailTemplateService struct {
	templateDir string
}

func NewEmailTemplateService() *EmailTemplateService {
	return &EmailTemplateService{
		templateDir: "email_templates",
	}
}

// Генерирует красивое HTML письмо для уведомления о заказе
func (s *EmailTemplateService) GeneratePaymentNotificationHTML(eventType string, data map[string]interface{}, arts []entity.Art) (string, []EmailAttachment, error) {
	templatePath := filepath.Join(s.templateDir, "payment_notification_template.html")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read template: %w", err)
	}

	html := string(templateContent)

	html = strings.ReplaceAll(html, "{{PAYMENT_INFO}}", s.generatePaymentInfo(eventType, data))
	html = strings.ReplaceAll(html, "{{CUSTOMER_INFO}}", s.generateCustomerInfo(data))
	html = strings.ReplaceAll(html, "{{SHIPPING_ADDRESS}}", s.generateShippingAddress(data))
	html = strings.ReplaceAll(html, "{{BILLING_ADDRESS}}", s.generateBillingAddress(data))
	html = strings.ReplaceAll(html, "{{ARTWORKS_INFO}}", s.generateArtworksInfo(arts))
	html = strings.ReplaceAll(html, "{{FOOTER_INFO}}", s.generateFooterInfo(data))

	// Собираем прикрепления с фотографиями
	attachments, err := s.collectArtPhotoAttachments(arts)
	if err != nil {
		log.Printf("Warning: Failed to collect photo attachments: %v", err)
		// Продолжаем без прикреплений
		attachments = []EmailAttachment{}
	}

	return html, attachments, nil
}

func (s *EmailTemplateService) generatePaymentInfo(eventType string, data map[string]interface{}) string {
	var html string

	if eventType == "checkout.session.completed" {
		amountInDollars := float64(data["amount_total"].(int64)) / 100.0
		html = fmt.Sprintf(`
			<div class="info-row">
				<span class="info-label">Сумма:</span>
				<span class="info-value amount">$%.2f %v</span>
			</div>
			<div class="info-row">
				<span class="info-label">ID сессии:</span>
				<span class="info-value">%v</span>
			</div>
			<div class="info-row">
				<span class="info-label">Статус:</span>
				<span class="info-value">%v</span>
			</div>`,
			amountInDollars, data["currency"], data["session_id"], data["payment_status"])
	} else if eventType == "payment_intent.succeeded" {
		amountInDollars := float64(data["amount"].(int64)) / 100.0
		html = fmt.Sprintf(`
			<div class="info-row">
				<span class="info-label">Сумма:</span>
				<span class="info-value amount">$%.2f %v</span>
			</div>
			<div class="info-row">
				<span class="info-label">Payment Intent ID:</span>
				<span class="info-value">%v</span>
			</div>
			<div class="info-row">
				<span class="info-label">Статус:</span>
				<span class="info-value">%v</span>
			</div>
			<div class="info-row">
				<span class="info-label">Описание:</span>
				<span class="info-value">%v</span>
			</div>`,
			amountInDollars, data["currency"], data["payment_intent_id"], data["status"], data["description"])
	}

	return html
}

func (s *EmailTemplateService) generateCustomerInfo(data map[string]interface{}) string {
	var html string

	if email, ok := data["customer_email"]; ok && email != nil {
		html += fmt.Sprintf(`<div class="info-row">
			<span class="info-label">Email:</span>
			<span class="info-value">%v</span>
		</div>`, email)
	}

	if name, ok := data["customer_name"]; ok && name != nil {
		html += fmt.Sprintf(`<div class="info-row">
			<span class="info-label">Имя:</span>
			<span class="info-value">%v</span>
		</div>`, name)
	}

	if phone, ok := data["customer_phone"]; ok && phone != nil && phone != "" {
		html += fmt.Sprintf(`<div class="info-row">
			<span class="info-label">Телефон:</span>
			<span class="info-value">%v</span>
		</div>`, phone)
	}

	return html
}

func (s *EmailTemplateService) generateShippingAddress(data map[string]interface{}) string {
	if shippingAddr, ok := data["shipping_address"]; ok && shippingAddr != nil {
		html := `<div class="section">
			<h2>📦 Адрес доставки</h2>`

		if shippingName, ok := data["shipping_name"]; ok && shippingName != nil {
			html += fmt.Sprintf(`<div class="info-row">
				<span class="info-label">Получатель:</span>
				<span class="info-value">%v</span>
			</div>`, shippingName)
		}

		html += `<div class="address-block">`
		if addr, ok := shippingAddr.(*stripe.Address); ok {
			if addr.Line1 != "" {
				html += fmt.Sprintf(`<div><strong>Адрес:</strong> %s</div>`, addr.Line1)
			}
			if addr.Line2 != "" {
				html += fmt.Sprintf(`<div><strong>Доп. адрес:</strong> %s</div>`, addr.Line2)
			}
			if addr.City != "" {
				html += fmt.Sprintf(`<div><strong>Город:</strong> %s</div>`, addr.City)
			}
			if addr.State != "" {
				html += fmt.Sprintf(`<div><strong>Штат/Регион:</strong> %s</div>`, addr.State)
			}
			if addr.PostalCode != "" {
				html += fmt.Sprintf(`<div><strong>Почтовый код:</strong> %s</div>`, addr.PostalCode)
			}
			if addr.Country != "" {
				html += fmt.Sprintf(`<div><strong>Страна:</strong> %s</div>`, addr.Country)
			}
		}
		html += `</div></div>`
		return html
	}
	return ""
}

func (s *EmailTemplateService) generateBillingAddress(data map[string]interface{}) string {
	if billingAddr, ok := data["billing_address"]; ok && billingAddr != nil {
		html := `<div class="section">
			<h2>🏠 Биллинговый адрес</h2>
			<div class="address-block">`

		if addr, ok := billingAddr.(*stripe.Address); ok {
			if addr.Line1 != "" {
				html += fmt.Sprintf(`<div><strong>Адрес:</strong> %s</div>`, addr.Line1)
			}
			if addr.Line2 != "" {
				html += fmt.Sprintf(`<div><strong>Доп. адрес:</strong> %s</div>`, addr.Line2)
			}
			if addr.City != "" {
				html += fmt.Sprintf(`<div><strong>Город:</strong> %s</div>`, addr.City)
			}
			if addr.State != "" {
				html += fmt.Sprintf(`<div><strong>Штат/Регион:</strong> %s</div>`, addr.State)
			}
			if addr.PostalCode != "" {
				html += fmt.Sprintf(`<div><strong>Почтовый код:</strong> %s</div>`, addr.PostalCode)
			}
			if addr.Country != "" {
				html += fmt.Sprintf(`<div><strong>Страна:</strong> %s</div>`, addr.Country)
			}
		}
		html += `</div></div>`
		return html
	}
	return ""
}

func (s *EmailTemplateService) generateArtworksInfo(arts []entity.Art) string {
	if len(arts) > 0 {
		html := `<div class="section">
			<h2>🖼️ Заказанные картины</h2>`

		for _, art := range arts {
			html += fmt.Sprintf(`<div class="artwork-card">
				<div class="artwork-header">
					<h3>%s (ID: %d)</h3>
				</div>
				<div class="artwork-details">
					<div>
						<div class="info-row">
							<span class="info-label">Цена:</span>
							<span class="info-value">$%d</span>
						</div>
						<div class="info-row">
							<span class="info-label">Год:</span>
							<span class="info-value">%d</span>
						</div>
						<div class="info-row">
							<span class="info-label">Размеры:</span>
							<span class="info-value">%d x %d см</span>
						</div>
					</div>
					<div>
						<div class="info-row">
							<span class="info-label">Автор:</span>
							<span class="info-value">%s</span>
						</div>
						<div class="info-row">
							<span class="info-label">Техника:</span>
							<span class="info-value">%s</span>
						</div>
						<div class="info-row">
							<span class="info-label">Материал:</span>
							<span class="info-value">%s</span>
						</div>
					</div>
				</div>
				<p><em>Все фотографии этой картины прикреплены к письму отдельными файлами.</em></p>
			</div>`,
				art.Name.EN, art.ID, art.Price, art.Year, art.DimensionX, art.DimensionY,
				func() string {
					if art.Author.Name.EN != "" {
						return art.Author.Name.EN
					}
					return "Неизвестный автор"
				}(),
				art.Technique.EN, art.Medium.EN)
		}

		html += `</div>`
		return html
	} else {
		return `<div class="section">
			<h2>⚠️ Внимание</h2>
			<p>Не удалось определить заказанные картины из данных платежа.</p>
		</div>`
	}
}

func (s *EmailTemplateService) generateFooterInfo(data map[string]interface{}) string {
	sessionId := "N/A"
	if id, ok := data["session_id"]; ok && id != nil {
		sessionId = fmt.Sprintf("%v", id)
	} else if id, ok := data["payment_intent_id"]; ok && id != nil {
		sessionId = fmt.Sprintf("%v", id)
	}
	return fmt.Sprintf("ID транзакции: %s", sessionId)
}

// Собирает все фотографии артов как прикрепления к письму
func (s *EmailTemplateService) collectArtPhotoAttachments(arts []entity.Art) ([]EmailAttachment, error) {
	var attachments []EmailAttachment

	for _, art := range arts {
		artPrefix := fmt.Sprintf("art_%d", art.ID)

		// Главная фотография
		if art.MainPhotoID != nil && art.MainPhoto.Path != "" {
			attachment, err := s.createPhotoAttachment(art.MainPhoto.Path, fmt.Sprintf("%s_main", artPrefix))
			if err != nil {
				log.Printf("Failed to create main photo attachment for art %d: %v", art.ID, err)
			} else {
				attachments = append(attachments, attachment)
			}
		}

		// Превью фотография
		if art.PreviewPhotoID != nil && art.PreviewPhoto.Path != "" {
			attachment, err := s.createPhotoAttachment(art.PreviewPhoto.Path, fmt.Sprintf("%s_preview", artPrefix))
			if err != nil {
				log.Printf("Failed to create preview photo attachment for art %d: %v", art.ID, err)
			} else {
				attachments = append(attachments, attachment)
			}
		}

		// Дополнительные фотографии
		for i, photo := range art.Photos {
			if !photo.IsMain && !photo.IsPreview && photo.Path != "" {
				attachment, err := s.createPhotoAttachment(photo.Path, fmt.Sprintf("%s_photo_%d", artPrefix, i+1))
				if err != nil {
					log.Printf("Failed to create photo attachment %d for art %d: %v", i+1, art.ID, err)
				} else {
					attachments = append(attachments, attachment)
				}
			}
		}
	}

	return attachments, nil
}

// Создает прикрепление для фотографии
func (s *EmailTemplateService) createPhotoAttachment(photoPath, filenamePrefix string) (EmailAttachment, error) {
	// Конвертируем относительный путь в абсолютный путь к файлу
	var fullPath string
	if strings.HasPrefix(photoPath, "/uploads/") {
		// Убираем /uploads/ из начала
		relativePath := strings.TrimPrefix(photoPath, "/uploads/")
		parts := strings.Split(relativePath, "/")
		if len(parts) >= 2 {
			fullPath = config.GetUploadFilePath(parts[0], strings.Join(parts[1:], "/"))
		} else {
			return EmailAttachment{}, fmt.Errorf("invalid photo path format: %s", photoPath)
		}
	} else {
		return EmailAttachment{}, fmt.Errorf("photo path does not start with /uploads/: %s", photoPath)
	}

	// Проверяем, что файл существует
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return EmailAttachment{}, fmt.Errorf("photo file not found: %s", fullPath)
	}

	// Читаем файл
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return EmailAttachment{}, fmt.Errorf("failed to read photo file %s: %w", fullPath, err)
	}

	// Определяем имя файла
	ext := filepath.Ext(photoPath)
	filename := fmt.Sprintf("%s%s", filenamePrefix, ext)

	// Определяем MIME тип
	mimeType := getMimeTypeFromExtension(ext)

	return EmailAttachment{
		Filename: filename,
		Content:  content,
		MimeType: mimeType,
	}, nil
}
