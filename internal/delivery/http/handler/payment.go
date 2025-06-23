package handler

import (
	"anastasia_gofman_backend/internal/delivery/http/dto"
	"net/http"
	"strconv"

	"anastasia_gofman_backend/pkg/service"

	"github.com/gin-gonic/gin"
)

// PaymentHandler обрабатывает запросы связанные с платежами
type PaymentHandler struct {
	stripeService *service.StripeService
}

// NewPaymentHandler создает новый PaymentHandler
func NewPaymentHandler(stripeService *service.StripeService) *PaymentHandler {
	return &PaymentHandler{
		stripeService: stripeService,
	}
}

// CreateProduct создает продукт в Stripe с ценой
// @Summary Create a product
// @Description Creates a product in Stripe with a price.
// @Tags Payments
// @Accept json
// @Produce json
// @Param product body dto.CreateProductDTO true "Product info"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/payments/products [post]
func (h *PaymentHandler) CreateProduct(c *gin.Context) {
	var req dto.CreateProductDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Currency == "" {
		req.Currency = "usd"
	}

	productWithPrice, err := h.stripeService.CreateProduct(
		req.Name,
		req.Description,
		req.ImageURLs,
		req.Price,
		req.Currency,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Продукт с ценой создан",
		"data":    productWithPrice,
	})
}

// UpdateProduct обновляет продукт
// @Summary Update a product
// @Description Updates a product in Stripe.
// @Tags Payments
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body dto.UpdateProductDTO true "Product update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/payments/products/{id} [put]
func (h *PaymentHandler) UpdateProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID продукта обязателен"})
		return
	}

	var req dto.UpdateProductDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	productWithPrice, err := h.stripeService.UpdateProduct(
		productID,
		req.Name,
		req.Description,
		req.ImageURLs,
		req.Price,
		req.Currency,
		req.Active,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Продукт обновлен",
		"data":    productWithPrice,
	})
}

// CreatePaymentLink создает платежную ссылку для продукта
// @Summary Create a payment link
// @Description Creates a payment link for a product.
// @Tags Payments
// @Accept json
// @Produce json
// @Param link_request body dto.CreatePaymentLinkDTO true "Payment link request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/payments/payment-links [post]
func (h *PaymentHandler) CreatePaymentLink(c *gin.Context) {
	var req dto.CreatePaymentLinkDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	paymentLink, err := h.stripeService.CreatePaymentLink(req.ProductID, req.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Платежная ссылка создана",
		"payment_link": paymentLink,
		"url":          paymentLink.URL,
	})
}

// GetProduct получает продукт по ID с его ценой
// @Summary Get a product
// @Description Get a product by ID with its price from Stripe.
// @Tags Payments
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/payments/products/{id} [get]
func (h *PaymentHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID продукта обязателен"})
		return
	}

	productWithPrice, err := h.stripeService.GetProduct(productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": productWithPrice,
	})
}

// DeleteProduct удаляет (архивирует) продукт
// @Summary Delete a product
// @Description Deletes (archives) a product in Stripe.
// @Tags Payments
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/payments/products/{id} [delete]
func (h *PaymentHandler) DeleteProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID продукта обязателен"})
		return
	}

	err := h.stripeService.DeleteProduct(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Продукт архивирован",
	})
}

// GetBalance получает баланс аккаунта
// @Summary Get balance
// @Description Get the account balance from Stripe.
// @Tags Payments
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/payments/balance [get]
func (h *PaymentHandler) GetBalance(c *gin.Context) {
	balance, err := h.stripeService.GetBalance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": balance,
		"mode":    map[string]bool{"test_mode": h.stripeService.IsTestMode()},
	})
}

// GetCustomers получает список покупателей
// @Summary Get customers
// @Description Get a list of customers from Stripe.
// @Tags Payments
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/payments/customers [get]
func (h *PaymentHandler) GetCustomers(c *gin.Context) {
	limit := int64(10)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil {
			limit = parsed
		}
	}

	customers, err := h.stripeService.ListCustomers(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customers": customers,
		"count":     len(customers),
	})
}

// GetPayments получает список платежей
// @Summary Get payments
// @Description Get a list of payment intents from Stripe.
// @Tags Payments
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param customer_id query string false "Filter by customer ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/payments/payment-intents [get]
func (h *PaymentHandler) GetPayments(c *gin.Context) {
	limit := int64(10)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil {
			limit = parsed
		}
	}

	customerID := c.Query("customer_id")

	payments, err := h.stripeService.ListPaymentIntents(customerID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payments": payments,
		"count":    len(payments),
	})
}

// CreateRefund создает возврат для платежа
// @Summary Create a refund
// @Description Creates a refund for a payment in Stripe.
// @Tags Payments
// @Accept json
// @Produce json
// @Param refund_request body dto.CreateRefundDTO true "Refund request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/payments/refunds [post]
func (h *PaymentHandler) CreateRefund(c *gin.Context) {
	var req dto.CreateRefundDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refund, err := h.stripeService.CreateRefund(req.PaymentIntentID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Возврат создан",
		"refund":  refund,
	})
}
