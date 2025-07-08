package handler

import (
	"anastasia_gofman_backend/internal/delivery/http/dto"
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/service"
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
