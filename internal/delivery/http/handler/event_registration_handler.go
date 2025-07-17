package handler

import (
	"anastasia_gofman_backend/internal/delivery/http/dto"
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/service"
	"anastasia_gofman_backend/pkg/config"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EventRegistrationHandler struct {
	registrationService *service.EventRegistrationService
	mailService         *service.MailService
}

type AuthorRegistrationRequestDTO struct {
	Email       string `json:"email"`
	FullName    string `json:"full_name"`
	Language    string `json:"language"`
	PhoneNumber string `json:"phone_number"`
	Portfolio   string `json:"portfolio"`
}

type AuthorRegistrationResponseDTO struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	FullName    string `json:"full_name"`
	Language    string `json:"language"`
	PhoneNumber string `json:"phone_number"`
	Portfolio   string `json:"portfolio"`
}

func ToAuthorRegistrationRequestDTO(registration entity.AuthorRegistration) AuthorRegistrationResponseDTO {
	return AuthorRegistrationResponseDTO{
		ID:          registration.ID,
		Email:       registration.Email,
		FullName:    registration.FullName,
		Language:    registration.Language,
		PhoneNumber: registration.PhoneNumber,
		Portfolio:   registration.Portfolio,
	}
}

func NewEventRegistrationHandler(
	registrationService *service.EventRegistrationService,
	mailService *service.MailService,
) *EventRegistrationHandler {
	return &EventRegistrationHandler{
		registrationService: registrationService,
		mailService:         mailService,
	}
}

// @Summary Register for event
// @Description Регистрирует пользователя на событие и отправляет письмо с подтверждением
// @Accept json
// @Produce json
// @Tags Event Registration
// @Param id path int true "Event ID"
// @Param registration body dto.EventRegistrationRequestDTO true "Registration details"
// @Success 201 {object} dto.EventRegistrationResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events/{id}/register [post]
func (h *EventRegistrationHandler) RegisterForEvent(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var req dto.EventRegistrationRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	registration, err := h.registrationService.RegisterForEvent(req.Email, req.FullName, req.Language, eventID)
	if err != nil {
		if err.Error() == "user already registered for this event" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToEventRegistrationResponseDTO(registration))
}

// @Summary Get event registrations
// @Description Получает все регистрации на конкретное событие
// @Accept json
// @Produce json
// @Tags Event Registration
// @Param id path int true "Event ID"
// @Success 200 {array} dto.EventRegistrationResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events/{id}/registrations [get]
func (h *EventRegistrationHandler) GetEventRegistrations(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	registrations, err := h.registrationService.GetRegistrationsByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []dto.EventRegistrationResponseDTO
	for _, registration := range registrations {
		response = append(response, dto.ToEventRegistrationResponseDTO(registration))
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get all registrations
// @Description Получает все регистрации на события
// @Accept json
// @Produce json
// @Tags Event Registration
// @Success 200 {array} dto.EventRegistrationResponseDTO
// @Failure 500 {object} map[string]string
// @Router /api/registrations [get]
func (h *EventRegistrationHandler) GetAllRegistrations(c *gin.Context) {
	registrations, err := h.registrationService.GetAllRegistrations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []dto.EventRegistrationResponseDTO
	for _, registration := range registrations {
		response = append(response, dto.ToEventRegistrationResponseDTO(registration))
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Create mail template
// @Description Создает новый шаблон письма для уведомлений о регистрации на событие, можно использовать вот таких ребят: - `full_name` -  полное имя мальчишки, которое он указал при регистрации, - `event_title` - название события на языке пользователя, - `event_start_date` - дата начала события, - `event_end_date` - дата окончания события, - `event_location` - место проведения события на языке пользователя
// @Accept json
// @Produce json
// @Tags Mail Templates
// @Param mail body dto.CreateMailDTO true "Mail template"
// @Success 201 {object} dto.MailDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/mail-templates [post]
func (h *EventRegistrationHandler) CreateMailTemplate(c *gin.Context) {
	var req dto.CreateMailDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mail, err := h.mailService.CreateMail(req.ToEntity())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToMailDTO(mail))
}

// @Summary Get mail templates
// @Description Получает все шаблоны писем
// @Accept json
// @Produce json
// @Tags Mail Templates
// @Success 200 {array} dto.MailDTO
// @Failure 500 {object} map[string]string
// @Router /api/mail-templates [get]
func (h *EventRegistrationHandler) GetMailTemplates(c *gin.Context) {
	mails, err := h.mailService.GetAllMails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []dto.MailDTO
	for _, mail := range mails {
		response = append(response, dto.ToMailDTO(mail))
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Register for author
// @Description Регистрирует автора и отправляет письмо с подтверждением ему и админу
// @Accept json
// @Produce json
// @Tags Author Registration
// @Param registration body AuthorRegistrationRequestDTO true "Registration details"
// @Success 201 {object} AuthorRegistrationResponseDTO
// @Router /api/author-registrations [post]
func (h *EventRegistrationHandler) RegisterForAuthor(c *gin.Context) {
	var req AuthorRegistrationRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	registration, err := h.registrationService.RegisterForAuthor(req.Email, req.FullName, req.Language, req.PhoneNumber, req.Portfolio)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ToAuthorRegistrationRequestDTO(registration))
}

// @Summary Get all author registrations
// @Description Получает все регистрации авторов
// @Accept json
// @Produce json
// @Tags Author Registration
// @Success 200 {array} AuthorRegistrationResponseDTO
// @Router /api/author-registrations [get]
func (h *EventRegistrationHandler) GetAllAuthorRegistrations(c *gin.Context) {
	registrations, err := h.registrationService.GetAllAuthorRegistrations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []AuthorRegistrationResponseDTO
	for _, registration := range registrations {
		response = append(response, ToAuthorRegistrationRequestDTO(registration))
	}

	c.JSON(http.StatusOK, response)
}

// // @Summary Get author registration by id
// // @Description Получает регистрацию автора по id
// // @Accept json
// // @Produce json
// // @Tags Author Registration
// // @Param id path int true "Author ID"
// func (h *EventRegistrationHandler) GetAuthorRegistrationByID(c *gin.Context) {
// 	id, err := strconv.Atoi(c.Param("id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid author id"})
// 		return
// 	}

// 	registration, err := h.registrationService.GetAuthorRegistrationByID(id)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, ToAuthorRegistrationRequestDTO(registration))
// }

// @Summary Subscribe to email
// @Description Добавляет email в список подписок, если он уже есть, то возвращает ошибку, статус по умолчанию active
// @Accept json
// @Produce json
// @Tags Email Subscription
// @Param subscription body dto.EmailSubscriptionRequestDTO true "Subscription details"
// @Success 201 {object} dto.EmailSubscriptionResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/subscriptions [post]
func (h *EventRegistrationHandler) SubscribeEmail(c *gin.Context) {
	var req dto.EmailSubscriptionRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.registrationService.SubscribeEmail(req.Email)
	if err != nil {
		if err.Error() == "email already subscribed" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToEmailSubscriptionResponseDTO(subscription))
}

// @Summary Get all email subscriptions
// @Description Получает всех мальчишек с пагинацией либо без, фильтрацией по статусу и сортировкой по дате создания
// @Accept json
// @Produce json
// @Tags Email Subscription
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Param with_pagination query bool false "With pagination" default(true)
// @Param status query string false "Filter by status"
// @Param created_at query string false "Sort by created_at date (ASC or DESC)" default("DESC")
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/subscriptions [get]
func (h *EventRegistrationHandler) GetAllEmailSubscriptions(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")
	withPaginationStr := c.DefaultQuery("with_pagination", "true")
	status := c.DefaultQuery("status", "")
	createdAtSort := c.DefaultQuery("created_at", "DESC")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page format"})
		return
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid size format"})
		return
	}

	withPagination, err := strconv.ParseBool(withPaginationStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid with_pagination format"})
		return
	}

	subscriptions, totalPages, totalItems, err := h.registrationService.GetAllEmailSubscriptions(page, size, withPagination, status, createdAtSort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []dto.EmailSubscriptionResponseDTO
	for _, subscription := range subscriptions {
		response = append(response, dto.ToEmailSubscriptionResponseDTO(subscription))
	}

	if withPagination {
		c.JSON(http.StatusOK, gin.H{
			"data": response,
			"pagination": gin.H{
				"total_pages":  totalPages,
				"current_page": page,
				"page_size":    size,
				"total_items":  totalItems,
			},
		})
	} else {
		c.JSON(http.StatusOK, gin.H{"data": response})
	}
}

// @Summary Update email subscription status
// @Description Обновляет статус подписки по id или по email, но будет ругаться, если ничего из этого не положить
// @Accept json
// @Produce json
// @Tags Email Subscription
// @Param subscription body dto.UpdateEmailSubscriptionStatusDTO true "New status and id or email"
// @Success 200 {object} dto.EmailSubscriptionResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/subscriptions/status [patch]
func (h *EventRegistrationHandler) UpdateEmailSubscriptionStatus(c *gin.Context) {
	var req dto.UpdateEmailSubscriptionStatusDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ID == nil && req.Email == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "хотя бы что-нибудь дай"})
		return
	}

	subscription, err := h.registrationService.UpdateEmailSubscriptionStatus(req.ID, req.Email, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToEmailSubscriptionResponseDTO(subscription))
}

// @Summary Delete email subscription
// @Description Удаляет подписку email
// @Accept json
// @Produce json
// @Tags Email Subscription
// @Param id path int true "Subscription ID"
// @Success 204
// @Router /api/subscriptions/{id} [delete]
func (h *EventRegistrationHandler) DeleteEmailSubscription(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}

	err = h.registrationService.DeleteEmailSubscription(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Create art request
// @Description Создает запрос на мальчишку
// @Accept json
// @Produce json
// @Tags Art Request
// @Param request body dto.ArtRequestRequestDTO true "Request details"
// @Success 201 {object} dto.ArtRequestResponseDTO
// @Router /api/arts/{id}/request [post]
func (h *EventRegistrationHandler) CreateArtRequest(c *gin.Context) {
	var req dto.ArtRequestRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	artID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid art id"})
		return
	}

	artRequest, err := h.registrationService.CreateArtRequest(req.Email, req.FullName, req.Language, req.PhoneNumber, req.Request, artID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	base_url := config.GetConfig().BaseURL
	c.JSON(http.StatusCreated, dto.ToArtRequestResponseDTO(artRequest, base_url))
}

// @Summary Get all art requests
// @Description Получает все запросы на мальчишек
// @Accept json
// @Produce json
// @Tags Art Request
// @Success 200 {array} dto.ArtRequestResponseDTO
// @Router /api/arts/requests [get]
func (h *EventRegistrationHandler) GetAllArtRequests(c *gin.Context) {
	requests, err := h.registrationService.GetAllArtRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []dto.ArtRequestResponseDTO
	base_url := config.GetConfig().BaseURL
	for _, request := range requests {
		response = append(response, dto.ToArtRequestResponseDTO(request, base_url))
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Delete art request
// @Description Удаляет запрос на мальчишку
// @Accept json
// @Produce json
// @Tags Art Request
// @Param id path int true "Request ID"
// @Success 204
// @Router /api/arts/requests/{id} [delete]
func (h *EventRegistrationHandler) DeleteArtRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id"})
		return
	}

	err = h.registrationService.DeleteArtRequest(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
