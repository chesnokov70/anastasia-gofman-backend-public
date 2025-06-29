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

type ArticleHandler struct {
	articleService service.PressAndArticleService
}

func NewArticleHandler(articleService service.PressAndArticleService) *ArticleHandler {
	return &ArticleHandler{articleService: articleService}
}

// @Description Получает все статьи, умеет в пагинацию, ордерд бай криэтед эт
// @Accept json
// @Produce json
// @Tags Articles
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param sorting query string false "Sorting type" Enums(NEW, CLOSEST, FARTHEST) default(NEW)
// @Param with_pagination query bool false "With pagination"
// @Router /api/articles [get]
func (h *ArticleHandler) GetAllArticles(c *gin.Context) {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Query("pageSize"))
	if err != nil {
		pageSize = 10
	}
	withPagination, err := strconv.ParseBool(c.Query("with_pagination"))
	if err != nil {
		withPagination = false
	}
	sorting := c.Query("sorting")
	if sorting == "" {
		sorting = "NEW"
	}

	_, articles, pages, total, err := h.articleService.GetAllPressAndArticles(page, pageSize, withPagination, "article", sorting)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	if withPagination {
		c.JSON(http.StatusOK, gin.H{
			"articles": dto.ToArticleResponseDTOs(articles, base_url),
			"pagination": gin.H{
				"total":       total,
				"total_pages": pages,
			},
		})
	} else {
		c.JSON(http.StatusOK, dto.ToArticleResponseDTOs(articles, base_url))
	}
}

// @Summary Get article by ID
// @Description Получает статью по ID
// @Accept json
// @Produce json
// @Tags Articles
// @Param id path int true "Article ID"
// @Router /api/articles/{id} [get]
func (h *ArticleHandler) GetArticleByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, article, err := h.articleService.GetPressOrArticleByID(uint(id), "article")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArticleResponseDTO(*article, base_url))
}

// @Summary Create article
// @Description Создает статью - фотки не суй плиз
// @Accept json
// @Produce json
// @Tags Articles
// @Param article body dto.CreateArticleDTO true "Article"
// @Success 200 {object} dto.ArticleResponseDTO
// @Router /api/articles [post]
func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	var article dto.CreateArticleDTO
	if err := c.ShouldBindJSON(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	article_entity := article.ToEntity(nil)
	r, article_response, err := h.articleService.CreatePressOrArticle("article", entity.Press{}, article_entity)
	log.Printf("r: %v", r)
	fmt.Printf("r: %v", r)
	fmt.Printf("article_response: %v", article_response)
	log.Printf("article_response: %v", article_response)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArticleResponseDTO(*article_response, base_url))
}

// @Summary Create article with photos
// @Description Создает статью и фотографии к ней - передается в формате multipart/form-data в поле main_photo, preview_photo, photos(массив)
// @Accept multipart/form-data
// @Produce json
// @Tags Articles
// @Param data formData string true "JSON с любыми полями из CreateArticleDTO" Extensions(x-example={"description": {"en": "Tung, tung, tung, tung, tung, tung, tung, tung, tung, Sahur.","es": "Golybini Shpionini","ru": "Тун, тун, тун, Сахур."},"full_text": {"en": "Tung, tung, tung, tung, tung, tung, tung, tung, tung, Sahur.","es": "Golybini Shpionini","ru": "Тун, тун, тун, Сахур."},"link": "https://example.com","position": 1,"title": {"en": "Tung, tung, tung, tung, tung, tung, tung, tung, tung, Sahur.","es": "Golybini Shpionini","ru": "Тун, тун, тун, Сахур."}})
// @Param main_photo formData file false "Main Photo"
// @Param preview_photo formData file false "Preview Photo"
// @Param photos formData []file false "Photos"
// @Success 201 {object} dto.ArticleResponseDTO
// @Failure 400 {object} map[string]string
// @Router /api/articles/with_photos [post]
func (h *ArticleHandler) CreateArticleWithPhotos(c *gin.Context) {
	var articleDTO dto.CreateArticleDTO

	jsonData := c.PostForm("data")
	if jsonData == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'data' field in form-data"})
		return
	}

	if err := json.Unmarshal([]byte(jsonData), &articleDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON in 'data' field: " + err.Error()})
		return
	}

	if err := c.Request.ParseMultipartForm(1 << 30); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form: " + err.Error()})
		return
	}

	_, createdArticle, err := h.articleService.CreatePressOrArticle("article", entity.Press{}, articleDTO.ToEntity(nil))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create article entity: " + err.Error()})
		return
	}
	currentArticleID := uint(createdArticle.ID)

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
		if _, _, err := h.articleService.AddMainOrPreviewPhotoToPressOrArticle("article", currentArticleID, mainPhotoFileHeader, true, false); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add main photo: " + err.Error()})
			return
		}
	}

	if previewPhotoFileHeader != nil {
		if _, _, err := h.articleService.AddMainOrPreviewPhotoToPressOrArticle("article", currentArticleID, previewPhotoFileHeader, false, true); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add preview photo: " + err.Error()})
			return
		}
	}

	photoFileHeaders := c.Request.MultipartForm.File["photos"]

	if len(photoFileHeaders) > 0 {
		if _, _, err := h.articleService.AddPhotosToPressOrArticle(currentArticleID, "article", photoFileHeaders); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add photos: " + err.Error()})
			return
		}
	}

	_, finalEvent, err := h.articleService.GetPressOrArticleByID(currentArticleID, "article")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve article after updates: " + err.Error()})
		return
	}

	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArticleResponseDTO(*finalEvent, base_url))
}

// @Summary Full update article
// @Description Обновляет всю статью - все поля необязательные, но перезаписывает все поля, те которые не передал - обнуляются (кроме фото)
// @Accept json
// @Produce json
// @Tags Articles
// @Param id path int true "Article ID"
// @Param data body dto.UpdateArticleDTO true "Article"
// @Success 200 {object} dto.ArticleResponseDTO
// @Router /api/articles/{id} [put]
func (h *ArticleHandler) FullUpdateArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	var articleDTO dto.UpdateArticleDTO
	if err := c.ShouldBindJSON(&articleDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	idUint := uint(id)
	_, article_response, err := h.articleService.FullUpdatePressOrArticle("article", entity.Press{}, articleDTO.ToEntity(&idUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArticleResponseDTO(*article_response, base_url))
}

// @Summary Partial update article
// @Description Обновляет часть статьи - все поля необязательные, поля, которые не передал не меняются, фото сюда не суй!
// @Accept json
// @Produce json
// @Tags Articles
// @Param id path int true "Article ID"
// @Param kwargs body dto.UpdateArticleDTO true "Article"
// @Success 200 {object} dto.ArticleResponseDTO
// @Router /api/articles/{id} [patch]
func (h *ArticleHandler) PartialUpdateArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	var articleDTO map[string]interface{}
	if err := c.ShouldBindJSON(&articleDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	idUint := uint(id)
	_, article_response, err := h.articleService.PartialUpdatePressOrArticle("article", idUint, articleDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArticleResponseDTO(*article_response, base_url))
}

// @Summary Delete article
// @Description Удаляет статью и фотки
// @Accept json
// @Produce json
// @Tags Articles
// @Param id path int true "Article ID"
// @Success 200 {object} map[string]string
// @Router /api/articles/{id} [delete]
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	err = h.articleService.DeletePressOrArticle("article", uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "article deleted"})
}

// @Summary Add main photo to article
// @Description Добавляет/обновляет главную фотографию к статье - передается фото в формате multipart/form-data в поле main_photo
// @Accept multipart/form-data
// @Produce json
// @Tags Articles
// @Param id path int true "Article ID"
// @Param is_preview query bool false "Is Preview"
// @Param main_photo formData file true "Main Photo"
// @Success 200 {object} dto.ArticleResponseDTO
// @Router /api/articles/{id}/main_photo [post]
func (h *ArticleHandler) AddMainPhotoToArticle(c *gin.Context) {
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
	_, article_response, err := h.articleService.AddMainOrPreviewPhotoToPressOrArticle("article", uint(id), main_photo, !is_preview, is_preview)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArticleResponseDTO(*article_response, base_url))
}

// @Summary Add photos to article
// @Description Добавляет фотографии к статье(имеющиеся не трогаются) - передается []base64 в формате multipart/form-data в поле photos
// @Accept multipart/form-data
// @Produce json
// @Tags Articles
// @Param id path int true "Article ID"
// @Param photos formData []file true "Photos"
// @Success 200 {object} dto.ArticleResponseDTO "Article"
// @Router /api/articles/{id}/photos [post]
func (h *ArticleHandler) AddPhotosToArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	_, _, err = h.articleService.GetPressOrArticleByID(uint(id), "article")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "article not found"})
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form"})
		return
	}
	photos := form.File["photos"]

	_, article, err := h.articleService.AddPhotosToPressOrArticle(uint(id), "article", photos)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cant add photos"})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArticleResponseDTO(*article, base_url))
}

// @Summary Patch article photos
// @Description Обновляет фотографии к статье(имеющиеся удаляются) - передается []base64 в формате multipart/form-data в поле photos
// @Accept multipart/form-data
// @Produce json
// @Tags Articles
// @Param id path int true "Article ID"
// @Param photos formData []file true "Photos"
// @Success 200 {object} dto.ArticleResponseDTO "Article"
// @Router /api/articles/{id}/photos [patch]
func (h *ArticleHandler) PatchArticlePhotos(c *gin.Context) {
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
	_, article, err := h.articleService.PatchPressOrArticlePhotos(uint(id), "article", photos)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cant patch photos"})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToArticleResponseDTO(*article, base_url))
}

// @Summary Get main photo
// @Description Получает главную фотографию статьи
// @Accept json
// @Produce json
// @Tags Articles
// @Param id path int true "Article ID"
// @Success 200 {object} entity.Photo
// @Router /api/articles/{id}/main_photo [get]
func (h *ArticleHandler) GetMainPhoto(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	photo, err := h.articleService.GetMainPhoto(uint(id), "article")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, photo)
}
