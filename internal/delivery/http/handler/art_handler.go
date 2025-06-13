package handler

import (
	"anastasia_gofman_backend/internal/delivery/http/dto"
	"anastasia_gofman_backend/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ArtHandler struct {
	artService service.ArtService
}

func NewArtHandler(artService service.ArtService) *ArtHandler {
	return &ArtHandler{artService: artService}
}

// @Summary Get all arts
// @Description Получает все картины
// @Accept json
// @Produce json
// @Tags Arts
// @Success 200 {array} []dto.ArtResponseDTO
// @Failure 500 {object} map[string]string
// @Router /api/arts [get]
func (h *ArtHandler) GetAllArts(c *gin.Context) {
	arts, err := h.artService.GetAllArts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToArtResponseDTOs(arts))
}

// @Summary Get art by id
// @Description Получает картину по id
// @Accept json
// @Produce json
// @Tags Arts
// @Param id path int true "Art ID"
// @Success 200 {object} dto.ArtResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
func (h *ArtHandler) GetArtByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	art, err := h.artService.GetArtByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(art))
}

// @Summary Create a new art
// @Description Создает новую картину, все поля необязательные кроме имени, по этому пути нельзя передавать ФОТО, но если передаешь id автора - убедись, что такой есть!
// @Accept json
// @Produce json
// @Tags Arts
// @Param data body dto.CreateArtDTO true "Art"
// @Success 201 {object} dto.ArtResponseDTO
// @Failure 400 {object} map[string]string
// @Router /api/arts [post]
func (h *ArtHandler) CreateArt(c *gin.Context) {
	var artDTO dto.CreateArtDTO
	if err := c.ShouldBindJSON(&artDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	art, err := h.artService.CreateArt(artDTO.ToEntity(nil))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToArtResponseDTO(art))
}

// @Summary Create art with photos
// @Description Создает картину и фотографии к ней - передается в формате multipart/form-data в поле main_photo, preview_photo, photos(массив)
// @Accept multipart/form-data
// @Produce json
// @Tags Arts
// @Param data formData string true "JSON с любыми полями из CreateArtDTO" Extensions(x-example={ "author_id": 1, "name": {"en": "Art Name", "ru": "Название Картины", "es": "Nombre del Arte"}, "title": {"en": "Art Title", "ru": "Заголовок", "es": "Título"}, "description": {"en": "Detailed description", "ru": "Детальное описание", "es": "Descripción detallada"}, "medium": {"en": "Oil on canvas", "ru": "Холст, масло", "es": "Óleo sobre lienzo"}, "technique": {"en": "Impasto", "ru": "Импасто", "es": "Empaste"}, "dimension_x": 120, "dimension_y": 90, "year": 2024, "frame": {"en": "Wooden frame", "ru": "Деревянная рама", "es": "Marco de madera"}, "price": 150000 })
// @Param main_photo formData file false "Main Photo"
// @Param preview_photo formData file false "Preview Photo"
// @Param photos formData []file false "Photos"
// @Success 201 {object} dto.ArtResponseDTO
// @Failure 400 {object} map[string]string
// @Router /api/arts/with_photos [post]
func (h *ArtHandler) CreateArtWithPhotos(c *gin.Context) {
	var artDTO dto.CreateArtWithPhotosDTO
	// JSONиз data
	jsonData := c.PostForm("data")
	if jsonData == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'data' field in form-data"})
		return
	}

	// JSON в artDTO
	if err := json.Unmarshal([]byte(jsonData), &artDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON in 'data' field: " + err.Error()})
		return
	}

	if err := c.Request.ParseMultipartForm(1 << 30); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form: " + err.Error()})
		return
	}

	createdArt, err := h.artService.CreateArt(artDTO.ToEntity(nil))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create art entity: " + err.Error()})
		return
	}
	currentArtID := uint(createdArt.ID)

	mainPhotoFileHeader, err1 := c.FormFile("main_photo")
	previewPhotoFileHeader, err2 := c.FormFile("preview_photo")
	if err1 != nil && err1 != http.ErrMissingFile && err2 != nil && err2 != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing main_photo: " + err1.Error()})
		return
	}

	if mainPhotoFileHeader != nil {
		if _, err := h.artService.AddMainOrPreviewPhotoToArt(currentArtID, mainPhotoFileHeader, true, false); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add main photo: " + err1.Error()})
			return
		}
	}

	if previewPhotoFileHeader != nil {
		if _, err := h.artService.AddMainOrPreviewPhotoToArt(currentArtID, previewPhotoFileHeader, false, true); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add preview photo: " + err2.Error()})
			return
		}
	}

	photoFileHeaders := c.Request.MultipartForm.File["photos"]

	if len(photoFileHeaders) > 0 {
		if _, err := h.artService.AddPhotosToArt(currentArtID, photoFileHeaders); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add photos: " + err.Error()})
			return
		}
	}

	finalEvent, err := h.artService.GetArtByID(currentArtID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve art after updates: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToArtResponseDTO(finalEvent))
}

// @Summary Full update art
// @Description Обновляет всю картину - все поля необязательные, но перезаписывает все поля, те которые не передал - обнуляются (кроме фото)
// @Accept json
// @Produce json
// @Tags Arts
// @Param id path int true "Art ID"
// @Param data body dto.UpdateArtDTO true "Art"
// @Success 200 {object} dto.ArtResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/arts/{id} [put]
func (h *ArtHandler) FullUpdateArt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	var artDTO dto.UpdateArtDTO
	if err := c.ShouldBindJSON(&artDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	idUint := uint(id)
	art, err := h.artService.FullUpdateArt(artDTO.ToEntity(&idUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(art))
}

// @Param kwargs body map[string]interface{} true "kwargs"
// @Param kwargs body map[string]interface{} true "kwargs"

// @Summary Partial update art
// @Description Обновляет часть картины - все поля необязательные, поля, которые не передал не меняются, фото сюда не суй!
// @Accept json
// @Produce json
// @Tags Arts
// @Param id path int true "Art ID"
// @Param kwargs body dto.UpdateArtDTO true "Art"
// @Success 200 {object} dto.ArtResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/arts/{id} [patch]
func (h *ArtHandler) PartialUpdateArt(c *gin.Context) {
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
	art, err := h.artService.PartialUpdateArt(uint(id), kwargs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(art))
}

// @Summary Delete art
// @Description Удаляет арт объект и его фотки
// @Accept json
// @Produce json
// @Tags Arts
// @Param id path int true "Art ID"
// @Success 200 {object} map[string]string
// @Router /api/arts/{id} [delete]
func (h *ArtHandler) DeleteArt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	err = h.artService.DeleteArt(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "art deleted"})
}

// @Summary Add main photo to art
// @Description Добавляет/обновляет главную фотографию к картине - передается фото в формате multipart/form-data в поле main_photo
// @Accept multipart/form-data
// @Produce json
// @Tags Arts
// @Param id path int true "Art ID"
// @Param is_preview query bool false "Is Preview"
// @Param main_photo formData file true "Main Photo"
// @Success 200 {object} dto.ArtResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/arts/{id}/main_photo [post]
func (h *ArtHandler) AddMainPhotoToArt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	is_preview_str := c.DefaultQuery("is_preview", "false")
	is_preview, err := strconv.ParseBool(is_preview_str)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	main_photo, err := c.FormFile("main_photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid photo format"})
		return
	}
	art, err := h.artService.AddMainOrPreviewPhotoToArt(uint(id), main_photo, !is_preview, is_preview)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(art))
}

// @Summary Add photos to art
// @Description Добавляет фотографии к картине(имеющиеся не трогаются) - передается []base64 в формате multipart/form-data в поле photos
// @Accept multipart/form-data
// @Produce json
// @Tags Arts
// @Param id path int true "Art ID"
// @Param photos formData []file true "Photos"
// @Success 200 {object} dto.ArtResponseDTO "Art"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/arts/{id}/photos [post]
func (h *ArtHandler) AddPhotosToArt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	_, err = h.artService.GetArtByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "art not found"})
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form"})
		return
	}
	photos := form.File["photos"]

	photos_result, err := h.artService.AddPhotosToArt(uint(id), photos)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cant add photos"})
		return
	}
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(photos_result))
}

// @Summary Patch art photos
// @Description Обновляет фотографии к картине(имеющиеся удаляются) - передается []base64 в формате multipart/form-data в поле photos
// @Accept multipart/form-data
// @Produce json
// @Tags Arts
// @Param id path int true "Art ID"
// @Param photos formData []file true "Photos"
// @Success 200 {object} dto.ArtResponseDTO "Art"
func (h *ArtHandler) PatchArtPhotos(c *gin.Context) {
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
	photos_result, err := h.artService.PatchArtPhotos(uint(id), photos)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cant patch photos"})
		return
	}
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(photos_result))
}

// @Summary Add author to art
// @Description Добавляет автора к картине
// @Accept json
// @Produce json
// @Tags Arts
// @Param id path int true "Art ID"
// @Param author_id path int true "Author ID"
// @Success 200 {object} dto.ArtResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/arts/{id}/author/{author_id} [post]
func (h *ArtHandler) AddAuthorToArt(c *gin.Context) {
	// fmt.Println("AddAuthorToArt")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	author_id, err := strconv.ParseUint(c.Param("author_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid author id format"})
		return
	}
	art, err := h.artService.AddAuthorToArt(uint(id), uint(author_id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(art))
}

// @Summary Get main photo
// @Description Получает главную фотографию картины
// @Accept json
// @Produce json
// @Tags Arts
// @Param id path int true "Art ID"
// @Success 200 {object} entity.Photo
// @Router /api/arts/{id}/main_photo [get]
func (h *ArtHandler) GetMainPhoto(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	photo, err := h.artService.GetMainPhoto(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, photo)
}
