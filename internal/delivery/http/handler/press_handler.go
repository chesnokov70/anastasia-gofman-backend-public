package handler

import (
	"anastasia_gofman_backend/internal/delivery/http/dto"
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/service"
	"anastasia_gofman_backend/pkg/config"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PressHandler struct {
	pressService service.PressAndArticleService
}

func NewPressHandler(pressService service.PressAndArticleService) *PressHandler {
	return &PressHandler{pressService: pressService}
}

// @Description Получает все пресс объекты, умеет в пагинацию, ордерд бай криэтед эт
// @Accept json
// @Produce json
// @Tags Press
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Param sorting query string false "Sorting type" Enums(NEW, CLOSEST, FARTHEST) default(NEW)
// @Param with_pagination query bool false "With pagination" default(true)
// @Router /api/press [get]
func (h *PressHandler) GetAllPress(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "10")
	sorting := c.Query("sorting")
	if sorting == "" {
		sorting = "NEW"
	}

	page_int, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page format"})
		return
	}
	size_int, err := strconv.Atoi(size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid size format"})
		return
	}

	with_pagination := c.DefaultQuery("with_pagination", "true")
	with_pagination_bool, err := strconv.ParseBool(with_pagination)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid with_pagination format"})
		return
	}

	press, _, pages, total, err := h.pressService.GetAllPressAndArticles(page_int, size_int, with_pagination_bool, "press", sorting)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	if with_pagination_bool {
		c.JSON(http.StatusOK, gin.H{
			"data": dto.ToPressResponseDTOs(press, base_url),
			"pagination": gin.H{
				"total_pages":  int(pages),
				"current_page": int(page_int),
				"page_size":    int(size_int),
				"total_items":  int(total),
			},
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"data": dto.ToPressResponseDTOs(press, base_url),
		})
	}
}

// @Summary Get press by ID
// @Description Получает пресс объект по ID
// @Accept json
// @Produce json
// @Tags Press
// @Param id path int true "Press ID"
// @Router /api/press/{id} [get]
func (h *PressHandler) GetPressByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	press, _, err := h.pressService.GetPressOrArticleByID(uint(id), "press")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToPressResponseDTO(*press, base_url))
}

// @Summary Create press
// @Description Создает пресс - фотки не суй плиз
// @Accept json
// @Produce json
// @Tags Press
// @Param press body dto.CreatePressDTO true "Press"
// @Success 200 {object} dto.PressResponseDTO
// @Router /api/press [post]
func (h *PressHandler) CreatePress(c *gin.Context) {
	var press dto.CreatePressDTO
	if err := c.ShouldBindJSON(&press); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	press_entity := press.ToEntity(nil)
	press_response, _, err := h.pressService.CreatePressOrArticle("press", press_entity, entity.Article{})
	log.Printf("press_response: %v", press_response)
	fmt.Printf("press_response: %v", press_response)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToPressResponseDTO(*press_response, base_url))
}

// @Summary Create press with photos
// @Description Создает пресс и фотографии к нему - передается в формате multipart/form-data в поле main_photo, preview_photo, photos(массив)
// @Accept multipart/form-data
// @Produce json
// @Tags Press
// @Param data formData string true "JSON с любыми полями из CreatePressDTO" Extensions(x-example={"description": {"en": "Tung, tung, tung, tung, tung, tung, tung, tung, tung, Sahur.","es": "Golybini Shpionini","ru": "Тун, тун, тун, Сахур."},"full_text": {"en": "Tung, tung, tung, tung, tung, tung, tung, tung, tung, Sahur.","es": "Golybini Shpionini","ru": "Тун, тун, тун, Сахур."},"link": "https://example.com","position": 1,"title": {"en": "Tung, tung, tung, tung, tung, tung, tung, tung, tung, Sahur.","es": "Golybini Shpionini","ru": "Тун, тун, тун, Сахур."}})
// @Param main_photo formData file false "Main Photo"
// @Param preview_photo formData file false "Preview Photo"
// @Param photos formData []file false "Photos"
// @Success 201 {object} dto.PressResponseDTO
// @Failure 400 {object} map[string]string
// @Router /api/press/with_photos [post]
func (h *PressHandler) CreatePressWithPhotos(c *gin.Context) {
	var pressDTO dto.CreatePressDTO

	jsonData := c.PostForm("data")
	if jsonData == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'data' field in form-data"})
		return
	}

	if err := json.Unmarshal([]byte(jsonData), &pressDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON in 'data' field: " + err.Error()})
		return
	}

	if err := c.Request.ParseMultipartForm(1 << 30); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form: " + err.Error()})
		return
	}

	createdPress, _, err := h.pressService.CreatePressOrArticle("press", pressDTO.ToEntity(nil), entity.Article{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create press entity: " + err.Error()})
		return
	}
	currentPressID := uint(createdPress.ID)

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
		if _, _, err := h.pressService.AddMainOrPreviewPhotoToPressOrArticle("press", currentPressID, mainPhotoFileHeader, true, false); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add main photo: " + err.Error()})
			return
		}
	}

	if previewPhotoFileHeader != nil {
		if _, _, err := h.pressService.AddMainOrPreviewPhotoToPressOrArticle("press", currentPressID, previewPhotoFileHeader, false, true); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add preview photo: " + err.Error()})
			return
		}
	}

	photoFileHeaders := c.Request.MultipartForm.File["photos"]

	if len(photoFileHeaders) > 0 {
		if _, _, err := h.pressService.AddPhotosToPressOrArticle(currentPressID, "press", photoFileHeaders); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add photos: " + err.Error()})
			return
		}
	}

	finalPress, _, err := h.pressService.GetPressOrArticleByID(currentPressID, "press")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve press after updates: " + err.Error()})
		return
	}

	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToPressResponseDTO(*finalPress, base_url))
}

// @Summary Full update press
// @Description Обновляет весь пресс - все поля необязательные, но перезаписывает все поля, те которые не передал - обнуляются (кроме фото)
// @Accept json
// @Produce json
// @Tags Press
// @Param id path int true "Press ID"
// @Param data body dto.UpdatePressDTO true "Press"
// @Success 200 {object} dto.PressResponseDTO
// @Router /api/press/{id} [put]
func (h *PressHandler) FullUpdatePress(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	var pressDTO dto.UpdatePressDTO
	if err := c.ShouldBindJSON(&pressDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	idUint := uint(id)
	press_response, _, err := h.pressService.FullUpdatePressOrArticle("press", pressDTO.ToEntity(&idUint), entity.Article{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToPressResponseDTO(*press_response, base_url))
}

// @Summary Partial update press
// @Description Обновляет часть пресса - все поля необязательные, поля, которые не передал не меняются, фото сюда не суй
// @Accept json
// @Produce json
// @Tags Press
// @Param id path int true "Press ID"
// @Param kwargs body dto.UpdatePressDTO true "Press"
// @Success 200 {object} dto.PressResponseDTO
// @Router /api/press/{id} [patch]
func (h *PressHandler) PartialUpdatePress(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	var pressDTO map[string]interface{}
	if err := c.ShouldBindJSON(&pressDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	idUint := uint(id)
	press_response, _, err := h.pressService.PartialUpdatePressOrArticle("press", idUint, pressDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToPressResponseDTO(*press_response, base_url))
}

// @Summary Delete press
// @Description Удаляет пресс и фотки
// @Accept json
// @Produce json
// @Tags Press
// @Param id path int true "Press ID"
// @Success 200 {object} map[string]string
// @Router /api/press/{id} [delete]
func (h *PressHandler) DeletePress(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	err = h.pressService.DeletePressOrArticle("press", uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "press deleted"})
}

// @Summary Add main photo to press
// @Description Добавляет/обновляет главную фотографию к прессу - передается фото в формате multipart/form-data в поле main_photo
// @Accept multipart/form-data
// @Produce json
// @Tags Press
// @Param id path int true "Press ID"
// @Param is_preview query bool false "Is Preview"
// @Param main_photo formData file true "Main Photo"
// @Success 200 {object} dto.PressResponseDTO
// @Router /api/press/{id}/main_photo [post]
func (h *PressHandler) AddMainPhotoToPress(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	is_preview_str := c.DefaultQuery("is_preview", "false")
	is_preview, err := strconv.ParseBool(is_preview_str)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid is_preview format"})
		return
	}
	main_photo, err := c.FormFile("main_photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid photo format"})
		return
	}
	press_response, _, err := h.pressService.AddMainOrPreviewPhotoToPressOrArticle("press", uint(id), main_photo, !is_preview, is_preview)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToPressResponseDTO(*press_response, base_url))
}

// @Summary Delete main photo from press
// @Description Удаляет главную/preview фотографию к прессу
// @Accept json
// @Produce json
// @Tags Press
// @Param id path int true "Press ID"
// @Param is_preview query bool false "Is Preview"
// @Success 200 {object} map[string]string
// @Router /api/press/{id}/main_photo [delete]
func (h *PressHandler) DeleteMainPhotoFromPress(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	is_preview_str := c.DefaultQuery("is_preview", "false")
	is_preview, err := strconv.ParseBool(is_preview_str)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid is_preview format"})
		return
	}
	err = h.pressService.DeleteMainOrPreviewPhoto(uint(id), "press", is_preview)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "photo deleted"})
}

// @Summary Add photos to press
// @Description Добавляет фотографии к прессу(имеющиеся не трогаются) - передается []base64 в формате multipart/form-data в поле photos
// @Accept multipart/form-data
// @Produce json
// @Tags Press
// @Param id path int true "Press ID"
// @Param photos formData []file true "Photos"
// @Success 200 {object} dto.PressResponseDTO "Press"
// @Router /api/press/{id}/photos [post]
func (h *PressHandler) AddPhotosToPress(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	_, _, err = h.pressService.GetPressOrArticleByID(uint(id), "press")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "press not found"})
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form"})
		return
	}
	photos := form.File["photos"]

	press, _, err := h.pressService.AddPhotosToPressOrArticle(uint(id), "press", photos)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cant add photos"})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToPressResponseDTO(*press, base_url))
}

// @Summary Get main photo
// @Description Получает главную фотографию пресса
// @Accept json
// @Produce json
// @Tags Press
// @Param id path int true "Press ID"
// @Success 200 {object} entity.Photo
// @Router /api/press/{id}/main_photo [get]
func (h *PressHandler) GetMainPhoto(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	photo, err := h.pressService.GetMainPhoto(uint(id), "press")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, photo)
}

// @Summary Patch press photos
// @Description Обновляет фотографии к прессу(имеющиеся удаляются) - передается []base64 в формате multipart/form-data в поле photos
// @Accept multipart/form-data
// @Produce json
// @Tags Press
// @Param id path int true "Press ID"
// @Param photos formData []file true "Photos"
// @Success 200 {object} dto.PressResponseDTO "Press"
// @Router /api/press/{id}/photos [patch]
// func (h *PressHandler) PatchPressPhotos(c *gin.Context) {
// 	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
// 		return
// 	}
// 	form, err := c.MultipartForm()
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form"})
// 		return
// 	}
// 	photos := form.File["photos"]
// 	press, _, err := h.pressService.PatchPressOrArticlePhotos(uint(id), "press", photos)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Cant patch photos"})
// 		return
// 	}
// 	base_url := config.GetBaseURL()
// 	c.JSON(http.StatusOK, dto.ToPressResponseDTO(*press, base_url))
// }

// @Summary Patch event photos
// @Description Обновляет фотографии события - передается JSON массив строк: URLs для существующих мальчишек, а base64 для новых мальчишек
// @Accept json
// @Produce json
// @Tags Press
// @Param id path int true "Press ID"
// @Param photos body []string true "Array of photo data: URLs for existing photos, base64 data URLs for new ones"
// @Success 200 {object} dto.PressResponseDTO "Press"
// @Failure 400 {object} map[string]string
// @Router /api/press/{id}/photos [patch]
func (h *PressHandler) PatchPressPhotos(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	var photoStrings []string
	if err := c.ShouldBindJSON(&photoStrings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse JSON: " + err.Error()})
		return
	}

	// if len(photoStrings) == 0 {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "no photos provided"})
	// 	return
	// }

	updatedEvent, _, err := h.pressService.PatchPhotosFromStrings("press", uint(id), photoStrings)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can't patch photos: " + err.Error()})
		return
	}

	baseUrl := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToPressResponseDTO(*updatedEvent, baseUrl))
}
