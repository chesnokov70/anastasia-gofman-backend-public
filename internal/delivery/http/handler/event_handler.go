package handler

import (
	"anastasia_gofman_backend/internal/delivery/http/dto"
	"anastasia_gofman_backend/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventService service.EventService
}

func NewEventHandler(eventService service.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

// @Summary Get all events
// @Description Получает все события
// @Accept json
// @Produce json
// @Tags Events
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/events [get]
func (h *EventHandler) GetAllEvents(c *gin.Context) {
	page := c.GetInt("page")
	size := c.GetInt("size")
	events, total_pages, err := h.eventService.GetAllEvents(page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       dto.ToEventResponseDTOs(events),
		"pagination": gin.H{"total_pages": total_pages, "current_page": page, "page_size": size},
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
// @Success 201 {object} dto.EventResponseDTO
// @Failure 400 {object} map[string]string
// @Router /api/events [post]
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var eventDTO dto.CreateEventDTO
	if err := c.ShouldBindJSON(&eventDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdEvent, err := h.eventService.CreateEvent(eventDTO.ToEntity(nil))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToEventResponseDTO(createdEvent))
}

// @Summary Create event with photos
// @Description Создает событие и фотографии к нему - передается в формате multipart/form-data. Поддерживаемые форматы дат: "2024-01-01T00:00:00Z", "01.01.2024", "01.01.2024 15:30"
// @Accept multipart/form-data
// @Produce json
// @Tags Events
// @Param data formData string true "JSON payload for Event details with flexible date formats." Extensions(x-example={ "title": {"en": "Event Title", "ru": "Заголовок События", "es": "Título del Evento"}, "description": {"en": "Event description", "ru": "Описание события", "es": "Descripción del evento"}, "location": {"en": "Event location", "ru": "Место проведения", "es": "Ubicación del evento"}, "start_date": "01.01.2024", "end_date": "02.01.2024" })
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

	c.JSON(http.StatusOK, dto.ToEventResponseDTO(finalEvent))
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

	c.JSON(http.StatusOK, dto.ToEventResponseDTO(updatedEvent))
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

	c.JSON(http.StatusOK, dto.ToEventResponseDTO(updatedEvent))
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
	c.JSON(http.StatusOK, dto.ToEventResponseDTO(updatedEvent))
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
	c.JSON(http.StatusOK, dto.ToEventResponseDTO(updatedEvent))
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

	c.JSON(http.StatusOK, dto.ToEventResponseDTO(updatedEvent))
}
