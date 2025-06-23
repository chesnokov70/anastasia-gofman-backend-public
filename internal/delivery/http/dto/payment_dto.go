package dto

// CreateProductDTO is used for creating a new Stripe product.
type CreateProductDTO struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	ImageURLs   []string `json:"image_urls"`
	Price       int64    `json:"price" binding:"required"`
	Currency    string   `json:"currency"`
}

// UpdateProductDTO is used for updating a Stripe product.
type UpdateProductDTO struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	ImageURLs   *[]string `json:"image_urls,omitempty"`
	Price       *int64    `json:"price,omitempty"`
	Currency    *string   `json:"currency,omitempty"`
	Active      *bool     `json:"active,omitempty"`
}

// CreatePaymentLinkDTO is used for creating a Stripe payment link.
type CreatePaymentLinkDTO struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int64  `json:"quantity"`
}

// CreateCustomerDTO is used for creating a new Stripe customer.
type CreateCustomerDTO struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateRefundDTO is used for creating a refund for a Stripe payment.
type CreateRefundDTO struct {
	PaymentIntentID string `json:"payment_intent_id" binding:"required"`
	Amount          int64  `json:"amount,omitempty"`
}
