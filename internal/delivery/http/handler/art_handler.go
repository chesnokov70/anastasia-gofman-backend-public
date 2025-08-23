package handler

import (
	"anastasia_gofman_backend/internal/delivery/http/dto"
	"anastasia_gofman_backend/internal/service"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"anastasia_gofman_backend/pkg/config"

	"github.com/gin-gonic/gin"
)

type ArtHandler struct {
	artService service.ArtService
}

func NewArtHandler(artService service.ArtService) *ArtHandler {
	return &ArtHandler{artService: artService}
}

// @Summary Get all arts
// @Description Получает все картины. В Sorting можно передать NEW, RATED(ХЗ что это - использую position), PRICE_HIGH, PRICE_LOW, если не передать - будет сортировка по position. В Filtering можно передать search - поиск по названию, Name Title Description Medium Technique Year Frame Style Direction DimensionStr. type = _common - все картины без type или с type = ""
// @Accept json
// @Produce json
// @Tags Arts
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Param with_pagination query bool false "With pagination" default(true)
// @Param without_collection query bool false "Without collection" default(true)
// @Param with_type_discrimination query bool false "With type discrimination" default(false)
// @Param sorting query string false "Sorting type" Enums(NEW, RATED, PRICE_HIGH, PRICE_LOW) default()
// @Param filtering query string false "JSON filter object. Example: {"price_from": 100, "price_to": 1000, "size": "MEDIUM", "direction": "EXCLUSIVE", "style": "abstract", "author": "author name", "type": "_common", "archive_type": "repeat", "search": {"name": "search text", "title": {"en": "title search"}, "description": {"en": "desc search"}, "medium": {"en": "medium search"}, "technique": {"en": "technique search"}, "year": "2024", "frame": {"en": "frame search"}, "style": "abstract", "direction": "EXCLUSIVE", "dimension_str": "100x100"}}"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/arts [get]
func (h *ArtHandler) GetAllArts(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "10")
	sorting := c.DefaultQuery("sorting", "")
	filtering := c.DefaultQuery("filtering", "")
	without_collection := c.DefaultQuery("without_collection", "true")
	without_collection_bool, err := strconv.ParseBool(without_collection)
	with_type_discrimination := c.DefaultQuery("with_type_discrimination", "false")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid without_collection format"})
		return
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
	with_type_discrimination_bool, err := strconv.ParseBool(with_type_discrimination)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid with_pagination format"})
		return
	}

	validSortings := map[string]bool{
		"NEW":        true,
		"RATED":      true,
		"PRICE_HIGH": true,
		"PRICE_LOW":  true,
		"":           true,
	}
	if !validSortings[sorting] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sorting parameter. Valid values: NEW, RATED, PRICE_HIGH, PRICE_LOW"})
		return
	}

	var filteringDTO *dto.ArtFilteringDTO
	if filtering != "" {
		filteringDTO = &dto.ArtFilteringDTO{}
		decoder := json.NewDecoder(strings.NewReader(filtering))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(filteringDTO); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filtering JSON format: " + err.Error() + "Example: {\"price_from\": 100, \"price_to\": 1000, \"size\": \"MEDIUM\", \"direction\": \"EXCLUSIVE\", \"style\": \"abstract\", \"author\": \"author name\", \"type\": \"_common\", \"archive_type\": \"repeat\", \"search\": {\"name\": {\"en\": \"search text\", \"ru\": \"текст поиска\", \"es\": \"texto de búsqueda\"}, \"title\": {\"en\": \"title search\"}, \"description\": {\"en\": \"desc search\"}, \"medium\": {\"en\": \"medium search\"}, \"technique\": {\"en\": \"technique search\"}, \"year\": \"2024\", \"frame\": {\"en\": \"frame search\"}, \"style\": \"abstract\", \"direction\": \"EXCLUSIVE\", \"dimension_str\": \"100x100\"}}"})
			return
		}

		if filteringDTO.Size != nil {
			validSizes := map[string]bool{"SMALL": true, "MEDIUM": true, "BIG": true}
			if !validSizes[*filteringDTO.Size] {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid size value. Valid values: SMALL, MEDIUM, BIG"})
				return
			}
		}
	}

	arts, total_pages, total_items, err := h.artService.GetAllArts(page_int, size_int, with_pagination_bool, sorting, filteringDTO.ToEntity(), without_collection_bool, with_type_discrimination_bool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	min_price, max_price, err := h.artService.GetMinAndMaxPrice()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	if with_pagination_bool {
		c.JSON(http.StatusOK, gin.H{
			"data":       dto.ToArtResponseDTOs(arts, base_url),
			"pagination": gin.H{"total_pages": int(total_pages), "current_page": int(page_int), "page_size": int(size_int), "total_items": int(total_items)},
			"min_price":  min_price,
			"max_price":  max_price,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"data":      dto.ToArtResponseDTOs(arts, base_url),
			"min_price": min_price,
			"max_price": max_price,
		})
	}
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
// @Router /api/arts/{id} [get]
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
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(art, base_url))
}

// // @Param with_stripe query bool false "With Stripe"

// @Summary Create a new art
// @Description Создает новую картину, все поля необязательные кроме имени, по этому пути нельзя передавать ФОТО, но если передаешь id автора - убедись, что такой есть! Если не передашь имя и дескрипшн для страйпа - возьму их из EN полей, но нельзя получается, чтобы поле name.en было пустым одновременно с nameForStripe
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
	// with_stripe := c.DefaultQuery("with_stripe", "false")
	with_stripe := "true"
	with_stripe_bool, err := strconv.ParseBool(with_stripe)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid with_stripe format"})
		return
	}

	art, err := h.artService.CreateArt(artDTO.ToEntity(nil), with_stripe_bool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusCreated, dto.ToArtResponseDTO(art, base_url))
}

// // @Param with_stripe query bool false "With Stripe"

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

	// with_stripe := c.DefaultQuery("with_stripe", "false")
	with_stripe := "true"
	with_stripe_bool, err := strconv.ParseBool(with_stripe)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid with_stripe format"})
		return
	}
	createdArt, err := h.artService.CreateArt(artDTO.ToEntity(nil), with_stripe_bool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create art entity: " + err.Error()})
		return
	}
	currentArtID := uint(createdArt.ID)

	mainPhotoFileHeader, err1 := c.FormFile("main_photo")
	previewPhotoFileHeader, err2 := c.FormFile("preview_photo")

	// Проверяем только если есть реальные ошибки (не просто отсутствие файла)
	if err1 != nil && err1 != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing main_photo: " + err1.Error()})
		return
	}
	if err2 != nil && err2 != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing preview_photo: " + err2.Error()})
		return
	}

	if mainPhotoFileHeader != nil {
		if _, err := h.artService.AddMainOrPreviewPhotoToArt(currentArtID, mainPhotoFileHeader, true, false); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add main photo: " + err.Error()})
			return
		}
	}

	if previewPhotoFileHeader != nil {
		if _, err := h.artService.AddMainOrPreviewPhotoToArt(currentArtID, previewPhotoFileHeader, false, true); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add preview photo: " + err.Error()})
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

	// h.artService.UpdatePhotosInStripe(currentArtID)
	// фон
	// go func(artID uint) {
	if with_stripe_bool {
		err := h.artService.UpdatePhotosInStripe(currentArtID)
		if err != nil {
			log.Printf("Failed to update photos in Stripe for art %d: %v", currentArtID, err)
		}
	}
	finalEvent, err := h.artService.GetArtByID(currentArtID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve art after updates: " + err.Error()})
		return
	}
	// if err := h.artService.UpdatePhotosInStripe(artID); err != nil {
	// 	log.Printf("Failed to update photos in Stripe for art %d: %v", artID, err)
	// }
	// }(uint(currentArtID))

	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(finalEvent, base_url))
}

// // @Param with_stripe query bool false "With Stripe"

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
	// with_stripe := c.DefaultQuery("with_stripe", "false")
	with_stripe := "true"
	with_stripe_bool, err := strconv.ParseBool(with_stripe)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid with_stripe format"})
		return
	}
	idUint := uint(id)
	art, err := h.artService.FullUpdateArt(artDTO.ToEntity(&idUint), with_stripe_bool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// h.artService.UpdatePhotosInStripe(idUint)
	// Запускаем в фоне
	if with_stripe_bool {
		// go func(artID uint) {
		if err := h.artService.UpdatePhotosInStripe(uint(id)); err != nil {
			log.Printf("Failed to update photos in Stripe for art %d: %v", uint(id), err)
		}
		// }(uint(id))
	}
	art, err = h.artService.GetArtByID(art.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve art after updates: " + err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(art, base_url))
}

// @Param kwargs body map[string]interface{} true "kwargs"
// @Param kwargs body map[string]interface{} true "kwargs"
// // @Param with_stripe query bool false "With Stripe"

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
	// with_stripe := c.DefaultQuery("with_stripe", "false")
	with_stripe := "true"
	with_stripe_bool, err := strconv.ParseBool(with_stripe)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid with_stripe format"})
		return
	}
	art, err := h.artService.PartialUpdateArt(uint(id), kwargs, with_stripe_bool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if with_stripe_bool {
		// go func(artID uint) {
		if err := h.artService.UpdatePhotosInStripe(uint(id)); err != nil {
			log.Printf("Failed to update photos in Stripe for art %d: %v", uint(id), err)
		}
		// }(uint(id))
	}
	art, err = h.artService.GetArtByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve art after updates: " + err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(art, base_url))
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
	// h.artService.DeleteProductInStripe(uint(id))
	// фон
	go func(artID uint) {
		if err := h.artService.DeleteProductInStripe(artID); err != nil {
			log.Printf("Failed to delete product in Stripe for art %d: %v", artID, err)
		}
	}(uint(id))
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
	// h.artService.UpdatePhotosInStripe(uint(id))
	// фон
	// go func(artID uint) {
	if err := h.artService.UpdatePhotosInStripe(uint(id)); err != nil {
		log.Printf("Failed to update photos in Stripe for art %d: %v", uint(id), err)
	}
	// }(uint(id))
	art, err = h.artService.GetArtByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve art after updates: " + err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(art, base_url))
}

// @Summary Delete main photo from art
// @Description Удаляет главную/preview фотографию к картине
// @Accept json
// @Produce json
// @Tags Arts
// @Param id path int true "Art ID"
// @Param is_preview query bool false "Is preview" default(false)
// @Success 200 {object} map[string]string
// @Router /api/arts/{id}/main_photo [delete]
func (h *ArtHandler) DeleteMainPhotoFromArt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	is_preview_str := c.DefaultQuery("is_preview", "false")
	is_preview, err := strconv.ParseBool(is_preview_str)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	err = h.artService.DeleteMainOrPreviewPhoto(uint(id), is_preview)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	art, err := h.artService.GetArtByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// h.artService.UpdatePhotosInStripe(uint(id))
	// фон
	// go func(artID uint) {
	if err := h.artService.UpdatePhotosInStripe(uint(id)); err != nil {
		log.Printf("Failed to update photos in Stripe for art %d: %v", uint(id), err)
	}
	// }(uint(id))
	art, err = h.artService.GetArtByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve art after updates: " + err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(art, base_url))
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
	// h.artService.UpdatePhotosInStripe(uint(id))
	// фон
	// go func(artID uint) {
	if err := h.artService.UpdatePhotosInStripe(uint(id)); err != nil {
		log.Printf("Failed to update photos in Stripe for art %d: %v", uint(id), err)
	}
	// }(uint(id))
	photos_result, err = h.artService.GetArtByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve art after updates: " + err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(photos_result, base_url))
}

// @Summary Patch art photos
// @Description Обновляет фотографии картины - передается JSON массив строк: URLs для существующих мальчишек, а base64 для новых мальчишек
// @Accept json
// @Produce json
// @Tags Arts
// @Param id path int true "Art ID"
// @Param photos body []string true "Array of photo data: URLs for existing photos, base64 data URLs for new ones"
// @Success 200 {object} dto.ArtResponseDTO "Art"
// @Router /api/arts/{id}/photos [patch]
func (h *ArtHandler) PatchArtPhotos(c *gin.Context) {
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

	updatedEvent, err := h.artService.PatchArtPhotosFromStrings(uint(id), photoStrings)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can't patch photos: " + err.Error()})
		return
	}
	// go func(artID uint) {
	if err := h.artService.UpdatePhotosInStripe(uint(id)); err != nil {
		log.Printf("Failed to update photos in Stripe for art %d: %v", uint(id), err)
	}
	// }(uint(id))
	updatedEvent, err = h.artService.GetArtByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve art after updates: " + err.Error()})
		return
	}
	baseUrl := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(updatedEvent, baseUrl))
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
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(art, base_url))
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

// @Summary Recreate all Stripe products for arts
// @Description Удаляет старые продукты в Stripe и пересоздаёт для всех или выбранных арт-объектов. Операции удаления необратимы — требуется confirm=true. Можно выполнить пробный прогон с dry_run=true (без изменений).
// @Tags Arts
// @Accept json
// @Produce json
// @Param confirm query bool true "Подтвердить необратимые удаления"
// @Param dry_run query bool false "Пробный прогон без изменений" default(false)
// @Param ids query string false "Список ID арт-объектов через запятую; если не указан — для всех"
// @Param del_from_stripe query bool false "Удалять ли продукт в старом аккаунте Stripe перед созданием нового" default(true)
// @Param not_create query bool false "Только удалить (и очистить ссылки в БД), новый продукт НЕ создавать" default(false)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /update-all-products [post]
func (h *ArtHandler) RecreateAllStripeProducts(c *gin.Context) {
	confirmStr := c.DefaultQuery("confirm", "false")
	dryRunStr := c.DefaultQuery("dry_run", "false")
	idsStr := c.DefaultQuery("ids", "")
	delFromStripeStr := c.DefaultQuery("del_from_stripe", "true")
	notCreateStr := c.DefaultQuery("not_create", "false")

	confirm, err := strconv.ParseBool(confirmStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid confirm parameter"})
		return
	}
	dryRun, err := strconv.ParseBool(dryRunStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dry_run parameter"})
		return
	}

	var ids []uint
	if idsStr != "" {
		parts := strings.Split(idsStr, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			val, convErr := strconv.ParseUint(p, 10, 32)
			if convErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id in ids: " + p})
				return
			}
			ids = append(ids, uint(val))
		}
	}

	delFromStripe, err := strconv.ParseBool(delFromStripeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid del_from_stripe parameter"})
		return
	}

	notCreate, err := strconv.ParseBool(notCreateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid not_create parameter"})
		return
	}

	results, err := h.artService.RecreateAllStripeProducts(confirm, dryRun, delFromStripe, notCreate, ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build per-art validation messages
	validations := make([]map[string]string, 0, len(results))
	for _, r := range results {
		msg := map[string]string{}
		artID := r["art_id"]
		msg["art_id"] = fmt.Sprintf("%v", artID)

		if v, ok := r["delete_ok"]; ok && v.(bool) {
			msg["delete"] = "Удаление в Stripe успешно; далее создаём продукт"
		} else if _, ok := r["delete_skipped"]; ok {
			msg["delete"] = "Удаление пропущено (нет старого продукта)"
		} else if errText, ok := r["delete_error"].(string); ok && errText != "" {
			msg["delete"] = "Ошибка удаления: " + errText + ". Перешли к следующему art"
		} else if v, ok := r["delete"].(string); ok && v != "" {
			msg["delete"] = v
		}

		if v, ok := r["recreate_ok"]; ok && v.(bool) {
			msg["create"] = "Создание нового продукта успешно; обновили ссылки в БД"
		} else if errText, ok := r["recreate_error"].(string); ok && errText != "" {
			msg["create"] = "Ошибка создания: " + errText + ". Проверьте входные данные/цену"
		} else if v, ok := r["recreate"].(string); ok && v != "" {
			msg["create"] = v
		} else if v, ok := r["recreate_skipped"].(string); ok && v != "" {
			msg["create"] = v
		}

		validations = append(validations, msg)
	}

	c.JSON(http.StatusOK, gin.H{
		"dry_run":         dryRun,
		"confirm":         confirm,
		"del_from_stripe": delFromStripe,
		"not_create":      notCreate,
		"results":         results,
		"validations":     validations,
	})
}

// // @Summary Get art by ID
// // @Description Получает арт по ID
// // @Accept json
// // @Produce json
// // @Tags Arts
// // @Param id path int true "Art ID"
// // @Success 200 {object} entity.Art
// // @Router /api/arts/{id} [get]
// func (h *ArtHandler) GetArtByID(c *gin.Context) {
// 	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
// 		return
// 	}
// 	art, err := h.artService.GetArtByID(uint(id))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, art)
// }

// // @Summary Patch art photos
// // @Description Обновляет фотографии к картине(имеющиеся удаляются) - передается []base64 в формате multipart/form-data в поле photos
// // @Accept multipart/form-data
// // @Produce json
// // @Tags Arts
// // @Param id path int true "Art ID"
// // @Param photos formData []file true "Photos"
// // @Success 200 {object} dto.ArtResponseDTO "Art"
// // @Router /api/arts/{id}/photos [patch]
// func (h *ArtHandler) PatchArtPhotos(c *gin.Context) {
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
// 	photos_result, err := h.artService.PatchArtPhotos(uint(id), photos)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Cant patch photos"})
// 		return
// 	}
// 	// h.artService.UpdatePhotosInStripe(uint(id))
// 	// фон
// 	go func(artID uint) {
// 		if err := h.artService.UpdatePhotosInStripe(artID); err != nil {
// 			log.Printf("Failed to update photos in Stripe for art %d: %v", artID, err)
// 		}
// 	}(uint(id))
// 	base_url := config.GetBaseURL()
// 	c.JSON(http.StatusOK, dto.ToArtResponseDTO(photos_result, base_url))
// }

// SetCarousel sets the carousel order.
// @Summary Set carousel
// @Description Получает массив айди мальчишек артов. индекс = позиция в карусели. Все прошлые карусели при этом сбрасываются.
// @Accept json
// @Produce json
// @Tags Arts
// @Param ids body []uint true "IDs в порядке карусели"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/arts/carousel [post]
func (h *ArtHandler) SetCarousel(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body: " + err.Error()})
		return
	}
	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids array must not be empty"})
		return
	}
	if err := h.artService.UpdateCarousel(ids); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "carousel updated"})
}

// GetCarousel returns current carousel list
// @Summary Get carousel
// @Description Возвращает список мальчишек в карусели, отсортированных по carousel_position, с id, main_photo и name
// @Produce json
// @Tags Arts
// @Success 200 {array} dto.CarouselArtDTO
// @Failure 500 {object} map[string]string
// @Router /api/arts/carousel [get]
func (h *ArtHandler) GetCarousel(c *gin.Context) {
	arts, err := h.artService.GetCarousel()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	baseUrl := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToCarouselArtDTOs(arts, baseUrl))
}
