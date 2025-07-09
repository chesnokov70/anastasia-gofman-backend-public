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

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/webhook"
)

type StripeWebhookHandler struct {
	emailService         *service.EmailService
	artService           service.ArtService
	emailTemplateService *service.EmailTemplateService
}

func NewStripeWebhookHandler(emailService *service.EmailService, artService service.ArtService) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		emailService:         emailService,
		artService:           artService,
		emailTemplateService: service.NewEmailTemplateService(),
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

	var processingError error
	switch event.Type {
	case "checkout.session.completed":
		processingError = h.handleCheckoutSessionCompleted(event)
	case "payment_intent.succeeded":
		processingError = h.handlePaymentIntentSucceeded(event)
	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}

	if processingError != nil {
		log.Printf("Error processing webhook: %v", processingError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process webhook",
			"retry": true,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

func (h *StripeWebhookHandler) handleCheckoutSessionCompleted(event stripe.Event) error {
	var sessionStub stripe.CheckoutSession
	err := json.Unmarshal(event.Data.Raw, &sessionStub)
	if err != nil {
		log.Printf("Error parsing checkout session stub: %v", err)
		return err
	}

	log.Printf("Checkout session completed event received for session ID: %s", sessionStub.ID)

	params := &stripe.CheckoutSessionParams{}
	params.AddExpand("line_items")
	session, err := session.Get(sessionStub.ID, params)
	if err != nil {
		log.Printf("Error retrieving full checkout session from Stripe: %v", err)
		return err
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

	orderData := map[string]interface{}{
		"session_id":     session.ID,
		"amount_total":   totalAmount,
		"currency":       session.Currency,
		"payment_status": session.PaymentStatus,
		"session":        session,
	}

	if session.CustomerDetails != nil {
		orderData["customer_email"] = session.CustomerDetails.Email
		orderData["customer_name"] = session.CustomerDetails.Name
		orderData["customer_phone"] = session.CustomerDetails.Phone
		orderData["billing_address"] = session.CustomerDetails.Address
	}

	if session.CollectedInformation != nil && session.CollectedInformation.ShippingDetails != nil {
		orderData["shipping_name"] = session.CollectedInformation.ShippingDetails.Name
		orderData["shipping_address"] = session.CollectedInformation.ShippingDetails.Address
	}

	for _, art := range purchasedArts {
		_, err := h.artService.ChangeArtTypeAfterBuy(art.StripeProductID)
		if err != nil {
			log.Printf("Error changing art type after buy: %v", err)
		}
	}
	return h.sendAdminNotificationWithArtInfo("checkout.session.completed", orderData, purchasedArts)
}

func (h *StripeWebhookHandler) handlePaymentIntentSucceeded(event stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	err := json.Unmarshal(event.Data.Raw, &paymentIntent)
	if err != nil {
		log.Printf("Error parsing payment intent: %v", err)
		return err
	}

	log.Printf("Payment intent succeeded: %s", paymentIntent.ID)

	var purchasedArts []entity.Art
	if productID, exists := paymentIntent.Metadata["product_id"]; exists {
		art, err := h.findArtByStripeProductID(productID)
		if err == nil {
			purchasedArts = append(purchasedArts, art)
		}
	}

	orderData := map[string]interface{}{
		"payment_intent_id": paymentIntent.ID,
		"amount":            paymentIntent.Amount,
		"currency":          paymentIntent.Currency,
		"status":            paymentIntent.Status,
		"description":       paymentIntent.Description,
	}

	go h.sendAdminNotificationWithArtInfo("payment_intent.succeeded", orderData, purchasedArts)
	return nil
}

func (h *StripeWebhookHandler) findArtByStripeProductID(productID string) (entity.Art, error) {
	return h.artService.GetArtByStripeProductID(productID)
}

func (h *StripeWebhookHandler) sendAdminNotificationWithArtInfo(eventType string, data map[string]interface{}, arts []entity.Art) error {
	adminEmails := config.GetConfig().Email.Admin
	if len(adminEmails) == 0 {
		log.Printf("No admin emails configured, skipping notification")
		return nil
	}

	subject := fmt.Sprintf("🎨 Новый заказ картины - Платеж подтвержден!")

	htmlBody, attachments, err := h.emailTemplateService.GeneratePaymentNotificationHTML(eventType, data, arts)
	if err != nil {
		log.Printf("Failed to generate email template: %v", err)
		htmlBody = h.generateSimpleFallbackEmail(eventType, data, arts)
		attachments = []service.EmailAttachment{}
	}

	err = h.emailService.SendToAllAdminsWithAttachments(subject, htmlBody, attachments)
	if err != nil {
		log.Printf("Failed to send admin notification emails: %v", err)
		return err // Вернуть ошибку для retry
	}

	log.Printf("Admin notification emails sent successfully to %d admin(s)", len(adminEmails))
	return nil
}

func (h *StripeWebhookHandler) generateSimpleFallbackEmail(eventType string, data map[string]interface{}, arts []entity.Art) string {
	html := `<html><body><h2>🎨 Новый заказ картины</h2>`

	if eventType == "checkout.session.completed" {
		amountInDollars := float64(data["amount_total"].(int64)) / 100.0
		html += fmt.Sprintf(`
			<p><strong>Сумма:</strong> $%.2f %v</p>
			<p><strong>ID сессии:</strong> %v</p>
			<p><strong>Email клиента:</strong> %v</p>
			<p><strong>Имя клиента:</strong> %v</p>`,
			amountInDollars, data["currency"], data["session_id"],
			data["customer_email"], data["customer_name"])
	}

	if len(arts) > 0 {
		html += `<h3>Заказанные картины:</h3><ul>`
		for _, art := range arts {
			html += fmt.Sprintf(`<li>%s (ID: %d) - $%d</li>`, art.Name.EN, art.ID, art.Price)
		}
		html += `</ul>`
	}

	html += `<p><strong>Внимание:</strong> Не удалось загрузить полный шаблон письма. Фотографии не прикреплены.</p>`
	html += `</body></html>`

	return html
}
