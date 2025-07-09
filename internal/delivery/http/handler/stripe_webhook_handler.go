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
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /recive-checkoutevent [post]
func (h *StripeWebhookHandler) HandleStripeWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	sig := c.GetHeader("Stripe-Signature")

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
	var session stripe.CheckoutSession
	err := json.Unmarshal(event.Data.Raw, &session)
	if err != nil {
		log.Printf("Error parsing checkout session: %v", err)
		return
	}

	log.Printf("Checkout session completed: %s", session.ID)

	var purchasedArts []entity.Art
	var totalAmount int64 = session.AmountTotal

	if len(session.LineItems.Data) > 0 {
		for _, item := range session.LineItems.Data {
			if item.Price != nil && item.Price.Product != nil {
				productID := item.Price.Product.ID
				art, err := h.findArtByStripeProductID(productID)
				if err == nil {
					purchasedArts = append(purchasedArts, art)
				}
			}
		}
	}

	go h.sendAdminNotificationWithArtInfo("checkout.session.completed", map[string]interface{}{
		"session_id":     session.ID,
		"amount_total":   totalAmount,
		"currency":       session.Currency,
		"customer_email": session.CustomerDetails.Email,
		"customer_name":  session.CustomerDetails.Name,
		"payment_status": session.PaymentStatus,
	}, purchasedArts)
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

	go h.sendAdminNotificationWithArtInfo("payment_intent.succeeded", map[string]interface{}{
		"payment_intent_id": paymentIntent.ID,
		"amount":            paymentIntent.Amount,
		"currency":          paymentIntent.Currency,
		"status":            paymentIntent.Status,
		"description":       paymentIntent.Description,
	}, purchasedArts)
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

	subject := fmt.Sprintf("New Payment Received - %s", eventType)

	var body string
	switch eventType {
	case "checkout.session.completed":
		body = fmt.Sprintf(`
			<h2>New Payment Completed</h2>
			<p><strong>Event Type:</strong> Checkout Session Completed</p>
			<p><strong>Session ID:</strong> %v</p>
			<p><strong>Amount:</strong> %v %v</p>
			<p><strong>Customer Email:</strong> %v</p>
			<p><strong>Customer Name:</strong> %v</p>
			<p><strong>Payment Status:</strong> %v</p>
		`,
			data["session_id"],
			data["amount_total"], data["currency"],
			data["customer_email"],
			data["customer_name"],
			data["payment_status"])

	case "payment_intent.succeeded":
		body = fmt.Sprintf(`
			<h2>New Payment Completed</h2>
			<p><strong>Event Type:</strong> Payment Intent Succeeded</p>
			<p><strong>Payment Intent ID:</strong> %v</p>
			<p><strong>Amount:</strong> %v %v</p>
			<p><strong>Status:</strong> %v</p>
			<p><strong>Description:</strong> %v</p>
		`,
			data["payment_intent_id"],
			data["amount"], data["currency"],
			data["status"],
			data["description"])
	}

	if len(arts) > 0 {
		body += "<h3>Purchased Artworks:</h3><ul>"
		for _, art := range arts {
			body += fmt.Sprintf(`
				<li>
					<strong>ID:</strong> %d<br>
					<strong>Name:</strong> %s<br>
					<strong>Price:</strong> %d USD<br>
					<strong>Author:</strong> %s<br>
					<strong>Year:</strong> %d<br>
					<strong>Dimensions:</strong> %dx%d<br>
				</li>
			`, art.ID, art.Name.EN, art.Price,
				func() string {
					if art.Author.Name.EN != "" {
						return art.Author.Name.EN
					} else {
						return "Unknown"
					}
				}(),
				art.Year, art.DimensionX, art.DimensionY)
		}
		body += "</ul>"
	} else {
		body += "<p><em>Unable to identify purchased artwork from payment data.</em></p>"
	}

	err := h.emailService.SendEmail(adminEmail, subject, body)
	if err != nil {
		log.Printf("Failed to send admin notification email: %v", err)
	} else {
		log.Printf("Admin notification email sent successfully for %s", eventType)
	}
}
