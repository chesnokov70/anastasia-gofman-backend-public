package handler

import (
	"net/http"

	"anastasia_gofman_backend/internal/delivery/http/dto"
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/service"
	"anastasia_gofman_backend/pkg/config"
	"strconv"

	"encoding/json"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type AuthorHandler struct {
	authorService service.AuthorService
}

func parseSpecializations(specializationStr string) pq.StringArray {
	if specializationStr == "" {
		return pq.StringArray{}
	}

	parts := strings.Split(specializationStr, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return pq.StringArray(result)
}

// @Summary Get all authors
// @Description Получаем всех авторов
// @Tags Authors
// @Accept json
// @Produce json
// @Param with_arts query bool false "With Arts"
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Param with_pagination query bool false "With pagination" default(true)
// @Param specialization query []string false "Filter by specialization" collectionFormat(csv)
// @Success 200 {array} dto.AuthorResponseWithArtsDTO
// @Router /api/authors [get]
func (h *AuthorHandler) GetAllAuthors(c *gin.Context) {
	page := c.GetInt("page")
	size := c.GetInt("size")
	with_pagination_str := c.DefaultQuery("with_pagination", "true")
	with_pagination, err := strconv.ParseBool(with_pagination_str)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid with_pagination format"})
		return
	}
	with_arts_str := c.DefaultQuery("with_arts", "false")
	with_arts, err := strconv.ParseBool(with_arts_str)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid with_arts format"})
		return
	}

	// Получаем параметры фильтрации по специализации
	specialization_param := c.Query("specialization")
	var specializations []string
	if specialization_param != "" {
		// Разделяем по запятой для поддержки нескольких специализаций
		specializations = strings.Split(specialization_param, ",")
		for i := range specializations {
			specializations[i] = strings.TrimSpace(specializations[i])
		}
	}

	var authors []entity.Author
	var arts map[uint][]entity.Art
	var total_pages, total_items int64

	// Используем фильтрацию по специализации если указана
	if len(specializations) > 0 {
		authors, arts, total_pages, total_items, err = h.authorService.GetAuthorsBySpecialization(specializations, with_arts, page, size, with_pagination)
	} else {
		authors, arts, total_pages, total_items, err = h.authorService.GetAllAuthors(with_arts, page, size, with_pagination)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	base_url := config.GetBaseURL()
	if with_pagination {
		response := gin.H{
			"data":       dto.ToAuthorResponseWithAllArtsDTOs(authors, arts, base_url),
			"pagination": gin.H{"total_pages": int(total_pages), "current_page": int(page), "page_size": int(size), "total_items": int(total_items)},
		}
		c.JSON(http.StatusOK, response)
	} else {
		response := gin.H{
			"data": dto.ToAuthorResponseWithAllArtsDTOs(authors, arts, base_url),
		}
		c.JSON(http.StatusOK, response)
	}
}

// @Summary Get author by ID
// @Description Получаем автора по ID
// @Tags Authors
// @Accept json
// @Produce json
// @Param id path int true "Author ID"
// @Param with_arts query bool false "With Arts"
// @Success 200 {object} dto.AuthorResponseWithArtsDTO
// @Router /api/authors/{id} [get]
func (h *AuthorHandler) GetAuthorByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	with_arts_str := c.DefaultQuery("with_arts", "false")
	with_arts, err := strconv.ParseBool(with_arts_str)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid with_arts format"})
		return
	}
	author, arts, err := h.authorService.GetAuthorByID(uint(id), with_arts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToAuthorResponseWithArtsDTO(author, arts[author.ID], base_url))
}

// @Summary Create author
// @Description С кайфом создаем автора, все поля необязательные кроме имени, ну и в mail пихать нужно mail нормальный, но можно ничего не пихать
// @Tags Authors
// @Accept json
// @Produce json
// @Param data body dto.CreateAuthorDTO true "Author data"
// @Success 201 {object} dto.AuthorResponseDTO
// @Failure 400 {object} map[string]string
// @Router /api/authors [post]
func (h *AuthorHandler) CreateAuthor(c *gin.Context) {
	var authorDTO dto.CreateAuthorDTO
	if err := c.ShouldBindJSON(&authorDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	author := entity.Author{
		Name:                  authorDTO.Name.ToEntity(),
		Bio:                   authorDTO.Bio.ToEntity(),
		Biography:             authorDTO.Biography.ToEntity(),
		EducationalBackground: authorDTO.EducationalBackground.ToEntity(),
		Exhibitions:           authorDTO.Exhibitions.ToEntity(),
		ContactInfo:           authorDTO.ContactInfo.ToEntity(),
		Specialization:        parseSpecializations(authorDTO.Specialization),
		AdditionalInfo:        authorDTO.AdditionalInfo,
		Contact: entity.ContactInfo{
			Email: authorDTO.Contact.Email,
			Phone: authorDTO.Contact.Phone,
			Links: entity.SocialLink{
				Instagram: authorDTO.Contact.Links.Instagram,
				Telegram:  authorDTO.Contact.Links.Telegram,
				Vkontakte: authorDTO.Contact.Links.Vkontakte,
				Facebook:  authorDTO.Contact.Links.Facebook,
				Twitter:   authorDTO.Contact.Links.Twitter,
				Youtube:   authorDTO.Contact.Links.Youtube,
				Linkedin:  authorDTO.Contact.Links.Linkedin,
				Whatsapp:  authorDTO.Contact.Links.Whatsapp,
				Pinterest: authorDTO.Contact.Links.Pinterest,
				Behance:   authorDTO.Contact.Links.Behance,
			},
		},
		IsActive: authorDTO.IsActive,
	}

	author, err := h.authorService.CreateAuthor(author)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusCreated, dto.ToAuthorResponseDTO(author, base_url))
}

// @Summary Delete author
// @Description Удаляем автора
// @Tags Authors
// @Accept json
// @Produce json
// @Param id path int true "Author ID"
// @Success 200 {object} map[string]string
// @Router /api/authors/{id} [delete]
func (h *AuthorHandler) DeleteAuthor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	err = h.authorService.DeleteAuthor(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "author deleted"})
}

// @Summary Partial update author
// @Description Обновляем автора частично, все поля необязательные, не передал = осталось прежним
// @Tags Authors
// @Accept json
// @Produce json
// @Param id path uint true "Author ID"
// @Param data body dto.UpdateAuthorDTO true "Author"
// @Success 200 {object} dto.AuthorResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/authors/{id} [patch]
func (h *AuthorHandler) PartialUpdateAuthor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if updateData["specialization"] != nil {
		updateData["specialization"] = parseSpecializations(updateData["specialization"].(string))
	}
	author, err := h.authorService.PartialUpdateAuthor(uint(id), updateData)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToAuthorResponseDTO(author, base_url))
}

// @Summary Update author
// @Description Обновляем автора полностью, все поля необязательные, но сотрется вся инфа, которая была раньше и которая не передана
// @Tags Authors
// @Accept json
// @Produce json
// @Param id path int true "Author ID"
// @Param data body dto.UpdateAuthorDTO true "Author data"
// @Success 200 {object} dto.AuthorResponseDTO
// @Failure 400 {object} map[string]string
// @Router /api/authors/{id} [put]
func (h *AuthorHandler) FullUpdateAuthor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	var authorDTO dto.UpdateAuthorDTO
	if err := c.ShouldBindJSON(&authorDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	author := entity.Author{
		ID:                    uint(id),
		Name:                  authorDTO.Name.ToEntity(),
		Bio:                   authorDTO.Bio.ToEntity(),
		Biography:             authorDTO.Biography.ToEntity(),
		EducationalBackground: authorDTO.EducationalBackground.ToEntity(),
		Exhibitions:           authorDTO.Exhibitions.ToEntity(),
		ContactInfo:           authorDTO.ContactInfo.ToEntity(),
		Specialization:        parseSpecializations(authorDTO.Specialization),
		AdditionalInfo:        authorDTO.AdditionalInfo,
		Contact: entity.ContactInfo{
			Email: authorDTO.Contact.Email,
			Phone: authorDTO.Contact.Phone,
			Links: entity.SocialLink{
				Instagram: authorDTO.Contact.Links.Instagram,
				Telegram:  authorDTO.Contact.Links.Telegram,
				Vkontakte: authorDTO.Contact.Links.Vkontakte,
				Facebook:  authorDTO.Contact.Links.Facebook,
				Twitter:   authorDTO.Contact.Links.Twitter,
				Youtube:   authorDTO.Contact.Links.Youtube,
				Linkedin:  authorDTO.Contact.Links.Linkedin,
				Whatsapp:  authorDTO.Contact.Links.Whatsapp,
				Pinterest: authorDTO.Contact.Links.Pinterest,
				Behance:   authorDTO.Contact.Links.Behance,
			},
		},
		IsActive: authorDTO.IsActive,
	}
	author, err = h.authorService.FullUpdateAuthor(author)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToAuthorResponseDTO(author, base_url))
}

// @Summary Create author with photos
// @Description Создает мальчишку c фотками - все пихается в multipart/form-data в поле main_photo, preview_photo, photos(массив)
// @Accept multipart/form-data
// @Produce json
// @Tags Authors
// @Param data formData string true "JSON который совпадает частично с CreateAuthorDTO" Extensions(x-example={"name": {"en": "John Doe", "es": "Juan Pérez", "ru": "Иван Иванов"}, "bio": {"en": "Some information", "es":"tralalelo tralala", "ru": "Инфа про мальчишку"}, "biography": {"en": "Some information", "es":"tralalelo tralala", "ru": "Инфа про мальчишку"}, "educational_background": {"en": "Some information", "es":"tralalelo tralala", "ru": "Инфа про мальчишку"}, "exhibitions": {"en": "Some information", "es":"tralalelo tralala", "ru": "Инфа про мальчишку"}, "contact_info": {"en": "Some information", "es":"tralalelo tralala", "ru": "Инфа про мальчишку"}, "contact": {"email": "author@example.com", "phone": "+1234567890", "links": {"instagram": "https://instagram.com/author", "telegram": "https://t.me/author", "vkontakte": "https://vk.com/author", "facebook": "https://facebook.com/author", "twitter": "https://twitter.com/author", "youtube": "https://youtube.com/author", "linkedin": "https://linkedin.com/in/author", "whatsapp": "+1234567890", "pinterest": "https://pinterest.com/author", "behance": "https://behance.net/author"}}, "is_active": true, "specialization": "sculptor,roker"})
// @Param main_photo formData file false "Main Photo"
// @Param preview_photo formData file false "Preview Photo"
// @Param photos formData []file false "Photos"
// @Success 201 {object} dto.AuthorResponseDTO
// @Failure 400 {object} map[string]string
// @Router /api/authors/with_photos [post]
func (h *AuthorHandler) CreateAuthorWithPhotos(c *gin.Context) {
	var authorDTO dto.CreateAuthorWithPhotosDTO
	jsonData := c.PostForm("data")
	if jsonData == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'data' field in form-data"})
		return
	}

	if err := json.Unmarshal([]byte(jsonData), &authorDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON in 'data' field: " + err.Error()})
		return
	}

	author := entity.Author{
		Name:                  authorDTO.Name.ToEntity(),
		Bio:                   authorDTO.Bio.ToEntity(),
		Biography:             authorDTO.Biography.ToEntity(),
		EducationalBackground: authorDTO.EducationalBackground.ToEntity(),
		Exhibitions:           authorDTO.Exhibitions.ToEntity(),
		ContactInfo:           authorDTO.ContactInfo.ToEntity(),
		Specialization:        parseSpecializations(authorDTO.Specialization),
		AdditionalInfo:        authorDTO.AdditionalInfo,
		Contact: entity.ContactInfo{
			Email: authorDTO.Contact.Email,
			Phone: authorDTO.Contact.Phone,
			Links: entity.SocialLink{
				Instagram: authorDTO.Contact.Links.Instagram,
				Telegram:  authorDTO.Contact.Links.Telegram,
				Vkontakte: authorDTO.Contact.Links.Vkontakte,
				Facebook:  authorDTO.Contact.Links.Facebook,
				Twitter:   authorDTO.Contact.Links.Twitter,
				Youtube:   authorDTO.Contact.Links.Youtube,
				Linkedin:  authorDTO.Contact.Links.Linkedin,
				Whatsapp:  authorDTO.Contact.Links.Whatsapp,
				Pinterest: authorDTO.Contact.Links.Pinterest,
				Behance:   authorDTO.Contact.Links.Behance,
			},
		},
		IsActive: authorDTO.IsActive,
	}

	if err := c.Request.ParseMultipartForm(1 << 30); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form: " + err.Error()})
		return
	}

	createdAuthor, err := h.authorService.CreateAuthor(author)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create author entity: " + err.Error()})
		return
	}
	currentAuthorID := uint(createdAuthor.ID)

	mainPhotoFileHeader, _ := c.FormFile("main_photo")
	previewPhotoFileHeader, _ := c.FormFile("preview_photo")

	if mainPhotoFileHeader != nil {
		if _, err := h.authorService.AddMainOrPreviewPhotoToAuthor(currentAuthorID, mainPhotoFileHeader, true, false); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add main photo: " + err.Error()})
			return
		}
	}

	if previewPhotoFileHeader != nil {
		if _, err := h.authorService.AddMainOrPreviewPhotoToAuthor(currentAuthorID, previewPhotoFileHeader, false, true); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add preview photo: " + err.Error()})
			return
		}
	}

	photoFileHeaders := c.Request.MultipartForm.File["photos"]

	if len(photoFileHeaders) > 0 {
		if _, err := h.authorService.AddPhotosToAuthor(currentAuthorID, photoFileHeaders); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add photos: " + err.Error()})
			return
		}
	}

	finalAuthor, _, err := h.authorService.GetAuthorByID(currentAuthorID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve author after updates: " + err.Error()})
		return
	}

	base_url := config.GetBaseURL()
	c.JSON(http.StatusCreated, dto.ToAuthorResponseDTO(finalAuthor, base_url))
}

// @Summary Add main photo to author
// @Description Добавляет/обновляет главную фотку - multipart/form-data в поле main_photo
// @Accept multipart/form-data
// @Produce json
// @Tags Authors
// @Param id path int true "Author ID"
// @Param is_preview query bool false "Is Preview"
// @Param main_photo formData file true "Main Photo"
// @Success 200 {object} dto.AuthorResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/authors/{id}/main_photo [post]
func (h *AuthorHandler) AddMainPhotoToAuthor(c *gin.Context) {
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
	author, err := h.authorService.AddMainOrPreviewPhotoToAuthor(uint(id), main_photo, !is_preview, is_preview)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToAuthorResponseDTO(author, base_url))
}

// @Summary Add photos to author
// @Description Добавляет фотографии к автору(имеющиеся не трогаются) - передается []base64 в формате multipart/form-data в поле photos
// @Accept multipart/form-data
// @Produce json
// @Tags Authors
// @Param id path int true "Author ID"
// @Param photos formData []file true "Photos"
// @Success 200 {object} dto.AuthorResponseDTO "Author"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/authors/{id}/photos [post]
func (h *AuthorHandler) AddPhotosToAuthor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	_, _, err = h.authorService.GetAuthorByID(uint(id), false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "author not found"})
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form"})
		return
	}
	photos := form.File["photos"]

	photos_result, err := h.authorService.AddPhotosToAuthor(uint(id), photos)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cant add photos"})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToAuthorResponseDTO(photos_result, base_url))
}

// // @Summary Patch author photos
// // @Description Обновляет фотографии к автору(имеющиеся удаляются) - multipart/form-data в поле photos
// // @Accept multipart/form-data
// // @Produce json
// // @Tags Authors
// // @Param id path int true "Author ID"
// // @Param photos formData []file true "Photos"
// // @Success 200 {object} dto.AuthorResponseDTO "Author"
// // @Router /api/authors/{id}/photos [patch]
// func (h *AuthorHandler) PatchAuthorPhotos(c *gin.Context) {
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
// 	photos_result, err := h.authorService.PatchAuthorPhotos(uint(id), photos)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Cant patch photos"})
// 		return
// 	}
// 	base_url := config.GetBaseURL()
// 	c.JSON(http.StatusOK, dto.ToAuthorResponseDTO(photos_result, base_url))
// }

// @Summary Patch author photos
// @Description Обновляет фотографии автора - передается JSON массив строк: URLs для существующих мальчишек, а base64 для новых мальчишек
// @Accept json
// @Produce json
// @Tags Authors
// @Param id path int true "Author ID"
// @Param photos body []string true "Array of photo data: URLs for existing photos, base64 data URLs for new ones"
// @Success 200 {object} dto.AuthorResponseDTO "Author"
// @Failure 400 {object} map[string]string
// @Router /api/authors/{id}/photos [patch]
func (h *AuthorHandler) PatchAuthorPhotos(c *gin.Context) {
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

	updatedEvent, err := h.authorService.PatchAuthorPhotosFromStrings(uint(id), photoStrings)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can't patch photos: " + err.Error()})
		return
	}

	baseUrl := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToAuthorResponseDTO(updatedEvent, baseUrl))
}

// @Summary Update main photo to author

// @Description Удаляет главную фотку автора - multipart/form-data в поле main_photo
// @Accept multipart/form-data
// @Produce json
// @Tags Authors
// @Param id path int true "Author ID"
// @Param is_preview query bool false "Is Preview"
// @Success 200 {object} dto.AuthorResponseDTO
// @Router /api/authors/{id}/main_photo [delete]
func (h *AuthorHandler) DeleteMainPhotoFromAuthor(c *gin.Context) {
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
	err = h.authorService.DeleteMainOrPreviewPhoto(uint(id), !is_preview, is_preview)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	res_str := "main"
	if is_preview {
		res_str = "preview"
	}
	c.JSON(http.StatusOK, gin.H{"message": res_str + " photo deleted"})
}

func NewAuthorHandler(authorService service.AuthorService) *AuthorHandler {
	return &AuthorHandler{authorService: authorService}
}

// @Summary Get author with arts
// @Description Получаем автора с его работами
// @Tags Authors
// @Accept json
// @Produce json
// @Param id path int true "Author ID"
// @Success 200 {object} dto.AuthorResponseWithArtsDTO
// @Router /api/authors/{id}/with_arts [get]
func (h *AuthorHandler) GetAuthorWithArts(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	author, arts, err := h.authorService.GetAuthorWithArts(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToAuthorResponseWithArtsDTO(author, arts, base_url))
}
