package handler

import (
	"anastasia_gofman_backend/internal/delivery/http/dto"
	"anastasia_gofman_backend/internal/service"
	"encoding/json"
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
// @Param filtering query string false "JSON filter object. Example: {\"price_from\": 100, \"price_to\": 1000, \"size\": \"MEDIUM\", \"direction\": \"EXCLUSIVE\", \"style\": \"abstract\", \"author\": \"author name\", \"type\": \"_common\", \"archive_type\": \"repeat\", \"search\": {\"name\": {\"en\": \"search text\", \"ru\": \"текст поиска\", \"es\": \"texto de búsqueda\"}, \"title\": {\"en\": \"title search\"}, \"description\": {\"en\": \"desc search\"}, \"medium\": {\"en\": \"medium search\"}, \"technique\": {\"en\": \"technique search\"}, \"year\": \"2024\", \"frame\": {\"en\": \"frame search\"}, \"style\": \"abstract\", \"direction\": \"EXCLUSIVE\", \"dimension_str\": \"100x100\"}}"
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

// @Summary Create a new art
// @Description Создает новую картину, все поля необязательные кроме имени, по этому пути нельзя передавать ФОТО, но если передаешь id автора - убедись, что такой есть! Если не передашь имя и дескрипшн для страйпа - возьму их из EN полей, но нельзя получается, чтобы поле name.en было пустым одновременно с nameForStripe
// @Accept json
// @Produce json
// @Tags Arts
// @Param data body dto.CreateArtDTO true "Art"
// @Param with_stripe query bool false "With Stripe"
// @Success 201 {object} dto.ArtResponseDTO
// @Failure 400 {object} map[string]string
// @Router /api/arts [post]
func (h *ArtHandler) CreateArt(c *gin.Context) {
	var artDTO dto.CreateArtDTO
	if err := c.ShouldBindJSON(&artDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	with_stripe := c.DefaultQuery("with_stripe", "false")
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

// @Summary Create art with photos
// @Description Создает картину и фотографии к ней - передается в формате multipart/form-data в поле main_photo, preview_photo, photos(массив)
// @Accept multipart/form-data
// @Produce json
// @Tags Arts
// @Param data formData string true "JSON с любыми полями из CreateArtDTO" Extensions(x-example={ "author_id": 1, "name": {"en": "Art Name", "ru": "Название Картины", "es": "Nombre del Arte"}, "title": {"en": "Art Title", "ru": "Заголовок", "es": "Título"}, "description": {"en": "Detailed description", "ru": "Детальное описание", "es": "Descripción detallada"}, "medium": {"en": "Oil on canvas", "ru": "Холст, масло", "es": "Óleo sobre lienzo"}, "technique": {"en": "Impasto", "ru": "Импасто", "es": "Empaste"}, "dimension_x": 120, "dimension_y": 90, "year": 2024, "frame": {"en": "Wooden frame", "ru": "Деревянная рама", "es": "Marco de madera"}, "price": 150000 })
// @Param main_photo formData file false "Main Photo"
// @Param preview_photo formData file false "Preview Photo"
// @Param photos formData []file false "Photos"
// @Param with_stripe query bool false "With Stripe"
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

	with_stripe := c.DefaultQuery("with_stripe", "false")
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

	finalEvent, err := h.artService.GetArtByID(currentArtID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve art after updates: " + err.Error()})
		return
	}

	// h.artService.UpdatePhotosInStripe(currentArtID)
	// фон
	go func(artID uint) {
		if with_stripe_bool {
			err := h.artService.UpdatePhotosInStripe(artID)
			if err != nil {
				log.Printf("Failed to update photos in Stripe for art %d: %v", artID, err)
			}
		}
		// if err := h.artService.UpdatePhotosInStripe(artID); err != nil {
		// 	log.Printf("Failed to update photos in Stripe for art %d: %v", artID, err)
		// }
	}(uint(currentArtID))

	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(finalEvent, base_url))
}

// @Summary Full update art
// @Description Обновляет всю картину - все поля необязательные, но перезаписывает все поля, те которые не передал - обнуляются (кроме фото)
// @Accept json
// @Produce json
// @Tags Arts
// @Param id path int true "Art ID"
// @Param data body dto.UpdateArtDTO true "Art"
// @Param with_stripe query bool false "With Stripe"
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
	with_stripe := c.DefaultQuery("with_stripe", "false")
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
		go func(artID uint) {
			if err := h.artService.UpdatePhotosInStripe(artID); err != nil {
				log.Printf("Failed to update photos in Stripe for art %d: %v", artID, err)
			}
		}(uint(id))
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(art, base_url))
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
// @Param with_stripe query bool false "With Stripe"
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
	with_stripe := c.DefaultQuery("with_stripe", "false")
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
		go func(artID uint) {
			if err := h.artService.UpdatePhotosInStripe(artID); err != nil {
				log.Printf("Failed to update photos in Stripe for art %d: %v", artID, err)
			}
		}(uint(id))
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
	go func(artID uint) {
		if err := h.artService.UpdatePhotosInStripe(artID); err != nil {
			log.Printf("Failed to update photos in Stripe for art %d: %v", artID, err)
		}
	}(uint(id))
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
	// h.artService.UpdatePhotosInStripe(uint(id))
	// фон
	go func(artID uint) {
		if err := h.artService.UpdatePhotosInStripe(artID); err != nil {
			log.Printf("Failed to update photos in Stripe for art %d: %v", artID, err)
		}
	}(uint(id))
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(photos_result, base_url))
}

// @Summary Patch art photos
// @Description Обновляет фотографии к картине(имеющиеся удаляются) - передается []base64 в формате multipart/form-data в поле photos
// @Accept multipart/form-data
// @Produce json
// @Tags Arts
// @Param id path int true "Art ID"
// @Param photos formData []file true "Photos"
// @Success 200 {object} dto.ArtResponseDTO "Art"
// @Router /api/arts/{id}/photos [patch]
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
	// h.artService.UpdatePhotosInStripe(uint(id))
	// фон
	go func(artID uint) {
		if err := h.artService.UpdatePhotosInStripe(artID); err != nil {
			log.Printf("Failed to update photos in Stripe for art %d: %v", artID, err)
		}
	}(uint(id))
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArtResponseDTO(photos_result, base_url))
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
