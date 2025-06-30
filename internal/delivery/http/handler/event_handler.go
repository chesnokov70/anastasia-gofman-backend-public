package handler

import (
	"anastasia_gofman_backend/internal/delivery/http/dto"
	"anastasia_gofman_backend/internal/service"
	"anastasia_gofman_backend/pkg/config"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventService       service.EventService
	translationService service.TranslationService
}

func NewEventHandler(eventService service.EventService, translationService service.TranslationService) *EventHandler {
	return &EventHandler{
		eventService:       eventService,
		translationService: translationService,
	}
}

// @Summary Get all events
// @Description Получает все события
// @Accept json
// @Produce json
// @Tags Events
// @Param offset query int false "Offset v pagination" default(0)
// @Param limit query int false "Limit v pagination" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/events [get]
func (h *EventHandler) GetAllEvents(c *gin.Context) {
	// Получаем offset и limit напрямую из query parameters
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset parameter"})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}

	// Валидация параметров
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}

	events, total_pages, total_items, err := h.eventService.GetAllEvents(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Рассчитываем текущую "страницу" для совместимости с пагинацией
	current_page := 1
	if limit > 0 {
		current_page = (offset / limit) + 1
	}
	baseUrl := config.GetBaseURL()
	c.JSON(http.StatusOK, gin.H{
		"data": dto.ToEventResponseDTOs(events, baseUrl),
		"pagination": gin.H{
			"total_items":  int(total_items),
			"offset":       offset,
			"limit":        limit,
			"current_page": current_page,
			"total_pages":  int(total_pages),
		},
	})
}

// @Summary Get event by ID
// @Description Получает событие по ID
// @Accept json
// @Produce json
// @Tags Events
// @Param id path int true "Event ID"
// @Success 200 {object} dto.EventResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events/{id} [get]
func (h *EventHandler) GetEventByID(c *gin.Context) {
	id := c.Param("id")
	id_uint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	event, err := h.eventService.GetEventByID(uint(id_uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

// @Summary Create event
// @Description Создает новое событие без фотографий - принимает JSON. Поддерживаемые форматы дат: "2024-01-01T00:00:00Z", "01.01.2024", "01.01.2024 15:30", 2006-01-02 15:04:05
// @Accept json
// @Produce json
// @Tags Events
// @Param event body dto.CreateEventDTO true "Event"
// @Param with_translation query bool false "Автоматически переводить отсутствующие языки для всех переданных в формате мультиязычных полей" default(false)
// @Success 201 {object} dto.EventResponseDTO
// @Failure 400 {object} map[string]string
// @Router /api/events [post]
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var eventDTO dto.CreateEventDTO
	if err := c.ShouldBindJSON(&eventDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	withTranslation := c.DefaultQuery("with_translation", "false")
	if withTranslation == "true" {
		if err := h.translationService.AutoCompleteEventTranslations(&eventDTO, 3); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to auto-translate event fields: " + err.Error(),
			})
			return
		}
	}

	createdEvent, err := h.eventService.CreateEvent(eventDTO.ToEntity(nil))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	baseUrl := config.GetBaseURL()
	c.JSON(http.StatusCreated, dto.ToEventResponseDTO(createdEvent, baseUrl))
}

// @Summary Create event with photos
// @Description Создает событие и фотографии к нему - передается в формате multipart/form-data. Поддерживаемые форматы дат: "2024-01-01T00:00:00Z", "01.01.2024", "01.01.2024 15:30"
// @Accept multipart/form-data
// @Produce json
// @Tags Events
// @Param data formData string true "JSON payload for Event details with flexible date formats." Extensions(x-example={ "title": {"en": "Event Title", "ru": "Заголовок События", "es": "Título del Evento"}, "description": {"en": "Event description", "ru": "Описание события", "es": "Descripción del evento"}, "location": {"en": "Event location", "ru": "Место проведения", "es": "Ubicación del evento"}, "start_date": "01.01.2024", "end_date": "02.01.2024" })
// @Param with_translation query bool false "Автоматически переводить отсутствующие языки" default(false)
// @Param main_photo formData file false "Main Photo"
// @Param preview_photo formData file false "Preview Photo"
// @Param photos formData []file false "Photos"
// @Success 201 {object} dto.EventResponseDTO
// @Failure 400 {object} map[string]string
// @Router /api/events/with_photos [post]
func (h *EventHandler) CreateEventWithPhotos(c *gin.Context) {
	var eventDTO dto.CreateEventWithPhotosDTO
	// JSON из data
	jsonData := c.PostForm("data")
	if jsonData == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'data' field in form-data"})
		return
	}

	// JSON в eventDTO
	if err := json.Unmarshal([]byte(jsonData), &eventDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON in 'data' field: " + err.Error()})
		return
	}

	withTranslation := c.DefaultQuery("with_translation", "false")
	if withTranslation == "true" {
		if err := h.translationService.AutoCompleteEventWithPhotosTranslations(&eventDTO, 3); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to auto-translate event fields: " + err.Error(),
			})
			return
		}
	}

	if err := c.Request.ParseMultipartForm(1 << 30); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form: " + err.Error()})
		return
	}

	createdEvent, err := h.eventService.CreateEvent(eventDTO.ToEntity(nil))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event entity: " + err.Error()})
		return
	}
	currentEventID := uint(createdEvent.ID)

	mainPhotoFileHeader, err1 := c.FormFile("main_photo")
	previewPhotoFileHeader, err2 := c.FormFile("preview_photo")

	if err1 != nil && err1 != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing main_photo: " + err1.Error()})
		return
	}
	if err2 != nil && err2 != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing preview_photo: " + err2.Error()})
		return
	}

	if mainPhotoFileHeader != nil {
		if _, err := h.eventService.AddMainOrPreviewPhotoToEvent(currentEventID, mainPhotoFileHeader, true, false); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add main photo: " + err.Error()})
			return
		}
	}

	if previewPhotoFileHeader != nil {
		if _, err := h.eventService.AddMainOrPreviewPhotoToEvent(currentEventID, previewPhotoFileHeader, false, true); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add preview photo: " + err.Error()})
			return
		}
	}

	photoFileHeaders := c.Request.MultipartForm.File["photos"]

	if len(photoFileHeaders) > 0 {
		if _, err := h.eventService.AddPhotosToEvent(currentEventID, photoFileHeaders); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add photos: " + err.Error()})
			return
		}
	}

	finalEvent, err := h.eventService.GetEventByID(currentEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event after updates: " + err.Error()})
		return
	}

	baseUrl := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToEventResponseDTO(finalEvent, baseUrl))
}

// @Summary Full update event
// @Description Обновляет все поля события - все поля необязательные, но перезаписывает все поля, те которые не передал - обнуляются (кроме фото)
// @Accept json
// @Produce json
// @Tags Events
// @Param id path int true "Event ID"
// @Param event body dto.UpdateEventDTO true "Event"
// @Success 200 {object} dto.EventResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events/{id} [put]
func (h *EventHandler) FullUpdateEvent(c *gin.Context) {
	id := c.Param("id")
	id_uint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	var eventDTO dto.UpdateEventDTO
	if err := c.ShouldBindJSON(&eventDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idInt := int(id_uint)
	updatedEvent, err := h.eventService.FullUpdateEvent(eventDTO.ToEntity(&idInt))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	baseUrl := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToEventResponseDTO(updatedEvent, baseUrl))
}

// @Summary Partial update event
// @Description Обновляет поля события, которые ты передал, остальные не трогает, фото сюда не суй!
// @Accept json
// @Produce json
// @Tags Events
// @Param id path int true "Event ID"
// @Param event body dto.UpdateEventDTO true "Event"
// @Success 200 {object} dto.EventResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events/{id} [patch]
func (h *EventHandler) PartialUpdateEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	var kwargs map[string]interface{}
	if err := c.ShouldBindJSON(&kwargs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedEvent, err := h.eventService.PartialUpdateEvent(uint(id), kwargs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	baseUrl := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToEventResponseDTO(updatedEvent, baseUrl))
}

// @Summary Delete event
// @Description Удаляет событие и фотки
// @Accept json
// @Produce json
// @Tags Events
// @Param id path int true "Event ID"
// @Success 200 {object} map[string]string
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	id := c.Param("id")
	id_uint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	err = h.eventService.DeleteEvent(uint(id_uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

// @Summary Get main photo
// @Description Получает главную фотографию события
// @Accept json
// @Produce json
// @Tags Events
// @Param id path int true "Event ID"
// @Success 200 {object} entity.Photo
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events/{id}/main_photo [get]
func (h *EventHandler) GetMainPhoto(c *gin.Context) {
	id := c.Param("id")
	id_uint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	photo, err := h.eventService.GetMainPhoto(uint(id_uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, photo)
}

// @Summary Add photos to event
// @Description Добавляет фотографии к событию(имеющиеся не трогаются) - передается []base64 в формате multipart/form-data в поле photos
// @Accept multipart/form-data
// @Produce json
// @Tags Events
// @Param id path int true "Event ID"
// @Param photos formData []file true "Photos"
// @Success 200 {object} dto.EventResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events/{id}/photos [post]
func (h *EventHandler) AddPhotosToEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	_, err = h.eventService.GetEventByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "event not found"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form"})
		return
	}
	photos := form.File["photos"]

	updatedEvent, err := h.eventService.AddPhotosToEvent(uint(id), photos)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cant add photos"})
		return
	}
	baseUrl := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToEventResponseDTO(updatedEvent, baseUrl))
}

// @Summary Patch event photos
// @Description Обновляет фотографии события(имеющиеся удаляются) - передается []base64 в формате multipart/form-data в поле photos
// @Accept multipart/form-data
// @Produce json
// @Tags Events
// @Param id path int true "Event ID"
// @Param photos formData []file true "Photos"
// @Success 200 {object} dto.EventResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events/{id}/photos [patch]
func (h *EventHandler) PatchEventPhotos(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form"})
		return
	}
	photos := form.File["photos"]
	updatedEvent, err := h.eventService.AddPhotosToEventReplaceOld(uint(id), photos)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cant patch photos"})
		return
	}
	baseUrl := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToEventResponseDTO(updatedEvent, baseUrl))
}

// @Summary Add main photo to event
// @Description Добавляет/обновляет главную фотографию к событию - передается фото в формате multipart/form-data в поле main_photo
// @Accept multipart/form-data
// @Produce json
// @Tags Events
// @Param id path int true "Event ID"
// @Param main_photo formData file true "Main Photo"
// @Success 200 {object} dto.EventResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events/{id}/main_photo [post]
func (h *EventHandler) AddMainPhotoToEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	main_photo, err := c.FormFile("main_photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid photo format"})
		return
	}
	updatedEvent, err := h.eventService.AddMainOrPreviewPhotoToEvent(uint(id), main_photo, true, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	baseUrl := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToEventResponseDTO(updatedEvent, baseUrl))
}

// @Summary Delete main photo from event
// @Description Удаляет главную/preview фотографию к событию
// @Accept json
// @Produce json
// @Tags Events
// @Param id path int true "Event ID"
// @Param is_preview query bool false "Is preview" default(false)
// @Success 200 {object} map[string]string
// @Router /api/events/{id}/main_photo [delete]
func (h *EventHandler) DeleteMainPhotoFromEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	is_preview_str := c.DefaultQuery("is_preview", "false")
	is_preview, err := strconv.ParseBool(is_preview_str)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	err = h.eventService.DeleteMainOrPreviewPhoto(uint(id), is_preview)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Main photo deleted successfully"})
}
