package service

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/balance"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/paymentlink"
	"github.com/stripe/stripe-go/v82/price"
	"github.com/stripe/stripe-go/v82/product"
	"github.com/stripe/stripe-go/v82/refund"
)

// StripeService представляет сервис для работы с Stripe API
type StripeService struct {
	testKey string
	prodKey string
	isLive  bool
}

// ProductWithPrice представляет продукт с его основной ценой
type ProductWithPriceAndLink struct {
	Product *stripe.Product     `json:"product"`
	Price   *stripe.Price       `json:"price"`
	Link    *stripe.PaymentLink `json:"link"`
}

// NewStripeService создает новый экземпляр StripeService
func NewStripeService(testKey, prodKey string, isLive bool) *StripeService {
	service := &StripeService{
		testKey: testKey,
		prodKey: prodKey,
		isLive:  isLive,
	}

	// Устанавливаем ключ в зависимости от режима
	if isLive {
		stripe.Key = prodKey
	} else {
		stripe.Key = testKey
	}

	return service
}

// ===== ПРОДУКТЫ =====

// CreateProduct создает продукт в Stripe с ценой
func (s *StripeService) CreateProduct(name, description string, imageURLs []string, priceAmount int64, currency string) (*ProductWithPriceAndLink, error) {
	if name == "" {
		return nil, errors.New("название продукта обязательно")
	}
	if priceAmount <= 0 {
		return nil, errors.New("цена должна быть больше 0")
	}
	if currency == "" {
		currency = "usd"
	}

	// 1. Создаем продукт
	var productParams *stripe.ProductParams
	if description != "" {
		productParams = &stripe.ProductParams{
			Name:        stripe.String(name),
			Description: stripe.String(description),
			Active:      stripe.Bool(true),
			Type:        stripe.String("good"),
			Shippable:   stripe.Bool(true),
		}
	} else {
		productParams = &stripe.ProductParams{
			Name:      stripe.String(name),
			Active:    stripe.Bool(true),
			Type:      stripe.String("good"),
			Shippable: stripe.Bool(true),
		}
	}

	// Добавляем изображения
	if len(imageURLs) > 0 {
		for _, imageURL := range imageURLs {
			productParams.Images = append(productParams.Images, stripe.String(imageURL))
		}
	}

	stripeProduct, err := product.New(productParams)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать продукт: %w", err)
	}

	priceParams := &stripe.PriceParams{
		Currency:   stripe.String(strings.ToLower(currency)),
		Product:    stripe.String(stripeProduct.ID),
		UnitAmount: stripe.Int64(priceAmount),
	}

	stripePrice, err := price.New(priceParams)
	if err != nil {
		s.DeleteProduct(stripeProduct.ID)
		return nil, fmt.Errorf("не удалось создать цену для продукта: %w", err)
	}

	paymentLink, err := s.CreatePaymentLink(stripeProduct.ID, 1)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать платежную ссылку: %w", err)
	}

	result := &ProductWithPriceAndLink{
		Product: stripeProduct,
		Price:   stripePrice,
		Link:    paymentLink,
	}

	log.Printf("Продукт с ценой и ссылкой создан: %s (ID: %s, Price: %d %s Link: %s)",
		name, stripeProduct.ID, priceAmount, currency, paymentLink.URL)

	return result, nil
}

// GetProduct получает продукт по ID с его основной ценой
func (s *StripeService) GetProduct(productID string) (*ProductWithPriceAndLink, error) {
	if productID == "" {
		return nil, errors.New("ID продукта обязателен")
	}

	stripeProduct, err := product.Get(productID, nil)
	if err != nil {
		return nil, fmt.Errorf("продукт не найден: %w", err)
	}

	// Получаем активную цену для продукта
	prices, err := s.ListPricesForProduct(productID, func(b bool) *bool { return &b }(true))
	if err != nil {
		return nil, fmt.Errorf("не удалось получить цену продукта: %w", err)
	}

	// paymentLink, err := s.GetPaymentLink(
	if err != nil {
		return nil, fmt.Errorf("не удалось создать платежную ссылку: %w", err)
	}

	var mainPrice *stripe.Price
	if len(prices) > 0 {
		mainPrice = prices[0] // Берем первую активную цену
	}

	return &ProductWithPriceAndLink{
		Product: stripeProduct,
		Price:   mainPrice,
		// Link:    paymentLink,
	}, nil
}

// UpdateProduct обновляет продукт и его цену
// Для удаления полей передайте пустую строку для description, пустой слайс для imageURLs
// priceAmount = 0 означает не обновлять цену
func (s *StripeService) UpdateProduct(productID string, name *string, description *string, imageURLs *[]string, priceAmount *int64, currency *string, active *bool) (*ProductWithPriceAndLink, error) {
	if productID == "" {
		return nil, errors.New("ID продукта обязателен")
	}

	currentProduct, err := s.GetProduct(productID)
	if err != nil {
		return nil, err
	}

	productParams := &stripe.ProductParams{}

	if name != nil && *name != "" {
		productParams.Name = stripe.String(*name)
	}

	if description != nil {
		if *description == "" {
			productParams.Description = stripe.String("")
		} else {
			productParams.Description = stripe.String(*description)
		}
	}

	if imageURLs != nil {
		if len(*imageURLs) == 0 {
			productParams.Images = []*string{}
		} else {
			for _, imageURL := range *imageURLs {
				productParams.Images = append(productParams.Images, stripe.String(imageURL))
			}
		}
	}

	if active != nil {
		productParams.Active = stripe.Bool(*active)
	}

	updatedProduct, err := product.Update(productID, productParams)
	if err != nil {
		return nil, fmt.Errorf("не удалось обновить продукт: %w", err)
	}

	var updatedPrice *stripe.Price = currentProduct.Price
	if priceAmount != nil && *priceAmount > 0 && (currentProduct.Price == nil || *priceAmount != currentProduct.Price.UnitAmount) {
		currencyValue := "usd"
		if currency != nil && *currency != "" {
			currencyValue = *currency
		}

		// Деактивируем старую цену
		if currentProduct.Price != nil {
			price.Update(currentProduct.Price.ID, &stripe.PriceParams{
				Active: stripe.Bool(false),
			})
		}

		// Создаем новую цену
		priceParams := &stripe.PriceParams{
			Currency:   stripe.String(strings.ToLower(currencyValue)),
			Product:    stripe.String(productID),
			UnitAmount: stripe.Int64(*priceAmount),
		}

		newPrice, err := price.New(priceParams)
		if err != nil {
			log.Printf("Не удалось создать новую цену: %v", err)
		} else {
			updatedPrice = newPrice
		}
	}

	if active != nil && !*active {
		result := &ProductWithPriceAndLink{
			Product: updatedProduct,
			Price:   updatedPrice,
		}

		log.Printf("Продукт обновлен: ID %s", productID)
		return result, nil
	}

	paymentLink, err := s.CreatePaymentLink(productID, 1)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать платежную ссылку: %w", err)
	}

	result := &ProductWithPriceAndLink{
		Product: updatedProduct,
		Price:   updatedPrice,
		Link:    paymentLink,
	}

	log.Printf("Продукт обновлен: ID %s", productID)
	return result, nil
}

// DeleteProduct архивирует продукт (в Stripe нельзя полностью удалить)
func (s *StripeService) DeleteProduct(productID string) error {
	if productID == "" {
		return errors.New("ID продукта обязателен")
	}

	// Сначала деактивируем все цены продукта
	prices, err := s.ListPricesForProduct(productID, nil)
	if err == nil {
		for _, p := range prices {
			price.Update(p.ID, &stripe.PriceParams{
				Active: stripe.Bool(false),
			})
		}
	}

	// Затем деактивируем продукт
	params := &stripe.ProductParams{
		Active: stripe.Bool(false),
	}

	_, err = product.Update(productID, params)
	if err != nil {
		return fmt.Errorf("не удалось архивировать продукт: %w", err)
	}

	log.Printf("Продукт архивирован: ID %s", productID)
	return nil
}

// ===== ЦЕНЫ (для внутреннего использования) =====

// CreatePrice создает цену для продукта (используется внутренне)
func (s *StripeService) CreatePrice(productID string, unitAmount int64, currency string) (*stripe.Price, error) {
	if productID == "" {
		return nil, errors.New("ID продукта обязателен")
	}
	if unitAmount <= 0 {
		return nil, errors.New("цена должна быть больше 0")
	}
	if currency == "" {
		currency = "usd"
	}

	params := &stripe.PriceParams{
		Currency:   stripe.String(strings.ToLower(currency)),
		Product:    stripe.String(productID),
		UnitAmount: stripe.Int64(unitAmount),
	}

	price, err := price.New(params)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать цену: %w", err)
	}

	log.Printf("Цена создана: %d %s для продукта %s (ID: %s)",
		unitAmount, currency, productID, price.ID)
	return price, nil
}

// GetPrice получает цену по ID
func (s *StripeService) GetPrice(priceID string) (*stripe.Price, error) {
	if priceID == "" {
		return nil, errors.New("ID цены обязателен")
	}

	price, err := price.Get(priceID, nil)
	if err != nil {
		return nil, fmt.Errorf("цена не найдена: %w", err)
	}

	return price, nil
}

// ListPricesForProduct получает все цены для продукта
func (s *StripeService) ListPricesForProduct(productID string, active *bool) ([]*stripe.Price, error) {
	if productID == "" {
		return nil, errors.New("ID продукта обязателен")
	}

	params := &stripe.PriceListParams{
		Product: stripe.String(productID),
	}

	if active != nil {
		params.Active = stripe.Bool(*active)
	}

	iter := price.List(params)
	var prices []*stripe.Price

	for iter.Next() {
		prices = append(prices, iter.Price())
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("ошибка получения цен: %w", err)
	}

	return prices, nil
}

// ===== ПЛАТЕЖНЫЕ ССЫЛКИ =====

// CreatePaymentLink создает платежную ссылку для продукта
func (s *StripeService) CreatePaymentLink(productID string, quantity int64) (*stripe.PaymentLink, error) {
	if productID == "" {
		return nil, errors.New("ID продукта обязателен")
	}
	if quantity <= 0 {
		quantity = 1
	}

	productWithPrice, err := s.GetProduct(productID)
	if err != nil {
		return nil, fmt.Errorf("не удалось найти продукт: %w", err)
	}

	if productWithPrice.Price == nil {
		return nil, errors.New("у продукта нет активной цены")
	}

	params := &stripe.PaymentLinkParams{
		LineItems: []*stripe.PaymentLinkLineItemParams{
			{
				Price:    stripe.String(productWithPrice.Price.ID),
				Quantity: stripe.Int64(quantity),
			},
		},
		ShippingAddressCollection: &stripe.PaymentLinkShippingAddressCollectionParams{
			AllowedCountries: []*string{
				stripe.String("US"), stripe.String("CA"), stripe.String("GB"),
				stripe.String("DE"), stripe.String("FR"), stripe.String("RU"),
			},
		},
	}

	paymentLink, err := paymentlink.New(params)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать платежную ссылку: %w", err)
	}

	// log.Printf("Платежная ссылка создана: %s (ID: %s)", paymentLink.URL, paymentLink.ID)
	return paymentLink, nil
}

// GetPaymentLink получает платежную ссылку по ID
func (s *StripeService) GetPaymentLink(linkID string) (*stripe.PaymentLink, error) {
	if linkID == "" {
		return nil, errors.New("ID ссылки обязателен")
	}

	link, err := paymentlink.Get(linkID, nil)
	if err != nil {
		return nil, fmt.Errorf("платежная ссылка не найдена: %w", err)
	}

	return link, nil
}

// ListCustomers получает список покупателей
func (s *StripeService) ListCustomers(limit int64) ([]*stripe.Customer, error) {
	params := &stripe.CustomerListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(limit),
		},
	}

	iter := customer.List(params)
	var customers []*stripe.Customer

	for iter.Next() {
		customers = append(customers, iter.Customer())
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("ошибка получения покупателей: %w", err)
	}

	return customers, nil
}

// ===== ПЛАТЕЖИ =====

// ListPaymentIntents получает список платежей
func (s *StripeService) ListPaymentIntents(customerID string, limit int64) ([]*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(limit),
		},
	}

	if customerID != "" {
		params.Customer = stripe.String(customerID)
	}

	iter := paymentintent.List(params)
	var payments []*stripe.PaymentIntent

	for iter.Next() {
		payments = append(payments, iter.PaymentIntent())
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("ошибка получения платежей: %w", err)
	}

	return payments, nil
}

// ===== УТИЛИТЫ =====
func (s *StripeService) CreateRefund(paymentIntentID string, amount int64) (*stripe.Refund, error) {
	if paymentIntentID == "" {
		return nil, errors.New("ID платежа обязателен")
	}

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}

	if amount > 0 {
		params.Amount = stripe.Int64(amount)
	}

	refund, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать возврат: %w", err)
	}

	log.Printf("Возврат создан: %s", refund.ID)
	return refund, nil
}

func (s *StripeService) GetBalance() (*stripe.Balance, error) {
	balance, err := balance.Get(nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить баланс: %w", err)
	}

	return balance, nil
}

func (s *StripeService) IsTestMode() bool {
	return !s.isLive
}

func (s *StripeService) SwitchToLiveMode() {
	s.isLive = true
	stripe.Key = s.prodKey
	log.Println("🔴 ПЕРЕКЛЮЧЕНО НА БОЕВОЙ РЕЖИМ")
}

func (s *StripeService) SwitchToTestMode() {
	s.isLive = false
	stripe.Key = s.testKey
	log.Println("🟡 Переключено на тестовый режим")
}

// // ListProducts получает список продуктов с их ценами
// func (s *StripeService) ListProducts(limit int64, active *bool) ([]*ProductWithPrice, error) {
// 	params := &stripe.ProductListParams{
// 		ListParams: stripe.ListParams{
// 			Limit: stripe.Int64(limit),
// 		},
// 		Type: stripe.String("good"),
// 	}

// 	if active != nil {
// 		params.Active = stripe.Bool(*active)
// 	}

// 	iter := product.List(params)
// 	var products []*ProductWithPrice

// 	for iter.Next() {
// 		stripeProduct := iter.Product()

// 		// Получаем активную цену для каждого продукта
// 		prices, _ := s.ListPricesForProduct(stripeProduct.ID, func(b bool) *bool { return &b }(true))
// 		var mainPrice *stripe.Price
// 		if len(prices) > 0 {
// 			mainPrice = prices[0]
// 		}

// 		products = append(products, &ProductWithPrice{
// 			Product: stripeProduct,
// 			Price:   mainPrice,
// 		})
// 	}

// 	if err := iter.Err(); err != nil {
// 		return nil, fmt.Errorf("ошибка получения продуктов: %w", err)
// 	}

// 	return products, nil
// }
