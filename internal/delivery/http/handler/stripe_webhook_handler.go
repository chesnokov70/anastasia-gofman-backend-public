package handler

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/service"
	"anastasia_gofman_backend/pkg/config"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/webhook"
)

type StripeWebhookHandler struct {
	emailService *service.EmailService
	artService   service.ArtService
}

func NewStripeWebhookHandler(emailService *service.EmailService, artService service.ArtService) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		emailService: emailService,
		artService:   artService,
	}
}

// @Summary Receive Stripe webhook events
// @Description Handle Stripe webhook events for payment confirmations
// @Tags payments
// @Accept json
// @Produce json
// @Router /recive-checkoutevent [post]
func (h *StripeWebhookHandler) HandleStripeWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	sig := c.GetHeader("Stripe-Signature")
	log.Printf("Received stripe webhook. Signature: %s", sig)
	log.Printf("Payload: %s", string(payload))

	webhookSecret := config.GetConfig().Stripe.WebhookSecret

	var event stripe.Event

	endpointSecret := webhookSecret
	event, err = webhook.ConstructEvent(payload, sig, endpointSecret)
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutSessionCompleted(event)
	case "payment_intent.succeeded":
		h.handlePaymentIntentSucceeded(event)
	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

func (h *StripeWebhookHandler) handleCheckoutSessionCompleted(event stripe.Event) {
	var sessionStub stripe.CheckoutSession
	err := json.Unmarshal(event.Data.Raw, &sessionStub)
	if err != nil {
		log.Printf("Error parsing checkout session stub: %v", err)
		return
	}

	log.Printf("Checkout session completed event received for session ID: %s", sessionStub.ID)

	params := &stripe.CheckoutSessionParams{}
	params.AddExpand("line_items")
	session, err := session.Get(sessionStub.ID, params)
	if err != nil {
		log.Printf("Error retrieving full checkout session from Stripe: %v", err)
		return
	}

	sessionJSON, _ := json.MarshalIndent(session, "", "  ")
	log.Printf("Retrieved full checkout session object from Stripe: %s", string(sessionJSON))

	var purchasedArts []entity.Art
	var totalAmount int64 = session.AmountTotal

	if session.LineItems != nil && len(session.LineItems.Data) > 0 {
		log.Printf("Processing %d line items...", len(session.LineItems.Data))
		for _, item := range session.LineItems.Data {
			if item.Price != nil && item.Price.Product != nil {
				productID := item.Price.Product.ID
				log.Printf("Found product ID in line item: %s", productID)
				art, err := h.findArtByStripeProductID(productID)
				if err == nil {
					purchasedArts = append(purchasedArts, art)
					log.Printf("Found art for product ID %s: %s", productID, art.Name.EN)
				} else {
					log.Printf("Error finding art for product ID %s: %v", productID, err)
				}
			} else {
				log.Println("Line item does not have price or product information.")
			}
		}
	} else {
		log.Println("No line items found in checkout session.")
	}

	// Собираем полную информацию о покупателе и адресах
	orderData := map[string]interface{}{
		"session_id":     session.ID,
		"amount_total":   totalAmount,
		"currency":       session.Currency,
		"payment_status": session.PaymentStatus,
		"session":        session,
	}

	// Извлекаем информацию о покупателе
	if session.CustomerDetails != nil {
		orderData["customer_email"] = session.CustomerDetails.Email
		orderData["customer_name"] = session.CustomerDetails.Name
		orderData["customer_phone"] = session.CustomerDetails.Phone
		orderData["billing_address"] = session.CustomerDetails.Address
	}

	// Извлекаем информацию об адресе доставки
	if session.CollectedInformation != nil && session.CollectedInformation.ShippingDetails != nil {
		orderData["shipping_name"] = session.CollectedInformation.ShippingDetails.Name
		orderData["shipping_address"] = session.CollectedInformation.ShippingDetails.Address
	}

	go h.sendAdminNotificationWithArtInfo("checkout.session.completed", orderData, purchasedArts)
}

func (h *StripeWebhookHandler) handlePaymentIntentSucceeded(event stripe.Event) {
	var paymentIntent stripe.PaymentIntent
	err := json.Unmarshal(event.Data.Raw, &paymentIntent)
	if err != nil {
		log.Printf("Error parsing payment intent: %v", err)
		return
	}

	log.Printf("Payment intent succeeded: %s", paymentIntent.ID)

	var purchasedArts []entity.Art
	if productID, exists := paymentIntent.Metadata["product_id"]; exists {
		art, err := h.findArtByStripeProductID(productID)
		if err == nil {
			purchasedArts = append(purchasedArts, art)
		}
	}

	// Собираем информацию о платеже
	orderData := map[string]interface{}{
		"payment_intent_id": paymentIntent.ID,
		"amount":            paymentIntent.Amount,
		"currency":          paymentIntent.Currency,
		"status":            paymentIntent.Status,
		"description":       paymentIntent.Description,
	}

	go h.sendAdminNotificationWithArtInfo("payment_intent.succeeded", orderData, purchasedArts)
}

func (h *StripeWebhookHandler) findArtByStripeProductID(productID string) (entity.Art, error) {
	return h.artService.GetArtByStripeProductID(productID)
}

func (h *StripeWebhookHandler) sendAdminNotificationWithArtInfo(eventType string, data map[string]interface{}, arts []entity.Art) {
	adminEmail := config.GetConfig().Email.Admin
	if adminEmail == "" {
		log.Printf("Admin email not configured, skipping notification")
		return
	}

	subject := fmt.Sprintf("🎨 Новый заказ картины - Платеж подтвержден!")
	body := h.generateBeautifulEmailHTML(eventType, data, arts)

	err := h.emailService.SendEmail(adminEmail, subject, body)
	if err != nil {
		log.Printf("Failed to send admin notification email: %v", err)
	} else {
		log.Printf("Admin notification email sent successfully for %s", eventType)
	}
}

func (h *StripeWebhookHandler) generateBeautifulEmailHTML(eventType string, data map[string]interface{}, arts []entity.Art) string {
	baseURL := config.GetBaseURL()

	// Начинаем с HTML структуры и CSS стилей
	html := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Новый заказ</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            background-color: #f8f9fa;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 28px;
            font-weight: 300;
        }
        .content {
            padding: 30px;
        }
        .section {
            margin-bottom: 30px;
            padding: 20px;
            border-radius: 8px;
            border-left: 4px solid #667eea;
            background-color: #f8f9fa;
        }
        .section h2 {
            margin-top: 0;
            color: #667eea;
            font-size: 20px;
            font-weight: 600;
        }
        .info-row {
            display: flex;
            margin-bottom: 10px;
            align-items: center;
        }
        .info-label {
            font-weight: 600;
            min-width: 120px;
            color: #555;
        }
        .info-value {
            color: #333;
        }
        .address-block {
            background: white;
            padding: 15px;
            border-radius: 6px;
            border: 1px solid #e9ecef;
            margin-top: 10px;
        }
        .artwork-card {
            border: 1px solid #e9ecef;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
            background: white;
        }
        .artwork-header {
            display: flex;
            align-items: center;
            margin-bottom: 15px;
        }
        .artwork-header h3 {
            margin: 0;
            color: #667eea;
            font-size: 18px;
        }
        .artwork-details {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 10px;
            margin-bottom: 15px;
        }
        .artwork-photos {
            margin-top: 15px;
        }
        .photo-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 10px;
            margin-top: 10px;
        }
        .photo-item {
            text-align: center;
        }
        .photo-item img {
            max-width: 100%;
            height: 120px;
            object-fit: cover;
            border-radius: 6px;
            border: 2px solid #e9ecef;
        }
        .photo-label {
            font-size: 12px;
            color: #666;
            margin-top: 5px;
        }
        .amount {
            font-size: 24px;
            font-weight: 700;
            color: #28a745;
        }
        .priority-notice {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 6px;
            padding: 15px;
            margin-top: 20px;
        }
        .priority-notice h3 {
            margin-top: 0;
            color: #856404;
        }
        .footer {
            background: #f8f9fa;
            padding: 20px;
            text-align: center;
            color: #666;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎨 Новый заказ картины</h1>
            <p>Поступил новый заказ, требуется отправка</p>
        </div>
        <div class="content">`

	// Информация о платеже
	html += `<div class="section">
            <h2>💳 Информация о платеже</h2>`

	if eventType == "checkout.session.completed" {
		amountInDollars := float64(data["amount_total"].(int64)) / 100.0
		html += fmt.Sprintf(`
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
		html += fmt.Sprintf(`
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

	html += `</div>`

	// Информация о покупателе
	html += `<div class="section">
            <h2>👤 Информация о покупателе</h2>`

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

	html += `</div>`

	// Адрес доставки
	if shippingAddr, ok := data["shipping_address"]; ok && shippingAddr != nil {
		html += `<div class="section">
                <h2>�� Адрес доставки</h2>
                <div class="address-block">`

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
	}

	// Биллинговый адрес (если отличается от адреса доставки)
	if billingAddr, ok := data["billing_address"]; ok && billingAddr != nil {
		html += `<div class="section">
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
	}

	if len(arts) > 0 {
		html += `<div class="section">
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
                    </div>`,
				art.Name.EN, art.ID, art.Price, art.Year, art.DimensionX, art.DimensionY,
				func() string {
					if art.Author.Name.EN != "" {
						return art.Author.Name.EN
					}
					return "Неизвестный автор"
				}(),
				art.Technique.EN, art.Medium.EN)

			// Добавляем фотографии
			photos := h.collectArtPhotos(art, baseURL)
			if len(photos) > 0 {
				html += `<div class="artwork-photos">
                        <h4>📸 Фотографии:</h4>
                        <div class="photo-grid">`

				for _, photo := range photos {
					html += fmt.Sprintf(`<div class="photo-item">
                            <img src="%s" alt="%s" />
                            <div class="photo-label">%s</div>
                        </div>`, photo.URL, photo.Label, photo.Label)
				}

				html += `</div></div>`
			}

			html += `</div>`
		}

		html += `</div>`
	} else {
		html += `<div class="section">
                <h2>⚠️ Внимание</h2>
                <p>Не удалось определить заказанные картины из данных платежа.</p>
            </div>`
	}

	html += `<div class="priority-notice">
            <h3>🚨 Требуется действие</h3>
            <p><strong>Необходимо подготовить картину к отправке по указанному адресу доставки.</strong></p>
            <p>Пожалуйста, свяжитесь с покупателем для уточнения деталей доставки и подтверждения адреса.</p>
        </div>`

	html += `</div>
        <div class="footer">
            <p>Это автоматическое уведомление от системы Anastasia Gofman Art</p>
            <p>Дата отправки: ` + fmt.Sprintf("%v", data["session_id"]) + `</p>
        </div>
    </div>
</body>
</html>`

	return html
}

type ArtPhoto struct {
	URL   string
	Label string
}

// Собираем все фотографии арта
func (h *StripeWebhookHandler) collectArtPhotos(art entity.Art, baseURL string) []ArtPhoto {
	var photos []ArtPhoto

	// Главная фотография
	if art.MainPhotoID != nil && art.MainPhoto.Path != "" {
		photoURL := baseURL + strings.TrimPrefix(art.MainPhoto.Path, "/")
		photos = append(photos, ArtPhoto{
			URL:   photoURL,
			Label: "Главное фото",
		})
	}

	// Превью фотография
	if art.PreviewPhotoID != nil && art.PreviewPhoto.Path != "" {
		photoURL := baseURL + strings.TrimPrefix(art.PreviewPhoto.Path, "/")
		photos = append(photos, ArtPhoto{
			URL:   photoURL,
			Label: "Превью",
		})
	}

	// Дополнительные фотографии
	for i, photo := range art.Photos {
		if !photo.IsMain && !photo.IsPreview && photo.Path != "" {
			photoURL := baseURL + strings.TrimPrefix(photo.Path, "/")
			photos = append(photos, ArtPhoto{
				URL:   photoURL,
				Label: fmt.Sprintf("Фото %d", i+1),
			})
		}
	}

	return photos
}
